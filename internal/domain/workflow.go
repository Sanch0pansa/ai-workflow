package domain

type Workflow struct {
	ID                   string                         `json:"id"`
	Name                 string                         `json:"name"`
	Stages               []Stage                        `json:"stages"`
	ArtifactDeclarations map[string]ArtifactDeclaration `json:"artifacts"`
}

func NewWorkflow(id string, name string, stages []Stage, artifacts map[string]ArtifactDeclaration) *Workflow {
	return &Workflow{
		ID:                   id,
		Name:                 name,
		Stages:               stages,
		ArtifactDeclarations: artifacts,
	}
}

func (w *Workflow) GetStageByID(stageID string) (*Stage, bool) {
	for _, stage := range w.Stages {
		if stage.ID == stageID {
			return &stage, true
		}
	}
	return nil, false
}

func (w *Workflow) GetArtifactDeclarationByID(artifactID string) (*ArtifactDeclaration, bool) {
	artifact, exists := w.ArtifactDeclarations[artifactID]
	if !exists {
		return nil, false
	}
	return &artifact, true
}
