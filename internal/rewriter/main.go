package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"golang.org/x/tools/imports"
)

func main() {
	// Define our flags
	inFile := flag.String("in", "", "Input file name (optional; uses stdin if not set)")
	typeName := flag.String("type-name", "", "Name of the type whose doc comment will be updated (required)")
	fieldName := flag.String("field-name", "", "Name of the struct field whose doc comment will be updated (optional)")
	flag.Parse()

	// The new doc comment text is expected as the first (and only) trailing argument
	if len(flag.Args()) != 1 {
		log.Fatalf("Usage: %s [flags] 'new doc comment text...'\n", os.Args[0])
	}
	newDocRaw := flag.Args()[0]

	if *typeName == "" {
		log.Fatal("--type-name is required")
	}

	// Read the input file or stdin
	var srcBytes []byte
	var err error
	if *inFile != "" {
		srcBytes, err = os.ReadFile(*inFile)
		if err != nil {
			log.Fatalf("Could not read file %s: %v", *inFile, err)
		}
	} else {
		// Read from stdin
		reader := bufio.NewReader(os.Stdin)
		srcBytes, err = io.ReadAll(reader)
		if err != nil {
			log.Fatalf("Failed to read stdin: %v", err)
		}
	}

	// Parse the Go file with comments
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "", srcBytes, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing Go file: %v", err)
	}

	// Update the doc comment
	if err := updateDocComment(astFile, *typeName, *fieldName, newDocRaw); err != nil {
		log.Fatalf("Error updating doc comment: %v", err)
	}

	// Write the AST back to source code
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, astFile); err != nil {
		log.Fatalf("Error printing AST: %v", err)
	}

	// Format the code (like goimports/gofmt)
	formatted, err := imports.Process("", buf.Bytes(), &imports.Options{
		Comments:   true,
		FormatOnly: true, // only format, no import reordering
	})
	if err != nil {
		log.Fatalf("Error formatting code: %v", err)
	}

	// Output to stdout
	os.Stdout.Write(formatted)
}

// updateDocComment finds a type by name and optionally a struct field
// by name, then replaces the doc comment with the given newDoc string.
func updateDocComment(file *ast.File, typeName, fieldName, newDoc string) error {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if typeSpec.Name.Name == typeName {
				// If no field name is specified, just update the type's doc
				if fieldName == "" {
					updateCommentGroup(&genDecl.Doc, newDoc)
					return nil
				}

				// Otherwise, we assume it's a struct type
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					return fmt.Errorf("type %q is not a struct, but --field-name was provided", typeName)
				}
				// Update the doc comment for the given field
				for _, field := range structType.Fields.List {
					for _, fieldIdent := range field.Names {
						if fieldIdent.Name == fieldName {
							updateCommentGroup(&field.Doc, newDoc)
							return nil
						}
					}
				}
				// If we got here, we didn't find the field
				return fmt.Errorf("could not find field %q in type %q", fieldName, typeName)
			}
		}
	}
	return fmt.Errorf("could not find type %q in the file", typeName)
}

// updateCommentGroup replaces a comment group with the given multi-line text.
// Each line in newDoc becomes a separate //-style comment line.
func updateCommentGroup(cg **ast.CommentGroup, newDoc string) {
	if *cg == nil {
		*cg = &ast.CommentGroup{}
	}
	lines := strings.Split(newDoc, "\n")
	commentList := make([]*ast.Comment, 0, len(lines))
	for _, line := range lines {
		// Trim to avoid weird leading spaces or possible trailing \n
		line = strings.TrimSpace(line)
		// Prefix with "// "
		commentList = append(commentList, &ast.Comment{
			Text: "// " + line,
		})
	}
	(*cg).List = commentList
}
