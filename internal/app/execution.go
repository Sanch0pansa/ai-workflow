package app

import (
	"context"
	"fmt"

	"ai-workflow/internal/domain/identity"
	"ai-workflow/internal/domain/run"
	"ai-workflow/internal/domain/workflow"

	"golang.org/x/sync/errgroup"
)

type StagesExecutionService struct {
	declarationStore DeclarationStore
	runStore         RunStore
	gateExecutors    map[workflow.GateType]GateExecutor
}

func NewExecuteCurrentStage(
	declarationStore DeclarationStore,
	runStore RunStore,
	gateExecutors map[workflow.GateType]GateExecutor,
) *StagesExecutionService {
	return &StagesExecutionService{
		declarationStore: declarationStore,
		runStore:         runStore,
		gateExecutors:    gateExecutors,
	}
}

func (e *StagesExecutionService) Execute(
	ctx context.Context,
	workflowID identity.WorkflowID,
	runID identity.WorkflowRunID,
) error {
	w, err := e.declarationStore.ReadDeclaration(workflowID)
	if err != nil {
		return err
	}
	r, err := e.runStore.ReadRun(runID)
	if err != nil {
		return err
	}
	stageIDs, err := r.ReadyStages(w)
	if err != nil {
		return err
	}
	if err := e.runStore.SaveRun(r); err != nil {
		return err
	}
	return e.executeStages(ctx, w, r, stageIDs)
}

func (e *StagesExecutionService) executeStages(
	ctx context.Context,
	w *workflow.Workflow,
	r *run.WorkflowRun,
	stageIDs []identity.StageID,
) error {
	g, stagesCtx := errgroup.WithContext(ctx)
	for _, stageID := range stageIDs {
		stage := w.StageDefinitions[stageID]
		g.Go(func() error {
			return e.executeStage(stagesCtx, stage, r)
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("execute stages: %w", err)
	}
	return nil
}

func (e *StagesExecutionService) executeStage(
	_ context.Context,
	stage workflow.StageDefinition,
	_ *run.WorkflowRun,
) error {
	// Agent execution and coordination of domain transitions are not implemented.
	return fmt.Errorf("stage %q: agent execution is not implemented", stage.ID)
}
