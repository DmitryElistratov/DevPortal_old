package handler

import (
	"net/http"

	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type GrafanaHandler struct {
	grafanaService *service.GrafanaService
}

func NewGrafanaHandler(grafanaService *service.GrafanaService) *GrafanaHandler {
	return &GrafanaHandler{grafanaService: grafanaService}
}

// ListDashboards handles GET /api/grafana/dashboards
func (h *GrafanaHandler) ListDashboards(c *gin.Context) {
	dashboards, err := h.grafanaService.ListDashboards(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboards)
}

// GetDashboard handles GET /api/grafana/dashboards/:uid
func (h *GrafanaHandler) GetDashboard(c *gin.Context) {
	uid := c.Param("uid")

	dashboard, err := h.grafanaService.GetDashboard(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// ListDatasources handles GET /api/grafana/datasources
func (h *GrafanaHandler) ListDatasources(c *gin.Context) {
	datasources, err := h.grafanaService.ListDatasources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, datasources)
}

// QueryMetrics handles POST /api/grafana/query
func (h *GrafanaHandler) QueryMetrics(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.grafanaService.QueryMetrics(c.Request.Context(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
