package service

import (
	"errors"

	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/internal/repository"
)

type ChatService struct {
	ChatRepo   *repository.ChatRepository
	TicketRepo *repository.TicketRepository
}

func NewChatService(
	chatRepo *repository.ChatRepository,
	ticketRepo *repository.TicketRepository,
) *ChatService {
	return &ChatService{
		ChatRepo:   chatRepo,
		TicketRepo: ticketRepo,
	}
}

//
// =======================
// SEND MESSAGE
// =======================
//
func (s *ChatService) SendMessage(
	ticketID uint,
	senderID uint,
	role string,
	message string,
) (*domain.Chat, error) {

	// 1. Ambil tiket
	ticket, err := s.TicketRepo.GetByID(ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// 2. AUTHORIZATION
	switch role {

	case domain.RoleUser:
		// User hanya boleh chat di tiket miliknya
		if ticket.UserID != senderID {
			return nil, errors.New("access denied: you do not own this ticket")
		}

	case domain.RoleCS:
		// CS boleh chat selama tiket masih aktif
		if ticket.Status == "CLOSED" || ticket.Status == "LOCKED" {
			return nil, errors.New("cannot chat on closed or locked tickets")
		}

	default:
		return nil, errors.New("invalid role")
	}

	// 3. Simpan chat
	chat := &domain.Chat{
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderRole: role,
		Message:    message,
	}

	if err := s.ChatRepo.CreateChat(chat); err != nil {
		return nil, err
	}

	return chat, nil
}

//
// =======================
// GET CHAT HISTORY
// =======================
//
func (s *ChatService) GetHistory(
	ticketID uint,
	requestorID uint,
	role string,
) ([]domain.Chat, error) {

	// 1. Ambil tiket
	ticket, err := s.TicketRepo.GetByID(ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// 2. AUTHORIZATION (HARUS SAMA DENGAN SEND)
	switch role {

	case domain.RoleUser:
		// User hanya boleh lihat chat tiket miliknya
		if ticket.UserID != requestorID {
			return nil, errors.New("access denied")
		}

	case domain.RoleCS:
		// CS BOLEH SELALU LIHAT CHAT SELAMA TIKET ADA
		// (tidak tergantung privilege reset password, dll)
		// optional: bisa tambahin cek assignment di sini

	default:
		return nil, errors.New("invalid role")
	}

	// 3. Ambil history chat
	return s.ChatRepo.GetChatHistory(ticketID)
}
