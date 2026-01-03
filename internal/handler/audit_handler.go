package handler

import (
    "net/http"
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