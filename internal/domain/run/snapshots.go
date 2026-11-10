package run

import (
	"maps"
	"slices"

	"ai-workflow/internal/domain/workflow"
)

func cloneFeedback(value Feedback) Feedback {
	value.GateResults = slices.Clone(value.GateResults)
	return value
}

func cloneAttempt(value StageAttempt) StageAttempt {
	value.Definition = cloneDefinition(value.Definition)
	value.ArtifactSpecs = maps.Clone(value.ArtifactSpecs)
	value.Inputs = maps.Clone(value.Inputs)
	value.PreviousOutputs = maps.Clone(value.PreviousOutputs)
	value.Outputs = maps.Clone(value.Outputs)
	value.Feedback = cloneFeedback(value.Feedback)
	value.GateResults = slices.Clone(value.GateResults)
	for i := range value.GateResults {
		value.GateResults[i].ArtifactIDs = slices.Clone(value.GateResults[i].ArtifactIDs)
	}
	return value
}

func cloneDefinition(value workflow.StageDefinition) workflow.StageDefinition {
	value.Inputs = maps.Clone(value.Inputs)
	value.Outputs = maps.Clone(value.Outputs)
	value.AllowedRevisits = slices.Clone(value.AllowedRevisits)
	value.Gates = maps.Clone(value.Gates)
	for id, gate := range value.Gates {
		gate.OutputPorts = slices.Clone(gate.OutputPorts)
		gate.Params = maps.Clone(gate.Params)
		value.Gates[id] = gate
	}
	return value
}
