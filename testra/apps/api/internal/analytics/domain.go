package analytics

import (
	"time"

	"github.com/google/uuid"
)

type Dashboard struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Type        string
	Config      map[string]interface{}
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DailyMetric struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	ProjectID   *uuid.UUID
	MetricDate  time.Time
	TotalRuns   int
	Passed      int
	Failed      int
	Skipped     int
	Blocked     int
	DurationMs  int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Summary struct {
	TotalRuns  int   `json:"total_runs"`
	Passed     int   `json:"passed"`
	Failed     int   `json:"failed"`
	Skipped    int   `json:"skipped"`
	Blocked    int   `json:"blocked"`
	DurationMs int64 `json:"duration_ms"`
}

type TrendPoint struct {
	Date       string `json:"date"`
	TotalRuns  int    `json:"total_runs"`
	Passed     int    `json:"passed"`
	Failed     int    `json:"failed"`
	Skipped    int    `json:"skipped"`
	Blocked    int    `json:"blocked"`
	DurationMs int64  `json:"duration_ms"`
}

type Metrics struct {
	TotalTestCases         int64                 `json:"total_test_cases"`
	TotalTestPlans         int64                 `json:"total_test_plans"`
	TotalTestRuns          int64                 `json:"total_test_runs"`
	ExecutionProgress      float64               `json:"execution_progress"`
	PassRate               float64               `json:"pass_rate"`
	FailRate               float64               `json:"fail_rate"`
	Blocked                int64                 `json:"blocked"`
	Retest                 int64                 `json:"retest"`
	Skipped                int64                 `json:"skipped"`
	AutomationCoverage     float64               `json:"automation_coverage"`
	APITestCoverage        float64               `json:"api_test_coverage"`
	ExecutionDurationMs    int64                 `json:"execution_duration_ms"`
	AverageExecutionTimeMs int64                 `json:"average_execution_time_ms"`
	TopFailedTestCases     []TopFailedItem       `json:"top_failed_test_cases"`
	TopFailedSuites        []TopFailedSuite      `json:"top_failed_suites"`
	TopFailedAPIs          []TopFailedAPI        `json:"top_failed_apis"`
	MostActiveQA           []ActiveUser          `json:"most_active_qa"`
	MostActiveAutomation   []ActiveUser          `json:"most_active_automation"`
	DefectDensity          float64               `json:"defect_density"`
	OpenDefects            int64                 `json:"open_defects"`
	ClosedDefects          int64                 `json:"closed_defects"`
	DefectAging            DefectAging           `json:"defect_aging"`
	BugReopenRate          float64               `json:"bug_reopen_rate"`
	RecentActivity         []Activity            `json:"recent_activity"`
	ExecutionTimeline      []TimelinePoint       `json:"execution_timeline"`
	WeeklyTrend            []TrendPoint          `json:"weekly_trend"`
	MonthlyTrend           []TrendPoint          `json:"monthly_trend"`
	ReleaseQualityTrend    []ReleaseQualityPoint `json:"release_quality_trend"`
}

type TopFailedItem struct {
	TestCaseID string `json:"test_case_id"`
	Title      string `json:"title"`
	Failures   int64  `json:"failures"`
}

type TopFailedSuite struct {
	SuiteID  string `json:"suite_id"`
	Name     string `json:"name"`
	Failures int64  `json:"failures"`
}

type TopFailedAPI struct {
	RequestID string `json:"request_id"`
	Name      string `json:"name"`
	Failures  int64  `json:"failures"`
}

type ActiveUser struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Count  int64  `json:"count"`
}

type DefectAging struct {
	AverageDays float64 `json:"average_days"`
	MaxDays     int64   `json:"max_days"`
}

type Activity struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type TimelinePoint struct {
	Date       string `json:"date"`
	TotalRuns  int    `json:"total_runs"`
	Passed     int    `json:"passed"`
	Failed     int    `json:"failed"`
	Skipped    int    `json:"skipped"`
	Blocked    int    `json:"blocked"`
	DurationMs int64  `json:"duration_ms"`
}

type ReleaseQualityPoint struct {
	Release string `json:"release"`
	Passed  int    `json:"passed"`
	Failed  int    `json:"failed"`
	Skipped int    `json:"skipped"`
	Blocked int    `json:"blocked"`
	Total   int    `json:"total"`
}
