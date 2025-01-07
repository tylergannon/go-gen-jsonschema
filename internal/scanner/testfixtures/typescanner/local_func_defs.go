package typescanner

type (
	NiceEnumType    string
	MarkerInterface interface {
		coolio()
	}
	TypeForSchemaMethod        []int
	PointerTypeForSchemaMethod struct{}

	TypeForSchemaFunction []string

	PointerTypeForSchemaFunction struct{}

	Type001 struct{}
	Type002 struct{}
	Type003 struct{}
	Type004 struct{}
)

func (t Type001) coolio()  {}
func (t Type002) coolio()  {}
func (t *Type003) coolio() {}
func (t *Type004) coolio() {}

var (
	_ MarkerInterface = Type001{}
	_ MarkerInterface = Type002{}
	_ MarkerInterface = &Type003{}
	_ MarkerInterface = &Type004{}
)

const (
	Val1 NiceEnumType = "val1"
	Val2 NiceEnumType = "val2"
	Val3 NiceEnumType = "val3"
	Val4 NiceEnumType = "val4"
	Val5 NiceEnumType = "val5"
)
