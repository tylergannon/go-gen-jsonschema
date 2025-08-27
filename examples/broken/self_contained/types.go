package self_contained

// Priority for tasks
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Severity for issues
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Issue with multiple enum fields
type Issue struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Severity    Severity `json:"severity"`
}
