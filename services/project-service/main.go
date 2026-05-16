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

	repo := NewProjectRepository(gormDB)
	service := NewProjectService(repo)
	handler := NewProjectHandler(service)

	r := gin.Default()

	r.GET("/projects", handler.ListProjects)
	r.GET("/projects/:id", handler.GetProject)
	r.GET("/projects/:id/participants", handler.ListParticipants)
	r.GET("/projects/:id/tags", handler.GetProjectTags)
	r.GET("/projects/:id/goals", handler.ListGoals)
	r.GET("/projects/:id/required-skills", handler.ListRequiredSkills)
	r.GET("/categories", handler.ListCategories)
	r.GET("/tag-catalog", handler.ListTagCatalog)
	r.GET("/tasks/:id/assignees", handler.ListTaskAssignees)

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/projects/my", handler.ListMyProjects)
		auth.GET("/projects/recommended", handler.Recommend)
		auth.POST("/projects/:id/participate", handler.CreateParticipationRequest)
		auth.GET("/projects/:id/requests", handler.ListProjectRequests)
		auth.GET("/projects/:id/my-request", handler.GetMyParticipationRequest)
		auth.POST("/requests/:id/approve", handler.ApproveRequest)
		auth.POST("/requests/:id/reject", handler.RejectRequest)
		auth.PUT("/projects/:id/participants/:userId/role", handler.UpdateParticipantRole)
		auth.DELETE("/projects/:id/tags/:tagName", handler.DeleteTag)
		auth.POST("/projects/:id/goals", handler.CreateGoal)
		auth.PUT("/goals/:goalId", handler.ToggleGoal)
		auth.DELETE("/goals/:goalId", handler.DeleteGoal)
		auth.POST("/projects/:id/required-skills", handler.AddRequiredSkill)
		auth.DELETE("/projects/:id/required-skills/:name", handler.DeleteRequiredSkill)
		auth.POST("/projects/:id/complete", handler.CompleteProject)
		auth.POST("/tasks/:id/assignees", handler.AddTaskAssignee)
		auth.DELETE("/tasks/:id/assignees/:userId", handler.RemoveTaskAssignee)

		auth.POST("/projects/:id/invitations", handler.CreateInvitations)
		auth.GET("/projects/:id/invitations", handler.ListProjectInvitations)
		auth.GET("/invitations/my", handler.ListMyInvitations)
		auth.GET("/invitations/count", handler.CountMyInvitations)
		auth.POST("/invitations/:id/accept", handler.AcceptInvitation)
		auth.POST("/invitations/:id/reject", handler.RejectInvitation)
		auth.POST("/invitations/:id/cancel", handler.CancelInvitation)
	}

	org := r.Group("/")
	org.Use(middleware.AuthMiddleware(), middleware.OrganizerMiddleware())
	{
		org.POST("/projects", handler.CreateProject)
		org.PUT("/projects/:id", handler.UpdateProject)
		org.POST("/projects/:id/archive", handler.ArchiveProject)
		org.POST("/projects/:id/tags", handler.AddTag)
	}

	admin := r.Group("/")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.DELETE("/admin/projects/:id", handler.AdminDeleteProject)
		admin.POST("/admin/projects/:id/restore", handler.AdminRestoreProject)

		admin.POST("/admin/categories", handler.AdminCreateCategory)
		admin.PUT("/admin/categories/:id", handler.AdminUpdateCategory)
		admin.DELETE("/admin/categories/:id", handler.AdminDeleteCategory)

		admin.POST("/admin/tag-catalog", handler.AdminCreateTag)
		admin.PUT("/admin/tag-catalog/:id", handler.AdminUpdateTag)
		admin.DELETE("/admin/tag-catalog/:id", handler.AdminDeleteTag)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	log.Printf("Project service starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
