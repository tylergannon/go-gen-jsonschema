//go:generate gen-jsonschema
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
