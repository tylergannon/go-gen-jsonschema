package iota_global

//go:generate go run ../../polytype/

// Priority represents task priority levels using iota
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityUrgent
)

// Task uses an iota-based enum
type Task struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Priority Priority `json:"priority"`
}

// Priority declares itself as an enum; the generator emits its typed constants.
func (Priority) enum() {}
