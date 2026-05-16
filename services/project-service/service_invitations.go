package main

import (
	"database/sql"
	"strings"
)

func (s *ProjectService) CreateInvitations(projectID, senderID string, req CreateInvitationsRequest) (int, error) {
	allowed, err := s.canManage(projectID, senderID, false)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if !allowed {
		return 0, ErrForbidden
	}

	created := 0
	emails := dedupeStrings(req.Emails)
	groups := dedupeStrings(req.GroupIDs)

	for _, email := range emails {
		email = strings.TrimSpace(strings.ToLower(email))
		if email == "" {
			continue
		}
		uid, err := s.repo.ResolveEmailToUser(email)
		var uidArg, emailArg *string
		if err == nil && uid != "" {
			uidArg = &uid
			if _, errP := s.repo.FindParticipation(projectID, uid); errP == nil {
				continue
			}
		} else if err == sql.ErrNoRows {
			e := email
			emailArg = &e
		} else {
			return created, err
		}
		exists, _ := s.repo.HasPendingInvitation(projectID, uidArg, emailArg)
		if exists {
			continue
		}
		if _, err := s.repo.CreateInvitation(projectID, senderID, uidArg, emailArg, nil, req.Message); err != nil {
			return created, err
		}
		created++
	}

	for _, gid := range groups {
		members, err := s.repo.GroupMemberIDs(gid)
		if err != nil {
			return created, err
		}
		groupArg := gid
		for _, uid := range members {
			if uid == senderID {
				continue
			}
			if _, errP := s.repo.FindParticipation(projectID, uid); errP == nil {
				continue
			}
			u := uid
			exists, _ := s.repo.HasPendingInvitation(projectID, &u, nil)
			if exists {
				continue
			}
			if _, err := s.repo.CreateInvitation(projectID, senderID, &u, nil, &groupArg, req.Message); err != nil {
				return created, err
			}
			created++
		}
	}
	return created, nil
}

func (s *ProjectService) ListIncomingInvitations(userID string) ([]InvitationWithProject, error) {
	email, err := s.repo.GetUserEmail(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListIncomingInvitations(userID, email)
}

func (s *ProjectService) CountIncomingInvitations(userID string) (int, error) {
	email, err := s.repo.GetUserEmail(userID)
	if err != nil {
		return 0, err
	}
	return s.repo.CountIncomingPending(userID, email)
}

func (s *ProjectService) ListProjectInvitations(projectID, userID string) ([]InvitationWithRecipient, error) {
	allowed, err := s.canManage(projectID, userID, false)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	return s.repo.ListProjectInvitations(projectID)
}

func (s *ProjectService) AcceptInvitation(invitationID, userID string) (string, error) {
	inv, err := s.repo.GetInvitation(invitationID)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if inv.Status != "pending" {
		return "", ErrForbidden
	}
	if inv.RecipientUserID != nil && *inv.RecipientUserID != userID {
		return "", ErrForbidden
	}
	if inv.RecipientUserID == nil && inv.RecipientEmail != nil {
		email, _ := s.repo.GetUserEmail(userID)
		if !strings.EqualFold(email, *inv.RecipientEmail) {
			return "", ErrForbidden
		}
	}
	projectID, err := s.repo.AcceptInvitation(invitationID, userID)
	if err == nil {
		pid := projectID
		s.repo.LogActivity(userID, "project_joined", &pid)
	}
	return projectID, err
}

func (s *ProjectService) RejectInvitation(invitationID, userID string) error {
	inv, err := s.repo.GetInvitation(invitationID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if inv.RecipientUserID != nil && *inv.RecipientUserID != userID {
		return ErrForbidden
	}
	if inv.RecipientUserID == nil && inv.RecipientEmail != nil {
		email, _ := s.repo.GetUserEmail(userID)
		if !strings.EqualFold(email, *inv.RecipientEmail) {
			return ErrForbidden
		}
	}
	return s.repo.RejectInvitation(invitationID, userID)
}

func (s *ProjectService) CancelInvitation(invitationID, userID string) error {
	inv, err := s.repo.GetInvitation(invitationID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if inv.SenderID != userID {
		allowed, err := s.canManage(inv.ProjectID, userID, false)
		if err != nil {
			return err
		}
		if !allowed {
			return ErrForbidden
		}
	}
	return s.repo.CancelInvitation(invitationID)
}

func dedupeStrings(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
