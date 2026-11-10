package run

import (
	"ai-workflow/internal/domain/identity"
)

type StageStatus string

const (
	StageStatusPending     StageStatus = "pending"
	StageStatusRunning     StageStatus = "running"
	StageStatusChecking    StageStatus = "checking"
	StageStatusCompleted   StageStatus = "completed"
	StageStatusFailed      StageStatus = "failed"
	StageStatusInvalidated StageStatus = "invalidated"
)

// StageRun is owned by WorkflowRun. Invalidation changes its readiness, not the
// historical verdicts or artifacts of its attempts.
type StageRun struct {
	StageID               identity.StageID        `json:"stage_id"`
	Status                StageStatus             `json:"status"`
	DefinitionFingerprint string                  `json:"definition_fingerprint"`
	CurrentAttemptID      identity.StageAttemptID `json:"current_attempt_id,omitempty"`
	Attempts              []StageAttempt          `json:"attempts"`
	PendingFeedback       Feedback                `json:"pending_feedback"`
}

func newStageRun(stageID identity.StageID, fingerprint string) *StageRun {
	return &StageRun{
		StageID:               stageID,
		Status:                StageStatusPending,
		DefinitionFingerprint: fingerprint,
	}
}

func (s *StageRun) attempt(id identity.StageAttemptID) *StageAttempt {
	for i := range s.Attempts {
		if s.Attempts[i].ID == id {
			return &s.Attempts[i]
		}
	}
	return nil
}

func (s *StageRun) previousAttemptWithOutputs() *StageAttempt {
	for i := len(s.Attempts) - 1; i >= 0; i-- {
		if len(s.Attempts[i].Outputs) > 0 {
			return &s.Attempts[i]
		}
	}
	return nil
}

func (s *StageRun) start(attempt StageAttempt) {
	s.Attempts = append(s.Attempts, attempt)
	s.CurrentAttemptID = attempt.ID
	s.Status = StageStatusRunning
	s.PendingFeedback = Feedback{}
}

func (s *StageRun) beginChecks(attempt *StageAttempt) {
	if len(attempt.Definition.Gates) == 0 {
		s.approve(attempt)
		return
	}
	attempt.Status = AttemptStatusChecking
	s.Status = StageStatusChecking
}

func (s *StageRun) approve(attempt *StageAttempt) {
	attempt.Status = AttemptStatusApproved
	s.Status = StageStatusCompleted
}

func (s *StageRun) needsAttempt() bool {
	switch s.Status {
	case StageStatusPending, StageStatusInvalidated:
		return true
	default:
		return false
	}
}

func (s *StageRun) isActive() bool {
	switch s.Status {
	case StageStatusRunning, StageStatusChecking:
		return true
	default:
		return false
	}
}

func (s *StageRun) attemptLimitReached(limit int) bool {
	if limit == 0 {
		return false
	}
	return len(s.Attempts) >= limit
}

func (s *StageRun) invalidate(feedback Feedback) {
	current := s.attempt(s.CurrentAttemptID)
	if current != nil {
		current.invalidate()
	}
	s.Status = StageStatusInvalidated
	s.CurrentAttemptID = "" // late results can no longer become current
	s.PendingFeedback = cloneFeedback(feedback)
}
