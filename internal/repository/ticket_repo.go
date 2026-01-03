package repository

import (
	"errors"
	"github.com/syukurgit/zta/internal/domain"
	"time"

	"gorm.io/gorm"
)

type TicketRepository struct {
	DB *gorm.DB
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{DB: db}
}

// Create menyimpan tiket baru dari User
func (r *TicketRepository) Create(ticket *domain.Ticket) error {
	return r.DB.Create(ticket).Error
}

// GetOpenTickets mengambil semua tiket yang belum dikerjakan (untuk Queue CS)
func (r *TicketRepository) GetOpenTickets() ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	// Preload User agar CS tahu siapa yang lapor (tapi hanya email/ID)
	err := r.DB.Preload("User").Where("status = ?", "OPEN").Find(&tickets).Error
	return tickets, err
}

// GetByID mengambil detail tiket
func (r *TicketRepository) GetByID(id uint) (*domain.Ticket, error) {
	var ticket domain.Ticket
	err := r.DB.Preload("User").First(&ticket, id).Error
	return &ticket, err
}

// AssignTicketToCS menangani logika "Claim" dengan transaksi aman
func (r *TicketRepository) AssignTicketToCS(ticketID, csID uint) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Cek apakah tiket masih OPEN? (PENTING: Mencegah race condition)
		var ticket domain.Ticket
		if err := tx.Where("id = ? AND status = ?", ticketID, "OPEN").First(&ticket).Error; err != nil {
			return errors.New("ticket is not available or already taken")
		}

		// 2. Buat assignment record
		assignment := domain.TicketAssignment{
			TicketID:   ticketID,
			CSID:       csID,
			AssignedAt: time.Now(),
		}
		if err := tx.Create(&assignment).Error; err != nil {
			return err
		}

		// 3. Update status tiket jadi IN_PROGRESS
		if err := tx.Model(&ticket).Update("status", "IN_PROGRESS").Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *TicketRepository) CountActiveTicketsByCS(csID uint) (int64, error) {
	var count int64
	// Kita harus join tabel assignment dengan tiket untuk cek statusnya
	err := r.DB.Table("ticket_assignments").
		Joins("JOIN tickets ON tickets.id = ticket_assignments.ticket_id").
		Where("ticket_assignments.cs_id = ? AND tickets.status = ?", csID, "IN_PROGRESS").
		Count(&count).Error
	
	return count, err
}

// internal/repository/ticket_repo.go

func (r *TicketRepository) UpdateStatus(ticketID uint, status string) error {
	return r.DB.Model(&domain.Ticket{}).Where("id = ?", ticketID).Update("status", status).Error
}