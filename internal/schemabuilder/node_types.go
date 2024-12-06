package schemabuilder

import (
	"github.com/dave/dst"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"github.com/tylergannon/structtag"
)

type StructType struct {
	StructNode *dst.StructType
	builder    *SchemaBuilder
}

type StructField struct {
	Name          string
	JSONFieldName string
	JSONSchemaTag *structtag.Tag
	Type          typeregistry.TypeSpec
}

func (s *StructType) Fields() ([]StructField, error) {
	var fields []StructField
	for _, field := range s.StructNode.Fields.List {
		var (
			tags, _       = structtag.Parse(field.Tag.Value)
			jsonTag, err  = tags.Get("json")
			jsonPropName  string
			JSONSchemaTag *structtag.Tag
		)
		if err == nil && jsonTag.HasOption("-") {
			continue
		} else if err != nil {
			jsonPropName = field.Names[0].Name
		} else {
			jsonPropName = jsonTag.Options[0]
		}
		JSONSchemaTag, _ = tags.Get("jsonschema")

		fields = append(fields, StructField{
			JSONFieldName: jsonPropName,
			JSONSchemaTag: JSONSchemaTag,
			Name:          field.Names[0].Name,
		})

	}
	return fields, nil
}
