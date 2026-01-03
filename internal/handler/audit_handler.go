package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/syukurgit/zta/internal/service"
)

type AuditHandler struct {
    Service *service.AuditService
}

func NewAuditHandler(s *service.AuditService) *AuditHandler {
    return &AuditHandler{Service: s}
}

func (h *AuditHandler) GetLogs(c *gin.Context) {
    logs, err := h.Service.GetAuditTrail()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
        return
    }
    
    // PERBAIKAN DI SINI:
    // Kirim logs langsung sebagai array, jangan dibungkus gin.H{"data": ...}
    // Ini agar konsisten dengan endpoint Ticket/Chat dan memudahkan Frontend axios.
    c.JSON(http.StatusOK, logs) 
}




func (h *AuditHandler) GetLogsByTicket(c *gin.Context) {
    ticketIDStr := c.Param("id")
    ticketID, _ := strconv.Atoi(ticketIDStr)

    // Langsung panggil repo via service (atau tambahkan method di service dulu idealnya, tapi ini shortcut aman)
    logs, err := h.Service.Repo.GetLogsByTicket(uint(ticketID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil log tiket"})
        return
    }
    c.JSON(http.StatusOK, logs)
}