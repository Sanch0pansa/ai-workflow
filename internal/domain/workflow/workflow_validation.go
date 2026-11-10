package workflow

import (
	"fmt"

	"ai-workflow/internal/domain/identity"
)

func (w *Workflow) Validate() error {
	if w == nil {
		return fmt.Errorf("workflow is required")
	}
	if w.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}
	if err := w.validateArtifactSpecs(); err != nil {
		return err
	}
	if err := w.validatePorts(w.Inputs); err != nil {
		return fmt.Errorf("workflow inputs: %w", err)
	}
	if err := w.validateStages(); err != nil {
		return err
	}
	if err := w.validateBindings(); err != nil {
		return err
	}
	if err := w.validateOutputs(); err != nil {
		return err
	}
	if err := w.validateAcyclic(); err != nil {
		return err
	}
	return w.validateRevisits()
}

func (w *Workflow) validateArtifactSpecs() error {
	for id, spec := range w.ArtifactSpecs {
		if id == "" {
			return fmt.Errorf("artifact spec key must not be empty")
		}
		if spec.ID != id {
			return fmt.Errorf("artifact spec %q: ID must match its map key", id)
		}
		if spec.Type == "" {
			return fmt.Errorf("artifact spec %q must have a type", id)
		}
	}
	return nil
}

func (w *Workflow) validatePorts(ports map[identity.PortName]identity.ArtifactSpecID) error {
	for name, specID := range ports {
		if name == "" {
			return fmt.Errorf("port name must not be empty")
		}
		_, exists := w.ArtifactSpecs[specID]
		if !exists {
			return fmt.Errorf("port %q: unknown artifact spec %q", name, specID)
		}
	}
	return nil
}

func (w *Workflow) validateOutputs() error {
	if len(w.Outputs) == 0 {
		return fmt.Errorf("workflow must have at least one output")
	}
	for port, source := range w.Outputs {
		if port == "" {
			return fmt.Errorf("workflow output name must not be empty")
		}
		if source.StageID == "" {
			return fmt.Errorf("workflow output %q must refer to a stage output", port)
		}
		if _, err := w.ArtifactSpecByPort(source); err != nil {
			return fmt.Errorf("workflow output %q: %w", port, err)
		}
	}
	return nil
}
