package run

import (
	"fmt"

	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

func (r *WorkflowRun) RecordGateResult(
	w *workflow.Workflow,
	stageID identity.StageID,
	attemptID identity.StageAttemptID,
	result GateResult,
) error {
	if err := r.reconcileForUpdate(w); err != nil {
		return err
	}
	stage, attempt, err := r.checkingAttempt(stageID, attemptID)
	if err != nil {
		return err
	}
	if err := attempt.validateGateResult(result); err != nil {
		return err
	}

	attempt.recordGateResult(result)
	r.applyGateVerdict(w, stage, attempt, result)
	r.refreshStatus(w)
	return nil
}

func (r *WorkflowRun) checkingAttempt(
	stageID identity.StageID,
	attemptID identity.StageAttemptID,
) (*StageRun, *StageAttempt, error) {
	stage, attempt, err := r.currentAttempt(stageID, attemptID)
	if err != nil {
		return nil, nil, err
	}
	if stage.Status != StageStatusChecking {
		return nil, nil, fmt.Errorf("stage %q is not awaiting checks", stageID)
	}
	if attempt.Status != AttemptStatusChecking {
		return nil, nil, fmt.Errorf("attempt %q is not awaiting checks", attemptID)
	}
	return stage, attempt, nil
}

func (r *WorkflowRun) applyGateVerdict(
	w *workflow.Workflow,
	stage *StageRun,
	attempt *StageAttempt,
	result GateResult,
) {
	if result.Status == GateStatusRejected {
		r.rejectAttempt(w, stage, attempt, result)
		return
	}
	if attempt.allGatesApproved() {
		stage.approve(attempt)
	}
}

func (r *WorkflowRun) rejectAttempt(w *workflow.Workflow, stage *StageRun, attempt *StageAttempt, result GateResult) {
	attempt.Status = AttemptStatusRejected
	stage.PendingFeedback = Feedback{
		Reason: result.Content,
		GateResults: []GateResultRef{{
			StageID:   stage.StageID,
			AttemptID: attempt.ID,
			ResultID:  result.ID,
		}},
	}
	r.markForRetry(w, stage)
}

func (r *WorkflowRun) hasGateResult(ref GateResultRef) bool {
	stage, exists := r.StageRuns[ref.StageID]
	if !exists {
		return false
	}
	attempt := stage.attempt(ref.AttemptID)
	if attempt == nil {
		return false
	}
	return attempt.hasGateResult(ref.ResultID)
}
