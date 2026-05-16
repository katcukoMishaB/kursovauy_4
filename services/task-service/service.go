package main

import (
	"database/sql"
	"errors"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
	ErrBadInput  = errors.New("bad input")
)

type TaskService struct {
	repo *TaskRepository
}

func NewTaskService(repo *TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) canManage(projectID, userID string) (bool, error) {
	organizerID, err := s.repo.GetOrganizerID(projectID)
	if err != nil {
		return false, err
	}
	if organizerID == userID {
		return true, nil
	}
	role, err := s.repo.GetParticipantRole(projectID, userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return role == "руководитель" || role == "заместитель", nil
}

func (s *TaskService) CreateTask(projectID, userID string, req CreateTaskRequest) (string, error) {
	allowed, err := s.canManage(projectID, userID)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if !allowed {
		return "", ErrForbidden
	}
	id, err := s.repo.CreateTask(projectID, req)
	if err != nil {
		return "", err
	}
	pid := projectID
	tid := id
	uid := userID
	s.repo.LogActivity(&uid, &pid, &tid, "task_created")
	return id, nil
}

func (s *TaskService) ListTasks(projectID, status string) ([]Task, error) {
	return s.repo.ListTasks(projectID, status)
}

func (s *TaskService) GetTask(id string) (Task, error) {
	t, err := s.repo.GetTask(id)
	if err == sql.ErrNoRows {
		return Task{}, ErrNotFound
	}
	return t, err
}

func (s *TaskService) UpdateTask(id, userID string, req UpdateTaskRequest) error {
	projectID, err := s.repo.GetTaskProject(id)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed, err := s.canManage(projectID, userID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return s.repo.UpdateTask(id, req)
}

func (s *TaskService) AssignTask(taskID, userID string, req AssignTaskRequest) error {
	projectID, err := s.repo.GetTaskProject(taskID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed, err := s.canManage(projectID, userID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	status := "новая"
	if req.AssignedTo != nil {
		status = "в работе"
	}
	if err := s.repo.AssignTask(taskID, req.AssignedTo, status); err != nil {
		return err
	}
	if req.AssignedTo != nil {
		uid := *req.AssignedTo
		pid := projectID
		tid := taskID
		s.repo.LogActivity(&uid, &pid, &tid, "task_assigned")
	}
	return nil
}

func (s *TaskService) UpdateTaskStatus(taskID, userID, status string) error {
	if status != "новая" && status != "в работе" && status != "на проверке" && status != "завершена" {
		return ErrBadInput
	}
	projectID, err := s.repo.GetTaskProject(taskID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	current, err := s.repo.GetTask(taskID)
	if err != nil {
		return err
	}

	canMgr, err := s.canManage(projectID, userID)
	if err != nil {
		return err
	}

	if !canMgr {
		isAssigned := current.AssignedTo != nil && *current.AssignedTo == userID
		if !isAssigned {
			extra, _ := s.repo.ListTaskExtraAssignees(taskID)
			for _, u := range extra {
				if u == userID {
					isAssigned = true
					break
				}
			}
		}
		if !isAssigned {
			return ErrForbidden
		}
		if status == "новая" {
			return ErrForbidden
		}
		if current.Status == "на проверке" && status == "в работе" {
			return ErrForbidden
		}
		if status == "завершена" {
			return ErrForbidden
		}
	}

	if err := s.repo.UpdateStatus(taskID, status); err != nil {
		return err
	}
	if status == "завершена" {
		uid := userID
		pid := projectID
		tid := taskID
		s.repo.LogActivity(&uid, &pid, &tid, "task_completed")
	}
	return nil
}

func (s *TaskService) RateTask(taskID, userID string, rating int) error {
	if rating < 1 || rating > 5 {
		return ErrBadInput
	}
	projectID, err := s.repo.GetTaskProject(taskID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed, err := s.canManage(projectID, userID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	if err := s.repo.RateTask(taskID, rating); err != nil {
		return err
	}
	uid := userID
	pid := projectID
	tid := taskID
	s.repo.LogActivity(&uid, &pid, &tid, "task_rated")
	return nil
}

func (s *TaskService) SetAttachment(taskID, userID, url string) error {
	projectID, err := s.repo.GetTaskProject(taskID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	task, err := s.repo.GetTask(taskID)
	if err != nil {
		return err
	}

	canMgr, err := s.canManage(projectID, userID)
	if err != nil {
		return err
	}
	if !canMgr {
		isAssigned := task.AssignedTo != nil && *task.AssignedTo == userID
		if !isAssigned {
			extra, _ := s.repo.ListTaskExtraAssignees(taskID)
			for _, u := range extra {
				if u == userID {
					isAssigned = true
					break
				}
			}
		}
		if !isAssigned {
			return ErrForbidden
		}
	}

	var ptr *string
	if url != "" {
		ptr = &url
	}
	return s.repo.UpdateAttachment(taskID, ptr)
}

func (s *TaskService) AddComment(taskID, userID string, req CommentRequest) (string, error) {
	id, err := s.repo.AddComment(taskID, userID, req.Content)
	if err != nil {
		return "", err
	}
	if pid, perr := s.repo.GetTaskProject(taskID); perr == nil {
		uid := userID
		pidv := pid
		tid := taskID
		s.repo.LogActivity(&uid, &pidv, &tid, "comment_added")
	}
	return id, nil
}

func (s *TaskService) ListComments(taskID string) ([]TaskComment, error) {
	return s.repo.ListComments(taskID)
}
