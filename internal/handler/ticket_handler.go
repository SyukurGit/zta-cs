package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/syukurgit/zta/internal/service"
)

type TicketHandler struct {
	Service *service.TicketService
}

func NewTicketHandler(s *service.TicketService) *TicketHandler {
	return &TicketHandler{Service: s}
}

// CreateTicket (USER Only)
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	// Ambil UserID dari Context (hasil AuthMiddleware)
	userID := c.GetUint("user_id")

	var input struct {
		Subject string `json:"subject" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.Service.CreateTicket(userID, input.Subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

// GetOpenTickets (CS Only)
func (h *TicketHandler) GetOpenTickets(c *gin.Context) {
	tickets, err := h.Service.GetOpenQueue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tickets"})
		return
	}
	c.JSON(http.StatusOK, tickets)
}

// ClaimTicket (CS Only)
func (h *TicketHandler) ClaimTicket(c *gin.Context) {
	csID := c.GetUint("user_id")
	ticketIDStr := c.Param("id")
	ticketID, _ := strconv.Atoi(ticketIDStr)

	err := h.Service.ClaimTicket(csID, uint(ticketID))
	if err != nil {
		// Bisa jadi error karena tiket sudah diambil orang lain
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket successfully claimed. You may now proceed."})
}


// ResetPasswordAction (CS Only)
func (h *TicketHandler) ResetPasswordAction(c *gin.Context) {
    csID := c.GetUint("user_id")
    ticketIDStr := c.Param("id")
    ticketID, _ := strconv.Atoi(ticketIDStr)

    // Panggil service yang baru kita buat
    newPass, err := h.Service.ExecuteResetPassword(csID, uint(ticketID))
    if err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "SUCCESS",
        "message": "Temporary access granted and used successfully.",
        "new_user_password": newPass,
        "info": "Share this password securely with the user via phone.",
    })
}