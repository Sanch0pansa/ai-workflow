package workflow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
)

type stageFingerprint struct {
	Stage    StageDefinition                                   `json:"stage"`
	Bindings []Binding                                         `json:"bindings"`
	Specs    map[identity.ArtifactSpecID]artifact.ArtifactSpec `json:"specs"`
}

// StageFingerprint tracks the definition, incoming bindings and used specs.
// It records provenance without preventing edits of an unfinished workflow.
func (w *Workflow) StageFingerprint(stageID identity.StageID) string {
	snapshot := stageFingerprint{
		Stage:    w.StageDefinitions[stageID],
		Bindings: w.IncomingBindings(stageID),
		Specs:    w.StageArtifactSpecs(stageID),
	}
	sort.Slice(snapshot.Bindings, func(i, j int) bool {
		return snapshot.Bindings[i].To.Port < snapshot.Bindings[j].To.Port
	})
	data, _ := json.Marshal(snapshot)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (w *Workflow) StageArtifactSpecs(stageID identity.StageID) map[identity.ArtifactSpecID]artifact.ArtifactSpec {
	stage := w.StageDefinitions[stageID]
	specs := map[identity.ArtifactSpecID]artifact.ArtifactSpec{}
	for _, ports := range []map[identity.PortName]identity.ArtifactSpecID{stage.Inputs, stage.Outputs} {
		for _, id := range ports {
			specs[id] = w.ArtifactSpecs[id]
		}
	}
	return specs
}
