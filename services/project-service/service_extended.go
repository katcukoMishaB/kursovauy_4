package main

import "database/sql"

func (s *ProjectService) AdminCreateCategory(name string) (string, error) {
	return s.repo.CreateCategory(name)
}
func (s *ProjectService) AdminUpdateCategory(id, name string) error {
	return s.repo.UpdateCategory(id, name)
}
func (s *ProjectService) AdminDeleteCategory(id string) error {
	return s.repo.DeleteCategory(id)
}

func (s *ProjectService) AdminCreateTag(name string) (string, error) {
	return s.repo.CreateTagCatalog(name)
}
func (s *ProjectService) AdminUpdateTag(id, name string) error {
	return s.repo.UpdateTagCatalog(id, name)
}
func (s *ProjectService) AdminDeleteTag(id string) error {
	return s.repo.DeleteTagCatalog(id)
}

func (s *ProjectService) ListGoals(projectID string) ([]ProjectGoal, error) {
	return s.repo.ListGoals(projectID)
}

func (s *ProjectService) CreateGoal(projectID, userID string, req CreateGoalRequest) (string, error) {
	allowed, err := s.canManage(projectID, userID, false)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if !allowed {
		return "", ErrForbidden
	}
	return s.repo.CreateGoal(projectID, req.Title, req.Description)
}

func (s *ProjectService) ToggleGoal(goalID, userID string, achieved bool) error {
	pid, err := s.repo.GetGoalProject(goalID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed, err := s.canManage(pid, userID, false)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	if err := s.repo.ToggleGoal(goalID, achieved); err != nil {
		return err
	}
	if achieved {
		s.repo.LogActivity(userID, "goal_achieved", &pid)
	}
	return nil
}

func (s *ProjectService) DeleteGoal(goalID, userID string) error {
	pid, err := s.repo.GetGoalProject(goalID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed, err := s.canManage(pid, userID, false)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return s.repo.DeleteGoal(goalID)
}

func (s *ProjectService) ListRequiredSkills(projectID string) ([]string, error) {
	return s.repo.ListRequiredSkills(projectID)
}

func (s *ProjectService) AddRequiredSkill(projectID, userID, name string) error {
	allowed, err := s.canManage(projectID, userID, false)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return s.repo.AddRequiredSkill(projectID, name)
}

func (s *ProjectService) DeleteRequiredSkill(projectID, userID, name string) error {
	allowed, err := s.canManage(projectID, userID, false)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return s.repo.DeleteRequiredSkill(projectID, name)
}

func (s *ProjectService) AdminDeleteProject(id string) error {
	if _, err := s.repo.GetProject(id); err == sql.ErrNoRows {
		return ErrNotFound
	}
	return s.repo.DeleteProject(id)
}

func (s *ProjectService) AdminRestoreProject(id string) error {
	if _, err := s.repo.GetProject(id); err == sql.ErrNoRows {
		return ErrNotFound
	}
	return s.repo.RestoreProject(id)
}

func (s *ProjectService) CompleteProject(projectID, userID string) error {
	allowed, err := s.canManage(projectID, userID, true)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	done, err := s.repo.AllTasksDone(projectID)
	if err != nil {
		return err
	}
	if !done {
		return ErrTasksOpen
	}
	if err := s.repo.CompleteProject(projectID); err != nil {
		return err
	}
	pid := projectID
	s.repo.LogActivity(userID, "project_completed", &pid)
	return nil
}

func (s *ProjectService) ListTagCatalog() ([]CatalogTag, error) {
	return s.repo.ListTagCatalog()
}

func (s *ProjectService) AddTaskAssignee(taskID, currentUserID, targetUserID string) error {
	pid, err := s.repo.GetProjectByTask(taskID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed, err := s.canManage(pid, currentUserID, false)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return s.repo.AddTaskAssignee(taskID, targetUserID)
}

func (s *ProjectService) RemoveTaskAssignee(taskID, currentUserID, targetUserID string) error {
	pid, err := s.repo.GetProjectByTask(taskID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed, err := s.canManage(pid, currentUserID, false)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return s.repo.RemoveTaskAssignee(taskID, targetUserID)
}

func (s *ProjectService) ListTaskAssignees(taskID string) ([]string, error) {
	return s.repo.ListTaskAssignees(taskID)
}

func (s *ProjectService) Recommend(userID string) ([]RecommendedProject, error) {
	hasMatch, err := s.repo.HasInterestsOrSkills(userID)
	if err != nil {
		return nil, err
	}
	if hasMatch {
		recs, err := s.repo.RecommendForUser(userID)
		if err != nil {
			return nil, err
		}
		if len(recs) == 0 {
			return s.repo.ListActiveFresh(userID, 6)
		}
		return recs, nil
	}
	return s.repo.ListActiveFresh(userID, 6)
}
