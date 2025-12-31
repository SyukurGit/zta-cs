package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/internal/repository"
	"github.com/syukurgit/zta/pkg/utils"
)

type VerificationService struct {
	Repo     *repository.VerificationRepository
	AuditSvc *AuditService // Injeksi Audit Service
}

// Constructor diperbarui menerima AuditService
func NewVerificationService(repo *repository.VerificationRepository, auditSvc *AuditService) *VerificationService {
	return &VerificationService{Repo: repo, AuditSvc: auditSvc}
}

func (s *VerificationService) StartVerification(ticketID uint, csID uint) error {
	// 1. Ambil Data User Target
	user, err := s.Repo.GetUserByTicket(ticketID)
	if err != nil {
		return errors.New("ticket or user not found")
	}

	// 2. POLICY CHECK: Risk Score
	if user.RiskScore >= 80 {
		// LOG: Security Alert
		s.AuditSvc.LogActivity(csID, "CS", "START_VERIFICATION", "DENIED", fmt.Sprintf("Ticket: %d, RiskScore: %d", ticketID, user.RiskScore))
		return errors.New("security alert: account is high risk. Escalation required.")
	}

	// 3. POLICY CHECK: Rate Limiting (Max 3/day)
	count, _ := s.Repo.CountRecentSessions(user.ID)
	if count >= 3 {
		// LOG: Rate Limit Hit
		s.AuditSvc.LogActivity(csID, "CS", "START_VERIFICATION", "DENIED", fmt.Sprintf("Ticket: %d, Reason: Rate Limit Exceeded", ticketID))
		return errors.New("limit exceeded: too many verification attempts today")
	}

	// 4. Generate Session ID (Secure Random UUID)
	sessionID := uuid.New().String()

	// 5. Pilih Pertanyaan (Sistem yang memilih, bukan CS)
	questions, err := s.Repo.GetSecureQuestionSet()
	if err != nil {
		return errors.New("system error: failed to generate question set")
	}

	// 6. Buat Objek Sesi
	session := &domain.VerificationSession{
		ID:        sessionID,
		TicketID:  ticketID,
		UserID:    user.ID,
		Status:    "PENDING",
		ExpiresAt: time.Now().Add(15 * time.Minute), // Expired dalam 15 menit
	}

	// 7. Simpan ke DB
	if err := s.Repo.CreateSession(session, questions); err != nil {
		return err
	}

	// LOG: Success Start
	s.AuditSvc.LogActivity(csID, "CS", "START_VERIFICATION", "SUCCESS", fmt.Sprintf("Ticket: %d, Session: %s", ticketID, sessionID))

	// 8. SIMULASI PENGIRIMAN LINK
	fmt.Printf("\n[EMAIL SERVICE] Sending to %s: Link: http://localhost:8080/verify/%s\n\n", user.Email, sessionID)

	return nil
}

// GetVerificationQuestions dipanggil saat User membuka link
func (s *VerificationService) GetVerificationQuestions(sessionID string) ([]domain.VerificationQuestion, error) {
	session, err := s.Repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	// Cek: Apakah sesi sudah kadaluarsa atau sudah selesai?
	if time.Now().After(session.ExpiresAt) || session.Status != "PENDING" {
		return nil, errors.New("session expired or already processed")
	}

	// Ambil pertanyaan (tapi nanti Handler harus pastikan jawaban/hash tidak dikirim ke JSON)
	return s.Repo.GetQuestionsBySession(sessionID)
}

// SubmitAnswers dipanggil saat User mengirim jawaban
func (s *VerificationService) SubmitAnswers(sessionID string, answers map[uint]string) (bool, error) {
	session, err := s.Repo.GetSessionByID(sessionID)
	if err != nil {
		return false, errors.New("invalid session")
	}

	if session.Status != "PENDING" {
		return false, errors.New("session is not pending")
	}

	// 1. Ambil Kunci Jawaban (Hash) dari DB
	questions, _ := s.Repo.GetQuestionsBySession(sessionID)

	allCorrect := true

	// 2. Periksa Jawaban Satu per Satu
	for _, q := range questions {
		userAnswer, provided := answers[q.ID]
		if !provided {
			allCorrect = false // Ada pertanyaan yang tidak dijawab
			break
		}

		// Bandingkan Jawaban User vs Hash di Database
		if !utils.CheckPasswordHash(userAnswer, q.AnswerHash) {
			allCorrect = false
			break // Salah satu saja, langsung FAIL
		}
	}

	// 3. DECISION MAKING (The Zero Trust Guard)
	if !allCorrect {
		// HUKUMAN: Risk Score +20
		s.Repo.UpdateSessionResult(sessionID, "FAILED", 20)
		
		// LOG: Verification Failed
		s.AuditSvc.LogActivity(session.UserID, "USER", "VERIFY_ANSWERS", "FAILED", fmt.Sprintf("Session: %s", sessionID))
		
		return false, nil // Return false tapi tidak error (artinya proses valid, hasilnya fail)
	}

	// 4. HADIAH (SUCCESS): Grant JIT Privilege
	csID, err := s.Repo.GetCSByTicket(session.TicketID)
	if err != nil {
		return true, errors.New("system error: ticket is not assigned to any CS")
	}

	privilege := &domain.TemporaryPrivilege{
		CSID:      csID,
		TicketID:  session.TicketID,
		Action:    "RESET_PASSWORD",
		Token:     utils.GenerateRandomToken(32),
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if err := s.Repo.SavePrivilege(privilege); err != nil {
		return true, err
	}

	// Tandai sesi lulus
	s.Repo.UpdateSessionResult(sessionID, "PASSED", 0)

	// LOG: Verification Passed & Privilege Granted
	s.AuditSvc.LogActivity(session.UserID, "USER", "VERIFY_ANSWERS", "PASSED", fmt.Sprintf("Session: %s, Privilege Granted to CS", sessionID))

	return true, nil
}