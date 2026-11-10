package run

import (
	"fmt"

	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

func NewWorkflowRun(
	id identity.WorkflowRunID,
	w *workflow.Workflow,
	inputs map[identity.PortName]artifact.Artifact,
) (*WorkflowRun, error) {
	if err := validateNewRun(id, w, inputs); err != nil {
		return nil, err
	}

	r := newRunState(id, w.ID)
	r.initializeInputs(w, inputs)
	r.initializeStages(w)
	return r, nil
}

func validateNewRun(
	id identity.WorkflowRunID,
	w *workflow.Workflow,
	inputs map[identity.PortName]artifact.Artifact,
) error {
	if id == "" {
		return fmt.Errorf("run ID is required")
	}
	if err := w.Validate(); err != nil {
		return err
	}
	return validateInitialInputs(w, inputs)
}

func validateInitialInputs(w *workflow.Workflow, inputs map[identity.PortName]artifact.Artifact) error {
	if len(inputs) != len(w.Inputs) {
		return fmt.Errorf("workflow input ports do not match")
	}
	artifacts := map[identity.ArtifactID]artifact.Artifact{}
	for port, specID := range w.Inputs {
		value, exists := inputs[port]
		if !exists {
			return fmt.Errorf("missing workflow input %q", port)
		}
		if err := validatePortArtifact(value, specID); err != nil {
			return fmt.Errorf("workflow input %q: %w", port, err)
		}
		previous, exists := artifacts[value.ID]
		if exists && previous != value {
			return fmt.Errorf("artifact ID %q has conflicting values", value.ID)
		}
		artifacts[value.ID] = value
	}
	return nil
}

func validatePortArtifact(value artifact.Artifact, specID identity.ArtifactSpecID) error {
	if value.ID == "" {
		return fmt.Errorf("artifact ID is required")
	}
	if value.ArtifactSpecID != specID {
		return fmt.Errorf("artifact %q must match spec %q", value.ID, specID)
	}
	return nil
}

func newRunState(id identity.WorkflowRunID, workflowID identity.WorkflowID) *WorkflowRun {
	return &WorkflowRun{
		ID:         id,
		WorkflowID: workflowID,
		Status:     RunStatusRunning,
		Inputs:     map[identity.PortName]identity.ArtifactID{},
		InputSpecs: map[identity.PortName]artifact.ArtifactSpec{},
		Outputs:    map[identity.PortName]identity.ArtifactID{},
		StageRuns:  map[identity.StageID]*StageRun{},
		Artifacts:  map[identity.ArtifactID]artifact.Artifact{},
	}
}

func (r *WorkflowRun) initializeInputs(w *workflow.Workflow, inputs map[identity.PortName]artifact.Artifact) {
	for port, value := range inputs {
		r.Inputs[port] = value.ID
		r.InputSpecs[port] = w.ArtifactSpecs[value.ArtifactSpecID]
		r.Artifacts[value.ID] = value
	}
}

func (r *WorkflowRun) initializeStages(w *workflow.Workflow) {
	for stageID := range w.StageDefinitions {
		r.StageRuns[stageID] = newStageRun(stageID, w.StageFingerprint(stageID))
	}
}
