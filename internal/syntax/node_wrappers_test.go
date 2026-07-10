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
