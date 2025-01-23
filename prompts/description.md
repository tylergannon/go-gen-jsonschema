Please read the following description of a software application.
After the description has been given, there will be instructions for how to
assist in the next steps of developing the application.  The description
begins and ends with a long string of `=` as a fence.

========================================================


# go-gen-jsonschema

go-gen-jsonschema is a companion for Golang RAG applications, that generates
code for communicating with the LLM.

It generates the JSON schemas for communcating structured responses, by
inspecting Go types statically.  It includes the following features not found
in other JSON schema generators:

1. Allows for generation of union types via the novel concept of "interface
   implementations", which are defined by marker functions added to the code.
2. Generates `json.Unmarshaler` implementations for interpreting union types
   generated via interface implementations.
3. Static analysis and generation includes compile-time checks for correctness,
   so that issues with schemas are found during the development cycle.
4. Generates embed.FS embeddings for generated schemas, as well as struct
   methods for marked structs, offering a shorthand for loading the structured
   response schema as a `json.RawMessage`.
5. Supports marking named basic types as "enums".  Doing so will find all
   declared `const` values of the marked type within the target package, and
   use those as enum values in schemas.

## Example

Given the following struct definitions:

types.go
```go
package interfaces

import (
	_ "github.com/dave/dst"
	_ "github.com/tylergannon/structtag"
)

//go:generate go run github.com/tylergannon/go-gen-jsonschema/gen-jsonschema/ --pretty

// Overall description for MyEnumType.
type MyEnumType string

const (
	// The first possible item
	Val1 MyEnumType = "val1"
	// Use this one second
	Val2 MyEnumType = "val2"
	// Use this one third
	Val3 MyEnumType = "val3"
	// Fourth option.
	Val4 MyEnumType = "val4"
)

type TestInterface interface {
	marker()
}

// Make this look pretty interesting.
type FancyStruct struct {
	// A list of enumVals that can be really meaningful when used correctly.
	EnumVal []MyEnumType `json:"enumVal"`

	// Something tells me this isn't going to make it into the document.
	IFace TestInterface `json:"iface"`
	// Here are the details.  Make sure you fill them out.
	Details [](*struct {
		Name      string `json:"-"`
		OtherName string `json:"-"`
		funk      int
		// Highly interesting stuff regarding Foo and Bar.
		Foo, Bar string

		EnumVal MyEnumType `json:"enumVal"`
	})
}

// Put this down when you feel really great about life.
type TestInterface1 struct {
	Field1 string `json:"field1"` // obvious
	Field2 string `json:"field2"` // oblivious
	Field3 int    `json:"field3"` // obsequious
}

func (t TestInterface1) marker() {}

// This is seriously silly, don't you imagine so?
type TestInterface2 struct {
	Fork3 int `json:"fork3"`
	Fork4 int `json:"fork4"`
	Fork5 int `json:"fork5"`
}

func (t TestInterface2) marker() {}

type PointerToTestInterface struct {
	Fork99 int `json:"fork99"`
	Fork10 int `json:"fork10"`
	Fork11 int `json:"fork11"`
}

func (t *PointerToTestInterface) marker() {}

var _ TestInterface = &PointerToTestInterface{}
```

Note the `go:generate` magic comment.

We then create a new file in which we will place stub methods and calls to
marker functions that identify the desired configuration:

```go
//go:build jsonschema
// +build jsonschema

package interfaces

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (FancyStruct) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

var (
	// identifies FancyStruct as a type that should be given a schema, and
	// the `Schema()` struct method as the one that should be wired to provide
	// the generated JSON schema.
	_ = jsonschema.NewJSONSchemaMethod(FancyStruct.Schema)
	// Identifies TestInterface as a marked interface having known
	// implementations.  In this case there are three implementations of the
	// TestInterface interface, which will go in to the union type.
	_ = jsonschema.NewInterfaceImpl[TestInterface](TestInterface1{}, TestInterface2{}, (*PointerToTestInterface)(nil))
	// Identifies MyEnumType as an enum.  Instances of MyEnumType will
	// therefore be described as enums in the schema.  All `const` values of
	// this type defined within the same package will become possible values.
	_ = jsonschema.NewEnumType[MyEnumType]()
)
```

Now we'll drop to a command-line and run `go generate` from within the package
directory.  That will invoke the `go-gen-jsonschema` generator, which will
produce the following artifacts:

- ./jsonschema/FancyStruct.json -- contains the schema for FancyStruct
- ./jsonschema_gen.go -- contains generated Go code for accessing the
  generated schema and, since our types contain marked interface
  implementations, it will also contain a custom `UnmarshalJSON()` method.

FancyStruct.json:
```json
{
  "type": "object",
  "description": "Make this look pretty interesting.",
  "properties": {
    "enumVal": {
      "type": "array",
      "description": "A list of enumVals that can be really meaningful when used correctly.",
      "items": {
        "type": "string",
        "description": "Overall description for MyEnumType.\n\nval1: \nThe first possible item\n\nval2: \nUse this one second\n\nval3: \nUse this one third\n\nval4: \nFourth option.",
        "enum": [
          "val1",
          "val2",
          "val3",
          "val4"
        ]
      }
    },
    "iface": {
      "anyOf": [
        {
          "type": "object",
          "description": "Put this down when you feel really great about life.",
          "properties": {
            "!type": {
              "type": "string",
              "const": "TestInterface1"
            },
            "field1": {
              "type": "string",
              "description": "obvious"
            },
            "field2": {
              "type": "string",
              "description": "oblivious"
            },
            "field3": {
              "type": "integer",
              "description": "obsequious"
            }
          },
          "required": [
            "!type",
            "field1",
            "field2",
            "field3"
          ]
        },
        {
          "type": "object",
          "description": "This is seriously silly, don't you imagine so?",
          "properties": {
            "!type": {
              "type": "string",
              "const": "TestInterface2"
            },
            "fork3": {
              "type": "integer"
            },
            "fork4": {
              "type": "integer"
            },
            "fork5": {
              "type": "integer"
            }
          },
          "required": [
            "!type",
            "fork3",
            "fork4",
            "fork5"
          ]
        },
        {
          "type": "object",
          "properties": {
            "!type": {
              "type": "string",
              "const": "PointerToTestInterface"
            },
            "fork99": {
              "type": "integer"
            },
            "fork10": {
              "type": "integer"
            },
            "fork11": {
              "type": "integer"
            }
          },
          "required": [
            "!type",
            "fork99",
            "fork10",
            "fork11"
          ]
        }
      ]
    },
    "Details": {
      "type": "array",
      "description": "Here are the details.  Make sure you fill them out.",
      "items": {
        "type": "object",
        "properties": {
          "Foo": {
            "type": "string",
            "description": "Highly interesting stuff regarding Foo and Bar."
          },
          "Bar": {
            "type": "string",
            "description": "Highly interesting stuff regarding Foo and Bar."
          },
          "enumVal": {
            "type": "string",
            "description": "Overall description for MyEnumType.\n\nval1: \nThe first possible item\n\nval2: \nUse this one second\n\nval3: \nUse this one third\n\nval4: \nFourth option.",
            "enum": [
              "val1",
              "val2",
              "val3",
              "val4"
            ]
          }
        },
        "required": [
          "Foo",
          "Bar",
          "enumVal"
        ]
      }
    }
  },
  "required": [
    "enumVal",
    "iface",
    "Details"
  ]
}
```

jsonschema_gen.go
```go
//go:build !jsonschema
// +build !jsonschema

// Code generated by go-gen-jsonschema. DO NOT EDIT.
package interfaces

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

func (FancyStruct) Schema() json.RawMessage {
	const fileName = "jsonschema/FancyStruct.json"
	data, err := __gen_jsonschema_fs.ReadFile(fileName)
	if err != nil {
		__gen_jsonschema_panic(fileName, err)
	}
	return data
}

// UnmarshalJSON is a generated custom json.Unmarshaler implementation for
// FancyStruct.
func (f *FancyStruct) UnmarshalJSON(b []byte) (err error) {
	type Wrapper struct {
		*FancyStruct
		IFace json.RawMessage `json:"iface"`
	}
	var wrapper Wrapper
	if err = json.Unmarshal(b, &wrapper); err != nil {
		return err
	} else if f.IFace, err = __jsonUnmarshal__interfaces__TestInterface(wrapper.IFace); err != nil {
		return err
	}
	return nil
}
func __jsonUnmarshal__interfaces__TestInterface(data []byte) (TestInterface, error) {
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
	case "TestInterface1":
		var obj TestInterface1
		if err = json.Unmarshal(data, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	case "TestInterface2":
		var obj TestInterface2
		if err = json.Unmarshal(data, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	case "PointerToTestInterface":
		var obj PointerToTestInterface
		if err = json.Unmarshal(data, &obj); err != nil {
			return nil, err
		}
		return &obj, nil
	default:
		return nil, fmt.Errorf("unknown discriminator: %s", discriminator)
	}
}

func __jsonschema__unmarshalDiscriminatorError(discriminator json.RawMessage, err error) error {
	return fmt.Errorf("unable to unmarshal discriminator value %v: %w", discriminator, err)
}
```

Note how a discriminator const value has been added to each of the interface
implementations, to ensure that it is easily differentiable from its peer
options.

========================================================
