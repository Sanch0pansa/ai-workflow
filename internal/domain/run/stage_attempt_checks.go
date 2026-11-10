package run

import (
	"fmt"
	"slices"

	"ai-workflow/internal/domain/identity"
)

func (a *StageAttempt) validateGateResult(result GateResult) error {
	if err := a.validateGateResultIdentity(result); err != nil {
		return err
	}
	if err := result.validateStatus(); err != nil {
		return err
	}
	return a.validateCheckedArtifacts(result)
}

func (a *StageAttempt) validateGateResultIdentity(result GateResult) error {
	if result.ID == "" {
		return fmt.Errorf("gate result ID is required")
	}
	if a.hasGateResult(result.ID) {
		return fmt.Errorf("gate result ID %q is already used in this attempt", result.ID)
	}
	_, exists := a.Definition.Gates[result.GateID]
	if !exists {
		return fmt.Errorf("unknown gate %q", result.GateID)
	}
	return nil
}

func (a *StageAttempt) validateCheckedArtifacts(result GateResult) error {
	expected := a.gateArtifacts(result.GateID)
	actual := slices.Clone(result.ArtifactIDs)
	slices.Sort(expected)
	slices.Sort(actual)
	if !slices.Equal(expected, actual) {
		return fmt.Errorf("gate %q result does not refer to its exact output artifacts", result.GateID)
	}
	return nil
}

func (a *StageAttempt) gateArtifacts(gateID identity.GateID) []identity.ArtifactID {
	gate := a.Definition.Gates[gateID]
	ids := make([]identity.ArtifactID, 0, len(gate.OutputPorts))
	for _, port := range gate.OutputPorts {
		ids = append(ids, a.Outputs[port])
	}
	return ids
}

func (a *StageAttempt) recordGateResult(result GateResult) {
	result.ArtifactIDs = slices.Clone(result.ArtifactIDs)
	a.GateResults = append(a.GateResults, result)
}

func (a *StageAttempt) allGatesApproved() bool {
	latest := a.latestGateStatuses()
	for id := range a.Definition.Gates {
		if latest[id] != GateStatusApproved {
			return false
		}
	}
	return true
}

func (a *StageAttempt) latestGateStatuses() map[identity.GateID]GateStatus {
	latest := map[identity.GateID]GateStatus{}
	for _, result := range a.GateResults {
		latest[result.GateID] = result.Status
	}
	return latest
}

func (a *StageAttempt) hasGateResult(id identity.GateResultID) bool {
	for _, result := range a.GateResults {
		if result.ID == id {
			return true
		}
	}
	return false
}
