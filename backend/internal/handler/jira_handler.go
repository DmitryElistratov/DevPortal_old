package handler

import (
	"net/http"

	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type JiraHandler struct {
	jiraService *service.JiraService
}

func NewJiraHandler(jiraService *service.JiraService) *JiraHandler {
	return &JiraHandler{jiraService: jiraService}
}

// GetIssue handles GET /api/jira/issues/:key
func (h *JiraHandler) GetIssue(c *gin.Context) {
	issueKey := c.Param("key")

	issue, err := h.jiraService.GetIssue(c.Request.Context(), issueKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, issue)
}

// SearchIssues handles POST /api/jira/issues/search
func (h *JiraHandler) SearchIssues(c *gin.Context) {
	var req struct {
		JQL string `json:"jql" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	issues, err := h.jiraService.SearchIssues(c.Request.Context(), req.JQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, issues)
}

// CreateIssue handles POST /api/jira/issues
func (h *JiraHandler) CreateIssue(c *gin.Context) {
	var req struct {
		ProjectKey  string `json:"projectKey" binding:"required"`
		Summary     string `json:"summary" binding:"required"`
		Description string `json:"description"`
		IssueType   string `json:"issueType" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	issue, err := h.jiraService.CreateIssue(
		c.Request.Context(),
		req.ProjectKey,
		req.Summary,
		req.Description,
		req.IssueType,
		userID.(string),
		userName.(string),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, issue)
}

// TransitionIssue handles POST /api/jira/issues/:key/transitions
func (h *JiraHandler) TransitionIssue(c *gin.Context) {
	issueKey := c.Param("key")

	var req struct {
		TransitionID string `json:"transitionId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	err := h.jiraService.TransitionIssue(
		c.Request.Context(),
		issueKey,
		req.TransitionID,
		userID.(string),
		userName.(string),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "issue transitioned successfully"})
}

// LinkIssueToRelease handles POST /api/jira/issues/:key/link
func (h *JiraHandler) LinkIssueToRelease(c *gin.Context) {
	issueKey := c.Param("key")

	var req struct {
		ReleaseID string `json:"releaseId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	err := h.jiraService.LinkIssueToRelease(
		c.Request.Context(),
		issueKey,
		req.ReleaseID,
		userID.(string),
		userName.(string),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "issue linked to release successfully"})
}
