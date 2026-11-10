package run

// Feedback selects relevant evidence instead of including the entire history.
type Feedback struct {
	Reason      string          `json:"reason"`
	GateResults []GateResultRef `json:"gate_results,omitempty"`
}
