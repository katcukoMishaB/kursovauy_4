package main

import "time"

type User struct {
	ID               string    `json:"id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	RegistrationDate time.Time `json:"registration_date"`
	Status           bool      `json:"status"`
}

type UserRole struct {
	UserID        string `json:"user_id"`
	IsParticipant bool   `json:"is_participant"`
	IsOrganizer   bool   `json:"is_organizer"`
	IsAdmin       bool   `json:"is_admin"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OrganizerRequest struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	ExperienceDesc string    `json:"experience_description"`
	SubmissionDate time.Time `json:"submission_date"`
	Status         string    `json:"status"`
	AdminComment   *string   `json:"admin_comment"`
}
