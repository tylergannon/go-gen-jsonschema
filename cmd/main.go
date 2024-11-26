package main

import (
	"flag"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/schemabuilder"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Command line flag for type names
var typeNames = flag.String("type", "", "comma-separated list of type names")

func main() {
	flag.Parse()

	if *typeNames == "" {
		log.Fatal("must specify at least one type name")
	}

	// Split type names into slice
	_types := strings.Split(*typeNames, ",")

	// Load the package
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, ".")
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

	fmt.Println("Here we go")

	if err := schemabuilder.GenerateSchemas(pkg, _types); err != nil {
		log.Fatal(err)
	}
}
