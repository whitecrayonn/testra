package automationhub

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/results"
)

type ResultsRepo interface {
	CreateRun(ctx context.Context, run *results.TestRun) error
	CreateItem(ctx context.Context, item *results.TestRunItem) error
	UpdateRun(ctx context.Context, run *results.TestRun) error
}

type Service struct {
	repo ResultsRepo
}

func NewService(repo ResultsRepo) *Service {
	return &Service{repo: repo}
}

type IngestInput struct {
	WorkspaceID uuid.UUID
	ProjectID   uuid.UUID
	SuiteID     *uuid.UUID
	Name        string
	Format      IngestionFormat
	Body        []byte
	CreatedBy   uuid.UUID
}

func (s *Service) Ingest(ctx context.Context, input IngestInput) (*IngestResult, error) {
	switch input.Format {
	case FormatJUnit:
		return s.ingestJUnit(ctx, input)
	case FormatPlaywright, FormatCypress:
		return s.ingestPlaywright(ctx, input)
	default:
		return nil, fmt.Errorf("unsupported format: %s", input.Format)
	}
}

func (s *Service) ingestJUnit(ctx context.Context, input IngestInput) (*IngestResult, error) {
	var suites JUnitTestSuites
	if err := xml.Unmarshal(input.Body, &suites); err != nil {
		return nil, fmt.Errorf("failed to parse JUnit XML: %w", err)
	}

	if len(suites.Suites) == 0 {
		return nil, fmt.Errorf("no test suites found in JUnit XML")
	}

	runID := uuid.New()
	now := nowUTC()

	total, passed, failed, skipped := 0, 0, 0, 0
	var totalDuration int64
	sortOrder := 0

	run := &results.TestRun{
		ID:          runID,
		WorkspaceID: input.WorkspaceID,
		ProjectID:   input.ProjectID,
		SuiteID:     input.SuiteID,
		Name:        input.Name,
		Status:      results.RunStatusRunning,
		Source:      results.RunSourceCI,
		CreatedBy:   input.CreatedBy,
		StartedAt:   &now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    map[string]interface{}{"format": string(input.Format)},
	}

	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	for _, suite := range suites.Suites {
		for _, tc := range suite.Cases {
			status := results.RunItemStatusPassed
			errMsg := ""
			stackTrace := ""

			if tc.Failure != nil {
				status = results.RunItemStatusFailed
				errMsg = tc.Failure.Message
				stackTrace = tc.Failure.Contents
				failed++
			} else if tc.Status == "skipped" || tc.Status == "disabled" {
				status = results.RunItemStatusSkipped
				skipped++
			} else {
				passed++
			}

			durMs := durationFromFloat(tc.Time)
			totalDuration += durMs
			total++

			item := &results.TestRunItem{
				ID:           uuid.New(),
				RunID:        runID,
				Title:        tc.Name,
				Status:       status,
				DurationMs:   durMs,
				ErrorMessage: errMsg,
				StackTrace:   stackTrace,
				SortOrder:    sortOrder,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := s.repo.CreateItem(ctx, item); err != nil {
				return nil, fmt.Errorf("failed to create run item: %w", err)
			}
			sortOrder++
		}
	}

	run.Total = total
	run.Passed = passed
	run.Failed = failed
	run.Skipped = skipped
	run.DurationMs = totalDuration
	run.Status = results.RunStatusPassed
	if failed > 0 {
		run.Status = results.RunStatusFailed
	}
	run.CompletedAt = &now
	run.UpdatedAt = now

	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to update run: %w", err)
	}

	return &IngestResult{
		RunID:      runID,
		Total:      total,
		Passed:     passed,
		Failed:     failed,
		Skipped:    skipped,
		DurationMs: totalDuration,
	}, nil
}

func (s *Service) ingestPlaywright(ctx context.Context, input IngestInput) (*IngestResult, error) {
	var report PlaywrightReport
	if err := json.Unmarshal(input.Body, &report); err != nil {
		return nil, fmt.Errorf("failed to parse Playwright/Cypress JSON: %w", err)
	}

	if len(report.Suites) == 0 {
		return nil, fmt.Errorf("no suites found in report")
	}

	runID := uuid.New()
	now := nowUTC()

	total, passed, failed, skipped := 0, 0, 0, 0
	var totalDuration int64
	sortOrder := 0

	run := &results.TestRun{
		ID:          runID,
		WorkspaceID: input.WorkspaceID,
		ProjectID:   input.ProjectID,
		SuiteID:     input.SuiteID,
		Name:        input.Name,
		Status:      results.RunStatusRunning,
		Source:      results.RunSourceCI,
		CreatedBy:   input.CreatedBy,
		StartedAt:   &now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    map[string]interface{}{"format": string(input.Format)},
	}

	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	for _, suite := range report.Suites {
		for _, test := range suite.Tests {
			status := results.RunItemStatusPassed
			errMsg := ""

			switch strings.ToLower(test.Status) {
			case "failed", "timedout":
				status = results.RunItemStatusFailed
				errMsg = test.Error
				failed++
			case "skipped":
				status = results.RunItemStatusSkipped
				skipped++
			default:
				passed++
			}

			totalDuration += test.Duration
			total++

			item := &results.TestRunItem{
				ID:           uuid.New(),
				RunID:        runID,
				Title:        test.Title,
				Status:       status,
				DurationMs:   test.Duration,
				ErrorMessage: errMsg,
				SortOrder:    sortOrder,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := s.repo.CreateItem(ctx, item); err != nil {
				return nil, fmt.Errorf("failed to create run item: %w", err)
			}
			sortOrder++
		}
	}

	run.Total = total
	run.Passed = passed
	run.Failed = failed
	run.Skipped = skipped
	run.DurationMs = totalDuration
	run.Status = results.RunStatusPassed
	if failed > 0 {
		run.Status = results.RunStatusFailed
	}
	run.CompletedAt = &now
	run.UpdatedAt = now

	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to update run: %w", err)
	}

	return &IngestResult{
		RunID:      runID,
		Total:      total,
		Passed:     passed,
		Failed:     failed,
		Skipped:    skipped,
		DurationMs: totalDuration,
	}, nil
}
