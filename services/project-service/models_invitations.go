package main

import "time"

type Invitation struct {
	ID                string     `json:"id"`
	ProjectID         string     `json:"project_id"`
	SenderID          string     `json:"sender_id"`
	RecipientUserID   *string    `json:"recipient_user_id"`
	RecipientEmail    *string    `json:"recipient_email"`
	RecipientGroupID  *string    `json:"recipient_group_id"`
	Message           string     `json:"message"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	RespondedAt       *time.Time `json:"responded_at"`
}

type InvitationWithProject struct {
	Invitation
	ProjectTitle       string  `json:"project_title"`
	ProjectImageURL    *string `json:"project_image_url"`
	SenderFirstName    string  `json:"sender_first_name"`
	SenderLastName     string  `json:"sender_last_name"`
	GroupName          *string `json:"group_name"`
}

type InvitationWithRecipient struct {
	Invitation
	RecipientFirstName *string `json:"recipient_first_name"`
	RecipientLastName  *string `json:"recipient_last_name"`
	GroupName          *string `json:"group_name"`
}

type CreateInvitationsRequest struct {
	Emails   []string `json:"emails"`
	GroupIDs []string `json:"group_ids"`
	Message  string   `json:"message"`
}
