package stringer_enum

import "fmt"

// LogLevel represents logging levels with Stringer implementation
type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
	Fatal
)

// String implements fmt.Stringer
func (l LogLevel) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	default:
		return fmt.Sprintf("LogLevel(%d)", l)
	}
}

// Config uses a Stringer enum
type Config struct {
	AppName  string   `json:"app_name"`
	LogLevel LogLevel `json:"log_level"`
}
