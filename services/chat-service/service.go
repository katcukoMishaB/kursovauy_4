package main

type ChatService struct {
	repo *ChatRepository
}

func NewChatService(repo *ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) CreateChat(projectID, name string) (string, error) {
	return s.repo.CreateChat(projectID, name)
}

func (s *ChatService) GetChats(projectID string) ([]Chat, error) {
	return s.repo.GetChats(projectID)
}

func (s *ChatService) GetMessages(chatID string) ([]ChatMessage, error) {
	return s.repo.GetMessages(chatID)
}

func (s *ChatService) SendMessage(chatID, userID, messageText string) (ChatMessage, error) {
	return s.repo.CreateMessage(chatID, userID, messageText)
}
