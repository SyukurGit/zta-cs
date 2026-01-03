package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"github.com/syukurgit/zta/internal/domain"
)

type VerificationRepository struct {
	DB *gorm.DB
}

func NewVerificationRepository(db *gorm.DB) *VerificationRepository {
	return &VerificationRepository{DB: db}
}

// GetUserRiskScore mengambil data user untuk pengecekan keamanan
func (r *VerificationRepository) GetUserByTicket(ticketID uint) (*domain.User, error) {
	var ticket domain.Ticket
	if err := r.DB.Preload("User").First(&ticket, ticketID).Error; err != nil {
		return nil, err
	}
	return &ticket.User, nil
}

// CountRecentSessions mengecek berapa kali user diverifikasi hari ini (Anti-Brute Force)
func (r *VerificationRepository) CountRecentSessions(userID uint) (int64, error) {
	var count int64
	err := r.DB.Model(&domain.VerificationSession{}).
		Where("user_id = ? AND created_at > ?", userID, time.Now().Add(-24*time.Hour)).
		Count(&count).Error
	return count, err
}

// GetRandomQuestions memilih 1 pertanyaan dari setiap kategori
func (r *VerificationRepository) GetSecureQuestionSet() ([]domain.VerificationQuestion, error) {
	var questions []domain.VerificationQuestion
	
	// Ambil 1 dari STATIC, 1 dari HISTORY, 1 dari USAGE
	// Menggunakan Raw SQL untuk random (MySQL specific: ORDER BY RAND())
	categories := []string{"STATIC", "HISTORY", "USAGE"}
	
	for _, cat := range categories {
		var q domain.VerificationQuestion
		// Hati-hati: ORDER BY RAND() lambat untuk data jutaan, tapi oke untuk ratusan soal.
		result := r.DB.Where("category = ?", cat).Order("RAND()").First(&q)
		if result.Error == nil {
			questions = append(questions, q)
		}
	}

	if len(questions) < 3 {
		return nil, errors.New("not enough questions in database")
	}
	return questions, nil
}

// CreateSession menyimpan sesi DAN pertanyaan yang terpilih ke database
func (r *VerificationRepository) CreateSession(session *domain.VerificationSession, questions []domain.VerificationQuestion) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Simpan Sesi
		if err := tx.Create(session).Error; err != nil {
			return err
		}

		// 2. Simpan Slot Pertanyaan (Attempts) untuk sesi ini
		for _, q := range questions {
			attempt := domain.VerificationAttempt{ // Pastikan struct ini ada di domain (lihat catatan bawah)
				SessionID:  session.ID,
				QuestionID: q.ID,
				IsCorrect:  false, // Default
			}
			if err := tx.Create(&attempt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}


// GetSessionByID mengambil data sesi beserta User-nya (untuk cek risk score/email)
// internal/repository/verification_repo.go

func (r *VerificationRepository) GetSessionByID(sessionID string) (*domain.VerificationSession, error) {
    var session domain.VerificationSession
    // Error "unsupported relations" muncul di sini jika struct di atas tidak punya field User
    if err := r.DB.Preload("User").First(&session, "id = ?", sessionID).Error; err != nil {
        return nil, err
    }
    return &session, nil
}

// GetQuestionsBySession mengambil daftar pertanyaan yang SUDAH dipilihkan untuk sesi ini
func (r *VerificationRepository) GetQuestionsBySession(sessionID string) ([]domain.VerificationQuestion, error) {
	var attempts []domain.VerificationAttempt
	var questions []domain.VerificationQuestion

	// 1. Ambil daftar ID pertanyaan dari tabel attempts
	if err := r.DB.Where("session_id = ?", sessionID).Find(&attempts).Error; err != nil {
		return nil, err
	}

	// 2. Ambil detail pertanyaan berdasarkan ID tersebut
	questionIDs := make([]uint, len(attempts))
	for i, a := range attempts {
		questionIDs[i] = a.QuestionID
	}

	if err := r.DB.Where("id IN ?", questionIDs).Find(&questions).Error; err != nil {
		return nil, err
	}

	return questions, nil
}

// UpdateSessionResult menyimpan hasil akhir: Status Sesi & Log Attempt
func (r *VerificationRepository) UpdateSessionResult(sessionID string, status string, riskIncrement int) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Update Status Sesi (PASSED / FAILED)
		if err := tx.Model(&domain.VerificationSession{}).Where("id = ?", sessionID).Update("status", status).Error; err != nil {
			return err
		}

		// 2. Jika Gagal, naikkan Risk Score User
		if status == "FAILED" && riskIncrement > 0 {
			// Kita butuh UserID, ambil dari sesi
			var session domain.VerificationSession
			tx.First(&session, "id = ?", sessionID)
			
			if err := tx.Model(&domain.User{}).Where("id = ?", session.UserID).
				Update("risk_score", gorm.Expr("risk_score + ?", riskIncrement)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// SavePrivilege (JIT) memberikan hak akses sementara ke CS
func (r *VerificationRepository) SavePrivilege(privilege *domain.TemporaryPrivilege) error {
	return r.DB.Create(privilege).Error
}

// GetCSByTicket Helper untuk mencari siapa CS yang memegang tiket ini
func (r *VerificationRepository) GetCSByTicket(ticketID uint) (uint, error) {
	var assignment domain.TicketAssignment
	err := r.DB.Where("ticket_id = ?", ticketID).First(&assignment).Error
	return assignment.CSID, err
}

