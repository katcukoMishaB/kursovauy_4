package main

import (
	"kursovauy_4/internal/database"
	"kursovauy_4/internal/middleware"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	gormDB, err := database.ConnectGORM()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	if sqlDB, err := gormDB.DB(); err == nil {
		defer sqlDB.Close()
	}

	repo := NewReportRepository(gormDB)
	service := NewReportService(repo)
	handler := NewReportHandler(service)

	r := gin.Default()

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/excel/project-kpi/:id", handler.ProjectKPIExcel)
		auth.GET("/projects/:id/dashboard", handler.ProjectDashboard)
	}

	admin := r.Group("/")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.GET("/users", handler.ListUserActivity)
		admin.GET("/users/:id", handler.GetUserActivity)
		admin.GET("/projects", handler.ListProjectEfficiency)
		admin.GET("/projects/:id", handler.GetProjectEfficiency)
		admin.GET("/summary", handler.GetSummary)

		admin.GET("/kpi/users", handler.UserKPIs)
		admin.GET("/kpi/projects", handler.ProjectKPIs)
		admin.GET("/kpi/dashboard", handler.AdminDashboard)
		admin.GET("/kpi/groups", handler.GroupKPIs)
		admin.GET("/kpi/user-types", handler.UserTypeKPIs)
		admin.GET("/excel/kpi", handler.KPIExcel)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	log.Printf("Report service starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
