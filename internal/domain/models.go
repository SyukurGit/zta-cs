package domain

import (
	"time"
	// "gorm.io/gorm"
)



const (
	RoleUser    = "USER"
	RoleCS      = "CS"
	RoleAuditor = "AUDITOR"
)
// 1. User: Aktor dalam sistem (User Biasa, CS, Auditor)
type User struct {
	ID           uint   `gorm:"primaryKey"`
    // PERUBAHAN DI SINI: Tambahkan type:varchar(255)
	Email        string `gorm:"type:varchar(255);uniqueIndex;not null"` 
	PasswordHash string `gorm:"not null"` 
	Role         string `gorm:"type:enum('USER','CS','AUDITOR');not null"` 
	RiskScore    int    `gorm:"default:0"` 
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// 2. Ticket: Kasus support
type Ticket struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null"`
	Subject   string `gorm:"type:varchar(255);not null"`
	Status    string `gorm:"type:enum('OPEN','IN_PROGRESS','CLOSED','LOCKED');default:'OPEN'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	
	// Relation
	User User `gorm:"foreignKey:UserID"`
}

// 3. TicketAssignment: Menegakkan aturan "1 CS per Tiket"
type TicketAssignment struct {
	TicketID   uint `gorm:"primaryKey"`
	CSID       uint `gorm:"not null"`
	AssignedAt time.Time
	
	// Relations
	Ticket Ticket `gorm:"foreignKey:TicketID"`
	CS     User   `gorm:"foreignKey:CSID"`
}

// 4. VerificationSession: Sesi verifikasi yang dikendalikan sistem
// internal/domain/models.go

type VerificationSession struct {
    ID           string    `gorm:"primaryKey;type:varchar(64)"` // UUID
    TicketID     uint      `gorm:"not null"`
    UserID       uint      `gorm:"not null"`
    Status       string    `gorm:"type:enum('PENDING','PASSED','FAILED','EXPIRED');default:'PENDING'"`
    AttemptCount int       `gorm:"default:0"` // Kolom yang baru ditambahkan
    ExpiresAt    time.Time `gorm:"not null"`
    CreatedAt    time.Time
    
    // Hubungan ini WAJIB ada agar Preload("User") berfungsi
    User         User      `gorm:"foreignKey:UserID"` 
}

// 5. VerificationQuestion: Bank soal (kategori statis)
type VerificationQuestion struct {
	ID           uint   `gorm:"primaryKey"`
	Category     string `gorm:"type:enum('STATIC','HISTORY','USAGE');not null"`
	QuestionText string `gorm:"type:text;not null"`
	AnswerHash   string `gorm:"not null"` // Jawaban disimpan sebagai hash, bukan plaintext
}

// 6. TemporaryPrivilege: INTI dari Just-In-Time (JIT) Access
type TemporaryPrivilege struct {
	ID        uint      `gorm:"primaryKey"`
	CSID      uint      `gorm:"not null"`
	TicketID  uint      `gorm:"not null"`
	Action    string    `gorm:"type:varchar(50);not null"` // e.g., RESET_PASSWORD
	Token     string    `gorm:"type:varchar(255);not null"` // System Token
	GrantedAt time.Time 
	ExpiresAt time.Time `gorm:"not null"` // Privilege mati otomatis setelah waktu ini
	IsUsed    bool      `gorm:"default:false"` // One-time use only
}

// 7. AuditLog: Log Immutable untuk Auditor
// internal/domain/models.go
// internal/domain/models.go

type AuditLog struct {
	ID        uint      `gorm:"primaryKey"`
	TicketID  uint      `gorm:"index;not null"` // Link ke Tiket
	ActorHash string    `gorm:"not null"`       // ID CS yang disamarkan
	ActorRole string    `gorm:"not null"`
	Action    string    `gorm:"not null"`
	Result    string    `gorm:"not null"`       // SUCCESS / DENIED
	Context   string    `gorm:"type:text"`      // Detail aktivitas
	Timestamp time.Time `gorm:"autoCreateTime"`
	
	// Relation untuk mempermudah pengambilan data
	Ticket Ticket `gorm:"foreignKey:TicketID"`
}

type VerificationAttempt struct {
    ID         uint   `gorm:"primaryKey"`
    SessionID  string `gorm:"size:64;not null"`
    QuestionID uint   `gorm:"not null"`
    IsCorrect  bool   `gorm:"default:false"`
    AttemptedAt time.Time `gorm:"autoCreateTime"`
}

type AnswerInput struct {
    QuestionID uint   `json:"question_id"`
    Answer     string `json:"answer"`
}

type Chat struct {
	ID        uint      `gorm:"primaryKey"`
	TicketID  uint      `gorm:"not null;index"` // Relasi ke Tiket
	SenderID  uint      `gorm:"not null"`       // ID User atau ID CS
	SenderRole string   `gorm:"type:enum('USER','CS');not null"` // Siapa yang kirim?
	Message   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}


