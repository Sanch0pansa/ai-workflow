package run

import (
	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

func (r *WorkflowRun) ReadyStages(w *workflow.Workflow) ([]identity.StageID, error) {
	if r.Status == RunStatusCompleted {
		return nil, nil
	}
	if err := r.Reconcile(w); err != nil {
		return nil, err
	}
	ready := []identity.StageID{}
	for _, stageID := range w.RequiredStageIDs() {
		if r.isReady(w, stageID) {
			ready = append(ready, stageID)
		}
	}
	return ready, nil
}

func (r *WorkflowRun) refreshStatus(w *workflow.Workflow) {
	outputs := r.acceptedWorkflowOutputs(w)
	if len(outputs) == len(w.Outputs) {
		r.Outputs = outputs
		r.Status = RunStatusCompleted
		return
	}

	r.Outputs = map[identity.PortName]identity.ArtifactID{}
	if r.canProgress(w) {
		r.Status = RunStatusRunning
	} else {
		r.Status = RunStatusBlocked
	}
}

func (r *WorkflowRun) acceptedWorkflowOutputs(w *workflow.Workflow) map[identity.PortName]identity.ArtifactID {
	outputs := map[identity.PortName]identity.ArtifactID{}
	for port, source := range w.Outputs {
		if r.StageRuns[source.StageID].Status == StageStatusCompleted {
			outputs[port] = r.sourceArtifact(source)
		}
	}
	return outputs
}

func (r *WorkflowRun) canProgress(w *workflow.Workflow) bool {
	for _, stageID := range w.RequiredStageIDs() {
		if r.StageRuns[stageID].isActive() {
			return true
		}
		if r.isReady(w, stageID) {
			return true
		}
	}
	return false
}

func (r *WorkflowRun) isReady(w *workflow.Workflow, stageID identity.StageID) bool {
	stage, exists := r.StageRuns[stageID]
	if !exists {
		return false
	}
	if !stage.needsAttempt() {
		return false
	}
	if stage.attemptLimitReached(w.StageDefinitions[stageID].AttemptLimit) {
		return false
	}
	return r.dependenciesCompleted(w, stageID)
}

func (r *WorkflowRun) dependenciesCompleted(w *workflow.Workflow, stageID identity.StageID) bool {
	for _, dependency := range w.DependencyStageIDs(stageID) {
		if r.StageRuns[dependency].Status != StageStatusCompleted {
			return false
		}
	}
	return true
}

func (r *WorkflowRun) sourceArtifact(ref workflow.PortRef) identity.ArtifactID {
	if ref.StageID == "" {
		return r.Inputs[ref.Port]
	}
	stage := r.StageRuns[ref.StageID]
	attempt := stage.attempt(stage.CurrentAttemptID)
	if attempt == nil {
		return ""
	}
	return attempt.Outputs[ref.Port]
}
