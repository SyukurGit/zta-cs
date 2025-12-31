package service

import (
	"fmt"
	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/internal/repository"
	"github.com/syukurgit/zta/pkg/utils"
)

type AuditService struct {
	Repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{Repo: repo}
}

// LogActivity adalah fungsi helper pusat untuk mencatat kegiatan
func (s *AuditService) LogActivity(actorID uint, role, action, result, contextData string) {
	// 1. Anonimalkan Identitas jika Aktornya adalah CS
	actorHash := ""
	if role == domain.RoleCS {
		actorHash = utils.AnonymizeID(actorID)
	} else {
		// Untuk User biasa atau System, boleh plain text atau hash juga (tergantung kebijakan)
		actorHash = fmt.Sprintf("USER-%d", actorID)
	}

	// 2. Buat Object Log
	log := &domain.AuditLog{
		ActorHash: actorHash,
		ActorRole: role,
		Action:    action,
		Result:    result,
		Context:   contextData,
	}

	// 3. Fire and Forget (Simpan ke DB)
	// Kita ignore error di sini agar logic utama tidak terganggu jika log gagal,
	// TAPI di sistem bank/finansial, ini harus blocking (wajib sukses).
	_ = s.Repo.CreateLog(log)
}

// GetAuditTrail untuk dashboard Auditor
func (s *AuditService) GetAuditTrail() ([]domain.AuditLog, error) {
	return s.Repo.GetAllLogs()
}