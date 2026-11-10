package app

import (
	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/workflow"
)

type DeclarationStore interface {
	ReadDeclaration(id identity.WorkflowID) (*workflow.Workflow, error)
}
