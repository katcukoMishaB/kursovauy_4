package main

import "time"

type Task struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	AssignedTo   *string   `json:"assigned_to"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	CreationDate time.Time `json:"creation_date"`
}

type TaskComment struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"task_id"`
	UserID          string    `json:"user_id"`
	Content         string    `json:"content"`
	PublicationDate time.Time `json:"publication_date"`
}
