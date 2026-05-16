package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

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

func nilIfNotValid(t sql.NullTime) interface{} {
	if !t.Valid {
		return nil
	}
	return t.Time
}

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func wrapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sql.ErrNoRows
	}
	return err
}

func (r *ProjectRepository) CreateProject(organizerID string, req CreateProjectRequest) (string, error) {
	end, err := parseDate(req.PlannedEndDate)
	if err != nil {
		return "", err
	}
	var imgArg interface{} = nil
	if req.ImageURL != nil && *req.ImageURL != "" {
		imgArg = *req.ImageURL
	}

	var id string
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(
			`INSERT INTO projects (organizer_id, title, short_description, full_description,
				goal_description, planned_end_date, image_url, status, creation_date)
			 VALUES (?, ?, ?, ?, ?, ?, ?, 'активен', CURRENT_DATE) RETURNING id`,
			organizerID, req.Title, req.ShortDescription, req.FullDescription,
			req.GoalDescription, nilIfNotValid(end), imgArg,
		).Scan(&id).Error; err != nil {
			return err
		}
		return tx.Exec(
			`INSERT INTO project_participations (project_id, user_id, role, join_date)
			 VALUES (?, ?, 'руководитель', CURRENT_DATE) ON CONFLICT (project_id, user_id) DO NOTHING`,
			id, organizerID,
		).Error
	})
	if err != nil {
		return "", err
	}
	cats := req.CategoryIDs
	if len(cats) == 0 && req.CategoryID != nil && *req.CategoryID != "" {
		cats = []string{*req.CategoryID}
	}
	if len(cats) > 0 {
		_ = r.ReplaceProjectCategories(id, cats)
	}
	return id, nil
}

func (r *ProjectRepository) ListProjects(status, categoryID string) ([]ProjectExtended, error) {
	q := `
		SELECT p.id, p.organizer_id, p.title, p.short_description, p.full_description,
			p.goal_description, p.status, p.creation_date, p.planned_end_date, p.completion_date,
			p.image_url,
			COALESCE(u.first_name || ' ' || u.last_name, 'Неизвестно') AS organizer_name,
			COALESCE(COUNT(DISTINCT pp.user_id), 0) AS participants_count,
			COALESCE((
				SELECT string_agg(c.name, ', ' ORDER BY c.name)
				FROM project_category_links pcl
				JOIN project_categories c ON c.id = pcl.category_id
				WHERE pcl.project_id = p.id
			), '') AS category_name
		FROM projects p
		LEFT JOIN users u ON p.organizer_id = u.id
		LEFT JOIN project_participations pp ON p.id = pp.project_id
		WHERE 1=1`
	args := []interface{}{}
	pos := 1
	if status != "" {
		q += fmt.Sprintf(" AND p.status = $%d", pos)
		args = append(args, status)
		pos++
	}
	if categoryID != "" {
		q += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM project_category_links pcl WHERE pcl.project_id = p.id AND pcl.category_id = $%d)", pos)
		args = append(args, categoryID)
	}
	q += " GROUP BY p.id, u.first_name, u.last_name ORDER BY p.creation_date DESC"

	rows, err := r.db.Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []ProjectExtended{}
	for rows.Next() {
		var p ProjectExtended
		if err := rows.Scan(&p.ID, &p.OrganizerID, &p.Title, &p.ShortDescription,
			&p.FullDescription, &p.GoalDescription, &p.Status, &p.CreationDate, &p.PlannedEndDate, &p.CompletionDate,
			&p.ImageURL, &p.OrganizerName, &p.ParticipantsCount, &p.CategoryName); err != nil {
			continue
		}
		p.Tags, _ = r.GetProjectTags(p.ID)
		p.CategoryIDs, _ = r.ListProjectCategoryIDs(p.ID)
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *ProjectRepository) ListMyProjects(userID string) ([]ProjectWithRole, error) {
	rows, err := r.db.Raw(`
		SELECT DISTINCT p.id, p.organizer_id, p.title, p.short_description, p.full_description,
			p.goal_description, p.status, p.creation_date, p.planned_end_date, p.completion_date,
			p.image_url,
			CASE
				WHEN p.organizer_id = ? THEN 'organizer'
				WHEN pp.role = 'руководитель' THEN 'leader'
				ELSE 'participant'
			END
		FROM projects p
		LEFT JOIN project_participations pp ON p.id = pp.project_id AND pp.user_id = ?
		WHERE p.organizer_id = ? OR pp.user_id = ?
		ORDER BY p.creation_date DESC
	`, userID, userID, userID, userID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []ProjectWithRole{}
	for rows.Next() {
		var p ProjectWithRole
		if err := rows.Scan(&p.ID, &p.OrganizerID, &p.Title, &p.ShortDescription,
			&p.FullDescription, &p.GoalDescription, &p.Status, &p.CreationDate, &p.PlannedEndDate, &p.CompletionDate,
			&p.ImageURL, &p.UserRole); err != nil {
			continue
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *ProjectRepository) GetProject(id string) (Project, error) {
	var p Project
	row := r.db.Raw(
		`SELECT p.id, p.organizer_id, p.title, p.short_description, p.full_description,
			p.goal_description, p.status, p.creation_date, p.planned_end_date, p.completion_date,
			p.image_url,
			COALESCE(u.first_name || ' ' || u.last_name, '')
		 FROM projects p
		 LEFT JOIN users u ON u.id = p.organizer_id
		 WHERE p.id = ?`, id).Row()
	err := row.Scan(&p.ID, &p.OrganizerID, &p.Title, &p.ShortDescription, &p.FullDescription,
		&p.GoalDescription, &p.Status, &p.CreationDate, &p.PlannedEndDate, &p.CompletionDate,
		&p.ImageURL, &p.OrganizerName)
	return p, err
}

func (r *ProjectRepository) GetOrganizerID(projectID string) (string, error) {
	var id string
	err := r.db.Raw(`SELECT organizer_id FROM projects WHERE id = ?`, projectID).Row().Scan(&id)
	return id, err
}

func (r *ProjectRepository) GetParticipantRole(projectID, userID string) (string, error) {
	var role string
	err := r.db.Raw(
		`SELECT role FROM project_participations WHERE project_id = ? AND user_id = ?`,
		projectID, userID,
	).Row().Scan(&role)
	return role, err
}

func (r *ProjectRepository) UpdateProject(id string, req CreateProjectRequest) error {
	end, err := parseDate(req.PlannedEndDate)
	if err != nil {
		return err
	}
	var imgArg interface{} = nil
	if req.ImageURL != nil && *req.ImageURL != "" {
		imgArg = *req.ImageURL
	}
	if err := r.db.Exec(
		`UPDATE projects SET title = ?, short_description = ?, full_description = ?,
			goal_description = ?, planned_end_date = ?, image_url = ? WHERE id = ?`,
		req.Title, req.ShortDescription, req.FullDescription,
		req.GoalDescription, nilIfNotValid(end), imgArg, id,
	).Error; err != nil {
		return err
	}
	cats := req.CategoryIDs
	if len(cats) == 0 && req.CategoryID != nil && *req.CategoryID != "" {
		cats = []string{*req.CategoryID}
	}
	return r.ReplaceProjectCategories(id, cats)
}

func (r *ProjectRepository) ArchiveProject(id string) error {
	return r.db.Exec(
		`UPDATE projects SET status = 'архивирован', completion_date = CURRENT_DATE WHERE id = ?`, id,
	).Error
}

func (r *ProjectRepository) ListCategories() ([]ProjectCategory, error) {
	categories := []ProjectCategory{}
	err := r.db.Raw(`SELECT id, name FROM project_categories ORDER BY name`).Scan(&categories).Error
	return categories, err
}

func (r *ProjectRepository) FindParticipation(projectID, userID string) (string, error) {
	var id string
	err := r.db.Raw(
		`SELECT id FROM project_participations WHERE project_id = ? AND user_id = ?`,
		projectID, userID,
	).Row().Scan(&id)
	return id, err
}

func (r *ProjectRepository) FindActiveRequest(projectID, userID string) (string, error) {
	var id string
	err := r.db.Raw(
		`SELECT id FROM project_participation_requests WHERE project_id = ? AND user_id = ? AND status = 'в рассмотрении' LIMIT 1`,
		projectID, userID,
	).Row().Scan(&id)
	return id, err
}

func (r *ProjectRepository) CreateParticipationRequest(projectID, userID, comment, resumeURL string) (string, error) {
	var id string
	err := r.db.Raw(
		`INSERT INTO project_participation_requests (project_id, user_id, comment, resume_url, submission_date, status)
		 VALUES (?, ?, ?, ?, CURRENT_DATE, 'в рассмотрении') RETURNING id`,
		projectID, userID, comment, resumeURL,
	).Scan(&id).Error
	return id, err
}

func (r *ProjectRepository) ListProjectRequests(projectID string) ([]ParticipationRequestExtended, error) {
	rows, err := r.db.Raw(`
		SELECT pr.id, pr.project_id, pr.user_id, pr.comment, pr.resume_url, pr.submission_date, pr.status,
			COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), COALESCE(u.email, '')
		FROM project_participation_requests pr
		LEFT JOIN users u ON u.id = pr.user_id
		WHERE pr.project_id = ?
		ORDER BY pr.submission_date DESC`,
		projectID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []ParticipationRequestExtended{}
	for rows.Next() {
		var rq ParticipationRequestExtended
		if err := rows.Scan(&rq.ID, &rq.ProjectID, &rq.UserID, &rq.Comment, &rq.ResumeURL,
			&rq.SubmissionDate, &rq.Status, &rq.FirstName, &rq.LastName, &rq.Email); err != nil {
			continue
		}
		skills := []string{}
		_ = r.db.Raw(
			`SELECT tc.name FROM user_skills us
			 JOIN tag_catalog tc ON tc.id = us.tag_id
			 WHERE us.user_id = ? ORDER BY tc.name`, rq.UserID).Scan(&skills).Error
		rq.Skills = skills
		requests = append(requests, rq)
	}
	return requests, nil
}

func (r *ProjectRepository) GetMyParticipationRequest(projectID, userID string) (ParticipationRequest, error) {
	var req ParticipationRequest
	row := r.db.Raw(
		`SELECT id, project_id, user_id, comment, resume_url, submission_date, status
		 FROM project_participation_requests WHERE project_id = ? AND user_id = ? ORDER BY submission_date DESC LIMIT 1`,
		projectID, userID,
	).Row()
	err := row.Scan(&req.ID, &req.ProjectID, &req.UserID, &req.Comment, &req.ResumeURL, &req.SubmissionDate, &req.Status)
	return req, err
}

func (r *ProjectRepository) GetRequestProjectAndUser(requestID string) (string, string, error) {
	var pid, uid string
	row := r.db.Raw(
		`SELECT project_id, user_id FROM project_participation_requests WHERE id = ?`, requestID,
	).Row()
	err := row.Scan(&pid, &uid)
	return pid, uid, err
}

func (r *ProjectRepository) ApproveRequest(requestID, projectID, userID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`UPDATE project_participation_requests SET status = 'одобрена' WHERE id = ?`, requestID).Error; err != nil {
			return err
		}
		return tx.Exec(
			`INSERT INTO project_participations (project_id, user_id, role, join_date)
			 VALUES (?, ?, 'участник', CURRENT_DATE) ON CONFLICT DO NOTHING`,
			projectID, userID,
		).Error
	})
}

func (r *ProjectRepository) RejectRequest(requestID string) error {
	return r.db.Exec(`UPDATE project_participation_requests SET status = 'отклонена' WHERE id = ?`, requestID).Error
}

func (r *ProjectRepository) ListParticipants(projectID string) ([]ParticipantWithUser, error) {
	rows, err := r.db.Raw(`
		SELECT pp.id, pp.project_id, pp.user_id, pp.role, pp.join_date,
			u.email, u.first_name, u.last_name
		FROM project_participations pp
		LEFT JOIN users u ON pp.user_id = u.id
		WHERE pp.project_id = ?
		ORDER BY pp.join_date`,
		projectID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := []ParticipantWithUser{}
	for rows.Next() {
		var p ParticipantWithUser
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.UserID, &p.Role, &p.JoinDate,
			&p.Email, &p.FirstName, &p.LastName); err == nil {
			skills := []string{}
			_ = r.db.Raw(
				`SELECT tc.name FROM user_skills us
				 JOIN tag_catalog tc ON tc.id = us.tag_id
				 WHERE us.user_id = ? ORDER BY tc.name`, p.UserID).Scan(&skills).Error
			p.Skills = skills
			participants = append(participants, p)
		}
	}
	return participants, nil
}

func (r *ProjectRepository) UpdateParticipantRole(projectID, userID, role string) error {
	return r.db.Exec(
		`UPDATE project_participations SET role = ? WHERE project_id = ? AND user_id = ?`,
		role, projectID, userID,
	).Error
}

func (r *ProjectRepository) AddTag(projectID, name string) error {
	var tagID string
	if err := r.db.Raw(`SELECT id FROM tag_catalog WHERE lower(name) = lower(?)`, name).Row().Scan(&tagID); err != nil {
		return err
	}
	return r.db.Exec(
		`INSERT INTO project_tags (project_id, tag_id, is_required) VALUES (?, ?, FALSE) ON CONFLICT DO NOTHING`,
		projectID, tagID,
	).Error
}

func (r *ProjectRepository) GetProjectTags(projectID string) ([]string, error) {
	tags := []string{}
	err := r.db.Raw(
		`SELECT tc.name FROM project_tags pt
		 JOIN tag_catalog tc ON tc.id = pt.tag_id
		 WHERE pt.project_id = ? AND pt.is_required = FALSE
		 ORDER BY tc.name`, projectID).Scan(&tags).Error
	return tags, err
}

func (r *ProjectRepository) DeleteTag(projectID, name string) error {
	return r.db.Exec(
		`DELETE FROM project_tags
		 WHERE project_id = ? AND is_required = FALSE
		   AND tag_id = (SELECT id FROM tag_catalog WHERE lower(name) = lower(?))`,
		projectID, name).Error
}

var _ = wrapNotFound
