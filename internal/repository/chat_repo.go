package repository

import (
	"github.com/syukurgit/zta/internal/domain"
	"gorm.io/gorm"
)

type ChatRepository struct {
	DB *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{DB: db}
}

// CreateChat menyimpan pesan baru
func (r *ChatRepository) CreateChat(chat *domain.Chat) error {
	return r.DB.Create(chat).Error
}

// GetChatHistory mengambil semua pesan dalam 1 tiket (urut dari lama ke baru)
func (r *ChatRepository) GetChatHistory(ticketID uint) ([]domain.Chat, error) {
	var chats []domain.Chat
	err := r.DB.Where("ticket_id = ?", ticketID).Order("created_at asc").Find(&chats).Error
	return chats, err
}