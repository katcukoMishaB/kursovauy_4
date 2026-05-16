package main

import (
	"database/sql"
	"errors"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrForbidden      = errors.New("forbidden")
	ErrAlreadyMember  = errors.New("already a member")
	ErrAlreadyRequest = errors.New("active request already exists")
	ErrInvalidRole    = errors.New("invalid role")
	ErrTasksOpen      = errors.New("есть незавершённые задачи")
)

type ProjectService struct {
	repo *ProjectRepository
}

func NewProjectService(repo *ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) canManage(projectID, userID string, allowLeaderOnly bool) (bool, error) {
	if s.repo.IsAdmin(userID) {
		return true, nil
	}
	organizerID, err := s.repo.GetOrganizerID(projectID)
	if err != nil {
		return false, err
	}
	if organizerID == userID {
		return true, nil
	}
	role, err := s.repo.GetParticipantRole(projectID, userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if allowLeaderOnly {
		return role == "руководитель", nil
	}
	return role == "руководитель" || role == "заместитель", nil
}

func (s *ProjectService) CreateProject(organizerID string, req CreateProjectRequest) (string, error) {
	id, err := s.repo.CreateProject(organizerID, req)
	if err == nil {
		s.repo.LogActivity(organizerID, "project_created", &id)
	}
	return id, err
}

func (s *ProjectService) ListProjects(status, categoryID string) ([]ProjectExtended, error) {
	return s.repo.ListProjects(status, categoryID)
}

func (s *ProjectService) ListMyProjects(userID string) ([]ProjectWithRole, error) {
	return s.repo.ListMyProjects(userID)
}

func (s *ProjectService) GetProject(id string) (Project, error) {
	p, err := s.repo.GetProject(id)
	if err == sql.ErrNoRows {
		return Project{}, ErrNotFound
	}
	if err == nil {
		p.CategoryIDs, _ = s.repo.ListProjectCategoryIDs(id)
	}
	return p, err
}

func (s *ProjectService) UpdateProject(projectID, userID string, req CreateProjectRequest) error {
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
	return s.repo.UpdateProject(projectID, req)
}

func (s *ProjectService) ArchiveProject(projectID, userID string) error {
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
	if err := s.repo.ArchiveProject(projectID); err != nil {
		return err
	}
	pid := projectID
	s.repo.LogActivity(userID, "project_archived", &pid)
	return nil
}

func (s *ProjectService) ListCategories() ([]ProjectCategory, error) {
	return s.repo.ListCategories()
}

func (s *ProjectService) CreateParticipationRequest(projectID, userID string, req ParticipationRequestBody) (string, error) {
	organizerID, err := s.repo.GetOrganizerID(projectID)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if organizerID == userID {
		return "", ErrAlreadyMember
	}
	if _, err := s.repo.FindParticipation(projectID, userID); err == nil {
		return "", ErrAlreadyMember
	} else if err != sql.ErrNoRows {
		return "", err
	}
	if _, err := s.repo.FindActiveRequest(projectID, userID); err == nil {
		return "", ErrAlreadyRequest
	} else if err != sql.ErrNoRows {
		return "", err
	}
	id, err := s.repo.CreateParticipationRequest(projectID, userID, req.Comment, req.ResumeURL)
	if err == nil {
		pid := projectID
		s.repo.LogActivity(userID, "participation_requested", &pid)
	}
	return id, err
}

func (s *ProjectService) ListProjectRequests(projectID, userID string) ([]ParticipationRequestExtended, error) {
	allowed, err := s.canManage(projectID, userID, false)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	return s.repo.ListProjectRequests(projectID)
}

func (s *ProjectService) GetMyParticipationRequest(projectID, userID string) (ParticipationRequest, error) {
	req, err := s.repo.GetMyParticipationRequest(projectID, userID)
	if err == sql.ErrNoRows {
		return ParticipationRequest{}, ErrNotFound
	}
	return req, err
}

func (s *ProjectService) ApproveRequest(requestID, userID string) error {
	projectID, applicantID, err := s.repo.GetRequestProjectAndUser(requestID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed, err := s.canManage(projectID, userID, true)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	if err := s.repo.ApproveRequest(requestID, projectID, applicantID); err != nil {
		return err
	}
	pid := projectID
	s.repo.LogActivity(applicantID, "project_joined", &pid)
	return nil
}

func (s *ProjectService) RejectRequest(requestID, userID string) error {
	projectID, _, err := s.repo.GetRequestProjectAndUser(requestID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed, err := s.canManage(projectID, userID, false)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return s.repo.RejectRequest(requestID)
}

func (s *ProjectService) ListParticipants(projectID string) ([]ParticipantWithUser, error) {
	return s.repo.ListParticipants(projectID)
}

func (s *ProjectService) UpdateParticipantRole(projectID, targetUserID, currentUserID, newRole string) error {
	if newRole != "участник" && newRole != "заместитель" && newRole != "руководитель" {
		return ErrInvalidRole
	}
	organizerID, err := s.repo.GetOrganizerID(projectID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if organizerID == targetUserID {
		return ErrInvalidRole
	}
	allowed, err := s.canManage(projectID, currentUserID, false)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return s.repo.UpdateParticipantRole(projectID, targetUserID, newRole)
}

func (s *ProjectService) AddTag(projectID, userID, name string) error {
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
	return s.repo.AddTag(projectID, name)
}

func (s *ProjectService) GetProjectTags(projectID string) ([]string, error) {
	return s.repo.GetProjectTags(projectID)
}

func (s *ProjectService) DeleteTag(projectID, userID, name string) error {
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
	return s.repo.DeleteTag(projectID, name)
}
