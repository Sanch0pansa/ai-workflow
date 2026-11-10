package workflow

import (
	"fmt"
)

func (w *Workflow) validateStages() error {
	for id, stage := range w.StageDefinitions {
		if id == "" {
			return fmt.Errorf("stage key must not be empty")
		}
		if stage.ID != id {
			return fmt.Errorf("stage %q: ID must match its map key", id)
		}
		if err := stage.validate(w); err != nil {
			return fmt.Errorf("stage %q: %w", id, err)
		}
	}
	return nil
}

func (s StageDefinition) validate(w *Workflow) error {
	if s.AttemptLimit < 0 {
		return fmt.Errorf("attempt limit must not be negative")
	}
	if err := w.validatePorts(s.Inputs); err != nil {
		return fmt.Errorf("inputs: %w", err)
	}
	if err := w.validatePorts(s.Outputs); err != nil {
		return fmt.Errorf("outputs: %w", err)
	}
	return s.validateGates()
}

func (s StageDefinition) validateGates() error {
	for id, gate := range s.Gates {
		if id == "" {
			return fmt.Errorf("gate key must not be empty")
		}
		if gate.ID != id {
			return fmt.Errorf("gate %q: ID must match its map key", id)
		}
		if err := gate.validate(s.Outputs); err != nil {
			return fmt.Errorf("gate %q: %w", id, err)
		}
	}
	return nil
}
