package domain

type GateType string

const (
	GateTypeCmd   GateType = "cmd"
	GateTypeHuman GateType = "human"
)

type Gate struct {
	ID     string
	Type   GateType
	Name   string
	Params map[string]string
}
