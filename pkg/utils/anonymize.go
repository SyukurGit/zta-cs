package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

// AnonymizeID mengubah ID integer menjadi Hash string yang konsisten tapi tidak bisa dibalik
func AnonymizeID(id uint) string {
	// Ambil secret key dari ENV untuk "garam" (Salt)
	secret := os.Getenv("SYSTEM_SECRET_KEY")
	
	// Gunakan HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("CS-ID-%d", id)))
	
	return hex.EncodeToString(h.Sum(nil))
}