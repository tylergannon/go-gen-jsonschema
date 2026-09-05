package enums_stringmode

//go:generate go run ../../polytype/

type Color int

const (
	ColorRed Color = iota
	ColorGreen
	ColorBlue
)

type Paint struct {
	C Color `json:"c"`
}
