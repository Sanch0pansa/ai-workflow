package app

import (
	"context"

	"ai-workflow/internal/domain/artifact"
	"ai-workflow/internal/domain/run"
	"ai-workflow/internal/domain/workflow"
)

type GateExecutor interface {
	Execute(ctx context.Context, gate workflow.Gate, artifacts []artifact.Artifact) (run.GateResult, error)
}
