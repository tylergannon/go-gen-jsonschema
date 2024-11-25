# go-gen-jsonschema

- Builds simple JSON schemas from static analysis of structs.
- Implements the subset of JSON Schema supported by OpenAI structured outputs.
   This is documented in `Requirements.txt`.
- Obeys the field names given in the `json:"someField"` annotation.
- Reads comments from code immediately before structs, and uses that as the
  description of any JSON Schema object (root or nested)
- Reads the comments from code immediately before field definitions, and uses
  them as the description for properties on object definitions for the schema.
- Follows fields across package boundaries, to their type declaration.
- Handles pointer fields, derivative types, and type aliases.

## Does *not* support

If a struct contains any of the following fields, they *MUST* be marked
`json:"-"` (aka *ignore*) or else generation will fail and no schema will be
emitted:

1. Private (unpublished) fields
2. Interface objects
3. Function object
4. Channel
5. `sync.Mutex`, `sync.Cond`, `sync.WaitGroup` etc

## Usage:

Types indicated on command line must be present in local package.

```go
package mypackage

//go:generate go-gen-jsonschema -type MyType,MyOtherType,MyThirdType
```
