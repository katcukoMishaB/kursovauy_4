package main

import (
	"database/sql"
	"encoding/json"
	"kursovauy_4/internal/database"
	"kursovauy_4/internal/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
)

var db *sql.DB

func main() {
	gormDB, err := database.ConnectGORM()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db, err = gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get sql db:", err)
	}
	defer db.Close()

	r := gin.Default()
	r.POST("/projects/:project_id/tasks", adaptWithParams(middleware.AuthMiddleware(createTaskHandler), "project_id"))
	r.GET("/projects/:project_id/tasks", adaptWithParams(getTasksHandler, "project_id"))
	r.GET("/tasks/:id", adaptWithParams(getTaskHandler, "id"))
	r.PUT("/tasks/:id", adaptWithParams(middleware.OrganizerMiddleware(updateTaskHandler), "id"))
	r.POST("/tasks/:id/assign", adaptWithParams(middleware.OrganizerMiddleware(assignTaskHandler), "id"))
	r.PUT("/tasks/:id/status", adaptWithParams(middleware.AuthMiddleware(updateTaskStatusHandler), "id"))
	r.POST("/tasks/:id/comments", adaptWithParams(middleware.AuthMiddleware(addCommentHandler), "id"))
	r.GET("/tasks/:id/comments", adaptWithParams(getCommentsHandler, "id"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	log.Printf("Task service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func adaptWithParams(handler http.HandlerFunc, params ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		vars := map[string]string{}
		for _, param := range params {
			vars[param] = c.Param(param)
		}
		req := mux.SetURLVars(c.Request, vars)
		handler(c.Writer, req)
	}
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["project_id"]

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var taskID string
	err := db.QueryRow(
		"INSERT INTO project_tasks (project_id, title, description, status, creation_date) VALUES ($1, $2, $3, 'новая', CURRENT_DATE) RETURNING id",
		projectID, req.Title, req.Description,
	).Scan(&taskID)

	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": taskID, "message": "Task created successfully"})
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["project_id"]
	status := r.URL.Query().Get("status")

	query := "SELECT id, project_id, assigned_to, title, description, status, creation_date FROM project_tasks WHERE project_id = $1"
	args := []interface{}{projectID}
	argPos := 2

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
		argPos++
	}

	query += " ORDER BY creation_date DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.ProjectID, &t.AssignedTo, &t.Title, &t.Description, &t.Status, &t.CreationDate)
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var t Task
	err := db.QueryRow(
		"SELECT id, project_id, assigned_to, title, description, status, creation_date FROM project_tasks WHERE id = $1",
		taskID,
	).Scan(&t.ID, &t.ProjectID, &t.AssignedTo, &t.Title, &t.Description, &t.Status, &t.CreationDate)

	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		"UPDATE project_tasks SET title = $1, description = $2 WHERE id = $3",
		req.Title, req.Description, taskID,
	)
	if err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Task updated successfully"})
}

func assignTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var projectID string
	err := db.QueryRow("SELECT project_id FROM project_tasks WHERE id = $1", taskID).Scan(&projectID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var organizerID string
	err = db.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&organizerID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	isOrganizer := organizerID == userID

	var userRole string
	err = db.QueryRow(
		"SELECT role FROM project_participations WHERE project_id = $1 AND user_id = $2",
		projectID, userID,
	).Scan(&userRole)

	isManager := err == nil && (userRole == "менеджер" || userRole == "руководитель")
	isLeader := err == nil && userRole == "руководитель"

	if !isOrganizer && !isManager && !isLeader {
		http.Error(w, "Access denied: only organizers, managers, and leaders can assign tasks", http.StatusForbidden)
		return
	}

	var req struct {
		AssignedTo *string `json:"assigned_to"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var assignedTo interface{} = req.AssignedTo
	if req.AssignedTo == nil {
		assignedTo = nil
	} else {
		assignedTo = *req.AssignedTo
	}

	var statusUpdate string
	if req.AssignedTo != nil {
		statusUpdate = "в работе"
	} else {
		statusUpdate = "новая"
	}

	_, err = db.Exec(
		"UPDATE project_tasks SET assigned_to = $1, status = $2 WHERE id = $3",
		assignedTo, statusUpdate, taskID,
	)
	if err != nil {
		http.Error(w, "Failed to assign task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Task assigned successfully"})
}

func updateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var assignedTo *string
	err := db.QueryRow("SELECT assigned_to FROM project_tasks WHERE id = $1", taskID).Scan(&assignedTo)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if assignedTo != nil && *assignedTo != userID {
		http.Error(w, "Not authorized to update this task", http.StatusForbidden)
		return
	}

	_, err = db.Exec("UPDATE project_tasks SET status = $1 WHERE id = $2", req.Status, taskID)
	if err != nil {
		http.Error(w, "Failed to update task status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Task status updated successfully"})
}

func addCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var commentID string
	err := db.QueryRow(
		"INSERT INTO task_comments (task_id, user_id, content, publication_date) VALUES ($1, $2, $3, CURRENT_DATE) RETURNING id",
		taskID, userID, req.Content,
	).Scan(&commentID)

	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": commentID, "message": "Comment added successfully"})
}

func getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	rows, err := db.Query(
		"SELECT id, task_id, user_id, content, publication_date FROM task_comments WHERE task_id = $1 ORDER BY publication_date",
		taskID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []TaskComment
	for rows.Next() {
		var c TaskComment
		err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.PublicationDate)
		if err != nil {
			continue
		}
		comments = append(comments, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
