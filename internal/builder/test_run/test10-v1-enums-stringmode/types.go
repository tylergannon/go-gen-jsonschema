package v1_enums_stringmode

//go:generate go run ./gen

type Color int

const (
	ColorRed Color = iota
	ColorGreen
	ColorBlue
)

type Paint struct {
	C Color `json:"c"`
}
