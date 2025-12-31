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