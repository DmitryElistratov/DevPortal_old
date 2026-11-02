package service

import (
	"context"
	"fmt"
	"log"

	"github.com/company/release-portal/internal/integration/grafana"
)

type GrafanaService struct {
	grafanaClient *grafana.Client
}

func NewGrafanaService(grafanaClient *grafana.Client) *GrafanaService {
	return &GrafanaService{grafanaClient: grafanaClient}
}

// ListDashboards returns list of dashboards
func (s *GrafanaService) ListDashboards(ctx context.Context) ([]grafana.Dashboard, error) {
	if s.grafanaClient == nil {
		return nil, fmt.Errorf("grafana client not configured")
	}

	log.Printf("[Grafana] Listing dashboards")
	dashboards, err := s.grafanaClient.ListDashboards(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list dashboards: %w", err)
	}

	log.Printf("[Grafana] Found %d dashboards", len(dashboards))
	return dashboards, nil
}

// GetDashboard returns details of a specific dashboard
func (s *GrafanaService) GetDashboard(ctx context.Context, uid string) (map[string]interface{}, error) {
	if s.grafanaClient == nil {
		return nil, fmt.Errorf("grafana client not configured")
	}

	log.Printf("[Grafana] Getting dashboard: %s", uid)
	dashboard, err := s.grafanaClient.GetDashboard(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	log.Printf("[Grafana] Successfully retrieved dashboard: %s", uid)
	return dashboard, nil
}

// ListDatasources returns list of datasources
func (s *GrafanaService) ListDatasources(ctx context.Context) ([]grafana.Datasource, error) {
	if s.grafanaClient == nil {
		return nil, fmt.Errorf("grafana client not configured")
	}

	log.Printf("[Grafana] Listing datasources")
	datasources, err := s.grafanaClient.ListDatasources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasources: %w", err)
	}

	log.Printf("[Grafana] Found %d datasources", len(datasources))
	return datasources, nil
}

// QueryMetrics executes a PromQL query
func (s *GrafanaService) QueryMetrics(ctx context.Context, query string) (*grafana.MetricResponse, error) {
	if s.grafanaClient == nil {
		return nil, fmt.Errorf("grafana client not configured")
	}

	log.Printf("[Grafana] Executing query: %s", query)
	result, err := s.grafanaClient.QueryMetrics(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	log.Printf("[Grafana] Query executed successfully")
	return result, nil
}
