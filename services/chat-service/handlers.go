package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	service  *ChatService
	upgrader websocket.Upgrader
}

func NewChatHandler(service *ChatService) *ChatHandler {
	return &ChatHandler{
		service: service,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	projectID := c.Param("project_id")
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	chatID, err := h.service.CreateChat(projectID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": chatID, "message": "Chat created successfully"})
}

func (h *ChatHandler) GetChats(c *gin.Context) {
	projectID := c.Param("project_id")
	chats, err := h.service.GetChats(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chats"})
		return
	}
	if chats == nil {
		chats = []Chat{}
	}
	c.JSON(http.StatusOK, chats)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	chatID := c.Param("id")
	messages, err := h.service.GetMessages(chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	if messages == nil {
		messages = []ChatMessage{}
	}
	c.JSON(http.StatusOK, messages)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	chatID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	var req struct {
		MessageText string `json:"message_text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	message, err := h.service.SendMessage(chatID, userID, req.MessageText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": message.ID, "message": "Message sent successfully"})
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	chatID := c.Param("id")
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var msg struct {
			UserID      string `json:"user_id"`
			MessageText string `json:"message_text"`
		}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		saved, err := h.service.SendMessage(chatID, msg.UserID, msg.MessageText)
		if err != nil {
			continue
		}
		if err := conn.WriteJSON(saved); err != nil {
			break
		}
	}
}
