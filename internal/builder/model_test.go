package builder_test

import (
	"encoding/json"
	"time"
)

type SomeInterfaceType interface {
	Foo()
	Bar()
	Baz()
}

type MyType struct {
	NormalField   int               `json:"normalField"`
	NormalString  string            `json:"normalString"`
	NormalBool    bool              `json:"normalBool"`
	OverrideField SomeInterfaceType `json:"overrideField"`
	OverrideTime  time.Time         `json:"overrideTime"`
}

func unmarshalSomeInterface(b []byte) (SomeInterfaceType, error) {
	panic("implement me")
}
func unmarshalTime(b []byte) (time.Time, error) {
	panic("implement me")
}

func (t *MyType) UnmarshalJSON(data []byte) error {
	type (
		Alias   MyType
		Wrapper struct {
			*Alias
			OverrideField json.RawMessage `json:"overrideField"`
			OverrideTime  json.RawMessage `json:"overrideTime"`
		}
	)

	var (
		unmarshalled = &Alias{}
		err          error
	)
	var temp = Wrapper{Alias: unmarshalled}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if unmarshalled.OverrideField, err = unmarshalSomeInterface(temp.OverrideField); err != nil {
		return err
	} else if unmarshalled.OverrideTime, err = unmarshalTime(temp.OverrideTime); err != nil {
		return err
	}
	*t = (MyType)(*unmarshalled)

	return nil
}

type InterfaceImpl struct{}

func (i InterfaceImpl) Foo() {}

func (i InterfaceImpl) Bar() {}

func (i InterfaceImpl) Baz() {}

var _ SomeInterfaceType = InterfaceImpl{}
