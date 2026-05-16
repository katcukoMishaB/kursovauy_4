package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Gateway struct {
	UserServiceURL    string
	ProjectServiceURL string
	TaskServiceURL    string
	ChatServiceURL    string
	ReportServiceURL  string
	Client            *http.Client
}

func main() {
	gateway := &Gateway{
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8001"),
		ProjectServiceURL: getEnv("PROJECT_SERVICE_URL", "http://localhost:8002"),
		TaskServiceURL:    getEnv("TASK_SERVICE_URL", "http://localhost:8003"),
		ChatServiceURL:    getEnv("CHAT_SERVICE_URL", "http://localhost:8004"),
		ReportServiceURL:  getEnv("REPORT_SERVICE_URL", "http://localhost:8005"),
		Client:            &http.Client{},
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/files/upload", gateway.uploadHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/files/{name}", gateway.fileHandler).Methods("GET")
	r.PathPrefix("/api/users").HandlerFunc(gateway.proxyHandler(gateway.UserServiceURL))
	r.PathPrefix("/api/projects").HandlerFunc(gateway.proxyHandler(gateway.ProjectServiceURL))
	r.PathPrefix("/api/tasks").HandlerFunc(gateway.proxyHandler(gateway.TaskServiceURL))
	r.PathPrefix("/api/chats").HandlerFunc(gateway.proxyWebSocketHandler(gateway.ChatServiceURL))
	r.PathPrefix("/api/reports").HandlerFunc(gateway.proxyHandler(gateway.ReportServiceURL))

	frontendPath := getFrontendPath()
	log.Printf("Serving frontend from: %s", frontendPath)

	fs := http.FileServer(http.Dir(frontendPath))

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" || path == "" {
			http.ServeFile(w, r, filepath.Join(frontendPath, "index.html"))
			return
		}
		fp := filepath.Join(frontendPath, path)
		if info, err := os.Stat(fp); err == nil && !info.IsDir() {
			http.StripPrefix("/", fs).ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(frontendPath, "index.html"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(r)))
}

func (g *Gateway) proxyHandler(targetURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api")
		var serviceURL string

		if strings.HasPrefix(path, "/users") {
			path = strings.TrimPrefix(path, "/users")
			serviceURL = g.UserServiceURL
		} else if strings.HasPrefix(path, "/projects") {
			path = strings.TrimPrefix(path, "/projects")
			serviceURL = g.ProjectServiceURL
		} else if strings.HasPrefix(path, "/tasks") {
			path = strings.TrimPrefix(path, "/tasks")
			serviceURL = g.TaskServiceURL
		} else if strings.HasPrefix(path, "/chats") {
			path = strings.TrimPrefix(path, "/chats")
			serviceURL = g.ChatServiceURL
		} else if strings.HasPrefix(path, "/reports") {
			path = strings.TrimPrefix(path, "/reports")
			serviceURL = g.ReportServiceURL
		} else {
			serviceURL = targetURL
		}

		url := serviceURL + path
		if r.URL.RawQuery != "" {
			url += "?" + r.URL.RawQuery
		}

		var body io.Reader
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				body = bytes.NewBuffer(bodyBytes)
			}
		}

		req, err := http.NewRequest(r.Method, url, body)
		if err != nil {
			http.Error(w, "Не удалось создать запрос", http.StatusInternalServerError)
			return
		}

		for key, values := range r.Header {
			if key != "Host" {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
		}

		resp, err := g.Client.Do(req)
		if err != nil {
			http.Error(w, "Не удалось перенаправить запрос", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (g *Gateway) proxyWebSocketHandler(targetURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if websocket.IsWebSocketUpgrade(r) {
			path := strings.TrimPrefix(r.URL.Path, "/api/chats")
			wsURL := strings.Replace(targetURL+path, "http://", "ws://", 1)
			wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}

			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Printf("WebSocket upgrade error: %v", err)
				return
			}
			defer conn.Close()

			targetConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				log.Printf("Failed to connect to target WebSocket: %v", err)
				return
			}
			defer targetConn.Close()

			go func() {
				for {
					messageType, message, err := conn.ReadMessage()
					if err != nil {
						break
					}
					if err := targetConn.WriteMessage(messageType, message); err != nil {
						break
					}
				}
			}()

			for {
				messageType, message, err := targetConn.ReadMessage()
				if err != nil {
					break
				}
				if err := conn.WriteMessage(messageType, message); err != nil {
					break
				}
			}
		} else {
			g.proxyHandler(targetURL)(w, r)
		}
	}
}

func uploadsDir() string {
	dir := getEnv("UPLOADS_DIR", "")
	if dir == "" {
		workDir, _ := os.Getwd()
		if strings.HasSuffix(workDir, "gateway") {
			dir = filepath.Join(workDir, "..", "uploads")
		} else {
			dir = filepath.Join(workDir, "uploads")
		}
	}
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	repl := func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-' || r == '.' || r == '_':
			return r
		case r >= 'А' && r <= 'я', r == 'Ё', r == 'ё':
			return r
		}
		return '_'
	}
	name = strings.Map(repl, name)
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	if len(name) > 80 {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		ext := filepath.Ext(name)
		if len(base) > 80-len(ext) {
			base = base[:80-len(ext)]
		}
		name = base + ext
	}
	if name == "" || name == "." {
		name = "file"
	}
	return name
}

func (g *Gateway) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Некорректная форма загрузки", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Файл не передан", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var idBytes [6]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		http.Error(w, "Ошибка генерации идентификатора", http.StatusInternalServerError)
		return
	}
	id := hex.EncodeToString(idBytes[:])

	safeName := sanitizeFilename(header.Filename)
	stored := id + "__" + safeName

	dst, err := os.Create(filepath.Join(uploadsDir(), stored))
	if err != nil {
		http.Error(w, "Ошибка сохранения файла", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Ошибка записи файла", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":  "/api/files/" + stored,
		"name": header.Filename,
	})
}

func (g *Gateway) fileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := filepath.Base(vars["name"])
	if name == "" || strings.Contains(name, "..") {
		http.Error(w, "Некорректное имя файла", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filepath.Join(uploadsDir(), name))
}

func getFrontendPath() string {
	workDir, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: failed to get working directory: %v", err)
		return "./frontend"
	}

	if strings.HasSuffix(workDir, "gateway") {
		return filepath.Join(workDir, "..", "frontend")
	}

	frontendPath := filepath.Join(workDir, "frontend")
	if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
		log.Printf("Warning: frontend directory not found at %s, trying relative path", frontendPath)
		return "./frontend"
	}

	return frontendPath
}
