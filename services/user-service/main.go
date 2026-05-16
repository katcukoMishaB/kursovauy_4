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

	repo := NewUserRepository(gormDB)
	service := NewUserService(repo)
	handler := NewUserHandler(service)

	r := gin.Default()
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)
	r.GET("/groups", handler.PublicListGroups)

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/profile", handler.GetProfile)
		auth.PUT("/profile", handler.UpdateProfile)
		auth.POST("/organizer-request", handler.CreateOrganizerRequest)
		auth.GET("/my-request", handler.GetMyRequest)
		auth.GET("/skills", handler.ListSkills)
		auth.POST("/skills", handler.AddSkill)
		auth.DELETE("/skills/:name", handler.DeleteSkill)
		auth.GET("/interests", handler.ListInterests)
		auth.POST("/interests", handler.AddInterest)
		auth.DELETE("/interests/:categoryId", handler.DeleteInterest)
	}

	admin := r.Group("/")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.GET("/organizer-requests", handler.ListOrganizerRequests)
		admin.POST("/organizer-requests/:id/approve", handler.ApproveOrganizerRequest)
		admin.POST("/organizer-requests/:id/reject", handler.RejectOrganizerRequest)
		admin.GET("/users", handler.ListUsers)
		admin.PUT("/users/:id", handler.UpdateUserStatus)

		admin.GET("/admin/users", handler.AdminListUsers)
		admin.POST("/admin/users", handler.AdminCreateUser)
		admin.PUT("/admin/users/:id", handler.AdminUpdateUser)
		admin.DELETE("/admin/users/:id", handler.AdminDeleteUser)
		admin.GET("/admin/dashboard", handler.AdminDashboard)

		admin.POST("/admin/groups", handler.AdminCreateGroup)
		admin.PUT("/admin/groups/:id", handler.AdminUpdateGroup)
		admin.DELETE("/admin/groups/:id", handler.AdminDeleteGroup)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	log.Printf("User service starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
