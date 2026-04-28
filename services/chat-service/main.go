package main

import (
	"kursovauy_4/internal/database"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.ConnectGORM()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	repo := NewChatRepository(db)
	service := NewChatService(repo)
	handler := NewChatHandler(service)

	router := gin.Default()
	router.POST("/projects/:project_id/chats", handler.CreateChat)
	router.GET("/projects/:project_id/chats", handler.GetChats)
	router.GET("/chats/:id/messages", handler.GetMessages)
	router.POST("/chats/:id/messages", handler.SendMessage)
	router.GET("/chats/:id/ws", handler.HandleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}

	log.Printf("Chat service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
