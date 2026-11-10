// Package workflow owns the editable definition of a workflow, including its
// stages, ports, bindings, artifact specs, gates, and permitted revisits.
package workflow

import (
	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
)

type Workflow struct {
	ID               identity.WorkflowID                               `json:"id" yaml:"id"`
	Name             string                                            `json:"name" yaml:"name"`
	StageDefinitions map[identity.StageID]StageDefinition              `json:"stage_definitions" yaml:"stage_definitions"`
	ArtifactSpecs    map[identity.ArtifactSpecID]artifact.ArtifactSpec `json:"artifact_specs" yaml:"artifact_specs"`
	Inputs           map[identity.PortName]identity.ArtifactSpecID     `json:"inputs" yaml:"inputs"`
	Outputs          map[identity.PortName]PortRef                     `json:"outputs" yaml:"outputs"`
	Bindings         []Binding                                         `json:"bindings" yaml:"bindings"`
}

// AssignIDs initializes model identities from their map keys after decoding a
// configuration. ID fields are omitted from YAML but retained in domain models
// and JSON snapshots. Validate checks that those identities remain consistent.
func (w *Workflow) AssignIDs() {
	for id, spec := range w.ArtifactSpecs {
		spec.ID = id
		w.ArtifactSpecs[id] = spec
	}
	for id, stage := range w.StageDefinitions {
		stage.ID = id
		stage.assignGateIDs()
		w.StageDefinitions[id] = stage
	}
}
