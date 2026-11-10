package workflow

import (
	"fmt"

	"ai-workflow/internal/domain/identity"
)

func (w *Workflow) validateAcyclic() error {
	visiting := map[identity.StageID]bool{}
	visited := map[identity.StageID]bool{}
	for id := range w.StageDefinitions {
		if err := w.checkDependencyCycles(id, visiting, visited); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) checkDependencyCycles(id identity.StageID, visiting, visited map[identity.StageID]bool) error {
	if visiting[id] {
		return fmt.Errorf("dependency cycle at stage %q", id)
	}
	if visited[id] {
		return nil
	}
	visiting[id] = true
	for _, dependency := range w.DependencyStageIDs(id) {
		if err := w.checkDependencyCycles(dependency, visiting, visited); err != nil {
			return err
		}
	}
	delete(visiting, id)
	visited[id] = true
	return nil
}

func (w *Workflow) validateRevisits() error {
	for id, stage := range w.StageDefinitions {
		for _, target := range stage.AllowedRevisits {
			if !w.isAncestor(target, id) {
				return fmt.Errorf("stage %q revisit target %q must be an ancestor", id, target)
			}
		}
	}
	return nil
}
