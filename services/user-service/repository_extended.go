package main

import (
	"database/sql"

	"gorm.io/gorm"
)

func (r *UserRepository) LogActivity(userID string, action string, projectID, taskID *string) {
	_ = r.db.Exec(
		`INSERT INTO activity_log (user_id, project_id, task_id, action) VALUES (?, ?, ?, ?)`,
		userID, projectID, taskID, action,
	).Error
}

type UserWithRoles struct {
	User
	IsParticipant bool `json:"is_participant"`
	IsOrganizer   bool `json:"is_organizer"`
	IsAdmin       bool `json:"is_admin"`
}

func (r *UserRepository) ListUsersWithRoles() ([]UserWithRoles, error) {
	users := []UserWithRoles{}
	err := r.db.Raw(`
		SELECT u.id, u.first_name, u.last_name, u.email, u.registration_date, u.status,
			u.user_type, u.group_id, g.name AS group_name,
			COALESCE(ur.is_participant, true) AS is_participant,
			COALESCE(ur.is_organizer, false) AS is_organizer,
			COALESCE(ur.is_admin, false) AS is_admin
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN groups g ON g.id = u.group_id
		ORDER BY u.registration_date DESC`).Scan(&users).Error
	return users, err
}

func (r *UserRepository) AdminCreateUser(req AdminCreateUserRequest, hashedPassword string) (string, error) {
	var id string
	userType := req.UserType
	if userType != "student" && userType != "teacher" && userType != "staff" {
		userType = "student"
	}
	var groupArg interface{}
	if req.GroupID != nil && *req.GroupID != "" {
		groupArg = *req.GroupID
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(
			`INSERT INTO users (first_name, last_name, email, password, registration_date, status, user_type, group_id)
			 VALUES (?, ?, ?, ?, CURRENT_DATE, true, ?, ?) RETURNING id`,
			req.FirstName, req.LastName, req.Email, hashedPassword, userType, groupArg,
		).Scan(&id).Error; err != nil {
			return err
		}
		return tx.Exec(
			`INSERT INTO user_roles (user_id, is_participant, is_organizer, is_admin) VALUES (?, ?, ?, ?)`,
			id, req.IsParticipant, req.IsOrganizer, req.IsAdmin,
		).Error
	})
	return id, err
}

func (r *UserRepository) AdminUpdateUser(id string, req AdminUpdateUserRequest, hashedPassword *string) error {
	userType := req.UserType
	if userType != "student" && userType != "teacher" && userType != "staff" {
		userType = "student"
	}
	var groupArg interface{}
	if req.GroupID != nil && *req.GroupID != "" {
		groupArg = *req.GroupID
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var err error
		if hashedPassword != nil {
			err = tx.Exec(
				`UPDATE users SET first_name = ?, last_name = ?, email = ?, password = ?, status = ?, user_type = ?, group_id = ? WHERE id = ?`,
				req.FirstName, req.LastName, req.Email, *hashedPassword, req.Status, userType, groupArg, id,
			).Error
		} else {
			err = tx.Exec(
				`UPDATE users SET first_name = ?, last_name = ?, email = ?, status = ?, user_type = ?, group_id = ? WHERE id = ?`,
				req.FirstName, req.LastName, req.Email, req.Status, userType, groupArg, id,
			).Error
		}
		if err != nil {
			return err
		}
		return tx.Exec(
			`INSERT INTO user_roles (user_id, is_participant, is_organizer, is_admin)
			 VALUES (?, ?, ?, ?)
			 ON CONFLICT (user_id) DO UPDATE SET
				is_participant = EXCLUDED.is_participant,
				is_organizer = EXCLUDED.is_organizer,
				is_admin = EXCLUDED.is_admin`,
			id, req.IsParticipant, req.IsOrganizer, req.IsAdmin,
		).Error
	})
}

func (r *UserRepository) AdminDeleteUser(id string) error {
	return r.db.Exec(`DELETE FROM users WHERE id = ?`, id).Error
}

func (r *UserRepository) UserExistsByEmail(email string) (bool, error) {
	var exists bool
	err := r.db.Raw(`SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`, email).Scan(&exists).Error
	return exists, err
}

type AdminDashboard struct {
	TotalUsers        int     `json:"total_users"`
	ActiveUsers       int     `json:"active_users"`
	BlockedUsers      int     `json:"blocked_users"`
	Organizers        int     `json:"organizers"`
	TotalProjects     int     `json:"total_projects"`
	ActiveProjects    int     `json:"active_projects"`
	ArchivedProjects  int     `json:"archived_projects"`
	CompletedProjects int     `json:"completed_projects"`
	TotalTasks        int     `json:"total_tasks"`
	CompletedTasks    int     `json:"completed_tasks"`
	OverdueTasks      int     `json:"overdue_tasks"`
	PendingRequests   int     `json:"pending_requests"`
	NewUsersWeek      int     `json:"new_users_week"`
	TasksDoneWeek     int     `json:"tasks_done_week"`
	AvgQuality        float64 `json:"avg_quality"`
}

func (r *UserRepository) Dashboard() (AdminDashboard, error) {
	var d AdminDashboard
	queries := []struct {
		q   string
		dst interface{}
	}{
		{"SELECT COUNT(*) FROM users", &d.TotalUsers},
		{"SELECT COUNT(*) FROM users WHERE status = true", &d.ActiveUsers},
		{"SELECT COUNT(*) FROM users WHERE status = false", &d.BlockedUsers},
		{"SELECT COUNT(*) FROM user_roles WHERE is_organizer = true", &d.Organizers},
		{"SELECT COUNT(*) FROM projects", &d.TotalProjects},
		{"SELECT COUNT(*) FROM projects WHERE status = 'активен'", &d.ActiveProjects},
		{"SELECT COUNT(*) FROM projects WHERE status = 'архивирован'", &d.ArchivedProjects},
		{"SELECT COUNT(*) FROM projects WHERE status = 'завершён'", &d.CompletedProjects},
		{"SELECT COUNT(*) FROM project_tasks", &d.TotalTasks},
		{"SELECT COUNT(*) FROM project_tasks WHERE status = 'завершена'", &d.CompletedTasks},
		{"SELECT COUNT(*) FROM project_tasks WHERE status != 'завершена' AND due_date IS NOT NULL AND due_date < CURRENT_DATE", &d.OverdueTasks},
		{"SELECT COUNT(*) FROM organizer_requests WHERE status = 'в рассмотрении'", &d.PendingRequests},
		{"SELECT COUNT(*) FROM users WHERE registration_date >= CURRENT_DATE - INTERVAL '7 days'", &d.NewUsersWeek},
		{"SELECT COUNT(*) FROM project_tasks WHERE status = 'завершена' AND completion_date >= CURRENT_DATE - INTERVAL '7 days'", &d.TasksDoneWeek},
		{"SELECT COALESCE(AVG(quality_rating), 0) FROM project_tasks WHERE quality_rating IS NOT NULL", &d.AvgQuality},
	}
	for _, qq := range queries {
		if err := r.db.Raw(qq.q).Scan(qq.dst).Error; err != nil && err != sql.ErrNoRows {
			return d, err
		}
	}
	return d, nil
}

func (r *UserRepository) ListSkills(userID string) ([]string, error) {
	skills := []string{}
	err := r.db.Raw(
		`SELECT tc.name FROM user_skills us
		 JOIN tag_catalog tc ON tc.id = us.tag_id
		 WHERE us.user_id = ? ORDER BY tc.name`, userID).Scan(&skills).Error
	return skills, err
}

func (r *UserRepository) AddSkill(userID, name string) error {
	var tagID string
	if err := r.db.Raw(`SELECT id FROM tag_catalog WHERE lower(name) = lower(?)`, name).Row().Scan(&tagID); err != nil {
		return err
	}
	return r.db.Exec(
		`INSERT INTO user_skills (user_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		userID, tagID).Error
}

func (r *UserRepository) DeleteSkill(userID, name string) error {
	return r.db.Exec(
		`DELETE FROM user_skills
		 WHERE user_id = ? AND tag_id = (SELECT id FROM tag_catalog WHERE lower(name) = lower(?))`,
		userID, name).Error
}

type InterestCategory struct {
	CategoryID string `json:"category_id" gorm:"column:category_id"`
	Name       string `json:"name" gorm:"column:name"`
}

func (r *UserRepository) ListInterests(userID string) ([]InterestCategory, error) {
	cats := []InterestCategory{}
	err := r.db.Raw(`
		SELECT pc.id AS category_id, pc.name AS name
		FROM user_interests ui
		JOIN project_categories pc ON pc.id = ui.category_id
		WHERE ui.user_id = ?
		ORDER BY pc.name`, userID).Scan(&cats).Error
	return cats, err
}

func (r *UserRepository) AddInterest(userID, categoryID string) error {
	return r.db.Exec(
		`INSERT INTO user_interests (user_id, category_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		userID, categoryID).Error
}

func (r *UserRepository) DeleteInterest(userID, categoryID string) error {
	return r.db.Exec(`DELETE FROM user_interests WHERE user_id = ? AND category_id = ?`, userID, categoryID).Error
}
