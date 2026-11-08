package app

import (
	"ai-workflow/internal/domain"
	"fmt"
)

type ExecuteCurrentStage struct {
	declarationStore DeclarationStore
	store            Store
}

func NewExecuteCurrentStage(declarationStore DeclarationStore, store Store) *ExecuteCurrentStage {
	return &ExecuteCurrentStage{
		declarationStore: declarationStore,
		store:            store,
	}
}

func (e *ExecuteCurrentStage) Execute(workflowID string, runID string) error {
	// Read the workflow declaration
	workflow, err := e.declarationStore.readDeclaration(workflowID)
	if err != nil {
		return err
	}

	// Read the current run
	run, err := e.store.readRun(runID)
	if err != nil {
		return err
	}

	// Get the current stage based on the run's current stage ID
	currentStage, found := workflow.GetStageByID(run.CurrentStageID)
	if !found {
		return fmt.Errorf("current stage with ID %s not found in workflow", run.CurrentStageID)
	}

	// Execute the current stage (this is a placeholder for actual execution logic)
	err = e.executeStage(currentStage, run)
	if err != nil {
		return err
	}

	// Update the run's stage status to completed
	stageRun, exists := run.Stages[run.CurrentStageID]
	if !exists {
		return fmt.Errorf("stage run with ID %s not found in run", run.CurrentStageID)
	}
	stageRun.StageStatus = domain.StageStatusCompleted
	run.Stages[run.CurrentStageID] = stageRun

	// Save the updated run
	err = e.store.saveRun(run)
	if err != nil {
		return err
	}

	return nil
}

func (e *ExecuteCurrentStage) executeStage(stage *domain.Stage, run *domain.Run) error {
	return nil
}
