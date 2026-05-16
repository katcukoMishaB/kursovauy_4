package main

import "time"

type User struct {
	ID               string    `json:"id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	RegistrationDate time.Time `json:"registration_date"`
	Status           bool      `json:"status"`
	UserType         string    `json:"user_type"`
	GroupID          *string   `json:"group_id"`
	GroupName        *string   `json:"group_name"`
}

type UserRole struct {
	UserID        string `json:"user_id"`
	IsParticipant bool   `json:"is_participant"`
	IsOrganizer   bool   `json:"is_organizer"`
	IsAdmin       bool   `json:"is_admin"`
}

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Password  string  `json:"password"`
	GroupID   *string `json:"group_id"`
}

type AdminCreateUserRequest struct {
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	Password      string  `json:"password"`
	IsParticipant bool    `json:"is_participant"`
	IsOrganizer   bool    `json:"is_organizer"`
	IsAdmin       bool    `json:"is_admin"`
	UserType      string  `json:"user_type"`
	GroupID       *string `json:"group_id"`
}

type AdminUpdateUserRequest struct {
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	Password      string  `json:"password"`
	Status        bool    `json:"status"`
	IsParticipant bool    `json:"is_participant"`
	IsOrganizer   bool    `json:"is_organizer"`
	IsAdmin       bool    `json:"is_admin"`
	UserType      string  `json:"user_type"`
	GroupID       *string `json:"group_id"`
}

type OrganizerRequest struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	ExperienceDesc string    `json:"experience_description"`
	ResumeURL      *string   `json:"resume_url"`
	SubmissionDate time.Time `json:"submission_date"`
	Status         string    `json:"status"`
	RequestType    string    `json:"request_type"`
}

type OrganizerRequestExtended struct {
	OrganizerRequest
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Skills    []string `json:"skills"`
	Interests []string `json:"interests"`
}

type CreateOrganizerRequestBody struct {
	ExperienceDescription string `json:"experience_description"`
	ResumeURL             string `json:"resume_url"`
	RequestType           string `json:"request_type"`
}
