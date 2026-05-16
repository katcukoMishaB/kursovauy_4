package main

import "time"

type Project struct {
	ID               string     `json:"id"`
	OrganizerID      string     `json:"organizer_id"`
	Title            string     `json:"title"`
	ShortDescription string     `json:"short_description"`
	FullDescription  string     `json:"full_description"`
	GoalDescription  *string    `json:"goal_description"`
	Status           string     `json:"status"`
	CreationDate     time.Time  `json:"creation_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`
	CompletionDate   *time.Time `json:"completion_date"`
	ImageURL         *string    `json:"image_url"`
	OrganizerName    string     `json:"organizer_name,omitempty"`
	CategoryIDs      []string   `json:"category_ids,omitempty"`
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
	ParticipantsCount int      `json:"participants_count"`
	CategoryName      string   `json:"category_name"`
	Tags              []string `json:"tags"`
}

type ProjectWithRole struct {
	Project
	UserRole string `json:"user_role"`
}

type ParticipantWithUser struct {
	Participation
	Email     *string  `json:"email"`
	FirstName *string  `json:"first_name"`
	LastName  *string  `json:"last_name"`
	Skills    []string `json:"skills"`
}

type CreateProjectRequest struct {
	CategoryID       *string  `json:"category_id"`
	CategoryIDs      []string `json:"category_ids"`
	Title            string   `json:"title"`
	ShortDescription string   `json:"short_description"`
	FullDescription  string   `json:"full_description"`
	GoalDescription  *string  `json:"goal_description"`
	PlannedEndDate   *string  `json:"planned_end_date"`
	ImageURL         *string  `json:"image_url"`
}

type ParticipationRequestExtended struct {
	ParticipationRequest
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Skills    []string `json:"skills"`
}

type CatalogTag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ParticipationRequestBody struct {
	Comment   string `json:"comment"`
	ResumeURL string `json:"resume_url"`
}

type ProjectGoal struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	Title         string     `json:"title"`
	Description   *string    `json:"description"`
	IsAchieved    bool       `json:"is_achieved"`
	CreationDate  time.Time  `json:"creation_date"`
	AchievedDate  *time.Time `json:"achieved_date"`
}

type CreateGoalRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

type RecommendedProject struct {
	ProjectExtended
	Score          int      `json:"score"`
	MatchedReasons []string `json:"matched_reasons"`
}
