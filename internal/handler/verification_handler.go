package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/syukurgit/zta/internal/service"
)

type VerificationHandler struct {
	Service *service.VerificationService
}

func NewVerificationHandler(s *service.VerificationService) *VerificationHandler {
	return &VerificationHandler{Service: s}
}

// StartVerification (CS Only)
func (h *VerificationHandler) StartVerification(c *gin.Context) {
	ticketIDStr := c.Param("id") // ID Tiket dari URL
	ticketID, _ := strconv.Atoi(ticketIDStr)
	csID := c.GetUint("user_id") // ID CS dari Token JWT

	err := h.Service.StartVerification(uint(ticketID), csID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Response untuk CS harus minim informasi
	c.JSON(http.StatusOK, gin.H{
		"message": "Verification session initialized. Secure link has been sent to the user's email.",
		"status":  "PENDING",
	})
}

// GetVerificationPage (Public)
func (h *VerificationHandler) GetVerificationPage(c *gin.Context) {
	sessionID := c.Param("token")
	
	questions, err := h.Service.GetVerificationQuestions(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Security: Hapus Hash Jawaban sebelum dikirim ke User!
	// Kita buat struct respons anonim agar bersih
	var response []gin.H
	for _, q := range questions {
		response = append(response, gin.H{
			"id":       q.ID,
			"category": q.Category,
			"question": q.QuestionText,
			// JANGAN KIRIM AnswerHash
		})
	}

	c.JSON(http.StatusOK, gin.H{"questions": response})
}

// SubmitVerification (Public)
func (h *VerificationHandler) SubmitVerification(c *gin.Context) {
	sessionID := c.Param("token")

	var input struct {
		Answers map[uint]string `json:"answers" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	passed, err := h.Service.SubmitAnswers(sessionID, input.Answers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System error processing answers"})
		return
	}

	if passed {
		c.JSON(http.StatusOK, gin.H{"status": "PASSED", "message": "Identity verified. Support agent has been notified."})
	} else {
		c.JSON(http.StatusForbidden, gin.H{"status": "FAILED", "message": "Verification failed. Access denied."})
	}
}