package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengubah password biasa menjadi hash terenkripsi
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // Cost 14 is strong
	return string(bytes), err
}

// CheckPasswordHash membandingkan password input dengan hash di DB
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}