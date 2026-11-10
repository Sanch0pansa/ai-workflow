package workflow_test

import (
	"testing"

	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

func TestInvalidDefinitions(t *testing.T) {
	cases := map[string]func(*workflow.Workflow){
		"no outputs":           func(w *workflow.Workflow) { w.Outputs = nil },
		"unknown source stage": func(w *workflow.Workflow) { w.Bindings[1].From.StageID = "missing" },
		"unknown source port":  func(w *workflow.Workflow) { w.Bindings[1].From.Port = "missing" },
		"unknown target port":  func(w *workflow.Workflow) { w.Bindings[1].To.Port = "missing" },
		"duplicate binding":    func(w *workflow.Workflow) { w.Bindings = append(w.Bindings, w.Bindings[1]) },
		"missing binding":      func(w *workflow.Workflow) { w.Bindings = w.Bindings[:1] },
		"cycle": func(w *workflow.Workflow) {
			w.Bindings[0].From = workflow.PortRef{StageID: "second", Port: "result"}
		},
		"unknown spec": func(w *workflow.Workflow) { w.Inputs["task"] = "missing" },
		"empty stage key": func(w *workflow.Workflow) {
			w.StageDefinitions[""] = workflow.StageDefinition{}
		},
		"empty gate key": func(w *workflow.Workflow) {
			w.StageDefinitions["second"].Gates[""] = workflow.Gate{
				Type: workflow.GateTypeHuman, OutputPorts: []identity.PortName{"result"},
			}
		},
		"empty spec key": func(w *workflow.Workflow) {
			w.ArtifactSpecs[""] = artifact.ArtifactSpec{ID: "other", Type: artifact.ArtifactTypeMD}
		},
		"mismatched specs": func(w *workflow.Workflow) {
			w.ArtifactSpecs["other"] = artifact.ArtifactSpec{ID: "other", Type: artifact.ArtifactTypeMD}
			w.StageDefinitions["second"].Inputs["task"] = "other"
		},
		"negative attempt limit": func(w *workflow.Workflow) {
			s := w.StageDefinitions["first"]
			s.AttemptLimit = -1
			w.StageDefinitions["first"] = s
		},
		"unknown gate output": func(w *workflow.Workflow) {
			w.StageDefinitions["second"].Gates["review"] = workflow.Gate{ID: "review", Type: workflow.GateTypeHuman, OutputPorts: []identity.PortName{"missing"}}
		},
		"duplicate gate output": func(w *workflow.Workflow) {
			w.StageDefinitions["second"].Gates["review"] = workflow.Gate{ID: "review", Type: workflow.GateTypeHuman, OutputPorts: []identity.PortName{"result", "result"}}
		},
		"invalid gate type": func(w *workflow.Workflow) {
			w.StageDefinitions["second"].Gates["review"] = workflow.Gate{ID: "review", Type: "missing", OutputPorts: []identity.PortName{"result"}}
		},
		"revisit to unrelated stage": func(w *workflow.Workflow) {
			s := w.StageDefinitions["second"]
			s.AllowedRevisits = []identity.StageID{"unrelated"}
			w.StageDefinitions["second"] = s
		},
		"revisit to self": func(w *workflow.Workflow) {
			s := w.StageDefinitions["second"]
			s.AllowedRevisits = []identity.StageID{"second"}
			w.StageDefinitions["second"] = s
		},
		"workflow output bypasses stage": func(w *workflow.Workflow) { w.Outputs["result"] = workflow.PortRef{Port: "task"} },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			w := definition()
			change(w)
			if err := w.Validate(); err == nil {
				t.Fatal("invalid definition was accepted")
			}
		})
	}
}
