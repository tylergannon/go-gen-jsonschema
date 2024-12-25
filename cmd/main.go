package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
	"github.com/tylergannon/go-gen-jsonschema/internal/loader"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Command line flags
var (
	typeNames = flag.String("type", "", "comma-separated list of type names")
	pretty    = flag.Bool("pretty", false, "output JSON with indentation")
	verbose   = flag.Bool("verbose", false, "print detailed processing information")
	subdir    = flag.String("subdir", "jsonschema", "subdirectory where to place schemas")
)

func main() {
	flag.Parse()

	if *typeNames == "" {
		log.Fatal("must specify at least one type name")
	}

	// Split type names into slice
	types := strings.Split(*typeNames, ",")

	pkgs, err := loader.Load(".")
	if err != nil {
		log.Fatalf("loading package: %v", err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	pkg := pkgs[0]

	// Check for package errors
	if len(pkg.Errors) > 0 {
		for _, err := range pkg.Errors {
			_, _ = fmt.Fprintf(log.Default().Writer(), "error in package: %v\n", err)
		}
		log.Fatal("package contains errors")
	}

	registry, err := typeregistry.NewRegistry(pkgs)
	if err != nil {
		log.Fatal(err)
	}

	if err = os.MkdirAll(*subdir, 0755); err != nil {
		log.Fatal(err)
	}

	var graphs []*typeregistry.SchemaGraph
	discriminators := builder.NewDiscriminatorMap()

	for _, typeName := range types {
		ts, ok := registry.GetTypeByName(typeName, pkg.PkgPath)
		if !ok {
			log.Fatalf("could not find type %q", typeName)
		}

		g, err := registry.GraphTypeForSchema(ts)
		if err != nil {
			log.Fatal(err)
		}

		graphs = append(graphs, g)

		schema := builder.New(g, discriminators).Render()

		var schemaBytes []byte
		if *pretty {
			schemaBytes, err = json.MarshalIndent(schema, "", "  ")
		} else {
			schemaBytes, err = json.Marshal(schema)
		}
		if err != nil {
			log.Fatal(err)
		}

		destFile := filepath.Join(*subdir, fmt.Sprintf("%s.json", typeName))
		if err = os.WriteFile(destFile, schemaBytes, 0644); err != nil {
			log.Fatal(err)
		}

		if *verbose {
			fmt.Printf("Processed type: %s (package: %s), output file: %s\n", typeName, pkg.PkgPath, destFile)
		}
	}
	_ = builder.RenderGoCode("jsonschema_gen.go", *subdir, graphs)
}
