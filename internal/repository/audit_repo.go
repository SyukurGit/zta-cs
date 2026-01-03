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

// --- TAMBAHAN BARU ---
func (r *AuditRepository) GetLogsByTicket(ticketID uint) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	// Ambil log khusus tiket ini, urutkan dari yang paling baru
	err := r.DB.Where("ticket_id = ?", ticketID).Order("timestamp desc").Find(&logs).Error
	return logs, err
}

