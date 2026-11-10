package workflow

import (
	"ai-workflow/internal/domain/identity"
)

type StageDefinition struct {
	ID          identity.StageID                              `json:"id" yaml:"-"`
	Name        string                                        `json:"name" yaml:"name"`
	Description string                                        `json:"description" yaml:"description"`
	Inputs      map[identity.PortName]identity.ArtifactSpecID `json:"inputs" yaml:"inputs"`   // port name -> artifact spec ID
	Outputs     map[identity.PortName]identity.ArtifactSpecID `json:"outputs" yaml:"outputs"` // port name -> artifact spec ID
	Gates       map[identity.GateID]Gate                      `json:"gates" yaml:"gates"`
	// AttemptLimit counts all attempts of this stage in one run; zero is unlimited.
	AttemptLimit    int                `json:"attempt_limit" yaml:"attempt_limit"`
	AllowedRevisits []identity.StageID `json:"allowed_revisits,omitempty" yaml:"allowed_revisits,omitempty"`
}

func (s *StageDefinition) assignGateIDs() {
	for id, gate := range s.Gates {
		gate.ID = id
		s.Gates[id] = gate
	}
}
