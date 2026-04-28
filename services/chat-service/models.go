package main

import "time"

type Chat struct {
	ID           string    `json:"id" gorm:"column:id;primaryKey"`
	ProjectID    string    `json:"project_id" gorm:"column:project_id"`
	Name         string    `json:"name" gorm:"column:name"`
	CreationDate time.Time `json:"creation_date" gorm:"column:creation_date"`
}

func (Chat) TableName() string {
	return "project_chats"
}

type ChatMessage struct {
	ID          string    `json:"id" gorm:"column:id;primaryKey"`
	ChatID      string    `json:"chat_id" gorm:"column:chat_id"`
	UserID      string    `json:"user_id" gorm:"column:user_id"`
	MessageText string    `json:"message_text" gorm:"column:message_text"`
	SentAt      time.Time `json:"sent_at" gorm:"column:sent_at"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
