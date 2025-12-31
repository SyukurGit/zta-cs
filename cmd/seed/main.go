package main

import (
	"fmt"
	"log"
	"github.com/syukurgit/zta/config"
	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/pkg/utils"

	"gorm.io/gorm"
)

func main() {
	// 1. Connect DB
	config.ConnectDB()

	// 2. Seed Users
	seedUsers(config.DB)

	// 3. Seed Questions
	seedQuestions(config.DB)

	fmt.Println("🌱 Database seeding completed successfully!")
}

func seedUsers(db *gorm.DB) {
	// Password default untuk semua user: "password123"
	hashedPassword, _ := utils.HashPassword("password123")

	users := []domain.User{
		{
			Email:        "user@example.com",
			PasswordHash: hashedPassword,
			Role:         "USER",
			RiskScore:    10, // Low risk
		},
		{
			Email:        "cs@company.com",
			PasswordHash: hashedPassword,
			Role:         "CS",
			RiskScore:    0,
		},
		{
			Email:        "auditor@company.com",
			PasswordHash: hashedPassword,
			Role:         "AUDITOR",
			RiskScore:    0,
		},
	}

	for _, u := range users {
		// FirstOrCreate mencegah duplikasi data jika script dijalankan 2x
		if err := db.Where("email = ?", u.Email).FirstOrCreate(&u).Error; err != nil {
			log.Printf("Failed to seed user %s: %v", u.Email, err)
		} else {
			fmt.Printf("✅ User seeded: %s\n", u.Email)
		}
	}
}

func seedQuestions(db *gorm.DB) {
	// Pertanyaan ini nanti dipilih sistem secara acak berdasarkan kategori
	questions := []domain.VerificationQuestion{
		{
			Category:     "STATIC",
			QuestionText: "Apa 4 digit terakhir NIK Anda?",
			AnswerHash:   hashAnswer("1234"), // Simulasi jawaban benar: 1234
		},
		{
			Category:     "HISTORY",
			QuestionText: "Bulan apa Anda terakhir mengganti password?",
			AnswerHash:   hashAnswer("juni"), // Simulasi jawaban benar: juni
		},
		{
			Category:     "USAGE",
			QuestionText: "Perangkat apa yang Anda gunakan login kemarin?",
			AnswerHash:   hashAnswer("iphone"), // Simulasi jawaban benar: iphone
		},
	}

	for _, q := range questions {
		if err := db.Where("question_text = ?", q.QuestionText).FirstOrCreate(&q).Error; err != nil {
			log.Printf("Failed to seed question: %v", err)
		} else {
			fmt.Printf("✅ Question seeded: %s\n", q.Category)
		}
	}
}

// Helper kecil untuk seeder ini saja
func hashAnswer(ans string) string {
	h, _ := utils.HashPassword(ans)
	return h
}