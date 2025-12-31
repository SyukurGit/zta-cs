package service

import (
	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/internal/repository"
)

type TicketService struct {
	Repo *repository.TicketRepository
}

func NewTicketService(repo *repository.TicketRepository) *TicketService {
	return &TicketService{Repo: repo}
}

func (s *TicketService) CreateTicket(userID uint, subject string) (*domain.Ticket, error) {
	ticket := &domain.Ticket{
		UserID:  userID,
		Subject: subject,
		Status:  "OPEN",
	}
	err := s.Repo.Create(ticket)
	return ticket, err
}

func (s *TicketService) GetOpenQueue() ([]domain.Ticket, error) {
	return s.Repo.GetOpenTickets()
}

func (s *TicketService) ClaimTicket(csID, ticketID uint) error {
	// Di sinilah nanti kita bisa tambahkan logika "Max 1 active ticket per CS" jika mau
	return s.Repo.AssignTicketToCS(ticketID, csID)
}