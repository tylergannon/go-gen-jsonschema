package builder

import (
	"fmt"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

type BuilderArgs struct {
	TargetDir string
	Pretty    bool
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
	if err = builder.RenderSchemas(); err != nil {
		return err
	}
	return nil
}
