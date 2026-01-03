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

// StartVerification: Memulai sesi dan mengirim link
func (s *VerificationService) StartVerification(ticketID uint, csID uint) (string, error) {
	// 1. Ambil Data User Target
	user, err := s.Repo.GetUserByTicket(ticketID)
	if err != nil {
		return "", errors.New("ticket or user not found")
	}

	// 2. POLICY CHECK: Risk Score
	if user.RiskScore >= 80 {
		s.AuditSvc.LogActivity(
			ticketID, // TicketID (Updated Signature)
			csID,
			"CS",
			"START_VERIFICATION",
			"DENIED",
			fmt.Sprintf("RiskScore: %d", user.RiskScore),
		)
		return "", errors.New("security alert: account is high risk. escalation required")
	}

	// 3. POLICY CHECK: Rate Limit
	count, _ := s.Repo.CountRecentSessions(user.ID)
	if count >= 200 {
		s.AuditSvc.LogActivity(
			ticketID,
			csID,
			"CS",
			"START_VERIFICATION",
			"DENIED",
			"Reason: Rate Limit Exceeded",
		)
		return "", errors.New("limit exceeded: too many verification attempts today")
	}

	// 4. Generate Session ID
	sessionID := uuid.New().String()

	// 5. Pilih Pertanyaan
	questions, err := s.Repo.GetSecureQuestionSet()
	if err != nil {
		return "", errors.New("system error: failed to generate question set")
	}

	// 6. Buat Session
	session := &domain.VerificationSession{
		ID:           sessionID,
		TicketID:     ticketID,
		UserID:       user.ID,
		Status:       "PENDING",
		AttemptCount: 0,
		ExpiresAt:    time.Now().Add(15 * time.Minute),
	}

	// 7. Simpan ke DB
	if err := s.Repo.CreateSession(session, questions); err != nil {
		return "", err
	}

	// 8. Audit Log
	s.AuditSvc.LogActivity(
		ticketID,
		csID,
		"CS",
		"START_VERIFICATION",
		"SUCCESS",
		fmt.Sprintf("Session Created: %s", sessionID),
	)

	// 9. BUILD VERIFICATION URL
	verificationURL := fmt.Sprintf(
		"http://localhost:3000/verify/%s",
		sessionID,
	)

	return verificationURL, nil
}

// GetVerificationQuestions dipanggil saat User membuka link
func (s *VerificationService) GetVerificationQuestions(sessionID string) ([]domain.VerificationQuestion, error) {
	session, err := s.Repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	// Cek: Apakah sesi sudah kadaluarsa atau sudah selesai?
	if time.Now().After(session.ExpiresAt) || session.Status != "PENDING" {
		return nil, errors.New("session expired, closed, or already processed")
	}

	// Ambil pertanyaan
	return s.Repo.GetQuestionsBySession(sessionID)
}

// SubmitAnswers dipanggil saat User mengirim jawaban (LOGIC 3 STRIKES)
func (s *VerificationService) SubmitAnswers(sessionID string, answers map[uint]string) (bool, error) {
	// 1. Ambil Session
	session, err := s.Repo.GetSessionByID(sessionID)
	if err != nil {
		return false, errors.New("invalid session")
	}

	if session.Status != "PENDING" {
		return false, errors.New("sesi sudah tidak aktif")
	}

	// 2. Ambil Kunci Jawaban
	questions, _ := s.Repo.GetQuestionsBySession(sessionID)
	allCorrect := true

	// 3. Periksa Jawaban
	for _, q := range questions {
		userAnswer, provided := answers[q.ID]
		if !provided {
			allCorrect = false
			break
		}
		if !utils.CheckPasswordHash(userAnswer, q.AnswerHash) {
			allCorrect = false
			break
		}
	}

	// 4. JIKA JAWABAN SALAH (Handle Attempt Count)
	if !allCorrect {
		session.AttemptCount++
		sisa := 3 - session.AttemptCount
		
		var msg string
		newStatus := "PENDING" // Default tetap pending jika masih ada sisa

		// Cek Sisa Percobaan
		if sisa <= 0 {
			msg = "Sesi ini ditutup, akses dikunci. User harus buat tiket baru."
			newStatus = "FAILED"
		} else {
			msg = fmt.Sprintf("User mengisi tapi salah. Sisa %d kali percobaan.", sisa)
		}

		// Update DB: AttemptCount & Status
		// Menggunakan map untuk update spesifik agar tidak menimpa field lain
		s.Repo.DB.Model(&session).Updates(map[string]interface{}{
			"attempt_count": session.AttemptCount,
			"status":        newStatus,
		})

		// Log Aktivitas Gagal
		s.AuditSvc.LogActivity(
			session.TicketID,
			session.UserID,
			"USER",
			"VERIFICATION_ATTEMPT",
			"FAILED",
			fmt.Sprintf("Session: %s, Attempt: %d, Result: %s", sessionID, session.AttemptCount, newStatus),
		)

		return false, errors.New(msg)
	}

	// 5. JIKA BERHASIL (SUCCESS): Berikan Privilege 'SEND_RESET_LINK' ke CS
	csID, err := s.Repo.GetCSByTicket(session.TicketID)
	if err != nil {
		return true, errors.New("system error: ticket is not assigned to any CS")
	}

	// Buat Privilege untuk CS agar bisa klik tombol "Send Reset Link"
	privilege := &domain.TemporaryPrivilege{
		CSID:      csID,
		TicketID:  session.TicketID,
		Action:    "SEND_RESET_LINK",
		Token:     utils.GenerateRandomToken(32),
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if err := s.Repo.SavePrivilege(privilege); err != nil {
		return true, err
	}

	// Tandai sesi lulus
	s.Repo.UpdateSessionResult(sessionID, "PASSED", 0)

	// LOG: Verification Passed
	s.AuditSvc.LogActivity(
		session.TicketID,
		session.UserID,
		"USER",
		"VERIFICATION_SUCCESS",
		"PASSED",
		"User berhasil menjawab pertanyaan. Akses dibuka untuk CS.",
	)

	return true, nil
}