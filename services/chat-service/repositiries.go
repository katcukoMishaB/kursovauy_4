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
	var id string
	err := r.db.Raw(
		`INSERT INTO project_chats (project_id, name, creation_date) VALUES (?, ?, ?) RETURNING id`,
		projectID, name, time.Now(),
	).Row().Scan(&id)
	return id, err
}

func (r *ChatRepository) GetChats(projectID string) ([]Chat, error) {
	var chats []Chat
	err := r.db.Where("project_id = ?", projectID).Order("creation_date").Find(&chats).Error
	return chats, err
}

func (r *ChatRepository) GetMessages(chatID string) ([]ChatMessage, error) {
	rows, err := r.db.Raw(`
		SELECT m.id, m.chat_id, m.user_id, m.message_text, m.sent_at,
			COALESCE(NULLIF(TRIM(COALESCE(u.first_name,'') || ' ' || COALESCE(u.last_name,'')), ''), 'Неизвестный') AS author_name,
			COALESCE(pp.role, '') AS author_role
		FROM chat_messages m
		LEFT JOIN users u ON u.id = m.user_id
		LEFT JOIN project_chats pc ON pc.id = m.chat_id
		LEFT JOIN project_participations pp ON pp.project_id = pc.project_id AND pp.user_id = m.user_id
		WHERE m.chat_id = ?
		ORDER BY m.sent_at`, chatID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := []ChatMessage{}
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.ChatID, &m.UserID, &m.MessageText, &m.SentAt,
			&m.AuthorName, &m.AuthorRole); err == nil {
			messages = append(messages, m)
		}
	}
	return messages, nil
}

func (r *ChatRepository) enrichMessage(m *ChatMessage) {
	_ = r.db.Raw(`
		SELECT COALESCE(NULLIF(TRIM(COALESCE(u.first_name,'') || ' ' || COALESCE(u.last_name,'')), ''), 'Неизвестный'),
			COALESCE(pp.role, '')
		FROM users u
		LEFT JOIN project_chats pc ON pc.id = ?
		LEFT JOIN project_participations pp ON pp.project_id = pc.project_id AND pp.user_id = u.id
		WHERE u.id = ?`, m.ChatID, m.UserID).Row().Scan(&m.AuthorName, &m.AuthorRole)
}

func (r *ChatRepository) CreateMessage(chatID, userID, messageText string) (ChatMessage, error) {
	now := time.Now()
	message := ChatMessage{
		ChatID:      chatID,
		UserID:      userID,
		MessageText: messageText,
		SentAt:      now,
	}
	if err := r.db.Raw(
		`INSERT INTO chat_messages (chat_id, user_id, message_text, sent_at)
		 VALUES (?, ?, ?, ?) RETURNING id`,
		chatID, userID, messageText, now,
	).Row().Scan(&message.ID); err != nil {
		return ChatMessage{}, err
	}
	r.enrichMessage(&message)
	r.db.Exec(`
		INSERT INTO activity_log (user_id, project_id, action)
		SELECT ?, project_id, 'message_sent' FROM project_chats WHERE id = ?`,
		userID, chatID)
	return message, nil
}
