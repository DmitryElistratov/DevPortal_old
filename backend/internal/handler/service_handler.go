package handler

import (
	"net/http"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	service *service.ServiceManagementService
}

func NewServiceHandler(svc *service.ServiceManagementService) *ServiceHandler {
	return &ServiceHandler{service: svc}
}

func (h *ServiceHandler) List(c *gin.Context) {
	services, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, services)
}

func (h *ServiceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	svc, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	c.JSON(http.StatusOK, svc)
}

func (h *ServiceHandler) Create(c *gin.Context) {
	var svc model.Service
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Create(&svc, userID.(string), userName.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, svc)
}

func (h *ServiceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var svc model.Service
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	svc.ID = id

	if err := h.service.Update(&svc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, svc)
}

func (h *ServiceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
