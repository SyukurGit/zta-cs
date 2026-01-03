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




// internal/handler/ticket_handler.go



func (h *TicketHandler) CloseTicket(c *gin.Context) {
	ticketIDStr := c.Param("id")
	ticketID, _ := strconv.Atoi(ticketIDStr)
	requestorID := c.GetUint("user_id")
	role := c.GetString("role")

	err := h.Service.CloseTicket(uint(ticketID), requestorID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket successfully closed. Access revoked."})
}

// --- TAMBAHAN UNTUK USER DASHBOARD ---

// GetUserTickets (USER Only) - GET /api/user/tickets
func (h *TicketHandler) GetUserTickets(c *gin.Context) {
    userID := c.GetUint("user_id") // Dari Middleware JWT

    // Panggil Service (Nanti kita buat di bawah)
    tickets, err := h.Service.GetUserTickets(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data tiket"})
        return
    }

    c.JSON(http.StatusOK, tickets)
}

// --- TAMBAHAN UNTUK CS DASHBOARD (Tab "Tiket Saya") ---

// GetCSActiveTickets (CS Only) - GET /api/cs/tickets/mine
func (h *TicketHandler) GetCSActiveTickets(c *gin.Context) {
    csID := c.GetUint("user_id")

    // Ambil tiket yang statusnya IN_PROGRESS dan di-handle oleh CS ini
    tickets, err := h.Service.GetCSActiveTickets(csID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil tiket aktif"})
        return
    }

    c.JSON(http.StatusOK, tickets)
}

// Tambahkan di ticket_handler.go

// GetTicketDetail (Bisa dipakai User & CS asal Middleware cek akses)
func (h *TicketHandler) GetTicketDetail(c *gin.Context) {
    ticketIDStr := c.Param("id")
    ticketID, _ := strconv.Atoi(ticketIDStr)

    // Logika Service GetByID biasa
    ticket, err := h.Service.Repo.GetByID(uint(ticketID)) 
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
        return
    }
    
    // (Opsional) Cek otorisasi user/cs di sini jika perlu strict

    c.JSON(http.StatusOK, ticket)
}


// internal/handler/ticket_handler.go

// GetCSHistory (CS Only) - GET /api/cs/tickets/history
func (h *TicketHandler) GetCSHistory(c *gin.Context) {
    csID := c.GetUint("user_id")

    tickets, err := h.Service.GetCSHistory(csID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil riwayat tiket"})
        return
    }

    c.JSON(http.StatusOK, tickets)
}