package domain

type StageType string

type Stage struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	InArtifacts  []ArtifactDeclaration `json:"inArtifacts"`
	OutArtifacts []ArtifactDeclaration `json:"outArtifacts"`
	Gates        map[string]Gate       `json:"gates"`
}
