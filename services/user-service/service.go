package main

import (
	"database/sql"
	"errors"
	"kursovauy_4/internal/auth"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidLogin    = errors.New("invalid credentials")
	ErrUserDisabled    = errors.New("user disabled")
	ErrUserNotFound    = errors.New("user not found")
	ErrRequestExists   = errors.New("active request exists")
	ErrRequestNotFound = errors.New("request not found")
)

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(req RegisterRequest) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	id, err := s.repo.CreateUser(req.FirstName, req.LastName, req.Email, string(hashed))
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return "", ErrEmailExists
		}
		return "", err
	}
	s.repo.LogActivity(id, "user_registered", nil, nil)
	return id, nil
}

func (s *UserService) Login(email, password string) (string, User, string, error) {
	user, hashed, err := s.repo.FindUserByEmail(email)
	if err == sql.ErrNoRows {
		return "", User{}, "", ErrInvalidLogin
	}
	if err != nil {
		return "", User{}, "", err
	}
	if !user.Status {
		return "", User{}, "", ErrUserDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return "", User{}, "", ErrInvalidLogin
	}
	role, err := s.repo.GetUserRole(user.ID)
	if err != nil {
		return "", User{}, "", err
	}
	roleStr := "participant"
	if role.IsAdmin {
		roleStr = "admin"
	} else if role.IsOrganizer {
		roleStr = "organizer"
	}
	token, err := auth.GenerateToken(user.ID, user.Email, roleStr)
	if err != nil {
		return "", User{}, "", err
	}
	s.repo.LogActivity(user.ID, "login", nil, nil)
	full, ferr := s.repo.FindUserByID(user.ID)
	if ferr == nil {
		user = full
	}
	return token, user, roleStr, nil
}

func (s *UserService) GetProfile(userID string) (User, error) {
	user, err := s.repo.FindUserByID(userID)
	if err == sql.ErrNoRows {
		return User{}, ErrUserNotFound
	}
	return user, err
}

func (s *UserService) UpdateProfile(userID string, req UpdateProfileRequest) error {
	var hashed *string
	if req.Password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		str := string(h)
		hashed = &str
	}
	return s.repo.UpdateProfile(userID, req.FirstName, req.LastName, hashed, req.GroupID)
}

func (s *UserService) CreateOrganizerRequest(userID string, req CreateOrganizerRequestBody) (string, error) {
	reqType := req.RequestType
	if reqType != "teacher" {
		reqType = "organizer"
	}
	if _, err := s.repo.FindActiveOrganizerRequest(userID, reqType); err == nil {
		return "", ErrRequestExists
	} else if err != sql.ErrNoRows {
		return "", err
	}
	return s.repo.CreateOrganizerRequest(userID, req.ExperienceDescription, req.ResumeURL, reqType)
}

func (s *UserService) GetMyLatestRequest(userID, requestType string) (OrganizerRequest, error) {
	rq, err := s.repo.GetLatestRequest(userID, requestType)
	if err == sql.ErrNoRows {
		return OrganizerRequest{}, ErrRequestNotFound
	}
	return rq, err
}

func (s *UserService) ListOrganizerRequests(filterType string) ([]OrganizerRequestExtended, error) {
	return s.repo.ListOrganizerRequests(filterType)
}

func (s *UserService) ApproveOrganizerRequest(requestID string) error {
	return s.repo.ApproveOrganizerRequest(requestID)
}

func (s *UserService) RejectOrganizerRequest(requestID string) error {
	return s.repo.RejectOrganizerRequest(requestID)
}

func (s *UserService) ListUsers() ([]User, error) {
	return s.repo.ListUsers()
}

func (s *UserService) UpdateUserStatus(userID string, status bool) error {
	return s.repo.UpdateUserStatus(userID, status)
}

func (s *UserService) ListGroups() ([]Group, error)          { return s.repo.ListGroups() }
func (s *UserService) CreateGroup(name string) (string, error) { return s.repo.CreateGroup(name) }
func (s *UserService) UpdateGroup(id, name string) error    { return s.repo.UpdateGroup(id, name) }
func (s *UserService) DeleteGroup(id string) error          { return s.repo.DeleteGroup(id) }
