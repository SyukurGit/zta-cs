package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
	"github.com/syukurgit/zta/internal/domain" // Import package domain yang baru dibuat
)

var DB *gorm.DB

func ConnectDB() {
    // Load .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )

    database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Connection Pooling (Penting untuk performa & stabilitas)
    sqlDB, _ := database.DB()
    
    // SetMaxIdleConns: Jumlah koneksi menganggur yang disimpan
    sqlDB.SetMaxIdleConns(10)
    
    // SetMaxOpenConns: Jumlah koneksi maksimum (mencegah DB overload)
    sqlDB.SetMaxOpenConns(100)

    DB = database
    fmt.Println("🚀 Database connected successfully (Zero Trust System Ready)")
}

func MigrateDB() {
	if DB == nil {
		log.Fatal("Database connection is not initialized")
	}

	fmt.Println("⏳ Running Database Migration...")

	err := DB.AutoMigrate(
		&domain.User{},
		&domain.Ticket{},
		&domain.TicketAssignment{},
		&domain.VerificationSession{},
		&domain.VerificationQuestion{},
		&domain.TemporaryPrivilege{},
		&domain.AuditLog{},
		&domain.VerificationAttempt{},
		&domain.VerificationSession{}, // Tabel anak
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("✅ Database Migration Completed Successfully!")
}