package run_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"

	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/run"
	"ai-workflow/internal/domain/workflow"
)

func exampleWorkflow() *workflow.Workflow {
	w := &workflow.Workflow{
		ID: "development",
		ArtifactSpecs: map[identity.ArtifactSpecID]artifact.ArtifactSpec{
			"document": {Type: artifact.ArtifactTypeMD},
			"code":     {Type: artifact.ArtifactTypeCode},
		},
		Inputs: map[identity.PortName]identity.ArtifactSpecID{"task": "document"},
		Outputs: map[identity.PortName]workflow.PortRef{
			"result": {StageID: "deliver", Port: "result"},
		},
		StageDefinitions: map[identity.StageID]workflow.StageDefinition{
			"plan": {
				Inputs:  map[identity.PortName]identity.ArtifactSpecID{"task": "document"},
				Outputs: map[identity.PortName]identity.ArtifactSpecID{"architecture": "document"},
			},
			"implement": {
				Inputs:          map[identity.PortName]identity.ArtifactSpecID{"architecture": "document"},
				Outputs:         map[identity.PortName]identity.ArtifactSpecID{"code": "code", "notes": "document"},
				AllowedRevisits: []identity.StageID{"plan"},
				Gates: map[identity.GateID]workflow.Gate{
					"tests":  {Type: workflow.GateTypeCmd, OutputPorts: []identity.PortName{"code"}},
					"review": {ID: "review", Type: workflow.GateTypeHuman, OutputPorts: []identity.PortName{"notes"}},
				},
			},
			"deliver": {
				Inputs:  map[identity.PortName]identity.ArtifactSpecID{"code": "code"},
				Outputs: map[identity.PortName]identity.ArtifactSpecID{"result": "code"},
			},
			// This stage is intentionally outside the workflow output dependency graph.
			"unrelated": {Outputs: map[identity.PortName]identity.ArtifactSpecID{"report": "document"}},
		},
		Bindings: []workflow.Binding{
			{From: workflow.PortRef{Port: "task"}, To: workflow.PortRef{StageID: "plan", Port: "task"}},
			{From: workflow.PortRef{StageID: "plan", Port: "architecture"}, To: workflow.PortRef{StageID: "implement", Port: "architecture"}},
			{From: workflow.PortRef{StageID: "implement", Port: "code"}, To: workflow.PortRef{StageID: "deliver", Port: "code"}},
		},
	}
	w.AssignIDs()
	return w
}

func value(id identity.ArtifactID, spec identity.ArtifactSpecID) artifact.Artifact {
	return artifact.Artifact{ID: id, ArtifactSpecID: spec, Content: "content of " + string(id)}
}

func newRun(t *testing.T, w *workflow.Workflow) *run.WorkflowRun {
	t.Helper()
	r, err := run.NewWorkflowRun("task-1", w, map[identity.PortName]artifact.Artifact{"task": value("task-description", "document")})
	must(t, err)
	return r
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func start(t *testing.T, w *workflow.Workflow, r *run.WorkflowRun, stage identity.StageID, attempt identity.StageAttemptID) run.StageAttempt {
	t.Helper()
	a, err := r.StartAttempt(w, stage, attempt)
	must(t, err)
	return a
}

func submitPlan(t *testing.T, w *workflow.Workflow, r *run.WorkflowRun, attempt identity.StageAttemptID, output identity.ArtifactID) {
	t.Helper()
	start(t, w, r, "plan", attempt)
	must(t, r.SubmitOutputs(w, "plan", attempt, map[identity.PortName]artifact.Artifact{"architecture": value(output, "document")}))
}

func submitImplementation(t *testing.T, w *workflow.Workflow, r *run.WorkflowRun, attempt identity.StageAttemptID, suffix string) {
	t.Helper()
	start(t, w, r, "implement", attempt)
	must(t, r.SubmitOutputs(w, "implement", attempt, map[identity.PortName]artifact.Artifact{
		"code": value(identity.ArtifactID("code-"+suffix), "code"), "notes": value(identity.ArtifactID("notes-"+suffix), "document"),
	}))
}

func gateResult(id identity.GateResultID, gate identity.GateID, status run.GateStatus, artifactID identity.ArtifactID) run.GateResult {
	return run.GateResult{ID: id, GateID: gate, Status: status, ArtifactIDs: []identity.ArtifactID{artifactID}, Content: "feedback: " + string(id)}
}

func TestAttemptUsesCurrentInputsPreviousOutputsAndSelectedFeedback(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	submitPlan(t, w, r, "plan-1", "architecture-1")
	submitImplementation(t, w, r, "implement-1", "1")
	must(t, r.RecordGateResult(w, "implement", "implement-1", gateResult("tests-ok", "tests", run.GateStatusApproved, "code-1")))
	must(t, r.RecordGateResult(w, "implement", "implement-1", gateResult("review-rejected", "review", run.GateStatusRejected, "notes-1")))

	a := start(t, w, r, "implement", "implement-2")
	if a.Inputs["architecture"] != "architecture-1" || a.PreviousOutputs["code"] != "code-1" || a.PreviousAttemptID != "implement-1" {
		t.Fatalf("unexpected retry context: %+v", a)
	}
	wantEvidence := []run.GateResultRef{{StageID: "implement", AttemptID: "implement-1", ResultID: "review-rejected"}}
	if !reflect.DeepEqual(a.Feedback.GateResults, wantEvidence) {
		t.Fatalf("feedback must select only rejected evidence: %+v", a.Feedback)
	}
	if len(r.StageRuns["implement"].Attempts[0].GateResults) != 2 {
		t.Fatal("previous check history was lost")
	}
	// Modifying a returned context must not alter inputs or definitions in history.
	a.Inputs["architecture"] = "wrong"
	a.PreviousOutputs["code"] = "wrong"
	gate := a.Definition.Gates["tests"]
	gate.OutputPorts[0] = "notes"
	stored := r.StageRuns["implement"].Attempts[1]
	if stored.Inputs["architecture"] != "architecture-1" || stored.PreviousOutputs["code"] != "code-1" || stored.Definition.Gates["tests"].OutputPorts[0] != "code" {
		t.Fatal("returned context aliases aggregate state")
	}
	must(t, r.SubmitOutputs(w, "implement", "implement-2", map[identity.PortName]artifact.Artifact{
		"code": value("code-2", "code"), "notes": value("notes-2", "document"),
	}))
	must(t, r.RecordGateResult(w, "implement", "implement-2", gateResult("review-ok-2", "review", run.GateStatusApproved, "notes-2")))
	if r.StageRuns["implement"].Status != run.StageStatusChecking {
		t.Fatal("previous attempt's test approval was reused for new artifacts")
	}
	if err := r.RecordGateResult(w, "implement", "implement-2", gateResult("stale-tests", "tests", run.GateStatusApproved, "code-1")); err == nil {
		t.Fatal("new attempt accepted approval of old artifacts")
	}
	must(t, r.RecordGateResult(w, "implement", "implement-2", gateResult("tests-ok-2", "tests", run.GateStatusApproved, "code-2")))
}

func TestChecksUseSelectedArtifactsAndTechnicalFailureDoesNotRetryStage(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	submitPlan(t, w, r, "plan-1", "architecture-1")
	submitImplementation(t, w, r, "implement-1", "1")
	if err := r.RecordGateResult(w, "implement", "implement-1", gateResult("wrong", "tests", run.GateStatusApproved, "notes-1")); err == nil {
		t.Fatal("check approved the wrong artifact")
	}
	failed := gateResult("tests-error", "tests", run.GateStatusError, "code-1")
	must(t, r.RecordGateResult(w, "implement", "implement-1", failed))
	failed.ArtifactIDs[0] = "mutated"
	if r.StageRuns["implement"].Attempts[0].GateResults[0].ArtifactIDs[0] != "code-1" {
		t.Fatal("check evidence aliases caller's slice")
	}
	if _, err := r.StartAttempt(w, "implement", "unexpected-retry"); err == nil {
		t.Fatal("technical check failure must not start a new stage attempt")
	}
	must(t, r.RecordGateResult(w, "implement", "implement-1", gateResult("tests-ok", "tests", run.GateStatusApproved, "code-1")))
	if r.StageRuns["implement"].Status != run.StageStatusChecking {
		t.Fatal("stage completed without human review")
	}
	must(t, r.RecordGateResult(w, "implement", "implement-1", gateResult("review-pending", "review", run.GateStatusPending, "notes-1")))
	must(t, r.RecordGateResult(w, "implement", "implement-1", gateResult("review-ok", "review", run.GateStatusApproved, "notes-1")))
	if r.StageRuns["implement"].Status != run.StageStatusCompleted || len(r.StageRuns["implement"].Attempts) != 1 {
		t.Fatal("expected approval of the original attempt")
	}
}

func TestRevisitPreservesHistoryAndResolvesNewInputs(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	submitPlan(t, w, r, "plan-1", "architecture-1")
	submitImplementation(t, w, r, "implement-1", "1")
	must(t, r.RecordGateResult(w, "implement", "implement-1", gateResult("review-rejected", "review", run.GateStatusRejected, "notes-1")))
	feedback := run.Feedback{
		Reason:      "architecture needs revision",
		GateResults: []run.GateResultRef{{StageID: "implement", AttemptID: "implement-1", ResultID: "review-rejected"}},
	}
	if err := r.Revisit(w, "implement", "plan", workflow.GateTypeCmd, feedback); err == nil {
		t.Fatal("a deterministic check must not request a revisit")
	}
	if err := r.Revisit(w, "deliver", "plan", workflow.GateTypeHuman, feedback); err == nil {
		t.Fatal("undeclared revisit was allowed")
	}
	bad := run.Feedback{Reason: "bad", GateResults: []run.GateResultRef{{StageID: "missing"}}}
	if err := r.Revisit(w, "implement", "plan", workflow.GateTypeHuman, bad); err == nil {
		t.Fatal("unknown evidence was accepted")
	}
	must(t, r.Revisit(w, "implement", "plan", workflow.GateTypeAgent, feedback))
	for _, stageID := range []identity.StageID{"plan", "implement", "deliver"} {
		if r.StageRuns[stageID].Status != run.StageStatusInvalidated {
			t.Fatalf("stage %q was not invalidated", stageID)
		}
	}
	if r.StageRuns["plan"].Attempts[0].Status != run.AttemptStatusApproved {
		t.Fatal("invalidation rewrote a historical verdict")
	}
	plan := start(t, w, r, "plan", "plan-2")
	if plan.PreviousOutputs["architecture"] != "architecture-1" || !reflect.DeepEqual(plan.Feedback, feedback) {
		t.Fatalf("revisit context: %+v", plan)
	}
	must(t, r.SubmitOutputs(w, "plan", "plan-2", map[identity.PortName]artifact.Artifact{"architecture": value("architecture-2", "document")}))
	implementation := start(t, w, r, "implement", "implement-2")
	if implementation.Inputs["architecture"] != "architecture-2" || implementation.PreviousOutputs["code"] != "code-1" {
		t.Fatalf("expected new upstream input and old own output: %+v", implementation)
	}
	if r.StageRuns["implement"].Attempts[0].Inputs["architecture"] != "architecture-1" {
		t.Fatal("old attempt's inputs were replaced")
	}
	if err := r.RecordGateResult(w, "implement", "implement-1", gateResult("late", "tests", run.GateStatusApproved, "code-1")); err == nil {
		t.Fatal("stale check result was accepted")
	}
}

func TestDefinitionEditInvalidatesRunningConsumersAndPreservesDefinitionHistory(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	submitPlan(t, w, r, "plan-1", "architecture-1")
	start(t, w, r, "implement", "implement-1")
	plan := w.StageDefinitions["plan"]
	plan.Description = "new architecture requirements"
	w.StageDefinitions["plan"] = plan
	must(t, r.Reconcile(w))
	if r.StageRuns["plan"].Status != run.StageStatusInvalidated || r.StageRuns["implement"].Status != run.StageStatusInvalidated {
		t.Fatal("definition edit did not invalidate consumers")
	}
	if r.StageRuns["plan"].Attempts[0].Definition.Description != "" {
		t.Fatal("historical definition changed")
	}
	if r.StageRuns["implement"].Attempts[0].Status != run.AttemptStatusInvalidated {
		t.Fatal("superseded work was left running in domain state")
	}
	if err := r.SubmitOutputs(w, "implement", "implement-1", map[identity.PortName]artifact.Artifact{
		"code": value("late-code", "code"), "notes": value("late-notes", "document"),
	}); err == nil {
		t.Fatal("late output from invalidated work was accepted")
	}
	if _, exists := r.Artifacts["late-code"]; exists {
		t.Fatal("rejected output was stored")
	}
}

func TestCompletionDependsOnDeclaredOutputsAndIsTerminal(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	ready, err := r.ReadyStages(w)
	must(t, err)
	if !reflect.DeepEqual(ready, []identity.StageID{"plan"}) {
		t.Fatalf("unrelated stage must not be scheduled: %v", ready)
	}
	submitPlan(t, w, r, "plan-1", "architecture-1")
	submitImplementation(t, w, r, "implement-1", "1")
	must(t, r.RecordGateResult(w, "implement", "implement-1", gateResult("tests-ok", "tests", run.GateStatusApproved, "code-1")))
	must(t, r.RecordGateResult(w, "implement", "implement-1", gateResult("review-ok", "review", run.GateStatusApproved, "notes-1")))
	start(t, w, r, "deliver", "deliver-1")
	must(t, r.SubmitOutputs(w, "deliver", "deliver-1", map[identity.PortName]artifact.Artifact{"result": value("result-1", "code")}))
	if r.Status != run.RunStatusCompleted || r.Outputs["result"] != "result-1" || r.StageRuns["unrelated"].Status != run.StageStatusPending {
		t.Fatalf("unexpected completion: %+v", r)
	}
	if err := r.Revisit(w, "implement", "plan", workflow.GateTypeHuman, run.Feedback{Reason: "new task"}); !errors.Is(err, run.ErrRunCompleted) {
		t.Fatalf("completed run reopened: %v", err)
	}
	w.StageDefinitions["plan"] = workflow.StageDefinition{}
	if err := r.Reconcile(w); !errors.Is(err, run.ErrRunCompleted) {
		t.Fatalf("completed history changed after workflow edit: %v", err)
	}
}

func TestAttemptLimitBlocksRunAndCanBeChangedInDefinition(t *testing.T) {
	w := exampleWorkflow()
	plan := w.StageDefinitions["plan"]
	plan.AttemptLimit = 1
	w.StageDefinitions["plan"] = plan
	r := newRun(t, w)
	start(t, w, r, "plan", "plan-1")
	must(t, r.FailAttempt(w, "plan", "plan-1", "agent stopped"))
	if r.Status != run.RunStatusBlocked || r.StageRuns["plan"].Status != run.StageStatusFailed {
		t.Fatal("exhausted attempts should block the run")
	}
	if _, err := r.StartAttempt(w, "plan", "plan-2"); err == nil {
		t.Fatal("attempt limit was ignored")
	}
	plan.AttemptLimit = 2
	w.StageDefinitions["plan"] = plan
	start(t, w, r, "plan", "plan-2")
}

func TestArtifactsCannotBeOverwrittenAndPartialOutputsAreNotStored(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	start(t, w, r, "plan", "plan-1")
	if err := r.SubmitOutputs(w, "plan", "plan-1", map[identity.PortName]artifact.Artifact{"architecture": value("task-description", "document")}); err == nil {
		t.Fatal("existing artifact ID was reused")
	}
	must(t, r.SubmitOutputs(w, "plan", "plan-1", map[identity.PortName]artifact.Artifact{"architecture": value("architecture-1", "document")}))
	start(t, w, r, "implement", "implement-1")
	if err := r.SubmitOutputs(w, "implement", "implement-1", map[identity.PortName]artifact.Artifact{"code": value("code-1", "code")}); err == nil {
		t.Fatal("missing output was accepted")
	}
	if _, exists := r.Artifacts["code-1"]; exists {
		t.Fatal("partial submission mutated the artifact collection")
	}
	if r.Artifacts["architecture-1"].ProducerAttemptID != "plan-1" || r.Artifacts["architecture-1"].ProducerRunID != r.ID {
		t.Fatal("artifact provenance was not recorded")
	}
}

func TestPersistedRunCanContinue(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	submitPlan(t, w, r, "plan-1", "architecture-1")
	submitImplementation(t, w, r, "implement-1", "1")
	data, err := json.Marshal(r)
	must(t, err)
	var restored run.WorkflowRun
	must(t, json.Unmarshal(data, &restored))
	must(t, restored.RecordGateResult(w, "implement", "implement-1", gateResult("tests-ok", "tests", run.GateStatusApproved, "code-1")))
	must(t, restored.RecordGateResult(w, "implement", "implement-1", gateResult("review-ok", "review", run.GateStatusApproved, "notes-1")))
	ready, err := restored.ReadyStages(w)
	must(t, err)
	if !reflect.DeepEqual(ready, []identity.StageID{"deliver"}) {
		t.Fatalf("restored run lost readiness: %v", ready)
	}
}

func TestIncompatibleDefinitionFailsWithoutDroppingHistory(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	submitPlan(t, w, r, "plan-1", "architecture-1")
	delete(w.StageDefinitions, "unrelated")
	if err := r.Reconcile(w); err == nil {
		t.Fatal("removing a known stage must fail explicitly")
	}
	if r.StageRuns["plan"].Status != run.StageStatusCompleted || len(r.Artifacts) != 2 {
		t.Fatal("failed reconciliation changed execution history")
	}
}

func TestEditingWorkflowOutputsCanCompleteRunButCannotAcceptLateWork(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	submitPlan(t, w, r, "plan-1", "architecture-1")
	start(t, w, r, "implement", "implement-1")
	w.Outputs = map[identity.PortName]workflow.PortRef{"result": {StageID: "plan", Port: "architecture"}}
	err := r.SubmitOutputs(w, "implement", "implement-1", map[identity.PortName]artifact.Artifact{
		"code": value("late-code", "code"), "notes": value("late-notes", "document"),
	})
	if !errors.Is(err, run.ErrRunCompleted) || r.Status != run.RunStatusCompleted || r.Outputs["result"] != "architecture-1" {
		t.Fatalf("expected terminal completion under edited workflow: %v, %+v", err, r)
	}
	if _, exists := r.Artifacts["late-code"]; exists {
		t.Fatal("late artifact was stored after completion")
	}
}

func TestParallelBranchesWaitForAllInputsAndInvalidateOnlyConsumers(t *testing.T) {
	w := exampleWorkflow()
	w.StageDefinitions["qa"] = workflow.StageDefinition{
		ID:     "qa",
		Inputs: map[identity.PortName]identity.ArtifactSpecID{"architecture": "document"}, Outputs: map[identity.PortName]identity.ArtifactSpecID{"report": "document"},
	}
	w.StageDefinitions["deliver"].Inputs["report"] = "document"
	w.Bindings = append(w.Bindings,
		workflow.Binding{From: workflow.PortRef{StageID: "plan", Port: "architecture"}, To: workflow.PortRef{StageID: "qa", Port: "architecture"}},
		workflow.Binding{From: workflow.PortRef{StageID: "qa", Port: "report"}, To: workflow.PortRef{StageID: "deliver", Port: "report"}},
	)
	r := newRun(t, w)
	submitPlan(t, w, r, "plan-1", "architecture-1")
	ready, err := r.ReadyStages(w)
	must(t, err)
	slices.Sort(ready)
	if !reflect.DeepEqual(ready, []identity.StageID{"implement", "qa"}) {
		t.Fatalf("expected parallel branches: %v", ready)
	}
	start(t, w, r, "qa", "qa-1")
	submitImplementation(t, w, r, "implement-1", "1")
	must(t, r.SubmitOutputs(w, "qa", "qa-1", map[identity.PortName]artifact.Artifact{"report": value("report-1", "document")}))
	if _, err := r.StartAttempt(w, "deliver", "too-early"); err == nil {
		t.Fatal("join stage ran before all input producers were approved")
	}
	implementation := w.StageDefinitions["implement"]
	implementation.Description = "updated implementation instructions"
	w.StageDefinitions["implement"] = implementation
	must(t, r.Reconcile(w))
	if r.StageRuns["qa"].Status != run.StageStatusCompleted || r.StageRuns["plan"].Status != run.StageStatusCompleted {
		t.Fatal("unaffected branch or ancestor was invalidated")
	}
	if r.StageRuns["implement"].Status != run.StageStatusInvalidated || r.StageRuns["deliver"].Status != run.StageStatusInvalidated {
		t.Fatal("edited branch and its consumer were not invalidated")
	}
}

func TestNewTaskImportsPreviousResultsAsInputsNotOwnOutputs(t *testing.T) {
	w := exampleWorkflow()
	imported := value("previous-task-result", "document")
	imported.ProducerRunID = "previous-run"
	imported.ProducerAttemptID = "previous-attempt"
	r, err := run.NewWorkflowRun("new-task", w, map[identity.PortName]artifact.Artifact{"task": imported})
	must(t, err)
	a := start(t, w, r, "plan", "plan-1")
	if a.Inputs["task"] != imported.ID || len(a.PreviousOutputs) != 0 || a.PreviousAttemptID != "" {
		t.Fatalf("imported result was confused with a previous own attempt: %+v", a)
	}
	if r.Artifacts[imported.ID].ProducerRunID != "previous-run" {
		t.Fatal("imported provenance was replaced")
	}
}

func TestNewGateRequiresNewApprovalAndKeepsOldDefinition(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	submitPlan(t, w, r, "plan-1", "architecture-1")
	plan := w.StageDefinitions["plan"]
	plan.Gates = map[identity.GateID]workflow.Gate{
		"review": {ID: "review", Type: workflow.GateTypeHuman, OutputPorts: []identity.PortName{"architecture"}},
	}
	w.StageDefinitions["plan"] = plan
	submitPlan(t, w, r, "plan-2", "architecture-2")
	if r.StageRuns["plan"].Status != run.StageStatusChecking {
		t.Fatal("newly added gate was implicitly approved")
	}
	if len(r.StageRuns["plan"].Attempts[0].Definition.Gates) != 0 {
		t.Fatal("adding a gate rewrote a previous attempt")
	}
	must(t, r.RecordGateResult(w, "plan", "plan-2", gateResult("new-review", "review", run.GateStatusApproved, "architecture-2")))
	if r.StageRuns["plan"].Status != run.StageStatusCompleted {
		t.Fatal("new approval did not accept the stage")
	}
}

func TestChangingExternalInputTypeIsIncompatible(t *testing.T) {
	w := exampleWorkflow()
	r := newRun(t, w)
	w.ArtifactSpecs["document"] = artifact.ArtifactSpec{ID: "document", Type: artifact.ArtifactTypeDiff}
	if err := r.Reconcile(w); err == nil {
		t.Fatal("external input interpretation changed without an error")
	}
}
