// Package run owns execution state for one task. Mutations belong to WorkflowRun;
// callers must serialize them even when agents and checks execute concurrently.
// Exported fields support persistence and inspection, not arbitrary mutation.
package run

import (
	"errors"

	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
)

var ErrRunCompleted = errors.New("completed workflow run cannot be changed; create a new run")

type RunStatus string

const (
	RunStatusRunning RunStatus = "running"
	// Blocked means required work cannot proceed, for example after exhausting attempts.
	RunStatusBlocked   RunStatus = "blocked"
	RunStatusCompleted RunStatus = "completed"
)

type WorkflowRun struct {
	ID         identity.WorkflowRunID                      `json:"id"`
	WorkflowID identity.WorkflowID                         `json:"workflow_id"`
	Status     RunStatus                                   `json:"status"`
	Inputs     map[identity.PortName]identity.ArtifactID   `json:"inputs"`
	InputSpecs map[identity.PortName]artifact.ArtifactSpec `json:"input_specs"`
	Outputs    map[identity.PortName]identity.ArtifactID   `json:"outputs"`
	StageRuns  map[identity.StageID]*StageRun              `json:"stages"`
	Artifacts  map[identity.ArtifactID]artifact.Artifact   `json:"artifacts"`
}
