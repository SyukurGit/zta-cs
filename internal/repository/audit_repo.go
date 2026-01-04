package repository

import (
	"github.com/syukurgit/zta/internal/domain"
	"gorm.io/gorm"
)

type AuditRepository struct {
	DB *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{DB: db}
}

// CreateLog menyimpan jejak aktivitas (Immutable / Gak bisa diedit)
func (r *AuditRepository) CreateLog(log *domain.AuditLog) error {
	return r.DB.Create(log).Error
}

// GetAllLogs mengambil semua log untuk dashboard Auditor
func (r *AuditRepository) GetAllLogs() ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	// Urutkan dari yang terbaru
	err := r.DB.Order("timestamp desc").Find(&logs).Error
	return logs, err
}



// internal/repository/audit_repo.go

// internal/repository/audit_repo.go

func (r *AuditRepository) GetAuditReports() ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	// Mengambil daftar tiket yang memiliki log audit
	err := r.DB.Preload("User").
		Joins("JOIN audit_logs ON audit_logs.ticket_id = tickets.id").
		Group("tickets.id").
		Order("tickets.updated_at desc").
		Find(&tickets).Error
	return tickets, err
}

func (r *AuditRepository) GetLogsByTicket(ticketID uint) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.DB.Where("ticket_id = ?", ticketID).Order("timestamp asc").Find(&logs).Error
	return logs, err
}