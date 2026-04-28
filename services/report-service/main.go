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
	r.GET("/users", adapt(middleware.AdminMiddleware(getUserActivityReportHandler)))
	r.GET("/users/:id", adaptWithParams(middleware.AdminMiddleware(getUserActivityDetailHandler), "id"))
	r.GET("/projects", adapt(middleware.AdminMiddleware(getProjectEfficiencyReportHandler)))
	r.GET("/projects/:id", adaptWithParams(middleware.AdminMiddleware(getProjectEfficiencyDetailHandler), "id"))
	r.GET("/summary", adapt(middleware.AdminMiddleware(getSummaryReportHandler)))

	r.GET("/excel/project/:id", adaptWithParams(middleware.AuthMiddleware(generateProjectExcelReportHandler), "id"))
	r.GET("/excel/all-projects", adapt(middleware.AdminMiddleware(generateAllProjectsExcelReportHandler)))
	r.GET("/excel/admin/project/:id", adaptWithParams(middleware.AdminMiddleware(generateAdminProjectExcelReportHandler), "id"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	log.Printf("Report service starting on port %s", port)
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

func getUserActivityReportHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT 
			u.id,
			u.first_name,
			u.last_name,
			u.email,
			u.registration_date,
			COUNT(DISTINCT pp.project_id) as projects_count,
			COUNT(DISTINCT CASE WHEN pt.status = 'завершена' THEN pt.id END) as tasks_completed,
			COUNT(DISTINCT cm.id) as messages_sent,
			COUNT(DISTINCT tc.id) as comments_count
		FROM users u
		LEFT JOIN project_participations pp ON u.id = pp.user_id
		LEFT JOIN project_tasks pt ON u.id = pt.assigned_to
		LEFT JOIN chat_messages cm ON u.id = cm.user_id
		LEFT JOIN task_comments tc ON u.id = tc.user_id
		GROUP BY u.id, u.first_name, u.last_name, u.email, u.registration_date
		ORDER BY u.registration_date DESC
	`)
	if err != nil {
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reports []UserActivityReport
	for rows.Next() {
		var report UserActivityReport
		err := rows.Scan(
			&report.UserID,
			&report.FirstName,
			&report.LastName,
			&report.Email,
			&report.RegistrationDate,
			&report.ProjectsCount,
			&report.TasksCompleted,
			&report.MessagesSent,
			&report.CommentsCount,
		)
		if err != nil {
			continue
		}
		reports = append(reports, report)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

func getUserActivityDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var report UserActivityReport
	err := db.QueryRow(`
		SELECT 
			u.id,
			u.first_name,
			u.last_name,
			u.email,
			u.registration_date,
			COUNT(DISTINCT pp.project_id) as projects_count,
			COUNT(DISTINCT CASE WHEN pt.status = 'завершена' THEN pt.id END) as tasks_completed,
			COUNT(DISTINCT cm.id) as messages_sent,
			COUNT(DISTINCT tc.id) as comments_count
		FROM users u
		LEFT JOIN project_participations pp ON u.id = pp.user_id
		LEFT JOIN project_tasks pt ON u.id = pt.assigned_to
		LEFT JOIN chat_messages cm ON u.id = cm.user_id
		LEFT JOIN task_comments tc ON u.id = tc.user_id
		WHERE u.id = $1
		GROUP BY u.id, u.first_name, u.last_name, u.email, u.registration_date
	`, userID).Scan(
		&report.UserID,
		&report.FirstName,
		&report.LastName,
		&report.Email,
		&report.RegistrationDate,
		&report.ProjectsCount,
		&report.TasksCompleted,
		&report.MessagesSent,
		&report.CommentsCount,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func getProjectEfficiencyReportHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT 
			p.id,
			p.title,
			p.status,
			p.creation_date,
			p.completion_date,
			COUNT(DISTINCT pp.user_id) as participants_count,
			COUNT(DISTINCT pt.id) as tasks_total,
			COUNT(DISTINCT CASE WHEN pt.status = 'завершена' THEN pt.id END) as tasks_completed,
			CASE 
				WHEN COUNT(DISTINCT pt.id) > 0 
				THEN (COUNT(DISTINCT CASE WHEN pt.status = 'завершена' THEN pt.id END)::float / COUNT(DISTINCT pt.id)::float) * 100
				ELSE 0
			END as completion_rate
		FROM projects p
		LEFT JOIN project_participations pp ON p.id = pp.project_id
		LEFT JOIN project_tasks pt ON p.id = pt.project_id
		GROUP BY p.id, p.title, p.status, p.creation_date, p.completion_date
		ORDER BY p.creation_date DESC
	`)
	if err != nil {
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reports []ProjectEfficiencyReport
	for rows.Next() {
		var report ProjectEfficiencyReport
		err := rows.Scan(
			&report.ProjectID,
			&report.Title,
			&report.Status,
			&report.CreationDate,
			&report.CompletionDate,
			&report.ParticipantsCount,
			&report.TasksTotal,
			&report.TasksCompleted,
			&report.CompletionRate,
		)
		if err != nil {
			continue
		}
		reports = append(reports, report)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

func getProjectEfficiencyDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	var report ProjectEfficiencyReport
	err := db.QueryRow(`
		SELECT 
			p.id,
			p.title,
			p.status,
			p.creation_date,
			p.completion_date,
			COUNT(DISTINCT pp.user_id) as participants_count,
			COUNT(DISTINCT pt.id) as tasks_total,
			COUNT(DISTINCT CASE WHEN pt.status = 'завершена' THEN pt.id END) as tasks_completed,
			CASE 
				WHEN COUNT(DISTINCT pt.id) > 0 
				THEN (COUNT(DISTINCT CASE WHEN pt.status = 'завершена' THEN pt.id END)::float / COUNT(DISTINCT pt.id)::float) * 100
				ELSE 0
			END as completion_rate
		FROM projects p
		LEFT JOIN project_participations pp ON p.id = pp.project_id
		LEFT JOIN project_tasks pt ON p.id = pt.project_id
		WHERE p.id = $1
		GROUP BY p.id, p.title, p.status, p.creation_date, p.completion_date
	`, projectID).Scan(
		&report.ProjectID,
		&report.Title,
		&report.Status,
		&report.CreationDate,
		&report.CompletionDate,
		&report.ParticipantsCount,
		&report.TasksTotal,
		&report.TasksCompleted,
		&report.CompletionRate,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func getSummaryReportHandler(w http.ResponseWriter, r *http.Request) {
	var summary struct {
		TotalUsers            int     `json:"total_users"`
		ActiveUsers           int     `json:"active_users"`
		TotalProjects         int     `json:"total_projects"`
		ActiveProjects        int     `json:"active_projects"`
		CompletedProjects     int     `json:"completed_projects"`
		TotalTasks            int     `json:"total_tasks"`
		CompletedTasks        int     `json:"completed_tasks"`
		AverageCompletionRate float64 `json:"average_completion_rate"`
	}

	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&summary.TotalUsers)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE status = true").Scan(&summary.ActiveUsers)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&summary.TotalProjects)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM projects WHERE status = 'активен'").Scan(&summary.ActiveProjects)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM projects WHERE status = 'завершён'").Scan(&summary.CompletedProjects)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM project_tasks").Scan(&summary.TotalTasks)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM project_tasks WHERE status = 'завершена'").Scan(&summary.CompletedTasks)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow(`
		SELECT 
			CASE 
				WHEN COUNT(*) > 0 
				THEN (COUNT(CASE WHEN status = 'завершена' THEN 1 END)::float / COUNT(*)::float) * 100
				ELSE 0
			END
		FROM project_tasks
	`).Scan(&summary.AverageCompletionRate)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
