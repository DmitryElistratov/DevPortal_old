package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/company/release-portal/internal/integration/gitlab"
	"github.com/company/release-portal/internal/repository"
	"github.com/gin-gonic/gin"
)

// CIHandler — проксирование GitLab CI/CD (пайплайны, джобы, запуск, MR, ветки)
// через сохранённую интеграцию gitlab. Тесты и релиз-процессы используют эти
// данные для отображения статуса QA в портале.
type CIHandler struct {
	repo        *repository.IntegrationRepository
	serviceRepo *repository.ServiceRepository
	fallback    *gitlab.Client
}

func NewCIHandler(repo *repository.IntegrationRepository, serviceRepo *repository.ServiceRepository, fallback *gitlab.Client) *CIHandler {
	return &CIHandler{repo: repo, serviceRepo: serviceRepo, fallback: fallback}
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

// resolveProject определяет GitLab-проект по :serviceId (репозиторий сервиса)
// или по query-параметру project. Возвращает ID проекта в виде строки.
func (h *CIHandler) resolveProject(c *gin.Context) (string, error) {
	if p := strings.TrimSpace(c.Query("project")); p != "" {
		return p, nil
	}
	if sid := c.Param("serviceId"); sid != "" && h.serviceRepo != nil {
		svc, err := h.serviceRepo.GetByID(sid)
		if err != nil || svc == nil {
			return "", fmt.Errorf("service %q not found", sid)
		}
		repo := strings.TrimSpace(svc.Repository)
		if repo == "" {
			return "", fmt.Errorf("service %q has no git repository configured", svc.Name)
		}
		// Из URL репозитория вытаскиваем path_with_namespace:
		// https://gitlab.example.com/group/sub/project(.git) -> group/sub/project
		path := repo
		if u, err := url.Parse(repo); err == nil && u.Host != "" {
			path = strings.Trim(u.Path, "/")
		} else {
			path = strings.TrimPrefix(strings.TrimPrefix(path, "git@"), "")
			if i := strings.Index(path, ":"); i >= 0 && !strings.Contains(path[:i], "/") {
				path = path[i+1:]
			}
		}
		path = strings.TrimSuffix(path, ".git")
		if path == "" {
			return "", fmt.Errorf("cannot derive GitLab project from repository %q", repo)
		}
		return url.PathEscape(path), nil
	}
	return "", fmt.Errorf("project or serviceId is required")
}

// ListBranchesForService GET /api/services/:serviceId/branches
// Тянет реальные ветки из GitLab для репозитория сервиса. Используется UI
// при выборе ветки для релиза/деплоя.
func (h *CIHandler) ListBranchesForService(c *gin.Context) {
	client, _, _ := h.client()
	if client == nil {
		h.noIntegration(c)
		return
	}
	project, err := h.resolveProject(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	branches, err := client.ListBranches(context.Background(), project)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	type branchInfo struct {
		Name   string `json:"name"`
		Commit string `json:"commit"`
	}
	out := make([]branchInfo, 0, len(branches))
	for _, b := range branches {
		out = append(out, branchInfo{Name: b.Name, Commit: b.Commit.ID})
	}
	c.JSON(http.StatusOK, out)
}

// ListBranches GET /api/ci/branches?project=<id|path>|serviceId=<id>
func (h *CIHandler) ListBranches(c *gin.Context) { h.branchesOrTags(c, false) }

// ListTags GET /api/ci/tags?project=<id|path>|serviceId=<id>
func (h *CIHandler) ListTags(c *gin.Context) { h.branchesOrTags(c, true) }

func (h *CIHandler) branchesOrTags(c *gin.Context, tags bool) {
	client, _, _ := h.client()
	if client == nil {
		h.noIntegration(c)
		return
	}
	project, err := h.resolveProject(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	type refInfo struct {
		Name   string `json:"name"`
		Commit string `json:"commit"`
	}
	out := make([]refInfo, 0)
	if tags {
		list, err := client.ListTags(context.Background(), project)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		for _, t := range list {
			out = append(out, refInfo{Name: t.Name, Commit: t.Commit.ID})
		}
	} else {
		list, err := client.ListBranches(context.Background(), project)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		for _, b := range list {
			out = append(out, refInfo{Name: b.Name, Commit: b.Commit.ID})
		}
	}
	c.JSON(http.StatusOK, out)
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
