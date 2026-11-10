package workflow

import (
	"fmt"

	"ai-workflow/internal/domain/identity"
)

func (g Gate) validate(outputs map[identity.PortName]identity.ArtifactSpecID) error {
	if err := g.validateType(); err != nil {
		return err
	}
	return g.validateOutputPorts(outputs)
}

func (g Gate) validateType() error {
	switch g.Type {
	case GateTypeCmd, GateTypeHuman, GateTypeAgent:
		return nil
	default:
		return fmt.Errorf("unknown gate type %q", g.Type)
	}
}

func (g Gate) validateOutputPorts(outputs map[identity.PortName]identity.ArtifactSpecID) error {
	if len(g.OutputPorts) == 0 {
		return fmt.Errorf("at least one output port is required")
	}
	seen := map[identity.PortName]bool{}
	for _, port := range g.OutputPorts {
		_, exists := outputs[port]
		if !exists {
			return fmt.Errorf("unknown output port %q", port)
		}
		if seen[port] {
			return fmt.Errorf("duplicate output port %q", port)
		}
		seen[port] = true
	}
	return nil
}
