package main

import (
	"gorm.io/gorm"
)

func (r *ProjectRepository) ListGoals(projectID string) ([]ProjectGoal, error) {
	rows, err := r.db.Raw(
		`SELECT id, project_id, title, description, is_achieved, creation_date, achieved_date
		 FROM project_goals WHERE project_id = ? ORDER BY creation_date`, projectID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	goals := []ProjectGoal{}
	for rows.Next() {
		var g ProjectGoal
		if err := rows.Scan(&g.ID, &g.ProjectID, &g.Title, &g.Description, &g.IsAchieved, &g.CreationDate, &g.AchievedDate); err == nil {
			goals = append(goals, g)
		}
	}
	return goals, nil
}

func (r *ProjectRepository) CreateGoal(projectID, title string, description *string) (string, error) {
	var id string
	err := r.db.Raw(
		`INSERT INTO project_goals (project_id, title, description) VALUES (?, ?, ?) RETURNING id`,
		projectID, title, description,
	).Row().Scan(&id)
	return id, wrapNotFound(err)
}

func (r *ProjectRepository) ToggleGoal(goalID string, achieved bool) error {
	if achieved {
		return r.db.Exec(
			`UPDATE project_goals SET is_achieved = true, achieved_date = CURRENT_DATE WHERE id = ?`, goalID).Error
	}
	return r.db.Exec(
		`UPDATE project_goals SET is_achieved = false, achieved_date = NULL WHERE id = ?`, goalID).Error
}

func (r *ProjectRepository) DeleteGoal(goalID string) error {
	return r.db.Exec(`DELETE FROM project_goals WHERE id = ?`, goalID).Error
}

func (r *ProjectRepository) GetGoalProject(goalID string) (string, error) {
	var pid string
	err := r.db.Raw(`SELECT project_id FROM project_goals WHERE id = ?`, goalID).Row().Scan(&pid)
	return pid, wrapNotFound(err)
}

func (r *ProjectRepository) ListRequiredSkills(projectID string) ([]string, error) {
	rows, err := r.db.Raw(
		`SELECT tc.name FROM project_tags pt
		 JOIN tag_catalog tc ON tc.id = pt.tag_id
		 WHERE pt.project_id = ? AND pt.is_required = TRUE
		 ORDER BY tc.name`, projectID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err == nil {
			skills = append(skills, s)
		}
	}
	return skills, nil
}

func (r *ProjectRepository) AddRequiredSkill(projectID, name string) error {
	var tagID string
	if err := r.db.Raw(`SELECT id FROM tag_catalog WHERE lower(name) = lower(?)`, name).Row().Scan(&tagID); err != nil {
		return err
	}
	return r.db.Exec(
		`INSERT INTO project_tags (project_id, tag_id, is_required) VALUES (?, ?, TRUE)
		 ON CONFLICT (project_id, tag_id) DO UPDATE SET is_required = TRUE`,
		projectID, tagID).Error
}

func (r *ProjectRepository) DeleteRequiredSkill(projectID, name string) error {
	return r.db.Exec(
		`DELETE FROM project_tags
		 WHERE project_id = ? AND is_required = TRUE
		   AND tag_id = (SELECT id FROM tag_catalog WHERE lower(name) = lower(?))`,
		projectID, name).Error
}

func (r *ProjectRepository) RecommendForUser(userID string) ([]RecommendedProject, error) {
	rows, err := r.db.Raw(`
		WITH
		user_int AS (SELECT category_id FROM user_interests WHERE user_id = ?),
		user_sk  AS (SELECT tag_id FROM user_skills WHERE user_id = ?),
		excluded AS (
			SELECT id FROM projects WHERE organizer_id = ?
			UNION
			SELECT project_id FROM project_participations WHERE user_id = ?
		),
		req_match AS (
			SELECT pt.project_id, COUNT(*) AS c
			FROM project_tags pt
			JOIN user_sk us ON us.tag_id = pt.tag_id
			WHERE pt.is_required = TRUE
			GROUP BY pt.project_id
		),
		tag_match AS (
			SELECT pt.project_id, COUNT(*) AS c
			FROM project_tags pt
			JOIN user_sk us ON us.tag_id = pt.tag_id
			WHERE pt.is_required = FALSE
			GROUP BY pt.project_id
		)
		SELECT
			p.id, p.organizer_id, p.title, p.short_description, p.full_description,
			p.goal_description, p.status, p.creation_date, p.planned_end_date, p.completion_date,
			COALESCE(u.first_name || ' ' || u.last_name, 'Неизвестно'),
			COALESCE((SELECT COUNT(*) FROM project_participations WHERE project_id = p.id), 0),
			COALESCE((
				SELECT string_agg(c.name, ', ' ORDER BY c.name)
				FROM project_category_links pcl
				JOIN project_categories c ON c.id = pcl.category_id
				WHERE pcl.project_id = p.id
			), ''),
			(CASE WHEN EXISTS (
					SELECT 1 FROM project_category_links pcl
					WHERE pcl.project_id = p.id AND pcl.category_id IN (SELECT category_id FROM user_int)
				) THEN 3 ELSE 0 END
				+ COALESCE(2 * (SELECT c FROM req_match WHERE project_id = p.id), 0)
				+ COALESCE((SELECT c FROM tag_match WHERE project_id = p.id), 0)) AS score,
			EXISTS (
				SELECT 1 FROM project_category_links pcl
				WHERE pcl.project_id = p.id AND pcl.category_id IN (SELECT category_id FROM user_int)
			) AS matched_category,
			COALESCE((SELECT c FROM req_match WHERE project_id = p.id), 0) AS matched_req,
			COALESCE((SELECT c FROM tag_match WHERE project_id = p.id), 0) AS matched_tag
		FROM projects p
		LEFT JOIN users u ON u.id = p.organizer_id
		WHERE p.status = 'активен' AND p.id NOT IN (SELECT id FROM excluded)
		ORDER BY score DESC, p.creation_date DESC
		LIMIT 12`, userID, userID, userID, userID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []RecommendedProject{}
	for rows.Next() {
		var rp RecommendedProject
		var matchedCat bool
		var matchedReq, matchedTag int
		if err := rows.Scan(
			&rp.ID, &rp.OrganizerID, &rp.Title, &rp.ShortDescription, &rp.FullDescription,
			&rp.GoalDescription, &rp.Status, &rp.CreationDate, &rp.PlannedEndDate, &rp.CompletionDate,
			&rp.OrganizerName, &rp.ParticipantsCount, &rp.CategoryName,
			&rp.Score, &matchedCat, &matchedReq, &matchedTag,
		); err != nil {
			continue
		}
		rp.Tags, _ = r.GetProjectTags(rp.ID)
		rp.CategoryIDs, _ = r.ListProjectCategoryIDs(rp.ID)
		rp.MatchedReasons = []string{}
		if matchedCat {
			rp.MatchedReasons = append(rp.MatchedReasons, "интересующая категория")
		}
		if matchedReq > 0 {
			rp.MatchedReasons = append(rp.MatchedReasons, "ваши навыки нужны проекту")
		}
		if matchedTag > 0 {
			rp.MatchedReasons = append(rp.MatchedReasons, "совпадение по тегам")
		}
		if rp.Score > 0 {
			results = append(results, rp)
		}
	}
	return results, nil
}

func (r *ProjectRepository) HasInterestsOrSkills(userID string) (bool, error) {
	var n int
	err := r.db.Raw(`
		SELECT (SELECT COUNT(*) FROM user_interests WHERE user_id = ?)
		     + (SELECT COUNT(*) FROM user_skills WHERE user_id = ?)`, userID, userID).Row().Scan(&n)
	return n > 0, wrapNotFound(err)
}

func (r *ProjectRepository) ListActiveFresh(userID string, limit int) ([]RecommendedProject, error) {
	rows, err := r.db.Raw(`
		SELECT
			p.id, p.organizer_id, p.title, p.short_description, p.full_description,
			p.goal_description, p.status, p.creation_date, p.planned_end_date, p.completion_date,
			COALESCE(u.first_name || ' ' || u.last_name, 'Неизвестно'),
			COALESCE((SELECT COUNT(*) FROM project_participations WHERE project_id = p.id), 0),
			COALESCE((
				SELECT string_agg(c.name, ', ' ORDER BY c.name)
				FROM project_category_links pcl
				JOIN project_categories c ON c.id = pcl.category_id
				WHERE pcl.project_id = p.id
			), '')
		FROM projects p
		LEFT JOIN users u ON u.id = p.organizer_id
		WHERE p.status = 'активен'
			AND p.organizer_id <> ?
			AND p.id NOT IN (SELECT project_id FROM project_participations WHERE user_id = ?)
		ORDER BY p.creation_date DESC
		LIMIT ?`, userID, userID, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []RecommendedProject{}
	for rows.Next() {
		var rp RecommendedProject
		if err := rows.Scan(
			&rp.ID, &rp.OrganizerID, &rp.Title, &rp.ShortDescription, &rp.FullDescription,
			&rp.GoalDescription, &rp.Status, &rp.CreationDate, &rp.PlannedEndDate, &rp.CompletionDate,
			&rp.OrganizerName, &rp.ParticipantsCount, &rp.CategoryName,
		); err != nil {
			continue
		}
		rp.Tags, _ = r.GetProjectTags(rp.ID)
		rp.CategoryIDs, _ = r.ListProjectCategoryIDs(rp.ID)
		rp.MatchedReasons = []string{"свежий проект"}
		results = append(results, rp)
	}
	return results, nil
}

func (r *ProjectRepository) IsAdmin(userID string) bool {
	if userID == "" {
		return false
	}
	var ok bool
	if err := r.db.Raw(`SELECT COALESCE(is_admin, false) FROM user_roles WHERE user_id = ?`, userID).Row().Scan(&ok); err != nil {
		return false
	}
	return ok
}

func (r *ProjectRepository) DeleteProject(id string) error {
	return r.db.Exec(`DELETE FROM projects WHERE id = ?`, id).Error
}

func (r *ProjectRepository) RestoreProject(id string) error {
	return r.db.Exec(
		`UPDATE projects SET status = 'активен', completion_date = NULL WHERE id = ?`, id).Error
}

func (r *ProjectRepository) LogActivity(userID, action string, projectID *string) {
	_ = r.db.Exec(
		`INSERT INTO activity_log (user_id, project_id, action) VALUES (?, ?, ?)`,
		userID, projectID, action,
	).Error
}

func (r *ProjectRepository) AllTasksDone(projectID string) (bool, error) {
	var total, done int
	if err := r.db.Raw(`SELECT COUNT(*) FROM project_tasks WHERE project_id = ?`, projectID).Row().Scan(&total); err != nil {
		return false, wrapNotFound(err)
	}
	if total == 0 {
		return false, nil
	}
	if err := r.db.Raw(`SELECT COUNT(*) FROM project_tasks WHERE project_id = ? AND status = 'завершена'`, projectID).Row().Scan(&done); err != nil {
		return false, wrapNotFound(err)
	}
	return done == total, nil
}

func (r *ProjectRepository) CompleteProject(id string) error {
	return r.db.Exec(
		`UPDATE projects SET status = 'завершён', completion_date = CURRENT_DATE WHERE id = ?`, id).Error
}

func (r *ProjectRepository) ListTagCatalog() ([]CatalogTag, error) {
	rows, err := r.db.Raw(`SELECT id, name FROM tag_catalog ORDER BY name`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []CatalogTag{}
	for rows.Next() {
		var t CatalogTag
		if err := rows.Scan(&t.ID, &t.Name); err == nil {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

func (r *ProjectRepository) CreateTagCatalog(name string) (string, error) {
	var id string
	err := r.db.Raw(`INSERT INTO tag_catalog (name) VALUES (?) RETURNING id`, name).Scan(&id).Error
	return id, err
}

func (r *ProjectRepository) UpdateTagCatalog(id, name string) error {
	return r.db.Exec(`UPDATE tag_catalog SET name = ? WHERE id = ?`, name, id).Error
}

func (r *ProjectRepository) DeleteTagCatalog(id string) error {
	return r.db.Exec(`DELETE FROM tag_catalog WHERE id = ?`, id).Error
}

func (r *ProjectRepository) CreateCategory(name string) (string, error) {
	var id string
	err := r.db.Raw(`INSERT INTO project_categories (name) VALUES (?) RETURNING id`, name).Scan(&id).Error
	return id, err
}

func (r *ProjectRepository) UpdateCategory(id, name string) error {
	return r.db.Exec(`UPDATE project_categories SET name = ? WHERE id = ?`, name, id).Error
}

func (r *ProjectRepository) DeleteCategory(id string) error {
	return r.db.Exec(`DELETE FROM project_categories WHERE id = ?`, id).Error
}

func (r *ProjectRepository) ListProjectCategoryIDs(projectID string) ([]string, error) {
	rows, err := r.db.Raw(`SELECT category_id FROM project_category_links WHERE project_id = ?`, projectID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *ProjectRepository) ReplaceProjectCategories(projectID string, categoryIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM project_category_links WHERE project_id = ?`, projectID).Error; err != nil {
			return err
		}
		for _, cid := range categoryIDs {
			if cid == "" {
				continue
			}
			if err := tx.Exec(
				`INSERT INTO project_category_links (project_id, category_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
				projectID, cid).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProjectRepository) AddTaskAssignee(taskID, userID string) error {
	return r.db.Exec(
		`INSERT INTO project_task_assignees (task_id, user_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		taskID, userID).Error
}

func (r *ProjectRepository) RemoveTaskAssignee(taskID, userID string) error {
	return r.db.Exec(
		`DELETE FROM project_task_assignees WHERE task_id = ? AND user_id = ?`, taskID, userID).Error
}

func (r *ProjectRepository) ListTaskAssignees(taskID string) ([]string, error) {
	rows, err := r.db.Raw(`SELECT user_id FROM project_task_assignees WHERE task_id = ?`, taskID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err == nil {
			out = append(out, u)
		}
	}
	return out, nil
}

func (r *ProjectRepository) GetProjectByTask(taskID string) (string, error) {
	var pid string
	err := r.db.Raw(`SELECT project_id FROM project_tasks WHERE id = ?`, taskID).Row().Scan(&pid)
	return pid, wrapNotFound(err)
}
