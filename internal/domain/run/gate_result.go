package run

import (
	"fmt"

	"ai-workflow/internal/domain/identity"
)

type GateStatus string

const (
	GateStatusPending  GateStatus = "pending"
	GateStatusApproved GateStatus = "approved"
	GateStatusRejected GateStatus = "rejected"
	// Error means the check could not run, not that the artifact was rejected.
	GateStatusError GateStatus = "error"
)

// GateResult records one check execution; retries append results to history.
type GateResult struct {
	ID          identity.GateResultID `json:"id"`
	GateID      identity.GateID       `json:"gate_id"`
	Status      GateStatus            `json:"status"`
	ArtifactIDs []identity.ArtifactID `json:"artifact_ids"`
	Content     string                `json:"content"`
}

type GateResultRef struct {
	StageID   identity.StageID        `json:"stage_id"`
	AttemptID identity.StageAttemptID `json:"attempt_id"`
	ResultID  identity.GateResultID   `json:"result_id"`
}

func (r GateResult) validateStatus() error {
	switch r.Status {
	case GateStatusPending, GateStatusApproved, GateStatusRejected, GateStatusError:
		return nil
	default:
		return fmt.Errorf("invalid gate status %q", r.Status)
	}
}
