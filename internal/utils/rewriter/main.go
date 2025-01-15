// go run main.go <file.go> <typeName> <newDocComment>
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("Usage: %s <file.go> <typeName> <newComment>", os.Args[0])
	}

	fileName := os.Args[1]
	typeName := os.Args[2]
	newComment := os.Args[3]

	// Load the single Go file using packages.
	// The pattern "file=<path>" is a special form recognized by go/packages.
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles,
	}
	directory := "./" + filepath.Dir(fileName)
	fmt.Printf("directory: %s\n", directory)
	pkgs, err := packages.Load(cfg, directory)
	if err != nil {
		log.Fatalf("failed to load package for file %q: %v", fileName, err)
	}
	if len(pkgs) == 0 {
		log.Fatalf("no packages found for file %q", fileName)
	}

	fmt.Printf("Found %d packages ==> %s\n", len(pkgs), pkgs[0].PkgPath)

	// Typically, pkgs[0] is the package containing our file.
	// We'll iterate over all syntax files in that package to find the matching file.
	var fileAST *ast.File
	var fset = pkgs[0].Fset

	for _, f := range pkgs[0].Syntax {
		fname := filepath.Base(fset.Position(f.Pos()).Filename)
		fmt.Println(fname)
		if fname == filepath.Base(fileName) {
			fileAST = f
			break
		}
	}
	if fileAST == nil {
		log.Fatalf("file %q not found in loaded package(s)", fileName)
	}

	// Walk the AST, find the TypeSpec for 'typeName'.
	found := false
	ast.Inspect(fileAST, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if ts.Name != nil && ts.Name.Name == typeName {
			// Replace (or set) the doc comment on this TypeSpec
			found = true
			ts.Doc = buildCommentGroup(ts.Pos(), newComment)
			fmt.Printf("found type %q in file %q, and new comment is\n%v\n", ts.Name.Name, fileName, ts.Doc)
			return false // We can stop searching once found.
		}
		return true
	})

	if !found {
		log.Printf("warning: type %q not found in file %q\n", typeName, fileName)
		// We'll still print out the file as-is,
		// but you may want to handle this differently in your real tool.
	}

	// Emit the modified file to stdout.
	var buf bytes.Buffer
	// Use printer.Fprint (rather than e.g. format.Source)
	// so we preserve the original layout & spacing as much as possible.
	if err := printer.Fprint(&buf, fset, fileAST); err != nil {
		log.Fatalf("failed to print AST: %v", err)
	}

	fmt.Println(buf.String())
}

// buildCommentGroup creates a *ast.CommentGroup containing the user-provided doc comment.
//
// For simplicity, we attach a single-line "// ... " comment.  If you'd like multi-line
// comment blocks or multiple comment lines, you could adapt accordingly.
func buildCommentGroup(pos token.Pos, commentText string) *ast.CommentGroup {
	// Ensure commentText starts with "// " to conform to typical doc styles.
	if !strings.HasPrefix(strings.TrimSpace(commentText), "//") {
		commentText = "// " + commentText
	}

	return &ast.CommentGroup{
		List: []*ast.Comment{
			{
				Slash: pos - 1, // Usually place the slash just before the type spec's position
				Text:  commentText,
			},
		},
	}
}
