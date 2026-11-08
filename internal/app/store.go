package app

import "ai-workflow/internal/domain"

type Store interface {
	saveArtifact(artifact *domain.Artifact) error
	readArtifact(id string) (*domain.Artifact, error)
	saveRun(run *domain.Run) error
	readRun(id string) (*domain.Run, error)
}
