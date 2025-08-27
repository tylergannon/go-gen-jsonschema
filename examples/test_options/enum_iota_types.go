package test_options

import "fmt"

// Color represents an iota-based enum for colors
type Color int

const (
	ColorRed Color = iota
	ColorGreen
	ColorBlue
	ColorYellow
	ColorPurple
)

// Implement Stringer for Color
func (c Color) String() string {
	switch c {
	case ColorRed:
		return "red"
	case ColorGreen:
		return "green"
	case ColorBlue:
		return "blue"
	case ColorYellow:
		return "yellow"
	case ColorPurple:
		return "purple"
	default:
		return fmt.Sprintf("Color(%d)", c)
	}
}

// Size represents an iota enum with custom values
type Size int

const (
	SizeSmall  Size = 10
	SizeMedium Size = 20
	SizeLarge  Size = 30
	SizeXLarge Size = 40
)

func (s Size) String() string {
	switch s {
	case SizeSmall:
		return "small"
	case SizeMedium:
		return "medium"
	case SizeLarge:
		return "large"
	case SizeXLarge:
		return "x-large"
	default:
		return fmt.Sprintf("Size(%d)", s)
	}
}

// LogLevel represents logging levels
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
	LogFatal
)

// String returns the string representation
func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[l]
}

// Product demonstrates iota enums with Stringer
type Product struct {
	Name     string   `json:"name"`
	Color    Color    `json:"color"`
	Size     Size     `json:"size"`
	LogLevel LogLevel `json:"log_level"`
}

// Configuration shows field-level enum config with iota
type Configuration struct {
	AppName  string   `json:"app_name"`
	LogLevel LogLevel `json:"log_level"`
	Theme    Color    `json:"theme"`
}
