// Package identity defines distinct string-based identifiers shared by domain
// aggregates. These are defined types, not aliases: mixing them requires an
// explicit conversion. Their JSON and YAML representation remains a string.
package identity

type WorkflowID string
type StageID string
type WorkflowRunID string
type StageAttemptID string
type ArtifactSpecID string
type ArtifactID string
type GateID string
type GateResultID string

// PortName identifies a port within its stage or workflow, not globally.
type PortName string
