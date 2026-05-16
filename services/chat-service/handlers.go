package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	service  *ChatService
	upgrader websocket.Upgrader
	hub      *Hub
}

func NewChatHandler(service *ChatService) *ChatHandler {
	return &ChatHandler{
		service: service,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		hub: NewHub(),
	}
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	projectID := c.Param("project_id")
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	chatID, err := h.service.CreateChat(projectID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать чат"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": chatID, "message": "Чат создан"})
}

func (h *ChatHandler) GetChats(c *gin.Context) {
	projectID := c.Param("project_id")
	chats, err := h.service.GetChats(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить чаты"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить сообщения"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные формы"})
		return
	}
	message, err := h.service.SendMessage(chatID, userID, req.MessageText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отправить сообщение"})
		return
	}
	h.hub.Broadcast(chatID, message)
	c.JSON(http.StatusOK, gin.H{"id": message.ID, "message": "Сообщение отправлено"})
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	chatID := c.Param("id")
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.hub.Join(chatID, conn)
	defer h.hub.Leave(chatID, conn)

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
		h.hub.Broadcast(chatID, saved)
	}
}
