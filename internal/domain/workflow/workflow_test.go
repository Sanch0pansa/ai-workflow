package workflow_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

func definition() *workflow.Workflow {
	w := &workflow.Workflow{
		ID: "example", ArtifactSpecs: map[identity.ArtifactSpecID]artifact.ArtifactSpec{
			"text": {Type: artifact.ArtifactTypeMD},
		},
		Inputs:  map[identity.PortName]identity.ArtifactSpecID{"task": "text"},
		Outputs: map[identity.PortName]workflow.PortRef{"result": {StageID: "second", Port: "result"}},
		StageDefinitions: map[identity.StageID]workflow.StageDefinition{
			"first": {
				Inputs: map[identity.PortName]identity.ArtifactSpecID{"task": "text"}, Outputs: map[identity.PortName]identity.ArtifactSpecID{"result": "text"},
			},
			"second": {
				Inputs: map[identity.PortName]identity.ArtifactSpecID{"task": "text"}, Outputs: map[identity.PortName]identity.ArtifactSpecID{"result": "text"},
				AllowedRevisits: []identity.StageID{"first"},
				Gates: map[identity.GateID]workflow.Gate{"review": {
					Type: workflow.GateTypeHuman, OutputPorts: []identity.PortName{"result"},
				}},
			},
			"unrelated": {},
		},
		Bindings: []workflow.Binding{
			{From: workflow.PortRef{Port: "task"}, To: workflow.PortRef{StageID: "first", Port: "task"}},
			{From: workflow.PortRef{StageID: "first", Port: "result"}, To: workflow.PortRef{StageID: "second", Port: "task"}},
		},
	}
	w.AssignIDs()
	return w
}

func TestDefinitionUsesMapKeysAsIdentity(t *testing.T) {
	data := []byte(`{
		"id": "example",
		"artifact_specs": {"document": {"type": "md"}},
		"stage_definitions": {
			"design": {
				"outputs": {"architecture": "document"},
				"gates": {
					"review": {"type": "human", "output_ports": ["architecture"]}
				}
			}
		},
		"outputs": {"result": {"stage_id": "design", "port": "architecture"}}
	}`)
	var w workflow.Workflow
	if err := json.Unmarshal(data, &w); err != nil {
		t.Fatal(err)
	}
	w.AssignIDs()
	if err := w.Validate(); err != nil {
		t.Fatalf("definition without duplicate IDs must be valid: %v", err)
	}
	if got := w.RequiredStageIDs(); !reflect.DeepEqual(got, []identity.StageID{"design"}) {
		t.Fatalf("stage identity was not resolved from its key: %v", got)
	}
	if w.StageDefinitions["design"].ID != "design" {
		t.Fatal("stage ID was not populated from its key")
	}
	if w.StageDefinitions["design"].Gates["review"].ID != "review" {
		t.Fatal("gate ID was not populated from its key")
	}
	if w.ArtifactSpecs["document"].ID != "document" {
		t.Fatal("artifact spec ID was not populated from its key")
	}
}
