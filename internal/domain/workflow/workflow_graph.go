package workflow

import (
	"fmt"
	"maps"
	"slices"

	"ai-workflow/internal/domain/identity"
)

func (w *Workflow) ArtifactSpecByPort(ref PortRef) (identity.ArtifactSpecID, error) {
	ports := w.Inputs
	if ref.StageID != "" {
		stage, exists := w.StageDefinitions[ref.StageID]
		if !exists {
			return "", fmt.Errorf("unknown stage %q", ref.StageID)
		}
		ports = stage.Outputs
	}
	specID, exists := ports[ref.Port]
	if !exists {
		return "", fmt.Errorf("unknown source port %+v", ref)
	}
	return specID, nil
}

func (w *Workflow) IncomingBindings(stageID identity.StageID) []Binding {
	bindings := []Binding{}
	for _, binding := range w.Bindings {
		if binding.To.StageID == stageID {
			bindings = append(bindings, binding)
		}
	}
	return bindings
}

// DependencyStageIDs returns unique predecessor IDs in unspecified order.
func (w *Workflow) DependencyStageIDs(stageID identity.StageID) []identity.StageID {
	set := map[identity.StageID]bool{}
	for _, binding := range w.IncomingBindings(stageID) {
		if binding.From.StageID != "" {
			set[binding.From.StageID] = true
		}
	}
	return slices.Collect(maps.Keys(set))
}

// RequiredStageIDs returns producers of workflow outputs and their ancestors
// in unspecified order.
// Call Validate before traversing a definition obtained from external input.
func (w *Workflow) RequiredStageIDs() []identity.StageID {
	stages := map[identity.StageID]bool{}
	for _, source := range w.Outputs {
		w.collectDependencies(source.StageID, stages)
	}
	return slices.Collect(maps.Keys(stages))
}

func (w *Workflow) collectDependencies(stageID identity.StageID, stages map[identity.StageID]bool) {
	if stages[stageID] {
		return
	}
	stages[stageID] = true
	for _, dependency := range w.DependencyStageIDs(stageID) {
		w.collectDependencies(dependency, stages)
	}
}

// DescendantStageIDs returns unique descendant IDs in unspecified order.
func (w *Workflow) DescendantStageIDs(stageID identity.StageID) []identity.StageID {
	stages := map[identity.StageID]bool{stageID: true}
	w.collectDescendants(stageID, stages)
	delete(stages, stageID)
	return slices.Collect(maps.Keys(stages))
}

func (w *Workflow) collectDescendants(stageID identity.StageID, stages map[identity.StageID]bool) {
	for _, binding := range w.Bindings {
		if binding.From.StageID != stageID {
			continue
		}
		child := binding.To.StageID
		if stages[child] {
			continue
		}
		stages[child] = true
		w.collectDescendants(child, stages)
	}
}

func (w *Workflow) isAncestor(ancestor, stageID identity.StageID) bool {
	for _, descendant := range w.DescendantStageIDs(ancestor) {
		if descendant == stageID {
			return true
		}
	}
	return false
}
