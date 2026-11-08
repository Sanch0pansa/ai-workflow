package domain

type ArtifactType string

const (
	ArtifactTypeMD   ArtifactType = "md"
	ArtifactTypeDiff ArtifactType = "diff"
	ArtifactTypeJSON ArtifactType = "json"
	ArtifactTypeCode ArtifactType = "code"
)

type ArtifactDeclaration struct {
	ID   string       `json:"id"`
	Type ArtifactType `json:"type"`
}

type Artifact struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}
