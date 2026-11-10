package workflow

import (
	"fmt"

	"ai-workflow/internal/domain/identity"
)

func (w *Workflow) validateBindings() error {
	bound := map[PortRef]bool{}
	for _, binding := range w.Bindings {
		if bound[binding.To] {
			return fmt.Errorf("duplicate binding target %+v", binding.To)
		}
		if err := w.validateBinding(binding); err != nil {
			return err
		}
		bound[binding.To] = true
	}
	return w.validateBoundInputs(bound)
}

func (w *Workflow) validateBinding(binding Binding) error {
	sourceSpec, err := w.ArtifactSpecByPort(binding.From)
	if err != nil {
		return err
	}
	targetSpec, err := w.inputSpec(binding.To)
	if err != nil {
		return err
	}
	if sourceSpec != targetSpec {
		return fmt.Errorf("artifact spec mismatch in binding %+v", binding)
	}
	return nil
}

func (w *Workflow) inputSpec(ref PortRef) (identity.ArtifactSpecID, error) {
	stage, exists := w.StageDefinitions[ref.StageID]
	if !exists {
		return "", fmt.Errorf("unknown target stage %q", ref.StageID)
	}
	specID, exists := stage.Inputs[ref.Port]
	if !exists {
		return "", fmt.Errorf("unknown target port %+v", ref)
	}
	return specID, nil
}

func (w *Workflow) validateBoundInputs(bound map[PortRef]bool) error {
	for id, stage := range w.StageDefinitions {
		for port := range stage.Inputs {
			ref := PortRef{StageID: id, Port: port}
			if !bound[ref] {
				return fmt.Errorf("stage %q input %q is not bound", id, port)
			}
		}
	}
	return nil
}
