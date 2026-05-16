package main

import "time"

type UserActivityReport struct {
	UserID           string    `json:"user_id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	RegistrationDate time.Time `json:"registration_date"`
	ProjectsCount    int       `json:"projects_count"`
	TasksCompleted   int       `json:"tasks_completed"`
	MessagesSent     int       `json:"messages_sent"`
	CommentsCount    int       `json:"comments_count"`
}

type ProjectEfficiencyReport struct {
	ProjectID         string     `json:"project_id"`
	Title             string     `json:"title"`
	Status            string     `json:"status"`
	CreationDate      time.Time  `json:"creation_date"`
	CompletionDate    *time.Time `json:"completion_date"`
	ParticipantsCount int        `json:"participants_count"`
	TasksTotal        int        `json:"tasks_total"`
	TasksCompleted    int        `json:"tasks_completed"`
	CompletionRate    float64    `json:"completion_rate"`
}

type Summary struct {
	TotalUsers            int     `json:"total_users"`
	ActiveUsers           int     `json:"active_users"`
	TotalProjects         int     `json:"total_projects"`
	ActiveProjects        int     `json:"active_projects"`
	CompletedProjects     int     `json:"completed_projects"`
	TotalTasks            int     `json:"total_tasks"`
	CompletedTasks        int     `json:"completed_tasks"`
	AverageCompletionRate float64 `json:"average_completion_rate"`
}

