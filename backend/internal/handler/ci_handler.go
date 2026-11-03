package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/company/release-portal/internal/integration/gitlab"
	"github.com/company/release-portal/internal/repository"
	"github.com/gin-gonic/gin"
)

// CIHandler — проксирование GitLab CI/CD (пайплайны, джобы, запуск, MR) через
// сохранённую интеграцию gitlab. Тесты и релиз-процессы используют эти данные
// для отображения статуса QA в портале.
type CIHandler struct {
	repo     *repository.IntegrationRepository
	fallback *gitlab.Client
}

func NewCIHandler(repo *repository.IntegrationRepository, fallback *gitlab.Client) *CIHandler {
	return &CIHandler{repo: repo, fallback: fallback}
}

// client возвращает клиента из активной интеграции gitlab (свежие credentials),
// либо fallback-клиент, интегрированный при старте.
func (h *CIHandler) client() (*gitlab.Client, string, int) {
	integrations, err := h.repo.GetEnabledByType("gitlab")
	if err == nil && len(integrations) > 0 {
		it := integrations[0]
		token, ok := requireConfigString(it.Config, "token", "apiToken", "apiKey", "access_token")
		if ok {
			return gitlab.NewClient(normalizeIntegrationURL(it.URL), token), it.Name, http.StatusOK
		}
	}
	if h.fallback != nil {
		return h.fallback, "startup", http.StatusOK
	}
	return nil, "", http.StatusServiceUnavailable
}

func (h *CIHandler) noIntegration(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error": "GitLab integration is not configured or has no token. Add it in Settings → Integrations.",
	})
}

// ListPipelines GET /api/ci/pipelines?project=<id|path>&ref=<branch>
func (h *CIHandler) ListPipelines(c *gin.Context) {
	client, _, _ := h.client()
	if client == nil {
		h.noIntegration(c)
		return
	}
	project := c.Query("project")
	if project == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project query param is required (numeric ID or URL-encoded path)"})
		return
	}
	pipelines, err := client.ListPipelines(context.Background(), project, c.Query("ref"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pipelines)
}

// GetPipelineJobs GET /api/ci/pipelines/:pipelineId/jobs?project=<id>
func (h *CIHandler) GetPipelineJobs(c *gin.Context) {
	client, _, _ := h.client()
	if client == nil {
		h.noIntegration(c)
		return
	}
	project := c.Query("project")
	if project == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project query param is required"})
		return
	}
	id, err := strconv.Atoi(c.Param("pipelineId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pipeline id"})
		return
	}
	jobs, err := client.GetPipelineJobs(context.Background(), project, id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

type triggerPipelineRequest struct {
	Project   string            `json:"project"`
	Ref       string            `json:"ref"`
	Variables map[string]string `json:"variables"`
}

// TriggerPipeline POST /api/ci/pipelines
func (h *CIHandler) TriggerPipeline(c *gin.Context) {
	client, _, _ := h.client()
	if client == nil {
		h.noIntegration(c)
		return
	}
	var req triggerPipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Project = strings.TrimSpace(req.Project)
	if req.Project == "" || req.Ref == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project and ref are required"})
		return
	}
	pipeline, err := client.TriggerPipeline(context.Background(), req.Project, req.Ref, req.Variables)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pipeline)
}

// ListMergeRequests GET /api/ci/merge-requests?project=<id>&state=opened
func (h *CIHandler) ListMergeRequests(c *gin.Context) {
	client, _, _ := h.client()
	if client == nil {
		h.noIntegration(c)
		return
	}
	project := c.Query("project")
	if project == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project query param is required"})
		return
	}
	mrs, err := client.ListMergeRequests(context.Background(), project, c.Query("state"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mrs)
}
