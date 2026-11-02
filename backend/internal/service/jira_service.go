package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/company/release-portal/internal/integration/jira"
	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/repository"
)

type JiraService struct {
	repo      *repository.ReleaseRepository
	auditRepo *repository.AuditRepository
	jiraClient *jira.Client
}

func NewJiraService(
	repo *repository.ReleaseRepository,
	auditRepo *repository.AuditRepository,
	jiraClient *jira.Client,
) *JiraService {
	return &JiraService{
		repo:       repo,
		auditRepo:  auditRepo,
		jiraClient: jiraClient,
	}
}

// GetIssue retrieves a Jira issue
func (s *JiraService) GetIssue(ctx context.Context, issueKey string) (*jira.Issue, error) {
	if s.jiraClient == nil {
		return nil, fmt.Errorf("jira client not configured")
	}

	log.Printf("[Jira] Getting issue: %s", issueKey)
	issue, err := s.jiraClient.GetIssue(ctx, issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue from Jira: %w", err)
	}

	log.Printf("[Jira] Successfully retrieved issue: %s", issueKey)
	return issue, nil
}

// SearchIssues searches for Jira issues using JQL
func (s *JiraService) SearchIssues(ctx context.Context, jql string) ([]jira.Issue, error) {
	if s.jiraClient == nil {
		return nil, fmt.Errorf("jira client not configured")
	}

	log.Printf("[Jira] Searching issues with JQL: %s", jql)
	issues, err := s.jiraClient.SearchIssues(ctx, jql)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues in Jira: %w", err)
	}

	log.Printf("[Jira] Found %d issues", len(issues))
	return issues, nil
}

// CreateIssue creates a new Jira issue
func (s *JiraService) CreateIssue(ctx context.Context, projectKey, summary, description, issueType, userID, userName string) (*jira.Issue, error) {
	if s.jiraClient == nil {
		return nil, fmt.Errorf("jira client not configured")
	}

	log.Printf("[Jira] Creating issue in project %s: %s", projectKey, summary)
	issue, err := s.jiraClient.CreateIssue(ctx, projectKey, summary, description, issueType)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue in Jira: %w", err)
	}

	// Audit log
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "create_jira_issue",
		Resource:   "jira_issue",
		ResourceID: issue.Key,
		Details:    fmt.Sprintf("Created Jira issue %s in project %s", issue.Key, projectKey),
	})

	log.Printf("[Jira] Successfully created issue: %s", issue.Key)
	return issue, nil
}

// TransitionIssue transitions a Jira issue to a new status
func (s *JiraService) TransitionIssue(ctx context.Context, issueKey, transitionID, userID, userName string) error {
	if s.jiraClient == nil {
		return fmt.Errorf("jira client not configured")
	}

	log.Printf("[Jira] Transitioning issue %s with transition %s", issueKey, transitionID)
	err := s.jiraClient.TransitionIssue(ctx, issueKey, transitionID)
	if err != nil {
		return fmt.Errorf("failed to transition issue in Jira: %w", err)
	}

	// Audit log
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "transition_jira_issue",
		Resource:   "jira_issue",
		ResourceID: issueKey,
		Details:    fmt.Sprintf("Transitioned Jira issue %s with transition %s", issueKey, transitionID),
	})

	log.Printf("[Jira] Successfully transitioned issue: %s", issueKey)
	return nil
}

// LinkIssueToRelease links a Jira issue to a release
func (s *JiraService) LinkIssueToRelease(ctx context.Context, issueKey, releaseID, userID, userName string) error {
	// Get release
	release, err := s.repo.GetByID(releaseID)
	if err != nil {
		return fmt.Errorf("release not found: %w", err)
	}

	// Get issue to verify it exists
	_, err = s.GetIssue(ctx, issueKey)
	if err != nil {
		return fmt.Errorf("issue not found: %w", err)
	}

	// Audit log
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "link_jira_issue",
		Resource:   "jira_issue",
		ResourceID: issueKey,
		Details:    fmt.Sprintf("Linked Jira issue %s to release %s (v%s)", issueKey, release.ID, release.Version),
	})

	log.Printf("[Jira] Successfully linked issue %s to release %s", issueKey, releaseID)
	return nil
}
