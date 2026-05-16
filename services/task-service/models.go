package main

import "time"

type Task struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"project_id"`
	AssignedTo     *string    `json:"assigned_to"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	Difficulty     int        `json:"difficulty"`
	QualityRating  *int       `json:"quality_rating"`
	AttachmentURL  *string    `json:"attachment_url"`
	CreationDate   time.Time  `json:"creation_date"`
	DueDate        *time.Time `json:"due_date"`
	CompletionDate *time.Time `json:"completion_date"`
	Overdue        bool       `json:"overdue"`
}

type TaskComment struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"task_id"`
	UserID          string    `json:"user_id"`
	Content         string    `json:"content"`
	PublicationDate time.Time `json:"publication_date"`
}

type CreateTaskRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Priority      string  `json:"priority"`
	Difficulty    int     `json:"difficulty"`
	DueDate       *string `json:"due_date"`
	AttachmentURL *string `json:"attachment_url"`
}

type UpdateTaskRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Priority      string  `json:"priority"`
	Difficulty    int     `json:"difficulty"`
	DueDate       *string `json:"due_date"`
	AttachmentURL *string `json:"attachment_url"`
}

type AssignTaskRequest struct {
	AssignedTo *string `json:"assigned_to"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type RateTaskRequest struct {
	QualityRating int `json:"quality_rating"`
}

type CommentRequest struct {
	Content string `json:"content"`
}
