package workflow

import (
	"ai-workflow/internal/domain/identity"
)

// PortRef refers to a stage port. An empty StageID denotes a workflow input.
type PortRef struct {
	StageID identity.StageID  `json:"stage_id,omitempty" yaml:"stage_id,omitempty"`
	Port    identity.PortName `json:"port" yaml:"port"`
}

type Binding struct {
	From PortRef `json:"from" yaml:"from"`
	To   PortRef `json:"to" yaml:"to"`
}
