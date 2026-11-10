package run

import (
	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

type AttemptStatus string

const (
	AttemptStatusRunning     AttemptStatus = "running"
	AttemptStatusChecking    AttemptStatus = "checking"
	AttemptStatusApproved    AttemptStatus = "approved"
	AttemptStatusRejected    AttemptStatus = "rejected"
	AttemptStatusFailed      AttemptStatus = "failed"
	AttemptStatusInvalidated AttemptStatus = "invalidated"
)

type StageAttempt struct {
	ID                    identity.StageAttemptID                           `json:"id"`
	Status                AttemptStatus                                     `json:"status"`
	DefinitionFingerprint string                                            `json:"definition_fingerprint"`
	Definition            workflow.StageDefinition                          `json:"definition"`
	ArtifactSpecs         map[identity.ArtifactSpecID]artifact.ArtifactSpec `json:"artifact_specs"`
	Inputs                map[identity.PortName]identity.ArtifactID         `json:"inputs"` // port -> artifact ID, fixed at start
	PreviousAttemptID     identity.StageAttemptID                           `json:"previous_attempt_id,omitempty"`
	PreviousOutputs       map[identity.PortName]identity.ArtifactID         `json:"previous_outputs"`
	Feedback              Feedback                                          `json:"feedback"`
	Outputs               map[identity.PortName]identity.ArtifactID         `json:"outputs"`
	GateResults           []GateResult                                      `json:"gate_results"`
}

func (a *StageAttempt) invalidate() {
	switch a.Status {
	case AttemptStatusRunning, AttemptStatusChecking:
		a.Status = AttemptStatusInvalidated
	}
}
