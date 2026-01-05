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

// LogActivity DIPERBARUI: Parameter pertama sekarang ticketID
// internal/service/audit_service.go

func (s *AuditService) LogActivity(ticketID uint, actorID uint, role, action, result, contextData string) {
	actorHash := ""
	if role == domain.RoleCS {
		// Anonymize CS ID menggunakan Hash
		actorHash = utils.AnonymizeID(actorID)
	} else {
		actorHash = fmt.Sprintf("USER-%d", actorID)
	}

	log := &domain.AuditLog{
		TicketID:  ticketID,
		ActorHash: actorHash,
		ActorRole: role,
		Action:    action,
		Result:    result,
		Context:   contextData,
	}
	_ = s.Repo.CreateLog(log)
}

func (s *AuditService) GetAuditTrail() ([]domain.AuditLog, error) {
	return s.Repo.GetAllLogs()
}