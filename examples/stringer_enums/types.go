package stringer_enums

import "fmt"

//go:generate gen-jsonschema

// LogLevel represents the severity of log messages.
// This is an integer-based enum with a Stringer implementation.
type LogLevel int

const (
	// LogDebug is for detailed diagnostic information
	LogDebug LogLevel = iota
	// LogInfo is for general informational messages
	LogInfo
	// LogWarning is for warning messages
	LogWarning
	// LogError is for error messages
	LogError
	// LogFatal is for fatal errors that cause termination
	LogFatal
)

// String implements the Stringer interface for LogLevel
func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarning:
		return "WARNING"
	case LogError:
		return "ERROR"
	case LogFatal:
		return "FATAL"
	default:
		return fmt.Sprintf("LogLevel(%d)", l)
	}
}

// Priority represents task priority levels with custom integer values
type Priority int

const (
	PriorityLow    Priority = 100
	PriorityNormal Priority = 200
	PriorityHigh   Priority = 300
	PriorityUrgent Priority = 400
)

// String implements the Stringer interface for Priority
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	default:
		return fmt.Sprintf("Priority(%d)", p)
	}
}

// ApplicationConfig demonstrates using Stringer enums in a struct.
// When WithStringerEnum is used, the JSON schema will use the string
// representations from the String() methods rather than the integer values.
type ApplicationConfig struct {
	// AppName is the name of the application
	AppName string `json:"app_name"`

	// LogLevel controls the verbosity of logging
	LogLevel LogLevel `json:"log_level"`

	// DefaultPriority is the default priority for new tasks
	DefaultPriority Priority `json:"default_priority"`

	// MaxConnections is the maximum number of concurrent connections
	MaxConnections int `json:"max_connections"`
}

// Task demonstrates another struct using the Stringer enums
type Task struct {
	// ID is the unique task identifier
	ID string `json:"id"`

	// Name is the task name
	Name string `json:"name"`

	// Priority determines the task's importance
	Priority Priority `json:"priority"`

	// LogLevel is the minimum log level for this task
	LogLevel LogLevel `json:"log_level"`
}
