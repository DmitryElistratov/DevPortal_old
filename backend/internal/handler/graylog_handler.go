package handler

import (
	"net/http"
	"time"

	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type GraylogHandler struct {
	graylogService *service.GraylogService
}

func NewGraylogHandler(graylogService *service.GraylogService) *GraylogHandler {
	return &GraylogHandler{graylogService: graylogService}
}

// SearchLogs handles POST /api/graylog/search
func (h *GraylogHandler) SearchLogs(c *gin.Context) {
	var req struct {
		Query string    `json:"query" binding:"required"`
		From  time.Time `json:"from"`
		To    time.Time `json:"to"`
		Limit int       `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.From.IsZero() {
		req.From = time.Now().Add(-24 * time.Hour)
	}
	if req.To.IsZero() {
		req.To = time.Now()
	}
	if req.Limit == 0 {
		req.Limit = 100
	}

	result, err := h.graylogService.SearchLogs(c.Request.Context(), req.Query, req.From, req.To, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListStreams handles GET /api/graylog/streams
func (h *GraylogHandler) ListStreams(c *gin.Context) {
	streams, err := h.graylogService.ListStreams(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, streams)
}

// GetStream handles GET /api/graylog/streams/:id
func (h *GraylogHandler) GetStream(c *gin.Context) {
	streamID := c.Param("id")

	stream, err := h.graylogService.GetStream(c.Request.Context(), streamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stream)
}
