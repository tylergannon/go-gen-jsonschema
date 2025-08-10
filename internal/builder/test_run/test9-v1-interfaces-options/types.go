package v1_interfaces_options

//go:generate go run ./gen

type IFace interface{ isIface() }

type Impl1 struct {
	X string `json:"x"`
}

func (Impl1) isIface() {}

type Impl2 struct {
	Y int `json:"y"`
}

func (Impl2) isIface() {}

type Owner struct {
	IF IFace `json:"if"`
}
