package app

import "ai-workflow/internal/domain"

type DeclarationStore interface {
	readDeclaration(id string) (*domain.Workflow, error)
}
