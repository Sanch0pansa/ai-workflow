package domain

type GateStatus string
type StageStatus string

const (
	GateStatusPending  GateStatus = "pending"
	GateStatusApproved GateStatus = "approved"
	GateStatusRejected GateStatus = "rejected"
)

const (
	StageStatusPending   StageStatus = "pending"
	StageStatusRunning   StageStatus = "running"
	StageStatusCompleted StageStatus = "completed"
)

type GateResult struct {
	GateID  string     `json:"gate_id"`
	Status  GateStatus `json:"status"`
	Content string     `json:"content"`
}

type StageRun struct {
	ID          string                `json:"id"`
	StageID     string                `json:"stage_id"`
	StageStatus StageStatus           `json:"stage_status"`
	GateStatus  map[string]GateResult `json:"gate_status"`
}

type Run struct {
	ID             string              `json:"id"`
	WorkflowID     string              `json:"workflow_id"`
	Stages         map[string]StageRun `json:"stages"`
	CurrentStageID string              `json:"current_stage_id"`
	Artifacts      map[string]Artifact `json:"artifacts"`
}
