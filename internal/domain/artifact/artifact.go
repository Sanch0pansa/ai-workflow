// Package artifact describes expected artifacts and concrete, immutable results.
package artifact

import "ai-workflow/internal/domain/identity"

type ArtifactType string

const (
	ArtifactTypeMD   ArtifactType = "md"
	ArtifactTypeDiff ArtifactType = "diff"
	ArtifactTypeJSON ArtifactType = "json"
	ArtifactTypeCode ArtifactType = "code"
)

// ArtifactSpec describes an expected result, not a particular version of it.
type ArtifactSpec struct {
	ID   identity.ArtifactSpecID `json:"id" yaml:"-"`
	Type ArtifactType            `json:"type" yaml:"type"`
}

// Artifact identifies one immutable version. Content is interpreted according to
// the spec's type: inline text, a commit ID, or a path to a preserved file, etc.
// Producer fields identify the originating run and attempt, when known.
// Imported artifacts can retain the producer fields from a previous run.
type Artifact struct {
	ID                identity.ArtifactID     `json:"id" yaml:"id"`
	ArtifactSpecID    identity.ArtifactSpecID `json:"artifact_spec_id" yaml:"artifact_spec_id"`
	Content           string                  `json:"content" yaml:"content"`
	ProducerRunID     identity.WorkflowRunID  `json:"producer_run_id,omitempty" yaml:"producer_run_id,omitempty"`
	ProducerAttemptID identity.StageAttemptID `json:"producer_attempt_id,omitempty" yaml:"producer_attempt_id,omitempty"`
}
