package main

import (
	"database/sql"

	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

const userActivityQuery = `
	SELECT u.id, u.first_name, u.last_name, u.email, u.registration_date,
		COUNT(DISTINCT pp.project_id) AS projects_count,
		COUNT(DISTINCT CASE WHEN pt.status = 'завершена' THEN pt.id END) AS tasks_completed,
		COUNT(DISTINCT cm.id) AS messages_sent,
		COUNT(DISTINCT tc.id) AS comments_count
	FROM users u
	LEFT JOIN project_participations pp ON u.id = pp.user_id
	LEFT JOIN project_tasks pt ON u.id = pt.assigned_to
	LEFT JOIN chat_messages cm ON u.id = cm.user_id
	LEFT JOIN task_comments tc ON u.id = tc.user_id`

func (r *ReportRepository) ListUserActivity() ([]UserActivityReport, error) {
	out := []UserActivityReport{}
	err := r.db.Raw(userActivityQuery + `
		GROUP BY u.id, u.first_name, u.last_name, u.email, u.registration_date
		ORDER BY u.registration_date DESC`).Scan(&out).Error
	return out, err
}

func (r *ReportRepository) GetUserActivity(userID string) (UserActivityReport, error) {
	var rep UserActivityReport
	err := r.db.Raw(userActivityQuery+` WHERE u.id = ?
		GROUP BY u.id, u.first_name, u.last_name, u.email, u.registration_date`, userID).Scan(&rep).Error
	if err != nil {
		return rep, err
	}
	if rep.UserID == "" {
		return rep, sql.ErrNoRows
	}
	return rep, nil
}

const projectEfficiencyQuery = `
	SELECT p.id AS project_id, p.title, p.status, p.creation_date, p.completion_date,
		COUNT(DISTINCT pp.user_id) AS participants_count,
		COUNT(DISTINCT pt.id) AS tasks_total,
		COUNT(DISTINCT CASE WHEN pt.status = 'завершена' THEN pt.id END) AS tasks_completed,
		CASE WHEN COUNT(DISTINCT pt.id) > 0
			THEN (COUNT(DISTINCT CASE WHEN pt.status = 'завершена' THEN pt.id END)::float / COUNT(DISTINCT pt.id)::float) * 100
			ELSE 0 END AS completion_rate
	FROM projects p
	LEFT JOIN project_participations pp ON p.id = pp.project_id
	LEFT JOIN project_tasks pt ON p.id = pt.project_id`

func (r *ReportRepository) ListProjectEfficiency() ([]ProjectEfficiencyReport, error) {
	out := []ProjectEfficiencyReport{}
	err := r.db.Raw(projectEfficiencyQuery + `
		GROUP BY p.id, p.title, p.status, p.creation_date, p.completion_date
		ORDER BY p.creation_date DESC`).Scan(&out).Error
	return out, err
}

func (r *ReportRepository) GetProjectEfficiency(projectID string) (ProjectEfficiencyReport, error) {
	var rep ProjectEfficiencyReport
	err := r.db.Raw(projectEfficiencyQuery+` WHERE p.id = ?
		GROUP BY p.id, p.title, p.status, p.creation_date, p.completion_date`, projectID).Scan(&rep).Error
	if err != nil {
		return rep, err
	}
	if rep.ProjectID == "" {
		return rep, sql.ErrNoRows
	}
	return rep, nil
}

func (r *ReportRepository) GetProjectInfo(projectID string) (title, organizerID string, err error) {
	var row struct {
		Title       string `gorm:"column:title"`
		OrganizerID string `gorm:"column:organizer_id"`
	}
	err = r.db.Raw(`SELECT title, organizer_id FROM projects WHERE id = ?`, projectID).Scan(&row).Error
	if err != nil {
		return "", "", err
	}
	if row.Title == "" && row.OrganizerID == "" {
		return "", "", sql.ErrNoRows
	}
	return row.Title, row.OrganizerID, nil
}

func (r *ReportRepository) GetParticipantRole(projectID, userID string) (string, error) {
	var role string
	err := r.db.Raw(
		`SELECT role FROM project_participations WHERE project_id = ? AND user_id = ?`,
		projectID, userID,
	).Scan(&role).Error
	if err != nil {
		return "", err
	}
	if role == "" {
		return "", sql.ErrNoRows
	}
	return role, nil
}

func (r *ReportRepository) GetSummary() (Summary, error) {
	var s Summary
	queries := []struct {
		q   string
		dst interface{}
	}{
		{"SELECT COUNT(*) FROM users", &s.TotalUsers},
		{"SELECT COUNT(*) FROM users WHERE status = true", &s.ActiveUsers},
		{"SELECT COUNT(*) FROM projects", &s.TotalProjects},
		{"SELECT COUNT(*) FROM projects WHERE status = 'активен'", &s.ActiveProjects},
		{"SELECT COUNT(*) FROM projects WHERE status = 'завершён'", &s.CompletedProjects},
		{"SELECT COUNT(*) FROM project_tasks", &s.TotalTasks},
		{"SELECT COUNT(*) FROM project_tasks WHERE status = 'завершена'", &s.CompletedTasks},
		{`SELECT CASE WHEN COUNT(*) > 0
			THEN (COUNT(CASE WHEN status = 'завершена' THEN 1 END)::float / COUNT(*)::float) * 100
			ELSE 0 END FROM project_tasks`, &s.AverageCompletionRate},
	}
	for _, q := range queries {
		if err := r.db.Raw(q.q).Scan(q.dst).Error; err != nil {
			return s, err
		}
	}
	return s, nil
}

