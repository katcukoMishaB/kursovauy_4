package main

import (
	"database/sql"
	"errors"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
)

type ReportService struct {
	repo *ReportRepository
}

func NewReportService(repo *ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) ListUserActivity() ([]UserActivityReport, error) {
	return s.repo.ListUserActivity()
}

func (s *ReportService) GetUserActivity(userID string) (UserActivityReport, error) {
	rep, err := s.repo.GetUserActivity(userID)
	if err == sql.ErrNoRows {
		return rep, ErrNotFound
	}
	return rep, err
}

func (s *ReportService) ListProjectEfficiency() ([]ProjectEfficiencyReport, error) {
	return s.repo.ListProjectEfficiency()
}

func (s *ReportService) GetProjectEfficiency(projectID string) (ProjectEfficiencyReport, error) {
	rep, err := s.repo.GetProjectEfficiency(projectID)
	if err == sql.ErrNoRows {
		return rep, ErrNotFound
	}
	return rep, err
}

func (s *ReportService) GetSummary() (Summary, error) {
	return s.repo.GetSummary()
}
