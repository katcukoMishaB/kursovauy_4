package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	r.POST("/projects", adapt(middleware.OrganizerMiddleware(createProjectHandler)))
	r.GET("/projects", adapt(getProjectsHandler))
	r.GET("/projects/my", adapt(middleware.AuthMiddleware(getMyProjectsHandler)))
	r.GET("/projects/:id", adaptWithParams(getProjectHandler, "id"))
	r.PUT("/projects/:id", adaptWithParams(middleware.OrganizerMiddleware(updateProjectHandler), "id"))
	r.POST("/projects/:id/archive", adaptWithParams(middleware.OrganizerMiddleware(archiveProjectHandler), "id"))
	r.GET("/categories", adapt(getCategoriesHandler))
	r.POST("/projects/:id/participate", adaptWithParams(middleware.AuthMiddleware(createParticipationRequestHandler), "id"))
	r.GET("/projects/:id/requests", adaptWithParams(middleware.AuthMiddleware(getParticipationRequestsHandler), "id"))
	r.GET("/projects/:id/my-request", adaptWithParams(middleware.AuthMiddleware(getMyParticipationRequestHandler), "id"))
	r.POST("/requests/:id/approve", adaptWithParams(middleware.AuthMiddleware(approveParticipationRequestHandler), "id"))
	r.POST("/requests/:id/reject", adaptWithParams(middleware.AuthMiddleware(rejectParticipationRequestHandler), "id"))
	r.GET("/projects/:id/participants", adaptWithParams(getParticipantsHandler, "id"))
	r.PUT("/projects/:id/participants/:userId/role", adaptWithParams(middleware.AuthMiddleware(updateParticipantRoleHandler), "id", "userId"))
	r.POST("/projects/:id/tags", adaptWithParams(middleware.OrganizerMiddleware(addTagHandler), "id"))
	r.GET("/projects/:id/tags", adaptWithParams(getProjectTagsHandler, "id"))
	r.DELETE("/projects/:id/tags/:tagName", adaptWithParams(middleware.AuthMiddleware(deleteTagHandler), "id", "tagName"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	log.Printf("Project service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func adapt(handler http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c.Writer, c.Request)
	}
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

func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	organizerID := r.Header.Get("X-User-ID")

	var req struct {
		CategoryID       *string `json:"category_id"`
		Title            string  `json:"title"`
		ShortDescription string  `json:"short_description"`
		FullDescription  string  `json:"full_description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create project"})
		return
	}
	defer tx.Rollback()

	var projectID string
	err = tx.QueryRow(
		"INSERT INTO projects (organizer_id, category_id, title, short_description, full_description, status, creation_date) VALUES ($1, $2, $3, $4, $5, 'активен', CURRENT_DATE) RETURNING id",
		organizerID, req.CategoryID, req.Title, req.ShortDescription, req.FullDescription,
	).Scan(&projectID)

	if err != nil {
		log.Printf("Error creating project: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create project"})
		return
	}

	_, err = tx.Exec(
		"INSERT INTO project_participations (project_id, user_id, role, join_date) VALUES ($1, $2, 'руководитель', CURRENT_DATE) ON CONFLICT (project_id, user_id) DO NOTHING",
		projectID, organizerID,
	)
	if err != nil {
		log.Printf("Error adding organizer as participant: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create project"})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create project"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": projectID, "message": "Project created successfully"})
}

func getProjectsHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	categoryID := r.URL.Query().Get("category_id")

	query := `
		SELECT 
			p.id, p.organizer_id, p.category_id, p.title, p.short_description, 
			p.full_description, p.status, p.creation_date, p.completion_date,
			COALESCE(u.first_name || ' ' || u.last_name, 'Неизвестно') as organizer_name,
			COALESCE(COUNT(DISTINCT pp.user_id), 0) as participants_count,
			COALESCE(c.name, '') as category_name
		FROM projects p
		LEFT JOIN users u ON p.organizer_id = u.id
		LEFT JOIN project_participations pp ON p.id = pp.project_id
		LEFT JOIN project_categories c ON p.category_id = c.id
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if status != "" {
		query += fmt.Sprintf(" AND p.status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	if categoryID != "" {
		query += fmt.Sprintf(" AND p.category_id = $%d", argPos)
		args = append(args, categoryID)
		argPos++
	}

	query += " GROUP BY p.id, u.first_name, u.last_name, c.id, c.name ORDER BY p.creation_date DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error fetching projects: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch projects"})
		return
	}
	defer rows.Close()

	projects := []ProjectExtended{}
	for rows.Next() {
		var p ProjectExtended
		err := rows.Scan(&p.ID, &p.OrganizerID, &p.CategoryID, &p.Title, &p.ShortDescription,
			&p.FullDescription, &p.Status, &p.CreationDate, &p.CompletionDate,
			&p.OrganizerName, &p.ParticipantsCount, &p.CategoryName)
		if err != nil {
			log.Printf("Error scanning project: %v", err)
			continue
		}

		// Получаем теги для проекта
		tagRows, err := db.Query("SELECT name FROM project_tags WHERE project_id = $1", p.ID)
		if err == nil {
			p.Tags = []string{}
			for tagRows.Next() {
				var tag string
				if err := tagRows.Scan(&tag); err == nil {
					p.Tags = append(p.Tags, tag)
				}
			}
			tagRows.Close()
		}

		projects = append(projects, p)
	}

	if projects == nil {
		projects = []ProjectExtended{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func getMyProjectsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	rows, err := db.Query(`
		SELECT DISTINCT p.id, p.organizer_id, p.category_id, p.title, p.short_description, p.full_description, p.status, p.creation_date, p.completion_date,
			CASE 
				WHEN p.organizer_id = $1 THEN 'organizer'
				WHEN pp.role = 'руководитель' THEN 'leader'
				ELSE 'participant'
			END as user_role
		FROM projects p
		LEFT JOIN project_participations pp ON p.id = pp.project_id AND pp.user_id = $1
		WHERE p.organizer_id = $1 OR pp.user_id = $1
		ORDER BY p.creation_date DESC
	`, userID)
	if err != nil {
		log.Printf("Error fetching my projects: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch projects"})
		return
	}
	defer rows.Close()

	type ProjectWithRole struct {
		Project
		UserRole string `json:"user_role"`
	}

	projects := []ProjectWithRole{}
	for rows.Next() {
		var p ProjectWithRole
		err := rows.Scan(&p.ID, &p.OrganizerID, &p.CategoryID, &p.Title, &p.ShortDescription, &p.FullDescription, &p.Status, &p.CreationDate, &p.CompletionDate, &p.UserRole)
		if err != nil {
			log.Printf("Error scanning project: %v", err)
			continue
		}
		projects = append(projects, p)
	}

	if projects == nil {
		projects = []ProjectWithRole{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func getProjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	var p Project
	err := db.QueryRow(
		"SELECT id, organizer_id, category_id, title, short_description, full_description, status, creation_date, completion_date FROM projects WHERE id = $1",
		projectID,
	).Scan(&p.ID, &p.OrganizerID, &p.CategoryID, &p.Title, &p.ShortDescription, &p.FullDescription, &p.Status, &p.CreationDate, &p.CompletionDate)

	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}
	if err != nil {
		log.Printf("Database error in getProject: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var req struct {
		CategoryID       *string `json:"category_id"`
		Title            string  `json:"title"`
		ShortDescription string  `json:"short_description"`
		FullDescription  string  `json:"full_description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingOrganizerID string
	err := db.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&existingOrganizerID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	isOrganizer := existingOrganizerID == userID

	var userRole string
	err = db.QueryRow(
		"SELECT role FROM project_participations WHERE project_id = $1 AND user_id = $2",
		projectID, userID,
	).Scan(&userRole)

	isManager := err == nil && (userRole == "менеджер" || userRole == "руководитель")
	isLeader := err == nil && userRole == "руководитель"

	if !isOrganizer && !isManager && !isLeader {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}

	_, err = db.Exec(
		"UPDATE projects SET category_id = $1, title = $2, short_description = $3, full_description = $4 WHERE id = $5",
		req.CategoryID, req.Title, req.ShortDescription, req.FullDescription, projectID,
	)
	if err != nil {
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Project updated successfully"})
}

func archiveProjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var existingOrganizerID string
	err := db.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&existingOrganizerID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	isOrganizer := existingOrganizerID == userID

	var userRole string
	err = db.QueryRow(
		"SELECT role FROM project_participations WHERE project_id = $1 AND user_id = $2",
		projectID, userID,
	).Scan(&userRole)

	isManager := err == nil && (userRole == "менеджер" || userRole == "руководитель")
	isLeader := err == nil && userRole == "руководитель"

	if !isOrganizer && !isManager && !isLeader {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}

	_, err = db.Exec(
		"UPDATE projects SET status = 'архивирован', completion_date = CURRENT_DATE WHERE id = $1",
		projectID,
	)
	if err != nil {
		http.Error(w, "Failed to archive project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Project archived successfully"})
}

func getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name FROM project_categories ORDER BY name")
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []ProjectCategory
	for rows.Next() {
		var c ProjectCategory
		err := rows.Scan(&c.ID, &c.Name)
		if err != nil {
			continue
		}
		categories = append(categories, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func createParticipationRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var req struct {
		Comment   string `json:"comment"`
		ResumeURL string `json:"resume_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	var organizerID string
	err := db.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&organizerID)
	if err != nil {
		log.Printf("Error fetching project: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	if organizerID == userID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Вы уже являетесь организатором этого проекта"})
		return
	}

	var existingParticipation string
	err = db.QueryRow(
		"SELECT id FROM project_participations WHERE project_id = $1 AND user_id = $2",
		projectID, userID,
	).Scan(&existingParticipation)

	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Вы уже являетесь участником этого проекта"})
		return
	}
	if err != sql.ErrNoRows {
		log.Printf("Error checking participation: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}

	var existingRequest string
	err = db.QueryRow(
		"SELECT id FROM project_participation_requests WHERE project_id = $1 AND user_id = $2 AND status = 'в рассмотрении'",
		projectID, userID,
	).Scan(&existingRequest)

	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "У вас уже есть активная заявка на этот проект"})
		return
	}
	if err != sql.ErrNoRows {
		log.Printf("Error checking existing request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}

	var requestID string
	err = db.QueryRow(
		"INSERT INTO project_participation_requests (project_id, user_id, comment, resume_url, submission_date, status) VALUES ($1, $2, $3, $4, CURRENT_DATE, 'в рассмотрении') RETURNING id",
		projectID, userID, req.Comment, req.ResumeURL,
	).Scan(&requestID)

	if err != nil {
		log.Printf("Error creating participation request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create participation request"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": requestID, "message": "Request created successfully"})
}

func getParticipationRequestsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var existingOrganizerID string
	err := db.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&existingOrganizerID)
	if err != nil {
		log.Printf("Error fetching project: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	var userRole string
	err = db.QueryRow("SELECT role FROM project_participations WHERE project_id = $1 AND user_id = $2", projectID, userID).Scan(&userRole)
	isOrganizer := existingOrganizerID == userID
	isManager := err == nil && (userRole == "менеджер" || userRole == "руководитель")

	if !isOrganizer && !isManager {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authorized"})
		return
	}

	rows, err := db.Query(
		"SELECT id, project_id, user_id, comment, resume_url, submission_date, status FROM project_participation_requests WHERE project_id = $1 ORDER BY submission_date DESC",
		projectID,
	)
	if err != nil {
		log.Printf("Error fetching requests: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch requests"})
		return
	}
	defer rows.Close()

	requests := []ParticipationRequest{}
	for rows.Next() {
		var req ParticipationRequest
		err := rows.Scan(&req.ID, &req.ProjectID, &req.UserID, &req.Comment, &req.ResumeURL, &req.SubmissionDate, &req.Status)
		if err != nil {
			log.Printf("Error scanning request: %v", err)
			continue
		}
		requests = append(requests, req)
	}

	if requests == nil {
		requests = []ParticipationRequest{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func approveParticipationRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]
	currentUserID := r.Header.Get("X-User-ID")

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Transaction error"})
		return
	}
	defer tx.Rollback()

	var projectID, userID string
	err = tx.QueryRow("SELECT project_id, user_id FROM project_participation_requests WHERE id = $1", requestID).Scan(&projectID, &userID)
	if err != nil {
		log.Printf("Error fetching request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Request not found"})
		return
	}

	var existingOrganizerID string
	err = tx.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&existingOrganizerID)
	if err != nil {
		log.Printf("Error fetching project: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	var userRole string
	err = tx.QueryRow("SELECT role FROM project_participations WHERE project_id = $1 AND user_id = $2", projectID, currentUserID).Scan(&userRole)
	isOrganizer := existingOrganizerID == currentUserID
	isLeader := err == nil && userRole == "руководитель"

	if !isOrganizer && !isLeader {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authorized"})
		return
	}

	_, err = tx.Exec("UPDATE project_participation_requests SET status = 'одобрена' WHERE id = $1", requestID)
	if err != nil {
		log.Printf("Error updating request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update request"})
		return
	}

	_, err = tx.Exec(
		"INSERT INTO project_participations (project_id, user_id, role, join_date) VALUES ($1, $2, 'участник', CURRENT_DATE) ON CONFLICT DO NOTHING",
		projectID, userID,
	)
	if err != nil {
		log.Printf("Error creating participation: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create participation"})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to approve request"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Request approved successfully"})
}

func rejectParticipationRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]
	currentUserID := r.Header.Get("X-User-ID")

	var projectID string
	err := db.QueryRow("SELECT project_id FROM project_participation_requests WHERE id = $1", requestID).Scan(&projectID)
	if err != nil {
		log.Printf("Error fetching request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Request not found"})
		return
	}

	var existingOrganizerID string
	err = db.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&existingOrganizerID)
	if err != nil {
		log.Printf("Error fetching project: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	var userRole string
	err = db.QueryRow("SELECT role FROM project_participations WHERE project_id = $1 AND user_id = $2", projectID, currentUserID).Scan(&userRole)
	isOrganizer := existingOrganizerID == currentUserID
	isManager := err == nil && (userRole == "менеджер" || userRole == "руководитель")

	if !isOrganizer && !isManager {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authorized"})
		return
	}

	_, err = db.Exec("UPDATE project_participation_requests SET status = 'отклонена' WHERE id = $1", requestID)
	if err != nil {
		log.Printf("Error rejecting request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to reject request"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Request rejected successfully"})
}

func getMyParticipationRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var request ParticipationRequest
	err := db.QueryRow(
		"SELECT id, project_id, user_id, comment, resume_url, submission_date, status FROM project_participation_requests WHERE project_id = $1 AND user_id = $2 ORDER BY submission_date DESC LIMIT 1",
		projectID, userID,
	).Scan(&request.ID, &request.ProjectID, &request.UserID, &request.Comment, &request.ResumeURL, &request.SubmissionDate, &request.Status)

	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "No request found"})
		return
	}

	if err != nil {
		log.Printf("Error fetching participation request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
}

func getParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	rows, err := db.Query(`
		SELECT pp.id, pp.project_id, pp.user_id, pp.role, pp.join_date,
		       u.email, u.first_name, u.last_name
		FROM project_participations pp
		LEFT JOIN users u ON pp.user_id = u.id
		WHERE pp.project_id = $1
		ORDER BY pp.join_date
	`, projectID)
	if err != nil {
		log.Printf("Error fetching participants: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch participants"})
		return
	}
	defer rows.Close()

	type ParticipantWithUser struct {
		Participation
		Email     *string `json:"email"`
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
	}

	participants := []ParticipantWithUser{}
	for rows.Next() {
		var p ParticipantWithUser
		err := rows.Scan(&p.ID, &p.ProjectID, &p.UserID, &p.Role, &p.JoinDate, &p.Email, &p.FirstName, &p.LastName)
		if err != nil {
			log.Printf("Error scanning participant: %v", err)
			continue
		}
		participants = append(participants, p)
	}

	if participants == nil {
		participants = []ParticipantWithUser{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participants)
}

func updateParticipantRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	userID := vars["userId"]
	currentUserID := r.Header.Get("X-User-ID")

	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Role != "участник" && req.Role != "менеджер" && req.Role != "руководитель" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid role. Must be участник, менеджер, or руководитель"})
		return
	}

	var existingOrganizerID string
	err := db.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&existingOrganizerID)
	if err != nil {
		log.Printf("Error fetching project: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	var userRole string
	err = db.QueryRow("SELECT role FROM project_participations WHERE project_id = $1 AND user_id = $2", projectID, currentUserID).Scan(&userRole)
	isOrganizer := existingOrganizerID == currentUserID
	isManager := err == nil && (userRole == "менеджер" || userRole == "руководитель")

	if !isOrganizer && !isManager {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authorized"})
		return
	}

	if existingOrganizerID == userID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot change organizer role"})
		return
	}

	_, err = db.Exec(
		"UPDATE project_participations SET role = $1 WHERE project_id = $2 AND user_id = $3",
		req.Role, projectID, userID,
	)
	if err != nil {
		log.Printf("Error updating participant role: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update participant role"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Participant role updated successfully"})
}

func addTagHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingOrganizerID string
	err := db.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&existingOrganizerID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	isOrganizer := existingOrganizerID == userID

	var userRole string
	err = db.QueryRow(
		"SELECT role FROM project_participations WHERE project_id = $1 AND user_id = $2",
		projectID, userID,
	).Scan(&userRole)

	isManager := err == nil && (userRole == "менеджер" || userRole == "руководитель")
	isLeader := err == nil && userRole == "руководитель"

	if !isOrganizer && !isManager && !isLeader {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}

	_, err = db.Exec(
		"INSERT INTO project_tags (project_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		projectID, req.Name,
	)
	if err != nil {
		http.Error(w, "Failed to add tag", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tag added successfully"})
}

func getProjectTagsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	rows, err := db.Query("SELECT name FROM project_tags WHERE project_id = $1", projectID)
	if err != nil {
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func deleteTagHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	tagName := vars["tagName"]
	currentUserID := r.Header.Get("X-User-ID")

	var existingOrganizerID string
	err := db.QueryRow("SELECT organizer_id FROM projects WHERE id = $1", projectID).Scan(&existingOrganizerID)
	if err != nil {
		log.Printf("Error fetching project: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	var userRole string
	err = db.QueryRow("SELECT role FROM project_participations WHERE project_id = $1 AND user_id = $2", projectID, currentUserID).Scan(&userRole)
	isOrganizer := existingOrganizerID == currentUserID
	isManager := err == nil && (userRole == "менеджер" || userRole == "руководитель")

	if !isOrganizer && !isManager {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authorized"})
		return
	}

	_, err = db.Exec(
		"DELETE FROM project_tags WHERE project_id = $1 AND name = $2",
		projectID, tagName,
	)
	if err != nil {
		log.Printf("Error deleting tag: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete tag"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tag deleted successfully"})
}
