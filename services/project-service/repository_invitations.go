package main

import (
	"database/sql"
	"strings"

	"gorm.io/gorm"
)

func (Invitation) TableName() string { return "project_invitations" }

func (r *ProjectRepository) ResolveEmailToUser(email string) (string, error) {
	var id string
	err := r.db.Raw(`SELECT id FROM users WHERE lower(email) = lower(?)`, strings.TrimSpace(email)).Scan(&id).Error
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", sql.ErrNoRows
	}
	return id, nil
}

func (r *ProjectRepository) GroupMemberIDs(groupID string) ([]string, error) {
	ids := []string{}
	err := r.db.Raw(`SELECT id FROM users WHERE group_id = ?`, groupID).Scan(&ids).Error
	return ids, err
}

func (r *ProjectRepository) CreateInvitation(projectID, senderID string,
	recipientUserID, recipientEmail, recipientGroupID *string, message string) (string, error) {
	var id string
	err := r.db.Raw(
		`INSERT INTO project_invitations
		   (project_id, sender_id, recipient_user_id, recipient_email, recipient_group_id, message, status)
		 VALUES (?, ?, ?, ?, ?, ?, 'pending') RETURNING id`,
		projectID, senderID, recipientUserID, recipientEmail, recipientGroupID, message,
	).Scan(&id).Error
	return id, err
}

func (r *ProjectRepository) HasPendingInvitation(projectID string, userID, email *string) (bool, error) {
	var exists bool
	var err error
	switch {
	case userID != nil:
		err = r.db.Raw(
			`SELECT EXISTS(
				SELECT 1 FROM project_invitations
				WHERE project_id = ? AND recipient_user_id = ? AND status = 'pending')`,
			projectID, *userID,
		).Scan(&exists).Error
	case email != nil:
		err = r.db.Raw(
			`SELECT EXISTS(
				SELECT 1 FROM project_invitations
				WHERE project_id = ? AND lower(recipient_email) = lower(?) AND status = 'pending')`,
			projectID, *email,
		).Scan(&exists).Error
	default:
		return false, nil
	}
	return exists, err
}

func (r *ProjectRepository) ListIncomingInvitations(userID, email string) ([]InvitationWithProject, error) {
	rows, err := r.db.Raw(`
		SELECT pi.id, pi.project_id, pi.sender_id, pi.recipient_user_id, pi.recipient_email,
		       pi.recipient_group_id, pi.message, pi.status, pi.created_at, pi.responded_at,
		       p.title, p.image_url,
		       COALESCE(u.first_name, ''), COALESCE(u.last_name, ''),
		       g.name AS group_name
		FROM project_invitations pi
		JOIN projects p ON p.id = pi.project_id
		LEFT JOIN users u ON u.id = pi.sender_id
		LEFT JOIN groups g ON g.id = pi.recipient_group_id
		WHERE pi.status = 'pending'
		  AND (pi.recipient_user_id = ? OR (pi.recipient_user_id IS NULL AND lower(pi.recipient_email) = lower(?)))
		ORDER BY pi.created_at DESC`,
		userID, email,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InvitationWithProject{}
	for rows.Next() {
		var inv InvitationWithProject
		if err := rows.Scan(&inv.ID, &inv.ProjectID, &inv.SenderID, &inv.RecipientUserID, &inv.RecipientEmail,
			&inv.RecipientGroupID, &inv.Message, &inv.Status, &inv.CreatedAt, &inv.RespondedAt,
			&inv.ProjectTitle, &inv.ProjectImageURL,
			&inv.SenderFirstName, &inv.SenderLastName,
			&inv.GroupName); err != nil {
			continue
		}
		out = append(out, inv)
	}
	return out, nil
}

func (r *ProjectRepository) CountIncomingPending(userID, email string) (int, error) {
	var n int
	err := r.db.Raw(`
		SELECT COUNT(*) FROM project_invitations
		WHERE status = 'pending'
		  AND (recipient_user_id = ? OR (recipient_user_id IS NULL AND lower(recipient_email) = lower(?)))`,
		userID, email,
	).Scan(&n).Error
	return n, err
}

func (r *ProjectRepository) ListProjectInvitations(projectID string) ([]InvitationWithRecipient, error) {
	rows, err := r.db.Raw(`
		SELECT pi.id, pi.project_id, pi.sender_id, pi.recipient_user_id, pi.recipient_email,
		       pi.recipient_group_id, pi.message, pi.status, pi.created_at, pi.responded_at,
		       u.first_name, u.last_name, g.name AS group_name
		FROM project_invitations pi
		LEFT JOIN users u ON u.id = pi.recipient_user_id
		LEFT JOIN groups g ON g.id = pi.recipient_group_id
		WHERE pi.project_id = ?
		ORDER BY pi.created_at DESC`,
		projectID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InvitationWithRecipient{}
	for rows.Next() {
		var inv InvitationWithRecipient
		if err := rows.Scan(&inv.ID, &inv.ProjectID, &inv.SenderID, &inv.RecipientUserID, &inv.RecipientEmail,
			&inv.RecipientGroupID, &inv.Message, &inv.Status, &inv.CreatedAt, &inv.RespondedAt,
			&inv.RecipientFirstName, &inv.RecipientLastName, &inv.GroupName); err != nil {
			continue
		}
		out = append(out, inv)
	}
	return out, nil
}

func (r *ProjectRepository) GetInvitation(id string) (Invitation, error) {
	var inv Invitation
	row := r.db.Raw(
		`SELECT id, project_id, sender_id, recipient_user_id, recipient_email,
		        recipient_group_id, message, status, created_at, responded_at
		 FROM project_invitations WHERE id = ?`, id,
	).Row()
	if err := row.Scan(&inv.ID, &inv.ProjectID, &inv.SenderID, &inv.RecipientUserID, &inv.RecipientEmail,
		&inv.RecipientGroupID, &inv.Message, &inv.Status, &inv.CreatedAt, &inv.RespondedAt); err != nil {
		return inv, err
	}
	return inv, nil
}

func (r *ProjectRepository) AcceptInvitation(invitationID, userID string) (string, error) {
	var projectID string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(
			`UPDATE project_invitations SET status='accepted', responded_at=NOW(),
			       recipient_user_id = COALESCE(recipient_user_id, ?)
			 WHERE id = ? RETURNING project_id`,
			userID, invitationID,
		).Scan(&projectID).Error; err != nil {
			return err
		}
		if projectID == "" {
			return sql.ErrNoRows
		}
		return tx.Exec(
			`INSERT INTO project_participations (project_id, user_id, role, join_date)
			 VALUES (?, ?, 'участник', CURRENT_DATE) ON CONFLICT DO NOTHING`,
			projectID, userID,
		).Error
	})
	return projectID, err
}

func (r *ProjectRepository) RejectInvitation(invitationID, userID string) error {
	res := r.db.Exec(
		`UPDATE project_invitations SET status='rejected', responded_at=NOW(),
		       recipient_user_id = COALESCE(recipient_user_id, ?)
		 WHERE id = ? AND status = 'pending'`,
		userID, invitationID,
	)
	return res.Error
}

func (r *ProjectRepository) CancelInvitation(invitationID string) error {
	return r.db.Exec(
		`UPDATE project_invitations SET status='cancelled', responded_at=NOW()
		 WHERE id = ? AND status = 'pending'`, invitationID,
	).Error
}

func (r *ProjectRepository) GetUserEmail(userID string) (string, error) {
	var email string
	err := r.db.Raw(`SELECT email FROM users WHERE id = ?`, userID).Scan(&email).Error
	return email, err
}
