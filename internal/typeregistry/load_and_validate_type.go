package typeregistry

type EnumType int

const (
	EnumTypeInt EnumType = iota
	EnumTypeString
	EnumTypeIota
)

type EnumField struct {
	TypeSpec
	Type        EnumType
	Description string
	Values      []string
}
