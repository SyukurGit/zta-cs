package service

import (
	"errors"
	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/internal/repository"
)

type ChatService struct {
	ChatRepo   *repository.ChatRepository
	TicketRepo *repository.TicketRepository // Butuh ini untuk cek pemilik tiket
}

func NewChatService(chatRepo *repository.ChatRepository, ticketRepo *repository.TicketRepository) *ChatService {
	return &ChatService{ChatRepo: chatRepo, TicketRepo: ticketRepo}
}

// SendMessage mengirim pesan dengan validasi kepemilikan
func (s *ChatService) SendMessage(ticketID uint, senderID uint, role string, message string) (*domain.Chat, error) {
	// 1. Ambil Data Tiket
	ticket, err := s.TicketRepo.GetByID(ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// 2. AUTHORIZATION CHECK (Penting!)
	if role == domain.RoleUser {
		// Jika User, pastikan dia pemilik tiket
		if ticket.UserID != senderID {
			return nil, errors.New("access denied: you do not own this ticket")
		}
	} else if role == domain.RoleCS {
		// Jika CS, pastikan tiket tidak CLOSED (Opsional: Bisa tambah cek Assignment)
		if ticket.Status == "CLOSED" || ticket.Status == "LOCKED" {
			return nil, errors.New("cannot chat on closed tickets")
		}
	}

	// 3. Simpan Chat
	chat := &domain.Chat{
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderRole: role,
		Message:    message,
	}

	err = s.ChatRepo.CreateChat(chat)
	return chat, err
}

func (s *ChatService) GetHistory(ticketID uint, requestorID uint, role string) ([]domain.Chat, error) {
	// Validasi akses baca history sama dengan validasi kirim pesan
	ticket, err := s.TicketRepo.GetByID(ticketID)
	if err != nil {
		return nil, err
	}

	if role == domain.RoleUser && ticket.UserID != requestorID {
		return nil, errors.New("access denied")
	}
	
	return s.ChatRepo.GetChatHistory(ticketID)
}