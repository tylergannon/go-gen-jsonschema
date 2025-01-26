package builder

import (
	"fmt"

	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

type BuilderArgs struct {
	TargetDir      string
	Pretty         bool
	GenerateTests  bool
	NumTestSamples int
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
	if err = builder.RenderSchemas(); err != nil {
		return err
	}
	if err = builder.RenderGoCode(); err != nil {
		return err
	}
	if args.GenerateTests {
		if err = builder.RenderTestCodeAnthropic(); err != nil {
			return err
		}
	}
	return nil
}
