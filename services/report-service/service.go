package main

type ReportService struct {
	repo *ReportRepository
}

func NewReportService(repo *ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}
