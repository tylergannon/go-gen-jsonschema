package builder

import (
	"fmt"
	"strings"

	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

type BuilderArgs struct {
	TargetDir      string
	Pretty         bool
	NumTestSamples int
	NoChanges      bool // If true, fail if any schema changes are detected
	Force          bool // If true, force regeneration of schemas even if no changes are detected
	Validate       bool // If true, generate ValidateJSON() methods and schema compilation
	// UnmarshalFormats selects generated union unmarshaler formats. The zero
	// value preserves the CLI default and generates JSON unmarshalers only.
	UnmarshalFormats UnmarshalFormats
}

type UnmarshalFormats string

const (
	UnmarshalFormatsJSON UnmarshalFormats = "json"
	UnmarshalFormatsYAML UnmarshalFormats = "yaml"
	UnmarshalFormatsBoth UnmarshalFormats = "both"
)

func (f UnmarshalFormats) generatesJSON() bool {
	return f == "" || f == UnmarshalFormatsJSON || f == UnmarshalFormatsBoth
}

func (f UnmarshalFormats) generatesYAML() bool {
	return f == UnmarshalFormatsYAML || f == UnmarshalFormatsBoth
}

func (f UnmarshalFormats) valid() bool {
	return f == "" || f == UnmarshalFormatsJSON || f == UnmarshalFormatsYAML || f == UnmarshalFormatsBoth
}

func Run(args BuilderArgs) (err error) {
	if !args.UnmarshalFormats.valid() {
		return fmt.Errorf("invalid unmarshal formats %q", args.UnmarshalFormats)
	}
	var (
		pkgs    []*decorator.Package
		builder SchemaBuilder
	)
	if pkgs, err = syntax.Load(args.TargetDir); err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found in %s", args.TargetDir)
	}
	if builder, err = New(pkgs[0]); err != nil {
		return err
	}
	builder.Pretty = args.Pretty
	builder.NumTestSamples = args.NumTestSamples
	builder.Validate = args.Validate
	builder.UnmarshalFormats = args.UnmarshalFormats

	// Allow registered transforms to mutate the model before render (no-ops by default)
	if err = (&builder).applyTransforms(); err != nil {
		return err
	}

	var changedSchemas map[string]bool
	if changedSchemas, err = builder.RenderSchemas(args.NoChanges, args.Force); err != nil {
		return err
	}

	// If NoChanges is set, fail if any schemas changed
	if args.NoChanges {
		var changedTypes []string
		for typeName, changed := range changedSchemas {
			if changed {
				changedTypes = append(changedTypes, typeName)
			}
		}
		if len(changedTypes) > 0 {
			return fmt.Errorf("schema changes detected for types: %s (and --no-changes or JSONSCHEMA_NO_CHANGES was set)", strings.Join(changedTypes, ", "))
		}
	}

	if err = builder.RenderGoCode(); err != nil {
		return err
	}
	return nil
}
