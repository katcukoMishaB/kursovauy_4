package main

import "time"

type Project struct {
	ID               string     `json:"id"`
	OrganizerID      string     `json:"organizer_id"`
	CategoryID       *string    `json:"category_id"`
	Title            string     `json:"title"`
	ShortDescription string     `json:"short_description"`
	FullDescription  string     `json:"full_description"`
	Status           string     `json:"status"`
	CreationDate     time.Time  `json:"creation_date"`
	CompletionDate   *time.Time `json:"completion_date"`
}

type ProjectCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ParticipationRequest struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	UserID         string    `json:"user_id"`
	Comment        string    `json:"comment"`
	ResumeURL      string    `json:"resume_url"`
	SubmissionDate time.Time `json:"submission_date"`
	Status         string    `json:"status"`
}

type Participation struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	JoinDate  time.Time `json:"join_date"`
}

type ProjectExtended struct {
	Project
	OrganizerName     string   `json:"organizer_name"`
	ParticipantsCount int      `json:"participants_count"`
	CategoryName      string   `json:"category_name"`
	Tags              []string `json:"tags"`
}
