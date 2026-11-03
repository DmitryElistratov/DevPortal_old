package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/company/release-portal/internal/model"
)

// TestOpsRepository — доступ к таблицам test_runs, checklists, checklist_executions.
type TestOpsRepository struct {
	db *sql.DB
}

func NewTestOpsRepository(db *sql.DB) *TestOpsRepository {
	return &TestOpsRepository{db: db}
}

func newTestOpsID(prefix string) string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b)
}

// ---------- Test runs ----------

func (r *TestOpsRepository) CreateTestRun(run *model.TestRun) error {
	if run.ID == "" {
		run.ID = newTestOpsID("tr")
	}
	_, err := r.db.Exec(`INSERT INTO test_runs
		(id, service_id, release_id, environment, suite, status, total, passed, failed, skipped,
		 duration_ms, report_url, pipeline_ref, triggered_by, details, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW(),NOW())`,
		run.ID, run.ServiceID, run.ReleaseID, run.Environment, run.Suite, run.Status,
		run.Total, run.Passed, run.Failed, run.Skipped, run.DurationMs,
		run.ReportURL, run.PipelineRef, run.TriggeredBy, run.Details)
	return err
}

const testRunColumns = `id, service_id, release_id, environment, suite, status, total, passed, failed, skipped,
	duration_ms, report_url, pipeline_ref, triggered_by, details, created_at, updated_at`

func scanTestRun(row *sql.Row) (*model.TestRun, error) {
	var t model.TestRun
	err := row.Scan(&t.ID, &t.ServiceID, &t.ReleaseID, &t.Environment, &t.Suite, &t.Status,
		&t.Total, &t.Passed, &t.Failed, &t.Skipped, &t.DurationMs, &t.ReportURL,
		&t.PipelineRef, &t.TriggeredBy, &t.Details, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TestOpsRepository) GetTestRun(id string) (*model.TestRun, error) {
	row := r.db.QueryRow(`SELECT `+testRunColumns+` FROM test_runs WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanTestRun(row)
}

func (r *TestOpsRepository) ListTestRuns(releaseID, serviceID string) ([]model.TestRun, error) {
	query := `SELECT ` + testRunColumns + ` FROM test_runs WHERE deleted_at IS NULL`
	args := []interface{}{}
	argN := 1
	if releaseID != "" {
		query += fmt.Sprintf(" AND release_id = $%d", argN)
		args = append(args, releaseID)
		argN++
	}
	if serviceID != "" {
		query += fmt.Sprintf(" AND service_id = $%d", argN)
		args = append(args, serviceID)
		argN++
	}
	query += " ORDER BY created_at DESC LIMIT 200"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.TestRun
	for rows.Next() {
		var t model.TestRun
		if err := rows.Scan(&t.ID, &t.ServiceID, &t.ReleaseID, &t.Environment, &t.Suite, &t.Status,
			&t.Total, &t.Passed, &t.Failed, &t.Skipped, &t.DurationMs, &t.ReportURL,
			&t.PipelineRef, &t.TriggeredBy, &t.Details, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TestOpsRepository) UpdateTestRun(run *model.TestRun) error {
	_, err := r.db.Exec(`UPDATE test_runs SET status=$2, total=$3, passed=$4, failed=$5, skipped=$6,
		duration_ms=$7, report_url=$8, pipeline_ref=$9, details=$10, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`,
		run.ID, run.Status, run.Total, run.Passed, run.Failed, run.Skipped,
		run.DurationMs, run.ReportURL, run.PipelineRef, run.Details)
	return err
}

// SoftDeleteTestRun помечает прогон удалённым (физическое удаление запрещено политикой).
func (r *TestOpsRepository) SoftDeleteTestRun(id string) error {
	_, err := r.db.Exec(`UPDATE test_runs SET deleted_at = NOW(), status = 'aborted', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

// ---------- Checklists ----------

func (r *TestOpsRepository) ListChecklists() ([]model.Checklist, error) {
	rows, err := r.db.Query(`SELECT id, name, description, steps, is_default, created_by, created_at, updated_at
		FROM checklists WHERE deleted_at IS NULL ORDER BY is_default DESC, created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Checklist
	for rows.Next() {
		var c model.Checklist
		var stepsRaw []byte
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &stepsRaw, &c.IsDefault, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.Steps = []model.ChecklistStep{}
		_ = json.Unmarshal(stepsRaw, &c.Steps)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *TestOpsRepository) GetChecklist(id string) (*model.Checklist, error) {
	var c model.Checklist
	var stepsRaw []byte
	err := r.db.QueryRow(`SELECT id, name, description, steps, is_default, created_by, created_at, updated_at
		FROM checklists WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&c.ID, &c.Name, &c.Description, &stepsRaw, &c.IsDefault, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	c.Steps = []model.ChecklistStep{}
	_ = json.Unmarshal(stepsRaw, &c.Steps)
	return &c, nil
}

func (r *TestOpsRepository) CreateChecklist(c *model.Checklist) error {
	if c.ID == "" {
		c.ID = newTestOpsID("cl")
	}
	stepsJSON, _ := json.Marshal(c.Steps)
	_, err := r.db.Exec(`INSERT INTO checklists (id, name, description, steps, is_default, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4::jsonb,$5,$6,NOW(),NOW())`,
		c.ID, c.Name, c.Description, string(stepsJSON), c.IsDefault, c.CreatedBy)
	return err
}

func (r *TestOpsRepository) UpdateChecklist(c *model.Checklist) error {
	stepsJSON, _ := json.Marshal(c.Steps)
	_, err := r.db.Exec(`UPDATE checklists SET name=$2, description=$3, steps=$4::jsonb, is_default=$5, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`,
		c.ID, c.Name, c.Description, string(stepsJSON), c.IsDefault)
	return err
}

// SoftDeleteChecklist — только пометка удаления, без физического DELETE.
func (r *TestOpsRepository) SoftDeleteChecklist(id string) error {
	_, err := r.db.Exec(`UPDATE checklists SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

// ---------- Checklist executions ----------

func (r *TestOpsRepository) CreateExecution(e *model.ChecklistExecution) error {
	if e.ID == "" {
		e.ID = newTestOpsID("ce")
	}
	stepsJSON, _ := json.Marshal(e.Steps)
	_, err := r.db.Exec(`INSERT INTO checklist_executions (id, checklist_id, release_id, steps, completed, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4::jsonb,$5,$6,NOW(),NOW())`,
		e.ID, e.ChecklistID, e.ReleaseID, string(stepsJSON), e.Completed, e.CreatedBy)
	return err
}

func (r *TestOpsRepository) GetExecution(id string) (*model.ChecklistExecution, error) {
	var e model.ChecklistExecution
	var stepsRaw []byte
	err := r.db.QueryRow(`SELECT id, checklist_id, release_id, steps, completed, created_by, created_at, updated_at
		FROM checklist_executions WHERE id = $1`, id).
		Scan(&e.ID, &e.ChecklistID, &e.ReleaseID, &stepsRaw, &e.Completed, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	e.Steps = []model.ChecklistExecutionStep{}
	_ = json.Unmarshal(stepsRaw, &e.Steps)
	return &e, nil
}

func (r *TestOpsRepository) ListExecutionsByRelease(releaseID string) ([]model.ChecklistExecution, error) {
	rows, err := r.db.Query(`SELECT id, checklist_id, release_id, steps, completed, created_by, created_at, updated_at
		FROM checklist_executions WHERE release_id = $1 ORDER BY created_at DESC`, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ChecklistExecution
	for rows.Next() {
		var e model.ChecklistExecution
		var stepsRaw []byte
		if err := rows.Scan(&e.ID, &e.ChecklistID, &e.ReleaseID, &stepsRaw, &e.Completed, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		e.Steps = []model.ChecklistExecutionStep{}
		_ = json.Unmarshal(stepsRaw, &e.Steps)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *TestOpsRepository) UpdateExecution(e *model.ChecklistExecution) error {
	stepsJSON, _ := json.Marshal(e.Steps)
	completed := true
	for _, s := range e.Steps {
		if !s.Checked {
			completed = false
			break
		}
	}
	_, err := r.db.Exec(`UPDATE checklist_executions SET steps=$2::jsonb, completed=$3, updated_at=NOW() WHERE id=$1`,
		e.ID, string(stepsJSON), completed)
	e.Completed = completed
	return err
}
