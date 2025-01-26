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
	GenerateTests  bool
	NumTestSamples int
	NoChanges      bool // If true, fail if any schema changes are detected
}

func Run(args BuilderArgs) (err error) {
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
	builder.GenerateTests = args.GenerateTests
	builder.NumTestSamples = args.NumTestSamples

	var changedSchemas map[string]bool
	if changedSchemas, err = builder.RenderSchemas(); err != nil {
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
	if args.GenerateTests {
		if err = builder.RenderTestCodeAnthropic(changedSchemas); err != nil {
			return err
		}
	}
	return nil
}
