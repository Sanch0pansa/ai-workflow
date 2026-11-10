package run

import (
	"fmt"

	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
)

func (r *WorkflowRun) validateOutputs(attempt *StageAttempt, outputs map[identity.PortName]artifact.Artifact) error {
	if len(outputs) != len(attempt.Definition.Outputs) {
		return fmt.Errorf("stage output ports do not match")
	}
	seen := map[identity.ArtifactID]bool{}
	for port, specID := range attempt.Definition.Outputs {
		value, exists := outputs[port]
		if !exists {
			return fmt.Errorf("missing stage output %q", port)
		}
		if err := r.validateOutputArtifact(attempt.ID, value, specID); err != nil {
			return fmt.Errorf("output %q: %w", port, err)
		}
		if seen[value.ID] {
			return fmt.Errorf("duplicate output artifact ID %q", value.ID)
		}
		seen[value.ID] = true
	}
	return nil
}

func (r *WorkflowRun) validateOutputArtifact(
	attemptID identity.StageAttemptID,
	value artifact.Artifact,
	specID identity.ArtifactSpecID,
) error {
	if err := validatePortArtifact(value, specID); err != nil {
		return err
	}
	_, exists := r.Artifacts[value.ID]
	if exists {
		return fmt.Errorf("artifact ID %q is already used", value.ID)
	}
	return r.validateOutputProducer(attemptID, value)
}

func (r *WorkflowRun) validateOutputProducer(attemptID identity.StageAttemptID, value artifact.Artifact) error {
	if value.ProducerAttemptID != "" && value.ProducerAttemptID != attemptID {
		return fmt.Errorf("artifact %q belongs to a different attempt", value.ID)
	}
	if value.ProducerRunID != "" && value.ProducerRunID != r.ID {
		return fmt.Errorf("artifact %q belongs to a different run", value.ID)
	}
	return nil
}

func (r *WorkflowRun) storeOutputs(attempt *StageAttempt, outputs map[identity.PortName]artifact.Artifact) {
	for port, value := range outputs {
		value.ProducerRunID = r.ID
		value.ProducerAttemptID = attempt.ID
		r.Artifacts[value.ID] = value
		attempt.Outputs[port] = value.ID
	}
}
