package handler

import (
	"net/http"

	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type KubernetesHandler struct {
	k8sService *service.KubernetesService
}

func NewKubernetesHandler(k8sService *service.KubernetesService) *KubernetesHandler {
	return &KubernetesHandler{k8sService: k8sService}
}

// ListPods handles GET /api/kubernetes/pods
func (h *KubernetesHandler) ListPods(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")

	pods, err := h.k8sService.ListPods(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pods)
}

// ListDeployments handles GET /api/kubernetes/deployments
func (h *KubernetesHandler) ListDeployments(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")

	deployments, err := h.k8sService.ListDeployments(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployments)
}

// ListServices handles GET /api/kubernetes/services
func (h *KubernetesHandler) ListServices(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")

	services, err := h.k8sService.ListServices(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, services)
}

// GetPodLogs handles GET /api/kubernetes/pods/:name/logs
func (h *KubernetesHandler) GetPodLogs(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")
	podName := c.Param("name")

	logs, err := h.k8sService.GetPodLogs(c.Request.Context(), namespace, podName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
