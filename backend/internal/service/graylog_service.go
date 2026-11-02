package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/company/release-portal/internal/integration/graylog"
)

type GraylogService struct {
	graylogClient *graylog.Client
}

func NewGraylogService(graylogClient *graylog.Client) *GraylogService {
	return &GraylogService{graylogClient: graylogClient}
}

// SearchLogs searches for log messages
func (s *GraylogService) SearchLogs(ctx context.Context, query string, from, to time.Time, limit int) (*graylog.SearchResponse, error) {
	if s.graylogClient == nil {
		return nil, fmt.Errorf("graylog client not configured")
	}

	log.Printf("[Graylog] Searching logs with query: %s", query)
	result, err := s.graylogClient.SearchLogs(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search logs: %w", err)
	}

	log.Printf("[Graylog] Found %d log messages", result.Total)
	return result, nil
}

// ListStreams returns list of streams
func (s *GraylogService) ListStreams(ctx context.Context) ([]graylog.Stream, error) {
	if s.graylogClient == nil {
		return nil, fmt.Errorf("graylog client not configured")
	}

	log.Printf("[Graylog] Listing streams")
	streams, err := s.graylogClient.ListStreams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list streams: %w", err)
	}

	log.Printf("[Graylog] Found %d streams", len(streams))
	return streams, nil
}

// GetStream returns details of a specific stream
func (s *GraylogService) GetStream(ctx context.Context, streamID string) (*graylog.Stream, error) {
	if s.graylogClient == nil {
		return nil, fmt.Errorf("graylog client not configured")
	}

	log.Printf("[Graylog] Getting stream: %s", streamID)
	stream, err := s.graylogClient.GetStream(ctx, streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}

	log.Printf("[Graylog] Successfully retrieved stream: %s", streamID)
	return stream, nil
}
