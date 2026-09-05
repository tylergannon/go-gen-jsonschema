package enums_stringmode

//go:generate go run ../../gen-jsonschema/

type Color int

const (
	ColorRed Color = iota
	ColorGreen
	ColorBlue
)

type Paint struct {
	C Color `json:"c"`
}
