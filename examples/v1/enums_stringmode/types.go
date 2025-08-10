package enums_stringmode

type Color int

const (
	ColorRed Color = iota
	ColorGreen
	ColorBlue
)

type Paint struct {
	C Color `json:"c"`
}
