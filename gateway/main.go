package main

import (
	"bytes"
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

	r.PathPrefix("/api/users").HandlerFunc(gateway.proxyHandler(gateway.UserServiceURL))
	r.PathPrefix("/api/projects").HandlerFunc(gateway.proxyHandler(gateway.ProjectServiceURL))
	r.PathPrefix("/api/tasks").HandlerFunc(gateway.proxyHandler(gateway.TaskServiceURL))
	r.PathPrefix("/api/chats").HandlerFunc(gateway.proxyWebSocketHandler(gateway.ChatServiceURL))
	r.PathPrefix("/api/reports").HandlerFunc(gateway.proxyHandler(gateway.ReportServiceURL))

	frontendPath := getFrontendPath()
	log.Printf("Serving frontend from: %s", frontendPath)

	fs := http.FileServer(http.Dir(frontendPath))

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			http.ServeFile(w, r, filepath.Join(frontendPath, "index.html"))
			return
		}
		http.StripPrefix("/", fs).ServeHTTP(w, r)
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
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
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
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
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

			// Подключение к целевому WebSocket серверу
			targetConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				log.Printf("Failed to connect to target WebSocket: %v", err)
				return
			}
			defer targetConn.Close()

			// Проксирование сообщений
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
			// Обычный HTTP запрос
			g.proxyHandler(targetURL)(w, r)
		}
	}
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
