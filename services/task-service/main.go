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

	repo := NewTaskRepository(gormDB)
	service := NewTaskService(repo)
	handler := NewTaskHandler(service)

	r := gin.Default()

	r.GET("/projects/:project_id/tasks", handler.ListTasks)
	r.GET("/tasks/:id", handler.GetTask)
	r.GET("/tasks/:id/comments", handler.ListComments)

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/projects/:project_id/tasks", handler.CreateTask)
		auth.PUT("/tasks/:id", handler.UpdateTask)
		auth.POST("/tasks/:id/assign", handler.AssignTask)
		auth.POST("/tasks/:id/rate", handler.RateTask)
		auth.PUT("/tasks/:id/status", handler.UpdateTaskStatus)
		auth.PUT("/tasks/:id/attachment", handler.SetAttachment)
		auth.POST("/tasks/:id/comments", handler.AddComment)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	log.Printf("Task service starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
