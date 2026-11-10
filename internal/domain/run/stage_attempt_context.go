package run

import (
	"maps"

	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

func (r *WorkflowRun) newAttempt(
	w *workflow.Workflow,
	stageID identity.StageID,
	attemptID identity.StageAttemptID,
) StageAttempt {
	stage := r.StageRuns[stageID]
	definition := w.StageDefinitions[stageID]
	attempt := StageAttempt{
		ID:                    attemptID,
		Status:                AttemptStatusRunning,
		DefinitionFingerprint: stage.DefinitionFingerprint,
		Definition:            cloneDefinition(definition),
		ArtifactSpecs:         w.StageArtifactSpecs(stageID),
		Inputs:                r.currentInputs(w, stageID),
		PreviousOutputs:       map[identity.PortName]identity.ArtifactID{},
		Outputs:               map[identity.PortName]identity.ArtifactID{},
		Feedback:              cloneFeedback(stage.PendingFeedback),
	}
	previous := stage.previousAttemptWithOutputs()
	if previous != nil {
		attempt.PreviousAttemptID = previous.ID
		attempt.PreviousOutputs = maps.Clone(previous.Outputs)
	}
	return attempt
}

func (r *WorkflowRun) currentInputs(
	w *workflow.Workflow,
	stageID identity.StageID,
) map[identity.PortName]identity.ArtifactID {
	inputs := map[identity.PortName]identity.ArtifactID{}
	for _, binding := range w.IncomingBindings(stageID) {
		inputs[binding.To.Port] = r.sourceArtifact(binding.From)
	}
	return inputs
}
