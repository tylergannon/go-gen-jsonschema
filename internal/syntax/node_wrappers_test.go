package syntax

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
	"github.com/stretchr/testify/require"
)

func TestStructFieldPropNames(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		wantNames []string
		wantSkip  bool
	}{
		{
			name:      "omitted name with omitzero",
			tag:       `json:",omitzero"`,
			wantNames: []string{"MaxRetries"},
		},
		{
			name:      "omitted name with omitempty",
			tag:       `json:",omitempty"`,
			wantNames: []string{"MaxRetries"},
		},
		{
			name:      "explicit name",
			tag:       `json:"max_retries,omitzero"`,
			wantNames: []string{"max_retries"},
		},
		{
			name:      "untagged exported field",
			wantNames: []string{"MaxRetries"},
		},
		{
			name:      "empty tag value",
			tag:       `json:""`,
			wantNames: []string{"MaxRetries"},
		},
		{
			name:      "skipped field",
			tag:       `json:"-"`,
			wantNames: []string{"-"},
			wantSkip:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := StructField{Field: &dst.Field{
				Names: []*dst.Ident{dst.NewIdent("MaxRetries")},
				Type:  dst.NewIdent("int"),
			}}
			if tt.tag != "" {
				field.Field.Tag = &dst.BasicLit{
					Kind:  token.STRING,
					Value: "`" + tt.tag + "`",
				}
			}

			require.Equal(t, tt.wantNames, field.PropNames())
			require.Equal(t, tt.wantSkip, field.Skip())
		})
	}
}

func TestStructFieldNamedVisibility(t *testing.T) {
	tests := []struct {
		name      string
		names     []string
		wantNames []string
		wantSkip  bool
	}{
		{name: "exported", names: []string{"Exported"}, wantNames: []string{"Exported"}},
		{name: "unexported", names: []string{"unexported"}, wantSkip: true},
		{name: "mixed group", names: []string{"unexported", "Exported"}, wantNames: []string{"Exported"}},
		{name: "unexported group", names: []string{"first", "second"}, wantSkip: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := StructField{Field: &dst.Field{Type: dst.NewIdent("string")}}
			for _, name := range tt.names {
				field.Field.Names = append(field.Field.Names, dst.NewIdent(name))
			}

			require.Equal(t, tt.wantNames, field.PropNames())
			require.Equal(t, tt.wantSkip, field.Skip())
		})
	}
}
