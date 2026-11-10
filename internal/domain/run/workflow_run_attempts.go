package run

import (
	"fmt"
	"slices"

	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

// StartAttempt resolves current inputs and adds previous own outputs and selected
// feedback automatically. Returned data is a copy, not mutable aggregate state.
func (r *WorkflowRun) StartAttempt(
	w *workflow.Workflow,
	stageID identity.StageID,
	attemptID identity.StageAttemptID,
) (StageAttempt, error) {
	if err := r.reconcileForUpdate(w); err != nil {
		return StageAttempt{}, err
	}
	if err := r.validateAttemptStart(w, stageID, attemptID); err != nil {
		return StageAttempt{}, err
	}

	attempt := r.newAttempt(w, stageID, attemptID)
	r.StageRuns[stageID].start(attempt)
	r.refreshStatus(w)
	return cloneAttempt(attempt), nil
}

// SubmitOutputs freezes all outputs together. Each output needs a new artifact
// ID, even when its content happens to match an older artifact.
func (r *WorkflowRun) SubmitOutputs(
	w *workflow.Workflow,
	stageID identity.StageID, attemptID identity.StageAttemptID,
	outputs map[identity.PortName]artifact.Artifact,
) error {
	if err := r.reconcileForUpdate(w); err != nil {
		return err
	}
	stage, attempt, err := r.runningAttempt(stageID, attemptID)
	if err != nil {
		return err
	}
	if err := r.validateOutputs(attempt, outputs); err != nil {
		return err
	}

	r.storeOutputs(attempt, outputs)
	stage.beginChecks(attempt)
	r.refreshStatus(w)
	return nil
}

// FailAttempt records an execution failure before outputs were produced.
func (r *WorkflowRun) FailAttempt(
	w *workflow.Workflow,
	stageID identity.StageID,
	attemptID identity.StageAttemptID,
	reason string,
) error {
	if err := r.reconcileForUpdate(w); err != nil {
		return err
	}
	stage, attempt, err := r.runningAttempt(stageID, attemptID)
	if err != nil {
		return err
	}

	attempt.Status = AttemptStatusFailed
	stage.PendingFeedback = Feedback{Reason: reason}
	r.markForRetry(w, stage)
	r.refreshStatus(w)
	return nil
}

func (r *WorkflowRun) validateAttemptStart(
	w *workflow.Workflow,
	stageID identity.StageID,
	attemptID identity.StageAttemptID,
) error {
	if !slices.Contains(w.RequiredStageIDs(), stageID) {
		return fmt.Errorf("stage %q is not required by workflow outputs", stageID)
	}
	if !r.isReady(w, stageID) {
		return fmt.Errorf("stage %q is not ready", stageID)
	}
	if attemptID == "" {
		return fmt.Errorf("attempt ID is required")
	}
	if r.hasAttempt(attemptID) {
		return fmt.Errorf("attempt ID %q is already used in this run", attemptID)
	}
	return nil
}

func (r *WorkflowRun) currentAttempt(
	stageID identity.StageID,
	attemptID identity.StageAttemptID,
) (*StageRun, *StageAttempt, error) {
	stage, exists := r.StageRuns[stageID]
	if !exists {
		return nil, nil, fmt.Errorf("unknown stage %q", stageID)
	}
	if stage.CurrentAttemptID != attemptID {
		return nil, nil, fmt.Errorf("attempt %q is not current for stage %q", attemptID, stageID)
	}
	attempt := stage.attempt(attemptID)
	if attempt == nil {
		return nil, nil, fmt.Errorf("unknown attempt %q", attemptID)
	}
	return stage, attempt, nil
}

func (r *WorkflowRun) runningAttempt(
	stageID identity.StageID,
	attemptID identity.StageAttemptID,
) (*StageRun, *StageAttempt, error) {
	stage, attempt, err := r.currentAttempt(stageID, attemptID)
	if err != nil {
		return nil, nil, err
	}
	if stage.Status != StageStatusRunning {
		return nil, nil, fmt.Errorf("stage %q is not running", stageID)
	}
	if attempt.Status != AttemptStatusRunning {
		return nil, nil, fmt.Errorf("attempt %q is not running", attemptID)
	}
	return stage, attempt, nil
}

func (r *WorkflowRun) hasAttempt(id identity.StageAttemptID) bool {
	for _, stage := range r.StageRuns {
		if stage.attempt(id) != nil {
			return true
		}
	}
	return false
}

func (r *WorkflowRun) markForRetry(w *workflow.Workflow, stage *StageRun) {
	stage.Status = StageStatusPending
	if stage.attemptLimitReached(w.StageDefinitions[stage.StageID].AttemptLimit) {
		stage.Status = StageStatusFailed
	}
}
