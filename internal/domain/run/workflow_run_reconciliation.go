package run

import (
	"fmt"
	"slices"

	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

// Reconcile applies an edited definition to an unfinished run. Changed stages
// and their consumers lose readiness; their attempt history is preserved.
func (r *WorkflowRun) Reconcile(w *workflow.Workflow) error {
	if r.Status == RunStatusCompleted {
		return ErrRunCompleted
	}
	if err := r.validateCompatibleWorkflow(w); err != nil {
		return err
	}

	for _, stageID := range r.synchronizeStages(w) {
		r.invalidate(w, stageID, Feedback{Reason: "stage definition changed"})
	}
	r.refreshStatus(w)
	return nil
}

func (r *WorkflowRun) validateCompatibleWorkflow(w *workflow.Workflow) error {
	if err := w.Validate(); err != nil {
		return err
	}
	if w.ID != r.WorkflowID {
		return fmt.Errorf("workflow identity no longer matches run")
	}
	if err := r.validateInputContract(w); err != nil {
		return err
	}
	return r.validateKnownStages(w)
}

func (r *WorkflowRun) validateInputContract(w *workflow.Workflow) error {
	if len(r.Inputs) != len(w.Inputs) {
		return fmt.Errorf("workflow input ports no longer match run")
	}
	for port, specID := range w.Inputs {
		value, exists := r.Artifacts[r.Inputs[port]]
		if !exists {
			return fmt.Errorf("workflow input %q is missing from run", port)
		}
		if err := validatePortArtifact(value, specID); err != nil {
			return fmt.Errorf("workflow input %q: %w", port, err)
		}
		if r.InputSpecs[port] != w.ArtifactSpecs[specID] {
			return fmt.Errorf("workflow input %q specification changed", port)
		}
	}
	return nil
}

func (r *WorkflowRun) validateKnownStages(w *workflow.Workflow) error {
	for stageID := range r.StageRuns {
		_, exists := w.StageDefinitions[stageID]
		if !exists {
			return fmt.Errorf("stage %q was removed from the workflow", stageID)
		}
	}
	return nil
}

func (r *WorkflowRun) synchronizeStages(w *workflow.Workflow) []identity.StageID {
	changed := []identity.StageID{}
	for stageID := range w.StageDefinitions {
		fingerprint := w.StageFingerprint(stageID)
		stage, exists := r.StageRuns[stageID]
		if !exists {
			r.StageRuns[stageID] = newStageRun(stageID, fingerprint)
			continue
		}
		if stage.DefinitionFingerprint != fingerprint {
			stage.DefinitionFingerprint = fingerprint
			changed = append(changed, stageID)
		}
	}
	slices.Sort(changed)
	return changed
}

// Reconciliation can itself complete a run when its declared outputs change.
// An update must therefore check openness again after applying the definition.
func (r *WorkflowRun) reconcileForUpdate(w *workflow.Workflow) error {
	if err := r.Reconcile(w); err != nil {
		return err
	}
	if r.Status == RunStatusCompleted {
		return ErrRunCompleted
	}
	return nil
}
