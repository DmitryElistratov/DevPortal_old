package handler

import (
	"net/http"
	"strings"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/repository"
	"github.com/gin-gonic/gin"
)

// TestOpsHandler — API для тест-ранов и QA чек-листов.
type TestOpsHandler struct {
	repo *repository.TestOpsRepository
}

func NewTestOpsHandler(repo *repository.TestOpsRepository) *TestOpsHandler {
	return &TestOpsHandler{repo: repo}
}

func currentUser(c *gin.Context) string {
	for _, key := range []string{"userName", "username"} {
		if v, ok := c.Get(key); ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				return s
			}
		}
	}
	return "system"
}

// ---------- Test runs ----------

func (h *TestOpsHandler) CreateTestRun(c *gin.Context) {
	var run model.TestRun
	if err := c.ShouldBindJSON(&run); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}
	if strings.TrimSpace(run.ServiceID) == "" && strings.TrimSpace(run.ReleaseID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "serviceId or releaseId is required"})
		return
	}
	if run.Status == "" {
		run.Status = "passed"
	}
	run.TriggeredBy = currentUser(c)
	if err := h.repo.CreateTestRun(&run); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, run)
}

func (h *TestOpsHandler) ListTestRuns(c *gin.Context) {
	runs, err := h.repo.ListTestRuns(c.Query("releaseId"), c.Query("serviceId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, runs)
}

func (h *TestOpsHandler) GetTestRun(c *gin.Context) {
	run, err := h.repo.GetTestRun(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test run not found"})
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *TestOpsHandler) UpdateTestRun(c *gin.Context) {
	existing, err := h.repo.GetTestRun(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test run not found"})
		return
	}
	var patch model.TestRun
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if patch.Environment != "" {
		existing.Environment = patch.Environment
	}
	if patch.Suite != "" {
		existing.Suite = patch.Suite
	}
	if patch.Status != "" {
		existing.Status = patch.Status
	}
	existing.Total = patch.Total
	existing.Passed = patch.Passed
	existing.Failed = patch.Failed
	existing.Skipped = patch.Skipped
	if patch.DurationMs != 0 {
		existing.DurationMs = patch.DurationMs
	}
	if patch.ReportURL != "" {
		existing.ReportURL = patch.ReportURL
	}
	if patch.Details != "" {
		existing.Details = patch.Details
	}
	if err := h.repo.UpdateTestRun(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DeleteTestRun — soft-delete (запись помечается удалённой, физически не удаляется).
func (h *TestOpsHandler) DeleteTestRun(c *gin.Context) {
	if err := h.repo.SoftDeleteTestRun(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "softDeleted": true})
}

// ---------- Checklists ----------

func (h *TestOpsHandler) ListChecklists(c *gin.Context) {
	items, err := h.repo.ListChecklists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *TestOpsHandler) GetChecklist(c *gin.Context) {
	item, err := h.repo.GetChecklist(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "checklist not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *TestOpsHandler) CreateChecklist(c *gin.Context) {
	var cl model.Checklist
	if err := c.ShouldBindJSON(&cl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(cl.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	cl.CreatedBy = currentUser(c)
	if err := h.repo.CreateChecklist(&cl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cl)
}

func (h *TestOpsHandler) UpdateChecklist(c *gin.Context) {
	existing, err := h.repo.GetChecklist(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "checklist not found"})
		return
	}
	var patch model.Checklist
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if patch.Name != "" {
		existing.Name = patch.Name
	}
	existing.Description = patch.Description
	if len(patch.Steps) > 0 {
		existing.Steps = patch.Steps
	}
	existing.IsDefault = patch.IsDefault
	if err := h.repo.UpdateChecklist(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DeleteChecklist — только soft-delete.
func (h *TestOpsHandler) DeleteChecklist(c *gin.Context) {
	if err := h.repo.SoftDeleteChecklist(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "softDeleted": true})
}

// ---------- Checklist executions (на релизах) ----------

type createExecutionRequest struct {
	ChecklistID string `json:"checklistId"`
	ReleaseID   string `json:"releaseId"`
}

func (h *TestOpsHandler) CreateExecution(c *gin.Context) {
	var req createExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ChecklistID == "" || req.ReleaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "checklistId and releaseId are required"})
		return
	}
	cl, err := h.repo.GetChecklist(req.ChecklistID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "checklist not found"})
		return
	}
	exec := &model.ChecklistExecution{
		ChecklistID: cl.ID,
		ReleaseID:   req.ReleaseID,
		CreatedBy:   currentUser(c),
	}
	for _, s := range cl.Steps {
		exec.Steps = append(exec.Steps, model.ChecklistExecutionStep{Title: s.Title})
	}
	if err := h.repo.CreateExecution(exec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, exec)
}

func (h *TestOpsHandler) ListExecutions(c *gin.Context) {
	releaseID := c.Query("releaseId")
	if releaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "releaseId query param is required"})
		return
	}
	items, err := h.repo.ListExecutionsByRelease(releaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

type updateExecutionRequest struct {
	Steps     []model.ChecklistExecutionStep `json:"steps"`
	Completed *bool                          `json:"completed"`
}

func (h *TestOpsHandler) UpdateExecution(c *gin.Context) {
	existing, err := h.repo.GetExecution(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}
	var req updateExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Steps) > 0 {
		user := currentUser(c)
		for i := range req.Steps {
			if req.Steps[i].Checked {
				if req.Steps[i].CheckedBy == "" {
					req.Steps[i].CheckedBy = user
				}
			} else {
				req.Steps[i].CheckedBy = ""
				req.Steps[i].CheckedAt = ""
			}
		}
		existing.Steps = req.Steps
	}
	if req.Completed != nil {
		existing.Completed = *req.Completed
	} else {
		done := true
		for _, s := range existing.Steps {
			if !s.Checked {
				done = false
				break
			}
		}
		existing.Completed = done
	}
	if err := h.repo.UpdateExecution(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}
