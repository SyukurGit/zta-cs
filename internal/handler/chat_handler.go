package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/syukurgit/zta/internal/service"
)

type ChatHandler struct {
	Service *service.ChatService
}

func NewChatHandler(s *service.ChatService) *ChatHandler {
	return &ChatHandler{Service: s}
}

// SendChat menangani pengiriman pesan (User & CS pakai endpoint yang sama/mirip)
func (h *ChatHandler) SendChat(c *gin.Context) {
	ticketIDStr := c.Param("id")
	ticketID, _ := strconv.Atoi(ticketIDStr)
	
	senderID := c.GetUint("user_id")
	role := c.GetString("role") // Diambil dari JWT

	var input struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	chat, err := h.Service.SendMessage(uint(ticketID), senderID, role, input.Message)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chat)
}

// GetHistory mengambil riwayat chat
func (h *ChatHandler) GetHistory(c *gin.Context) {
	ticketIDStr := c.Param("id")
	ticketID, _ := strconv.Atoi(ticketIDStr)

	requestorID := c.GetUint("user_id")
	role := c.GetString("role")

	chats, err := h.Service.GetHistory(uint(ticketID), requestorID, role)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chats)
}