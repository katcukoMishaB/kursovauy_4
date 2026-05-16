package main

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (Task) TableName() string        { return "project_tasks" }
func (TaskComment) TableName() string { return "task_comments" }

const taskColumns = `id, project_id, assigned_to, title, description, status, priority, difficulty, quality_rating, attachment_url, creation_date, due_date, completion_date`

func decorateOverdue(t *Task) {
	if t.DueDate != nil && t.Status != "завершена" && t.DueDate.Before(time.Now().Truncate(24*time.Hour)) {
		t.Overdue = true
	}
}

func parseDate(s *string) (sql.NullTime, error) {
	if s == nil || *s == "" {
		return sql.NullTime{}, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return sql.NullTime{}, err
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}

func (r *TaskRepository) CreateTask(projectID string, req CreateTaskRequest) (string, error) {
	due, err := parseDate(req.DueDate)
	if err != nil {
		return "", err
	}
	priority := req.Priority
	if priority == "" {
		priority = "средний"
	}
	difficulty := req.Difficulty
	if difficulty == 0 {
		difficulty = 3
	}
	var dueArg interface{} = nil
	if due.Valid {
		dueArg = due.Time
	}
	var attArg interface{} = nil
	if req.AttachmentURL != nil && *req.AttachmentURL != "" {
		attArg = *req.AttachmentURL
	}
	var id string
	err = r.db.Raw(
		`INSERT INTO project_tasks (project_id, title, description, status, priority, difficulty, due_date, attachment_url, creation_date)
		 VALUES (?, ?, ?, 'новая', ?, ?, ?, ?, CURRENT_DATE) RETURNING id`,
		projectID, req.Title, req.Description, priority, difficulty, dueArg, attArg,
	).Scan(&id).Error
	return id, err
}

func (r *TaskRepository) ListTasks(projectID, status string) ([]Task, error) {
	q := `SELECT ` + taskColumns + ` FROM project_tasks WHERE project_id = ?`
	args := []interface{}{projectID}
	if status != "" {
		q += ` AND status = ?`
		args = append(args, status)
	}
	q += ` ORDER BY
		CASE priority WHEN 'высокий' THEN 1 WHEN 'средний' THEN 2 ELSE 3 END,
		due_date NULLS LAST, creation_date DESC`

	tasks := []Task{}
	if err := r.db.Raw(q, args...).Scan(&tasks).Error; err != nil {
		return nil, err
	}
	for i := range tasks {
		decorateOverdue(&tasks[i])
	}
	return tasks, nil
}

func (r *TaskRepository) GetTask(id string) (Task, error) {
	var t Task
	err := r.db.Raw(`SELECT `+taskColumns+` FROM project_tasks WHERE id = ?`, id).Scan(&t).Error
	if err != nil {
		return t, err
	}
	if t.ID == "" {
		return t, sql.ErrNoRows
	}
	decorateOverdue(&t)
	return t, nil
}

func (r *TaskRepository) UpdateTask(id string, req UpdateTaskRequest) error {
	due, err := parseDate(req.DueDate)
	if err != nil {
		return err
	}
	priority := req.Priority
	if priority == "" {
		priority = "средний"
	}
	difficulty := req.Difficulty
	if difficulty == 0 {
		difficulty = 3
	}
	var dueArg interface{} = nil
	if due.Valid {
		dueArg = due.Time
	}
	var attArg interface{} = nil
	if req.AttachmentURL != nil && *req.AttachmentURL != "" {
		attArg = *req.AttachmentURL
	}
	return r.db.Exec(
		`UPDATE project_tasks SET title = ?, description = ?, priority = ?, difficulty = ?, due_date = ?, attachment_url = ? WHERE id = ?`,
		req.Title, req.Description, priority, difficulty, dueArg, attArg, id,
	).Error
}

func (r *TaskRepository) AssignTask(taskID string, assignedTo *string, status string) error {
	var v interface{} = nil
	if assignedTo != nil {
		v = *assignedTo
	}
	return r.db.Exec(
		`UPDATE project_tasks SET assigned_to = ?, status = ? WHERE id = ?`,
		v, status, taskID,
	).Error
}

func (r *TaskRepository) UpdateStatus(id, status string) error {
	if status == "завершена" {
		return r.db.Exec(
			`UPDATE project_tasks SET status = ?, completion_date = CURRENT_DATE WHERE id = ?`,
			status, id,
		).Error
	}
	return r.db.Exec(
		`UPDATE project_tasks SET status = ?, completion_date = NULL WHERE id = ?`,
		status, id,
	).Error
}

func (r *TaskRepository) RateTask(id string, rating int) error {
	return r.db.Exec(`UPDATE project_tasks SET quality_rating = ? WHERE id = ?`, rating, id).Error
}

func (r *TaskRepository) GetTaskAssignee(id string) (*string, error) {
	var assignedTo sql.NullString
	if err := r.db.Raw(`SELECT assigned_to FROM project_tasks WHERE id = ?`, id).Scan(&assignedTo).Error; err != nil {
		return nil, err
	}
	if !assignedTo.Valid {
		return nil, nil
	}
	v := assignedTo.String
	return &v, nil
}

func (r *TaskRepository) GetTaskProject(id string) (string, error) {
	var pid string
	err := r.db.Raw(`SELECT project_id FROM project_tasks WHERE id = ?`, id).Scan(&pid).Error
	if err != nil {
		return "", err
	}
	if pid == "" {
		return "", sql.ErrNoRows
	}
	return pid, nil
}

func (r *TaskRepository) GetOrganizerID(projectID string) (string, error) {
	var id string
	err := r.db.Raw(`SELECT organizer_id FROM projects WHERE id = ?`, projectID).Scan(&id).Error
	return id, err
}

func (r *TaskRepository) GetParticipantRole(projectID, userID string) (string, error) {
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

func (r *TaskRepository) AddComment(taskID, userID, content string) (string, error) {
	var id string
	err := r.db.Raw(
		`INSERT INTO task_comments (task_id, user_id, content, publication_date)
		 VALUES (?, ?, ?, CURRENT_DATE) RETURNING id`,
		taskID, userID, content,
	).Scan(&id).Error
	return id, err
}

func (r *TaskRepository) ListComments(taskID string) ([]TaskComment, error) {
	comments := []TaskComment{}
	err := r.db.Raw(
		`SELECT id, task_id, user_id, content, publication_date FROM task_comments WHERE task_id = ? ORDER BY publication_date`,
		taskID,
	).Scan(&comments).Error
	return comments, err
}

func (r *TaskRepository) LogActivity(userID, projectID, taskID *string, action string) {
	_ = r.db.Exec(
		`INSERT INTO activity_log (user_id, project_id, task_id, action) VALUES (?, ?, ?, ?)`,
		userID, projectID, taskID, action,
	).Error
}

func (r *TaskRepository) ListTaskExtraAssignees(taskID string) ([]string, error) {
	out := []string{}
	err := r.db.Raw(`SELECT user_id FROM project_task_assignees WHERE task_id = ?`, taskID).Scan(&out).Error
	return out, err
}

func (r *TaskRepository) UpdateAttachment(taskID string, url *string) error {
	return r.db.Exec(`UPDATE project_tasks SET attachment_url = ? WHERE id = ?`, url, taskID).Error
}

var _ = fmt.Sprintf

var _ = (*gorm.DB)(nil)
