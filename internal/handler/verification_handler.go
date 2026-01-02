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
	ticketIDStr := c.Param("id")
	ticketID, _ := strconv.Atoi(ticketIDStr)
	csID := c.GetUint("user_id")

	verificationURL, err := h.Service.StartVerification(uint(ticketID), csID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "PENDING",
		"verification_url": verificationURL,
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
		Answers map[string]string `json:"answers" binding:"required"`
	}

	// 1. Bind JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	// 2. Convert key string -> uint
	answers := make(map[uint]string)

	for k, v := range input.Answers {
		id, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
			return
		}
		answers[uint(id)] = v
	}

	// 3. Submit ke service
	passed, err := h.Service.SubmitAnswers(sessionID, answers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "System error processing answers",
		})
		return
	}

	// 4. Response akhir
	if passed {
		c.JSON(http.StatusOK, gin.H{
			"status":  "PASSED",
			"message": "Identity verified. Support agent has been notified.",
		})
	} else {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "FAILED",
			"message": "Verification failed. Access denied.",
		})
	}
}
