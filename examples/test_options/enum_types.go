package test_options

import "fmt"

// Status represents a string-based enum
type Status string

const (
	StatusPending  Status = "pending"
	StatusActive   Status = "active"
	StatusComplete Status = "complete"
	StatusCanceled Status = "canceled"
)

// Priority represents a string-based priority enum for comparison
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Severity represents an iota enum with custom string values
type Severity int

const (
	SeverityInfo Severity = iota + 1
	SeverityWarning
	SeverityError
	SeverityCritical
)

// Implement Stringer interface for Severity
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return fmt.Sprintf("severity(%d)", s)
	}
}

// WeekDay represents days of the week with custom iota values
type WeekDay int

const (
	Sunday WeekDay = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// Implement Stringer for WeekDay
func (d WeekDay) String() string {
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if d >= 0 && int(d) < len(days) {
		return days[d]
	}
	return fmt.Sprintf("WeekDay(%d)", d)
}

// Task demonstrates usage of various enum types
type Task struct {
	Title    string   `json:"title"`
	Status   Status   `json:"status"`
	Priority Priority `json:"priority"`
	Severity Severity `json:"severity"`
	DueDay   WeekDay  `json:"due_day"`
}

// WorkItem demonstrates field-level enum configuration (v1 pattern)
type WorkItem struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Priority Priority `json:"priority"`
	Level    Severity `json:"level"`
}
