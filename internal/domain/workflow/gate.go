package workflow

import (
	"ai-workflow/internal/domain/identity"
)

type GateType string

const (
	GateTypeCmd   GateType = "cmd"
	GateTypeHuman GateType = "human"
	GateTypeAgent GateType = "agent"
)

// Gate checks selected output ports. Every declared gate is required.
type Gate struct {
	ID          identity.GateID     `json:"id" yaml:"-"`
	Type        GateType            `json:"type" yaml:"type"`
	Name        string              `json:"name" yaml:"name"`
	OutputPorts []identity.PortName `json:"output_ports" yaml:"output_ports"`
	Params      map[string]string   `json:"params,omitempty" yaml:"params,omitempty"`
}
