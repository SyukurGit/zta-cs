package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/internal/repository"
	"github.com/syukurgit/zta/pkg/utils"
)

type TicketService struct {
	Repo     *repository.TicketRepository
	AuditSvc *AuditService // Injeksi Audit Service
}

// NewTicketService: Constructor diperbarui menerima AuditService
func NewTicketService(repo *repository.TicketRepository, auditSvc *AuditService) *TicketService {
	return &TicketService{Repo: repo, AuditSvc: auditSvc}
}

// CreateTicket: User membuat tiket baru
func (s *TicketService) CreateTicket(userID uint, subject string) (*domain.Ticket, error) {
	ticket := &domain.Ticket{
		UserID:  userID,
		Subject: subject,
		Status:  "OPEN",
	}
	err := s.Repo.Create(ticket)
	return ticket, err
}

// GetOpenQueue: Mengambil tiket yang belum diambil CS
func (s *TicketService) GetOpenQueue() ([]domain.Ticket, error) {
	return s.Repo.GetOpenTickets()
}

// ClaimTicket: CS mengambil tiket dari antrian
func (s *TicketService) ClaimTicket(csID, ticketID uint) error {
	// 1. POLICY CHECK: Max 1 Active Ticket per CS
	activeCount, err := s.Repo.CountActiveTicketsByCS(csID)
	if err != nil {
		return errors.New("system error: failed to check active tickets")
	}

	if activeCount > 0 {
		// LOG: Policy Violation
		s.AuditSvc.LogActivity(
			ticketID, // TicketID
			csID,     // ActorID
			"CS",     // Role
			"CLAIM_TICKET",
			"DENIED",
			"Reason: Active Limit Reached (Max 1)",
		)
		return errors.New("policy violation: you have an active ticket. Please finish or close it first.")
	}

	// 2. Lanjutkan proses Claim
	err = s.Repo.AssignTicketToCS(ticketID, csID)
	if err == nil {
		// LOG: Success Claim
		s.AuditSvc.LogActivity(
			ticketID,
			csID,
			"CS",
			"CLAIM_TICKET",
			"SUCCESS",
			"CS claimed the ticket",
		)
	}
	return err
}

// ExecuteResetPassword: CS membuat LINK reset password (bukan mereset password langsung)
func (s *TicketService) ExecuteResetPassword(csID, ticketID uint) (string, error) {
	// 1. Cek Privilege 'SEND_RESET_LINK' (Diberikan oleh VerificationService jika lulus)
	var privilege domain.TemporaryPrivilege
	err := s.Repo.DB.Where("cs_id = ? AND ticket_id = ? AND action = ? AND expires_at > ? AND is_used = ?",
		csID, ticketID, "SEND_RESET_LINK", time.Now(), false).First(&privilege).Error

	if err != nil {
		s.AuditSvc.LogActivity(
			ticketID,
			csID,
			"CS",
			"GENERATE_RESET_LINK",
			"DENIED",
			"Reason: No Valid Privilege (User not verified)",
		)
		return "", errors.New("AKSES DITOLAK: User belum lulus verifikasi.")
	}

	// 2. Buat Token Rahasia untuk User (agar User bisa ganti password sendiri)
	userResetToken := utils.GenerateRandomToken(64)
	
	// Simpan token ini sebagai privilege milik SYSTEM/USER untuk nanti divalidasi saat submit password baru
	userPriv := &domain.TemporaryPrivilege{
		TicketID:  ticketID,
		CSID:      0, // 0 menandakan ini token milik User/System, bukan CS spesifik
		Action:    "USER_SET_PASSWORD",
		Token:     userResetToken,
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	
	if err := s.Repo.DB.Create(userPriv).Error; err != nil {
		return "", errors.New("failed to generate user token")
	}

	// 3. Hanguskan Privilege CS (One-time Use)
	s.Repo.DB.Model(&privilege).Update("is_used", true)

	// LOG: Success
	s.AuditSvc.LogActivity(
		ticketID,
		csID,
		"CS",
		"GENERATE_RESET_LINK",
		"SUCCESS",
		"Reset link generated for user",
	)

	// 4. Kembalikan Link untuk dikirim via Chat
	return fmt.Sprintf("http://localhost:3000/reset-password/%s", userResetToken), nil
}

// CloseTicket: Menutup tiket dan mencabut akses
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
	
	// (Opsional) Jika CS, bisa tambahkan cek apakah dia yang handle tiket ini

	// 3. Update Status
	err = s.Repo.UpdateStatus(ticketID, "CLOSED")
	if err == nil {
		// LOG: Audit Trail
		s.AuditSvc.LogActivity(
			ticketID,
			requestorID,
			role,
			"CLOSE_TICKET",
			"SUCCESS",
			"Ticket closed manually",
		)
	}
	return err
}

// GetUserTickets: Mengambil semua tiket milik user tertentu
func (s *TicketService) GetUserTickets(userID uint) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	err := s.Repo.DB.Preload("User").Where("user_id = ?", userID).Order("created_at desc").Find(&tickets).Error
	return tickets, err
}

// GetCSActiveTickets: Mengambil tiket yang sedang dikerjakan CS tertentu (IN_PROGRESS)
func (s *TicketService) GetCSActiveTickets(csID uint) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	
	// Gunakan JOIN ke tabel ticket_assignments
	err := s.Repo.DB.Preload("User").
		Joins("JOIN ticket_assignments ON ticket_assignments.ticket_id = tickets.id").
		Where("ticket_assignments.cs_id = ? AND tickets.status = ?", csID, "IN_PROGRESS").
		Order("tickets.updated_at desc").
		Find(&tickets).Error
		
	return tickets, err
}

// GetCSHistory: Mengambil tiket yang SUDAH diselesaikan (CLOSED) oleh CS tertentu
func (s *TicketService) GetCSHistory(csID uint) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	
	// Ambil tiket yang statusnya CLOSED dan pernah di-assign ke CS ini
	err := s.Repo.DB.Preload("User").
		Joins("JOIN ticket_assignments ON ticket_assignments.ticket_id = tickets.id").
		Where("ticket_assignments.cs_id = ? AND tickets.status = ?", csID, "CLOSED").
		Order("tickets.updated_at desc").
		Find(&tickets).Error
		
	return tickets, err
}