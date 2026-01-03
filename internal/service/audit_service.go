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

	// Cetak error di console kalau gagal simpan (buat debug)
	if err := s.Repo.CreateLog(log); err != nil {
		fmt.Printf("❌ Gagal simpan audit log: %v\n", err)
	}
}

func (s *AuditService) GetAuditTrail() ([]domain.AuditLog, error) {
	return s.Repo.GetAllLogs()
}