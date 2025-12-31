package service

import (
	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/internal/repository"
	"errors"
    "time"
    "github.com/syukurgit/zta/pkg/utils" // Pastikan import ini ada
    "gorm.io/gorm"
	
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

// ExecuteResetPassword: CS melakukan aksi sensitif menggunakan JIT Privilege
func (s *TicketService) ExecuteResetPassword(csID, ticketID uint) (string, error) {
    // 1. CEK PRIVILEGE (Zero Trust Core Logic)
    // Cari privilege yang valid: milik CS ini, untuk tiket ini, belum expired, dan belum dipakai.
    var privilege domain.TemporaryPrivilege
    err := s.Repo.DB.Where("cs_id = ? AND ticket_id = ? AND action = ? AND expires_at > ? AND is_used = ?", 
        csID, ticketID, "RESET_PASSWORD", time.Now(), false).First(&privilege).Error
    
    if err != nil {
        return "", errors.New("ACCESS DENIED: No active privilege found. User verification required.")
    }

    // 2. Ambil Data Tiket untuk tahu User-nya siapa
    var ticket domain.Ticket
    if err := s.Repo.DB.First(&ticket, ticketID).Error; err != nil {
        return "", errors.New("ticket not found")
    }

    // 3. Generate Password Baru (Random)
    newPassword := utils.GenerateRandomToken(12) // Password sementara
    hashedPassword, _ := utils.HashPassword(newPassword)

    // 4. Update Password & Hanguskan Privilege (Transaction)
    err = s.Repo.DB.Transaction(func(tx *gorm.DB) error {
        // Update password user
        if err := tx.Model(&domain.User{}).Where("id = ?", ticket.UserID).Update("password_hash", hashedPassword).Error; err != nil {
            return err
        }

        // Tandai privilege sudah dipakai (One-time Use)
        if err := tx.Model(&privilege).Update("is_used", true).Error; err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        return "", err
    }

    return newPassword, nil
}