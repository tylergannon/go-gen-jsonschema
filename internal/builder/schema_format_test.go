package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

func TestWriteSchemaUsesSemanticNewlinesWithoutPrettyIndentation(t *testing.T) {
	typeID := syntax.TypeID{PkgPath: "example.com/test", TypeName: "Example"}
	builder := SchemaBuilder{
		schemas: make(schemaMap),
	}
	builder.AddSchema(typeID, ObjectNode{
		Desc: "A deliberately long description, with punctuation, stays on one semantic line.",
		Properties: ObjectPropSet{
			{
				Name: "name",
				Schema: PropertyNode[string]{
					Typ:  "string",
					Desc: "Display name.",
				},
			},
		},
		TypeID_: typeID,
	})

	targetDir := t.TempDir()
	changed, err := builder.writeSchema(typeID, targetDir, false)
	require.NoError(t, err)
	require.True(t, changed)

	generated, err := os.ReadFile(filepath.Join(targetDir, "Example.json"))
	require.NoError(t, err)
	require.Equal(t, `{
"type": "object",
"description": "A deliberately long description, with punctuation, stays on one semantic line.",
"properties": {
"name": {
"type": "string",
"description": "Display name."
}
},
"required": [
"name"
],
"additionalProperties": false
}
`, string(generated))

	changed, err = builder.writeSchema(typeID, targetDir, false)
	require.NoError(t, err)
	require.False(t, changed)
}
