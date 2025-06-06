//go:build !jsonschema
// +build !jsonschema

// Code generated by go-gen-jsonschema. DO NOT EDIT.
package messages

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
)

//go:embed jsonschema
var __gen_jsonschema_fs embed.FS

var errNoDiscriminator = errors.New("no discriminator property '!type' found")

func __gen_jsonschema_panic(fname string, err error) {
	panic(fmt.Sprintf("error reading %s from embedded FS: %s", fname, err.Error()))
}

func (ToolFuncGetTypeInfo) Schema() json.RawMessage {
	const fileName = "jsonschema/ToolFuncGetTypeInfo.json"
	data, err := __gen_jsonschema_fs.ReadFile(fileName)
	if err != nil {
		__gen_jsonschema_panic(fileName, err)
	}
	return data
}

func (GeneratedTestResponse) Schema() json.RawMessage {
	const fileName = "jsonschema/GeneratedTestResponse.json"
	data, err := __gen_jsonschema_fs.ReadFile(fileName)
	if err != nil {
		__gen_jsonschema_panic(fileName, err)
	}
	return data
}

// UnmarshalJSON is a generated custom json.Unmarshaler implementation for
// Assertion.
func (a *Assertion) UnmarshalJSON(b []byte) (err error) {
	type Alias Assertion
	type Wrapper struct {
		Alias
		Value json.RawMessage `json:"value"`
	}
	var wrapper Wrapper
	if err = json.Unmarshal(b, &wrapper); err != nil {
		return err
	}
	*a = Assertion(wrapper.Alias)
	if a.Value, err = __jsonUnmarshal__messages__AssertionValue(wrapper.Value); err != nil {
		return err
	}
	return nil
}
func __jsonUnmarshal__messages__AssertionValue(data []byte) (AssertionValue, error) {
	var (
		temp          map[string]json.RawMessage
		discriminator string
		err           = json.Unmarshal(data, &temp)
	)

	if err != nil {
		return nil, err
	} else if _tempDiscriminator, ok := temp["!type"]; !ok {
		return nil, errNoDiscriminator
	} else if err = json.Unmarshal(_tempDiscriminator, &discriminator); err != nil {
		return nil, __jsonschema__unmarshalDiscriminatorError(_tempDiscriminator, err)
	}
	switch discriminator {
	case "AssertNumericValue":
		var obj AssertNumericValue
		if err = json.Unmarshal(data, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	case "AssertStringValue":
		var obj AssertStringValue
		if err = json.Unmarshal(data, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	case "AssertBoolValue":
		var obj AssertBoolValue
		if err = json.Unmarshal(data, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	case "AssertType":
		var obj AssertType
		if err = json.Unmarshal(data, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	case "AssertArrayLength":
		var obj AssertArrayLength
		if err = json.Unmarshal(data, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("unknown discriminator: %s", discriminator)
	}
}

func __jsonschema__unmarshalDiscriminatorError(discriminator json.RawMessage, err error) error {
	return fmt.Errorf("unable to unmarshal discriminator value %v: %w", discriminator, err)
}
