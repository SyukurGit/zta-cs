package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/internal/repository"
	"github.com/syukurgit/zta/pkg/utils"
	"gorm.io/gorm"
)

type TicketService struct {
	Repo     *repository.TicketRepository
	AuditSvc *AuditService // Injeksi Audit Service
}

// Constructor diperbarui menerima AuditService
func NewTicketService(repo *repository.TicketRepository, auditSvc *AuditService) *TicketService {
	return &TicketService{Repo: repo, AuditSvc: auditSvc}
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
	// 1. POLICY CHECK: Max 1 Active Ticket per CS
	activeCount, err := s.Repo.CountActiveTicketsByCS(csID)
	if err != nil {
		return errors.New("system error: failed to check active tickets")
	}

	if activeCount > 0 {
		// LOG: Policy Violation
		s.AuditSvc.LogActivity(csID, "CS", "CLAIM_TICKET", "DENIED", fmt.Sprintf("Ticket: %d, Reason: Active Limit Reached", ticketID))
		return errors.New("policy violation: you have an active ticket. Please finish or close it first.")
	}

	// 2. Lanjutkan proses Claim
	err = s.Repo.AssignTicketToCS(ticketID, csID)
	if err == nil {
		// LOG: Success Claim
		s.AuditSvc.LogActivity(csID, "CS", "CLAIM_TICKET", "SUCCESS", fmt.Sprintf("Ticket: %d", ticketID))
	}
	return err
}

// ExecuteResetPassword: CS melakukan aksi sensitif menggunakan JIT Privilege
func (s *TicketService) ExecuteResetPassword(csID, ticketID uint) (string, error) {
	// 1. CEK PRIVILEGE (Zero Trust Core Logic)
	var privilege domain.TemporaryPrivilege
	err := s.Repo.DB.Where("cs_id = ? AND ticket_id = ? AND action = ? AND expires_at > ? AND is_used = ?",
		csID, ticketID, "RESET_PASSWORD", time.Now(), false).First(&privilege).Error

	if err != nil {
		// LOG: Access Denied (Critical)
		s.AuditSvc.LogActivity(csID, "CS", "RESET_PASSWORD", "DENIED", fmt.Sprintf("Ticket: %d, Reason: No Valid Privilege", ticketID))
		return "", errors.New("ACCESS DENIED: No active privilege found. User verification required.")
	}

	// 2. Ambil Data Tiket
	var ticket domain.Ticket
	if err := s.Repo.DB.First(&ticket, ticketID).Error; err != nil {
		return "", errors.New("ticket not found")
	}

	// 3. Generate Password Baru
	newPassword := utils.GenerateRandomToken(12)
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

	// LOG: Success Action (Critical)
	s.AuditSvc.LogActivity(csID, "CS", "RESET_PASSWORD", "SUCCESS", fmt.Sprintf("Ticket: %d", ticketID))

	return newPassword, nil
}