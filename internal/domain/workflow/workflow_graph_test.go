package workflow_test

import (
	"reflect"
	"slices"
	"testing"

	"ai-workflow/internal/domain/identity"
)

func TestGraphDerivedFromPortBindings(t *testing.T) {
	w := definition()
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := w.DependencyStageIDs("second"); !reflect.DeepEqual(got, []identity.StageID{"first"}) {
		t.Fatalf("dependencies: %v", got)
	}
	got := w.RequiredStageIDs()
	slices.Sort(got)
	if !reflect.DeepEqual(got, []identity.StageID{"first", "second"}) {
		t.Fatalf("required stages: %v", got)
	}
	if got := w.DescendantStageIDs("first"); !reflect.DeepEqual(got, []identity.StageID{"second"}) {
		t.Fatalf("descendants: %v", got)
	}
}
