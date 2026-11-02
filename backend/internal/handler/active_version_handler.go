package handler

import (
	"net/http"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type ActiveVersionHandler struct {
	service *service.ActiveVersionService
}

func NewActiveVersionHandler(svc *service.ActiveVersionService) *ActiveVersionHandler {
	return &ActiveVersionHandler{service: svc}
}

func (h *ActiveVersionHandler) Deploy(c *gin.Context) {
	var version model.ActiveVersion
	if err := c.ShouldBindJSON(&version); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Deploy(c.Request.Context(), &version, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, version)
}

func (h *ActiveVersionHandler) GetByService(c *gin.Context) {
	serviceID := c.Param("serviceId")
	environment := c.DefaultQuery("environment", "production")

	versions, err := h.service.GetByServiceID(serviceID, environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, versions)
}

func (h *ActiveVersionHandler) UpdateTraffic(c *gin.Context) {
	versionID := c.Param("versionId")
	var req struct {
		TrafficPercentage int `json:"trafficPercentage" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.UpdateTrafficPercentage(c.Request.Context(), versionID, req.TrafficPercentage, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "traffic updated"})
}

func (h *ActiveVersionHandler) Rollback(c *gin.Context) {
	versionID := c.Param("versionId")
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Rollback(c.Request.Context(), versionID, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "version rolled back"})
}

func (h *ActiveVersionHandler) Promote(c *gin.Context) {
	versionID := c.Param("versionId")
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Promote(c.Request.Context(), versionID, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "version promoted"})
}
