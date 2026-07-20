package automationhub

import (
	"time"

	"github.com/google/uuid"
)

type IngestionFormat string

const (
	FormatJUnit       IngestionFormat = "junit"
	FormatPytestJUnit IngestionFormat = "pytest-junit"
	FormatPlaywright  IngestionFormat = "playwright"
	FormatCypress     IngestionFormat = "cypress"
	FormatNewman      IngestionFormat = "newman"
	FormatRobot       IngestionFormat = "robot"
)

func IsValidFormat(s string) bool {
	switch IngestionFormat(s) {
	case FormatJUnit, FormatPytestJUnit, FormatPlaywright, FormatCypress, FormatNewman, FormatRobot:
		return true
	}
	return false
}

type IngestResult struct {
	ExecutionID uuid.UUID
	RunID       uuid.UUID
	Total       int
	Passed      int
	Failed      int
	Skipped     int
	DurationMs  int64
}

type AutomationProject struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	ProjectID     *uuid.UUID
	Name          string
	Framework     string
	RepositoryURL string
	Branch        string
	Command       string
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AutomationExecution struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	WorkspaceID  uuid.UUID
	TestRunID    *uuid.UUID
	Name         string
	Status       string
	ReportFormat string
	ReportPath   string
	RetryOf      *uuid.UUID
	DurationMs   int64
	Total        int
	Passed       int
	Failed       int
	Skipped      int
	Blocked      int
	CreatedBy    uuid.UUID
	TriggeredBy  uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ArtifactKind string

const (
	ArtifactKindReport     ArtifactKind = "report"
	ArtifactKindLog        ArtifactKind = "log"
	ArtifactKindScreenshot ArtifactKind = "screenshot"
	ArtifactKindArtifact   ArtifactKind = "artifact"
)

func IsValidArtifactKind(s string) bool {
	switch ArtifactKind(s) {
	case ArtifactKindReport, ArtifactKindLog, ArtifactKindScreenshot, ArtifactKindArtifact:
		return true
	}
	return false
}

type AutomationArtifact struct {
	ID            uuid.UUID
	ExecutionID   uuid.UUID
	WorkspaceID   uuid.UUID
	TestRunItemID *uuid.UUID
	Kind          ArtifactKind
	Name          string
	FilePath      string
	MimeType      string
	FileSize      int64
	Metadata      map[string]interface{}
	CreatedAt     time.Time
}

type AutomationLog struct {
	ID          uuid.UUID
	ExecutionID uuid.UUID
	WorkspaceID uuid.UUID
	Level       string
	Message     string
	LoggedAt    time.Time
	CreatedAt   time.Time
}

type ParsedReport struct {
	Total      int
	Passed     int
	Failed     int
	Skipped    int
	Blocked    int
	DurationMs int64
	Suites     []ParsedSuite
}

type ParsedSuite struct {
	Name  string
	Cases []ParsedCase
}

type ParsedCase struct {
	Name         string
	Status       string
	DurationMs   int64
	ErrorMessage string
	StackTrace   string
	Logs         []string
	Screenshots  []string
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func durationFromFloat(seconds float64) int64 {
	return int64(seconds * 1000)
}
