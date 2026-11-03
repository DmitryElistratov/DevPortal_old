package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type ReleaseHandler struct {
	service   *service.ReleaseService
	deploySvc *service.DeployService // strict GitOps deploy/rollback (git commit -> ArgoCD sync)
}

func NewReleaseHandler(svc *service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{service: svc}
}

// WithDeployService enables the GitOps deployment pipeline. When it is not
// set, deploys fall back to status bookkeeping only (legacy behaviour).
func (h *ReleaseHandler) WithDeployService(ds *service.DeployService) *ReleaseHandler {
	h.deploySvc = ds
	return h
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

	if err := h.service.Create(&rel, toString(userID), toString(userName)); err != nil {
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
	AppName     string `json:"appName,omitempty"`
}

// Deploy performs a strict GitOps deployment: the portal commits the desired
// state into the git ops repository (via GitLab API) and triggers an ArgoCD
// sync — it never applies anything to Kubernetes directly. Without the
// deploy service configured we only record status (legacy mode).
func (h *ReleaseHandler) Deploy(c *gin.Context) {
	id := c.Param("id")
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if h.deploySvc != nil {
		result, err := h.deploySvc.Deploy(c.Request.Context(), service.DeployRequest{
			ReleaseID:   id,
			Environment: req.Environment,
			AppName:     req.AppName,
		}, toString(userID), toString(userName))
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, service.ErrNoGitOps) || strings.Contains(err.Error(), "no ArgoCD application") {
				status = http.StatusUnprocessableEntity
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	if err := h.service.Deploy(id, req.Environment, toString(userID), toString(userName)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deployment initiated"})
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// Rollback reverts via GitOps tooling (ArgoCD rollback of the application),
// never by touching the cluster directly.
func (h *ReleaseHandler) Rollback(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if h.deploySvc != nil {
		result, err := h.deploySvc.Rollback(c.Request.Context(), id, toString(userID), toString(userName))
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, service.ErrNoGitOps) || strings.Contains(err.Error(), "no ArgoCD application") {
				status = http.StatusUnprocessableEntity
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	if err := h.service.Rollback(id, toString(userID), toString(userName)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rollback initiated"})
}

func (h *ReleaseHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	if err := h.service.Approve(id, toString(userID), toString(userName)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "approved"})
}
