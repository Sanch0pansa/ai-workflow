package app

import (
	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/run"
)

type RunStore interface {
	SaveArtifact(value *artifact.Artifact) error
	ReadArtifact(id identity.ArtifactID) (*artifact.Artifact, error)
	SaveRun(value *run.WorkflowRun) error
	ReadRun(id identity.WorkflowRunID) (*run.WorkflowRun, error)
}
