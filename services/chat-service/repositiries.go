package main

import (
	"time"

	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateChat(projectID, name string) (string, error) {
	chat := Chat{
		ProjectID:    projectID,
		Name:         name,
		CreationDate: time.Now(),
	}
	if err := r.db.Create(&chat).Error; err != nil {
		return "", err
	}
	return chat.ID, nil
}

func (r *ChatRepository) GetChats(projectID string) ([]Chat, error) {
	var chats []Chat
	err := r.db.Where("project_id = ?", projectID).Order("creation_date").Find(&chats).Error
	return chats, err
}

func (r *ChatRepository) GetMessages(chatID string) ([]ChatMessage, error) {
	var messages []ChatMessage
	err := r.db.Where("chat_id = ?", chatID).Order("sent_at").Find(&messages).Error
	return messages, err
}

func (r *ChatRepository) CreateMessage(chatID, userID, messageText string) (ChatMessage, error) {
	message := ChatMessage{
		ChatID:      chatID,
		UserID:      userID,
		MessageText: messageText,
		SentAt:      time.Now(),
	}
	if err := r.db.Create(&message).Error; err != nil {
		return ChatMessage{}, err
	}
	return message, nil
}
