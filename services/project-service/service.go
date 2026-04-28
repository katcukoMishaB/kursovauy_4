package main

type ProjectService struct {
	repo *ProjectRepository
}

func NewProjectService(repo *ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}
