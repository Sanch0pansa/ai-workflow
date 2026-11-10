package run

import (
	"fmt"
	"slices"

	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

// Revisit is an explicit human/agent decision, separate from a check verdict.
func (r *WorkflowRun) Revisit(
	w *workflow.Workflow,
	fromStageID, targetStageID identity.StageID,
	actor workflow.GateType,
	feedback Feedback,
) error {
	if err := r.reconcileForUpdate(w); err != nil {
		return err
	}
	if err := validateRevisit(w, fromStageID, targetStageID, actor); err != nil {
		return err
	}
	if err := r.validateRevisitFeedback(feedback); err != nil {
		return err
	}

	r.invalidate(w, targetStageID, feedback)
	r.refreshStatus(w)
	return nil
}

func validateRevisit(w *workflow.Workflow, fromStageID, targetStageID identity.StageID, actor workflow.GateType) error {
	if actor != workflow.GateTypeHuman && actor != workflow.GateTypeAgent {
		return fmt.Errorf("only a human or agent may request a revisit")
	}
	source, exists := w.StageDefinitions[fromStageID]
	if !exists {
		return fmt.Errorf("unknown revisit source %q", fromStageID)
	}
	if !slices.Contains(source.AllowedRevisits, targetStageID) {
		return fmt.Errorf("revisit from %q to %q is not permitted", fromStageID, targetStageID)
	}
	return nil
}

func (r *WorkflowRun) validateRevisitFeedback(feedback Feedback) error {
	if feedback.Reason == "" {
		return fmt.Errorf("revisit reason is required")
	}
	for _, ref := range feedback.GateResults {
		if !r.hasGateResult(ref) {
			return fmt.Errorf("unknown feedback evidence %+v", ref)
		}
	}
	return nil
}

func (r *WorkflowRun) invalidate(w *workflow.Workflow, stageID identity.StageID, feedback Feedback) {
	r.StageRuns[stageID].invalidate(feedback)
	for _, descendant := range w.DescendantStageIDs(stageID) {
		r.StageRuns[descendant].invalidate(Feedback{Reason: "upstream result was invalidated"})
	}
	r.Outputs = map[identity.PortName]identity.ArtifactID{}
}
