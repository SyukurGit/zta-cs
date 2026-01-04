// internal/handler/audit_handler.go

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

// GetLogs: Mengambil semua log mentah (untuk dashboard lama)
func (h *AuditHandler) GetLogs(c *gin.Context) {
	logs, err := h.Service.GetAuditTrail()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// GetAuditReports: Daftar laporan per tiket (Fungsi yang tadi undefined)
func (h *AuditHandler) GetAuditReports(c *gin.Context) {
	reports, err := h.Service.Repo.GetAuditReports() // Panggil langsung via repo atau service
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan audit"})
		return
	}
	c.JSON(http.StatusOK, reports)
}

// GetLogsByTicket: Timeline detail log per tiket
func (h *AuditHandler) GetLogsByTicket(c *gin.Context) {
	ticketIDStr := c.Param("id")
	ticketID, _ := strconv.Atoi(ticketIDStr)

	logs, err := h.Service.Repo.GetLogsByTicket(uint(ticketID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil timeline log"})
		return
	}
	c.JSON(http.StatusOK, logs)
}