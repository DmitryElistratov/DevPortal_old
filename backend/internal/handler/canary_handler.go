package handler

import (
	"net/http"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type CanaryHandler struct {
	service *service.CanaryService
}

func NewCanaryHandler(svc *service.CanaryService) *CanaryHandler {
	return &CanaryHandler{service: svc}
}

func (h *CanaryHandler) Create(c *gin.Context) {
	var canary model.CanaryDeployment
	if err := c.ShouldBindJSON(&canary); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Create(c.Request.Context(), &canary, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, canary)
}

func (h *CanaryHandler) Get(c *gin.Context) {
	id := c.Param("id")
	canary, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "canary deployment not found"})
		return
	}
	c.JSON(http.StatusOK, canary)
}

func (h *CanaryHandler) GetByRelease(c *gin.Context) {
	releaseID := c.Param("releaseId")
	canary, err := h.service.GetByReleaseID(releaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "canary deployment not found"})
		return
	}
	c.JSON(http.StatusOK, canary)
}

func (h *CanaryHandler) GetActive(c *gin.Context) {
	serviceID := c.Param("serviceId")
	canary, err := h.service.GetActiveByServiceID(serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active canary deployment"})
		return
	}
	c.JSON(http.StatusOK, canary)
}

func (h *CanaryHandler) UpdatePercentage(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Percentage int `json:"percentage" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.UpdatePercentage(c.Request.Context(), id, req.Percentage, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "percentage updated"})
}

func (h *CanaryHandler) Promote(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Promote(c.Request.Context(), id, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "canary promoted"})
}

func (h *CanaryHandler) Rollback(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Rollback(c.Request.Context(), id, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "canary rolled back"})
}

func (h *CanaryHandler) List(c *gin.Context) {
	serviceID := c.Param("serviceId")
	canaries, err := h.service.ListByServiceID(serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, canaries)
}
