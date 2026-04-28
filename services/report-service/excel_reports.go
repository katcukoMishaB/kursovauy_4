package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
)

func generateProjectExcelReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]
	userID := r.Header.Get("X-User-ID")

	var projectTitle string
	var organizerID string
	err := db.QueryRow("SELECT title, organizer_id FROM projects WHERE id = $1", projectID).Scan(&projectTitle, &organizerID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	var userRole string
	err = db.QueryRow(`
		SELECT role FROM project_participations 
		WHERE project_id = $1 AND user_id = $2
	`, projectID, userID).Scan(&userRole)
	isLeader := err == nil && (userRole == "руководитель" || userRole == "менеджер")

	if organizerID != userID && !isLeader {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	sheet1 := "Общая информация"
	f.NewSheet(sheet1)
	f.DeleteSheet("Sheet1")

	var tasksCompleted, tasksNotCompleted int
	err = db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN status = 'завершена' THEN 1 END) as completed,
			COUNT(CASE WHEN status != 'завершена' THEN 1 END) as not_completed
		FROM project_tasks 
		WHERE project_id = $1
	`, projectID).Scan(&tasksCompleted, &tasksNotCompleted)
	if err != nil {
		http.Error(w, "Failed to fetch task statistics", http.StatusInternalServerError)
		return
	}

	f.SetCellValue(sheet1, "A1", "Проект")
	f.SetCellValue(sheet1, "B1", projectTitle)
	f.SetCellValue(sheet1, "A2", "Задач завершено")
	f.SetCellValue(sheet1, "B2", tasksCompleted)
	f.SetCellValue(sheet1, "A3", "Задач не завершено")
	f.SetCellValue(sheet1, "B3", tasksNotCompleted)

	sheet2 := "Участники"
	f.NewSheet(sheet2)

	rows, err := db.Query(`
		SELECT 
		COALESCE(u.email, '') as email,
		COALESCE(u.first_name || ' ' || u.last_name, '') as name,
		pp.role,
		COALESCE(completed_tasks.cnt, 0) as tasks_completed,
		COALESCE(not_completed_tasks.cnt, 0) as tasks_not_completed
	FROM project_participations pp
	JOIN users u ON pp.user_id = u.id
	LEFT JOIN (
		SELECT assigned_to, COUNT(*) as cnt
		FROM project_tasks 
		WHERE project_id = $1 AND status = 'завершена'
		GROUP BY assigned_to
	) completed_tasks ON u.id = completed_tasks.assigned_to
	LEFT JOIN (
		SELECT assigned_to, COUNT(*) as cnt
		FROM project_tasks 
		WHERE project_id = $1 
			AND status IS NOT NULL 
			AND status != 'завершена'
		GROUP BY assigned_to
	) not_completed_tasks ON u.id = not_completed_tasks.assigned_to
	WHERE pp.project_id = $1
	ORDER BY pp.role, u.email
	`, projectID)
	if err != nil {
		http.Error(w, "Failed to fetch participants", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	f.SetCellValue(sheet2, "A1", "Email")
	f.SetCellValue(sheet2, "B1", "Имя")
	f.SetCellValue(sheet2, "C1", "Роль")
	f.SetCellValue(sheet2, "D1", "Задач выполнено")
	f.SetCellValue(sheet2, "E1", "Задач не выполнено")

	row := 2
	for rows.Next() {
		var email, name, role string
		var tasksCompleted, tasksNotCompleted int
		err := rows.Scan(&email, &name, &role, &tasksCompleted, &tasksNotCompleted)
		if err != nil {
			continue
		}
		f.SetCellValue(sheet2, fmt.Sprintf("A%d", row), email)
		f.SetCellValue(sheet2, fmt.Sprintf("B%d", row), name)
		f.SetCellValue(sheet2, fmt.Sprintf("C%d", row), role)
		f.SetCellValue(sheet2, fmt.Sprintf("D%d", row), tasksCompleted)
		f.SetCellValue(sheet2, fmt.Sprintf("E%d", row), tasksNotCompleted)
		row++
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=project_report_%s.xlsx", projectID))
	if err := f.Write(w); err != nil {
		log.Printf("Error writing file: %v", err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}
}

func generateAdminProjectExcelReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	var projectTitle, organizerID string
	err := db.QueryRow("SELECT title, organizer_id FROM projects WHERE id = $1", projectID).Scan(&projectTitle, &organizerID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	var organizerEmail string
	err = db.QueryRow("SELECT email FROM users WHERE id = $1", organizerID).Scan(&organizerEmail)
	if err != nil {
		organizerEmail = "Неизвестно"
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	sheet1 := "Общая информация"
	f.NewSheet(sheet1)
	f.DeleteSheet("Sheet1")

	var tasksCompleted, tasksNotCompleted int
	err = db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN status = 'завершена' THEN 1 END) as completed,
			COUNT(CASE WHEN status != 'завершена' THEN 1 END) as not_completed
		FROM project_tasks 
		WHERE project_id = $1
	`, projectID).Scan(&tasksCompleted, &tasksNotCompleted)
	if err != nil {
		http.Error(w, "Failed to fetch task statistics", http.StatusInternalServerError)
		return
	}

	f.SetCellValue(sheet1, "A1", "ID проекта")
	f.SetCellValue(sheet1, "B1", projectID)
	f.SetCellValue(sheet1, "A2", "Название")
	f.SetCellValue(sheet1, "B2", projectTitle)
	f.SetCellValue(sheet1, "A3", "Руководитель")
	f.SetCellValue(sheet1, "B3", organizerEmail)
	f.SetCellValue(sheet1, "A4", "Задач завершено")
	f.SetCellValue(sheet1, "B4", tasksCompleted)
	f.SetCellValue(sheet1, "A5", "Задач не завершено")
	f.SetCellValue(sheet1, "B5", tasksNotCompleted)

	sheet2 := "Участники"
	f.NewSheet(sheet2)

	rows, err := db.Query(`
    SELECT 
        COALESCE(u.email, '') as email,
        COALESCE(u.first_name || ' ' || u.last_name, '') as name,
        pp.role,
        COALESCE((
            SELECT COUNT(*) 
            FROM project_tasks 
            WHERE project_id = $1 
                AND assigned_to = u.id 
                AND status = 'завершена'
        ), 0) as tasks_completed,
        COALESCE((
            SELECT COUNT(*) 
            FROM project_tasks 
            WHERE project_id = $1 
                AND assigned_to = u.id 
                AND status IS NOT NULL 
                AND status != 'завершена'
        ), 0) as tasks_not_completed
    FROM project_participations pp
    JOIN users u ON pp.user_id = u.id
    WHERE pp.project_id = $1
    ORDER BY 
        CASE pp.role 
            WHEN 'руководитель' THEN 1
            WHEN 'менеджер' THEN 2
            ELSE 3 
        END,
        u.email
`, projectID)
	if err != nil {
		http.Error(w, "Failed to fetch participants", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	f.SetCellValue(sheet2, "A1", "Email")
	f.SetCellValue(sheet2, "B1", "Имя")
	f.SetCellValue(sheet2, "C1", "Роль")
	f.SetCellValue(sheet2, "D1", "Задач выполнено")
	f.SetCellValue(sheet2, "E1", "Задач не выполнено")

	row := 2
	for rows.Next() {
		var email, name, role string
		var tasksCompleted, tasksNotCompleted int
		err := rows.Scan(&email, &name, &role, &tasksCompleted, &tasksNotCompleted)
		if err != nil {
			continue
		}
		f.SetCellValue(sheet2, fmt.Sprintf("A%d", row), email)
		f.SetCellValue(sheet2, fmt.Sprintf("B%d", row), name)
		f.SetCellValue(sheet2, fmt.Sprintf("C%d", row), role)
		f.SetCellValue(sheet2, fmt.Sprintf("D%d", row), tasksCompleted)
		f.SetCellValue(sheet2, fmt.Sprintf("E%d", row), tasksNotCompleted)
		row++
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=admin_project_report_%s.xlsx", projectID))
	if err := f.Write(w); err != nil {
		log.Printf("Error writing file: %v", err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}
}

func generateAllProjectsExcelReportHandler(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	sheet1 := "Общая статистика"
	f.NewSheet(sheet1)
	f.DeleteSheet("Sheet1")

	var totalProjects, totalTasks, completedTasks, notCompletedTasks, totalUsers int
	err := db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&totalProjects)
	if err != nil {
		http.Error(w, "Failed to fetch statistics", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM project_tasks").Scan(&totalTasks)
	if err != nil {
		http.Error(w, "Failed to fetch statistics", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM project_tasks WHERE status = 'завершена'").Scan(&completedTasks)
	if err != nil {
		http.Error(w, "Failed to fetch statistics", http.StatusInternalServerError)
		return
	}

	notCompletedTasks = totalTasks - completedTasks

	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		http.Error(w, "Failed to fetch statistics", http.StatusInternalServerError)
		return
	}

	f.SetCellValue(sheet1, "A1", "Всего проектов")
	f.SetCellValue(sheet1, "B1", totalProjects)
	f.SetCellValue(sheet1, "A2", "Всего задач")
	f.SetCellValue(sheet1, "B2", totalTasks)
	f.SetCellValue(sheet1, "A3", "Задач выполнено")
	f.SetCellValue(sheet1, "B3", completedTasks)
	f.SetCellValue(sheet1, "A4", "Задач не выполнено")
	f.SetCellValue(sheet1, "B4", notCompletedTasks)
	f.SetCellValue(sheet1, "A5", "Всего пользователей")
	f.SetCellValue(sheet1, "B5", totalUsers)

	sheet2 := "Проекты"
	f.NewSheet(sheet2)

	rows, err := db.Query(`
		SELECT 
			p.id,
			p.title,
			COALESCE(participants.cnt, 0) as participants_count,
			COALESCE(tasks_completed.cnt, 0) as tasks_completed,
			COALESCE(tasks_not_completed.cnt, 0) as tasks_not_completed
		FROM projects p
		LEFT JOIN (
			SELECT project_id, COUNT(*) as cnt
			FROM project_participations 
			GROUP BY project_id
		) participants ON p.id = participants.project_id
		LEFT JOIN (
			SELECT project_id, COUNT(*) as cnt
			FROM project_tasks 
			WHERE status = 'завершена'
			GROUP BY project_id
		) tasks_completed ON p.id = tasks_completed.project_id
		LEFT JOIN (
			SELECT project_id, COUNT(*) as cnt
			FROM project_tasks 
			WHERE status IS NOT NULL AND status != 'завершена'
			GROUP BY project_id
		) tasks_not_completed ON p.id = tasks_not_completed.project_id
		ORDER BY p.creation_date DESC
	`)
	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	f.SetCellValue(sheet2, "A1", "ID проекта")
	f.SetCellValue(sheet2, "B1", "Название")
	f.SetCellValue(sheet2, "C1", "Участников")
	f.SetCellValue(sheet2, "D1", "Задач выполнено")
	f.SetCellValue(sheet2, "E1", "Задач не выполнено")

	row := 2
	for rows.Next() {
		var projectID, title string
		var participantsCount, tasksCompleted, tasksNotCompleted int
		err := rows.Scan(&projectID, &title, &participantsCount, &tasksCompleted, &tasksNotCompleted)
		if err != nil {
			continue
		}
		f.SetCellValue(sheet2, fmt.Sprintf("A%d", row), projectID)
		f.SetCellValue(sheet2, fmt.Sprintf("B%d", row), title)
		f.SetCellValue(sheet2, fmt.Sprintf("C%d", row), participantsCount)
		f.SetCellValue(sheet2, fmt.Sprintf("D%d", row), tasksCompleted)
		f.SetCellValue(sheet2, fmt.Sprintf("E%d", row), tasksNotCompleted)
		row++
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=all_projects_report_%s.xlsx", time.Now().Format("20060102")))
	if err := f.Write(w); err != nil {
		log.Printf("Error writing file: %v", err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}
}
