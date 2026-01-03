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


// internal/service/ticket_service.go

func (s *TicketService) CloseTicket(ticketID uint, requestorID uint, role string) error {
	// 1. Ambil data tiket
	ticket, err := s.Repo.GetByID(ticketID)
	if err != nil {
		return errors.New("ticket not found")
	}

	// 2. Cek Otoritas
	if role == domain.RoleUser && ticket.UserID != requestorID {
		return errors.New("unauthorized: you don't own this ticket")
	}
    
    // Jika CS, pastikan dia yang memegang tiket ini (opsional tapi bagus untuk ZTA)
    if role == domain.RoleCS {
        // Logika tambahan bisa ditambahkan di sini untuk cek assignment
    }

	// 3. Update Status
	err = s.Repo.UpdateStatus(ticketID, "CLOSED")
	if err == nil {
		// LOG: Audit Trail (Penting!)
		s.AuditSvc.LogActivity(requestorID, role, "CLOSE_TICKET", "SUCCESS", fmt.Sprintf("Ticket: %d", ticketID))
	}
	return err
}


// GetUserTickets mengambil semua tiket milik user tertentu
func (s *TicketService) GetUserTickets(userID uint) ([]domain.Ticket, error) {
    var tickets []domain.Ticket
    // Preload User supaya data user tidak null (opsional, tapi bagus)
    err := s.Repo.DB.Preload("User").Where("user_id = ?", userID).Order("created_at desc").Find(&tickets).Error
    return tickets, err
}

// GetCSActiveTickets mengambil tiket yang sedang dikerjakan CS tertentu
// GetCSActiveTickets mengambil tiket yang sedang dikerjakan CS tertentu
func (s *TicketService) GetCSActiveTickets(csID uint) ([]domain.Ticket, error) {
    var tickets []domain.Ticket
    
    // PERBAIKAN: Gunakan JOIN ke tabel ticket_assignments
    // Karena AssignTicketToCS menyimpan relasi di tabel assignments, bukan di kolom tickets.cs_id
    err := s.Repo.DB.Preload("User").
        Joins("JOIN ticket_assignments ON ticket_assignments.ticket_id = tickets.id").
        Where("ticket_assignments.cs_id = ? AND tickets.status = ?", csID, "IN_PROGRESS").
        Order("tickets.updated_at desc").
        Find(&tickets).Error
        
    return tickets, err
}

// internal/service/ticket_service.go

// GetCSHistory mengambil tiket yang SUDAH diselesaikan (CLOSED) oleh CS tertentu
func (s *TicketService) GetCSHistory(csID uint) ([]domain.Ticket, error) {
    var tickets []domain.Ticket
    
    // Logic: Ambil tiket yang statusnya CLOSED dan pernah di-assign ke CS ini
    err := s.Repo.DB.Preload("User").
        Joins("JOIN ticket_assignments ON ticket_assignments.ticket_id = tickets.id").
        Where("ticket_assignments.cs_id = ? AND tickets.status = ?", csID, "CLOSED").
        Order("tickets.updated_at desc").
        Find(&tickets).Error
        
    return tickets, err
}