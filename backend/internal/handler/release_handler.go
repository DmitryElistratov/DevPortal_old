package handler

import (
	"net/http"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type ReleaseHandler struct {
	service *service.ReleaseService
}

func NewReleaseHandler(svc *service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{service: svc}
}

func (h *ReleaseHandler) List(c *gin.Context) {
	serviceID := c.Query("serviceId")
	releases, err := h.service.List(serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, releases)
}

func (h *ReleaseHandler) Get(c *gin.Context) {
	id := c.Param("id")
	rel, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
		return
	}
	c.JSON(http.StatusOK, rel)
}

func (h *ReleaseHandler) Create(c *gin.Context) {
	var rel model.Release
	if err := c.ShouldBindJSON(&rel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Create(&rel, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rel)
}

func (h *ReleaseHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var rel model.Release
	if err := c.ShouldBindJSON(&rel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rel.ID = id

	if err := h.service.Update(&rel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rel)
}

type DeployRequest struct {
	Environment string `json:"environment" binding:"required"`
}

func (h *ReleaseHandler) Deploy(c *gin.Context) {
	id := c.Param("id")
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Deploy(id, req.Environment, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deployment initiated"})
}

func (h *ReleaseHandler) Rollback(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Rollback(id, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rollback initiated"})
}

func (h *ReleaseHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Approve(id, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "approved"})
}
