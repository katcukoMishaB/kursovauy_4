package main

type TaskService struct {
	repo *TaskRepository
}

func NewTaskService(repo *TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}
