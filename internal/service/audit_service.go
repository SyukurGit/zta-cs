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
// internal/service/audit_service.go
func (s *AuditService) LogActivity(ticketID uint, actorID uint, role, action, result, contextData string) {
    actorHash := ""
    if role == domain.RoleCS {
        actorHash = utils.AnonymizeID(actorID)
    } else {
        actorHash = fmt.Sprintf("USER-%d", actorID)
    }

    log := &domain.AuditLog{
        TicketID:  ticketID, // Simpan ID Tiket
        ActorHash: actorHash,
        ActorRole: role,
        Action:    action,
        Result:    result,
        Context:   contextData,
    }
    _ = s.Repo.CreateLog(log)
}

// GetAuditTrail untuk dashboard Auditor
func (s *AuditService) GetAuditTrail() ([]domain.AuditLog, error) {
	return s.Repo.GetAllLogs()
}

