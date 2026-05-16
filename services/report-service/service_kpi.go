package main

import (
	"database/sql"
	"time"
)

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func makeRange(from, to string) DateRange {
	return DateRange{From: parseDate(from), To: parseDate(to)}
}

func (s *ReportService) UserKPIs(projectID, from, to string) ([]UserKPI, error) {
	return s.repo.UserKPIs(projectID, makeRange(from, to))
}

func (s *ReportService) UserKPIsFiltered(projectID, from, to, groupID, userType string) ([]UserKPI, error) {
	return s.repo.UserKPIsFiltered(projectID, makeRange(from, to), UserKPIFilter{
		GroupID: groupID, UserType: userType,
	})
}

func (s *ReportService) GroupKPIs(from, to string) ([]GroupKPI, error) {
	return s.repo.GroupKPIs(makeRange(from, to))
}

func (s *ReportService) UserTypeKPIs(from, to string) ([]UserTypeKPI, error) {
	return s.repo.UserTypeKPIs(makeRange(from, to))
}

func (s *ReportService) ProjectKPIs(from, to string) ([]ProjectKPI, error) {
	return s.repo.ProjectKPIs(makeRange(from, to))
}

func (s *ReportService) ProjectKPI(projectID, userID string, isAdmin bool) (ProjectKPI, error) {
	if !isAdmin {
		title, organizerID, err := s.repo.GetProjectInfo(projectID)
		if err == sql.ErrNoRows {
			return ProjectKPI{}, ErrNotFound
		}
		if err != nil {
			return ProjectKPI{}, err
		}
		_ = title
		if organizerID != userID {
			role, err := s.repo.GetParticipantRole(projectID, userID)
			if err == sql.ErrNoRows || (role != "руководитель" && role != "заместитель") {
				return ProjectKPI{}, ErrForbidden
			}
			if err != nil {
				return ProjectKPI{}, err
			}
		}
	}
	k, err := s.repo.ProjectKPI(projectID)
	if err == sql.ErrNoRows {
		return ProjectKPI{}, ErrNotFound
	}
	return k, err
}

type DashboardData struct {
	Timeseries             Timeseries             `json:"timeseries"`
	ProjectStatusBreakdown ProjectStatusBreakdown `json:"project_status_breakdown"`
	TopUsers               []UserKPI              `json:"top_users"`
}

func (s *ReportService) DashboardData() (DashboardData, error) {
	var d DashboardData
	ts, err := s.repo.TimeseriesLast30()
	if err != nil {
		return d, err
	}
	d.Timeseries = ts

	psb, err := s.repo.GlobalProjectStatusBreakdown()
	if err != nil {
		return d, err
	}
	d.ProjectStatusBreakdown = psb

	users, err := s.repo.UserKPIs("", DateRange{})
	if err != nil {
		return d, err
	}
	if len(users) > 5 {
		users = users[:5]
	}
	d.TopUsers = users
	return d, nil
}

type ProjectDashboardData struct {
	KPI               ProjectKPI       `json:"kpi"`
	StatusBreakdown   StatusBreakdown  `json:"status_breakdown"`
	UserKPIs          []UserKPI        `json:"user_kpis"`
}

func (s *ReportService) ProjectDashboard(projectID, userID string, isAdmin bool) (ProjectDashboardData, error) {
	return s.ProjectDashboardFiltered(projectID, userID, isAdmin, "", "")
}

func (s *ReportService) ProjectDashboardFiltered(projectID, userID string, isAdmin bool, from, to string) (ProjectDashboardData, error) {
	var d ProjectDashboardData
	kpi, err := s.ProjectKPI(projectID, userID, isAdmin)
	if err != nil {
		return d, err
	}
	d.KPI = kpi

	sb, err := s.repo.ProjectStatusBreakdown(projectID)
	if err != nil {
		return d, err
	}
	d.StatusBreakdown = sb

	uk, err := s.repo.UserKPIs(projectID, makeRange(from, to))
	if err != nil {
		return d, err
	}
	d.UserKPIs = filterActive(uk)
	return d, nil
}

func filterActive(uk []UserKPI) []UserKPI {
	out := []UserKPI{}
	for _, u := range uk {
		if u.TasksAssigned > 0 || u.CommentsCount > 0 || u.MessagesCount > 0 {
			out = append(out, u)
		}
	}
	return out
}
