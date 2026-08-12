package defects

import (
	"time"

	"github.com/google/uuid"
)

type DefectStatus string

type DefectSeverity string

type DefectPriority string

const (
	DefectStatusOpen       DefectStatus = "open"
	DefectStatusInProgress DefectStatus = "in_progress"
	DefectStatusResolved   DefectStatus = "resolved"
	DefectStatusClosed     DefectStatus = "closed"
	DefectStatusRejected   DefectStatus = "rejected"
)

const (
	DefectSeverityLow      DefectSeverity = "low"
	DefectSeverityMedium   DefectSeverity = "medium"
	DefectSeverityHigh     DefectSeverity = "high"
	DefectSeverityCritical DefectSeverity = "critical"
)

const (
	DefectPriorityLow      DefectPriority = "low"
	DefectPriorityMedium   DefectPriority = "medium"
	DefectPriorityHigh     DefectPriority = "high"
	DefectPriorityCritical DefectPriority = "critical"
)

func IsValidStatus(s DefectStatus) bool {
	switch s {
	case DefectStatusOpen, DefectStatusInProgress, DefectStatusResolved, DefectStatusClosed, DefectStatusRejected:
		return true
	}
	return false
}

func IsValidSeverity(s DefectSeverity) bool {
	switch s {
	case DefectSeverityLow, DefectSeverityMedium, DefectSeverityHigh, DefectSeverityCritical:
		return true
	}
	return false
}

func IsValidPriority(p DefectPriority) bool {
	switch p {
	case DefectPriorityLow, DefectPriorityMedium, DefectPriorityHigh, DefectPriorityCritical:
		return true
	}
	return false
}

type Defect struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	ProjectID     uuid.UUID
	TestRunItemID *uuid.UUID
	Title         string
	Description   string
	Severity      DefectSeverity
	Priority      DefectPriority
	Status        DefectStatus
	AssignedTo    *uuid.UUID
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
