//go:generate go run ../../polytype/
package template_rendering

// Status is a simple string enum
type Status string

const (
	StatusPending  Status = "pending"
	StatusActive   Status = "active"
	StatusComplete Status = "complete"
)

// WorkItem demonstrates field-level enum configuration
type WorkItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status Status `json:"status"`
}

// Status declares itself as an enum; the generator emits its typed constants.
//
//lint:ignore U1000 enum marker method, read by the polytype generator
func (Status) enum() {}
