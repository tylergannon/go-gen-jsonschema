package custom_names

// Color represents color choices
type Color int

const (
	ColorRed Color = iota
	ColorGreen
	ColorBlue
	ColorYellow
)

// Theme configuration
type Theme struct {
	Name           string `json:"name"`
	PrimaryColor   Color  `json:"primary_color"`
	SecondaryColor Color  `json:"secondary_color"`
}
