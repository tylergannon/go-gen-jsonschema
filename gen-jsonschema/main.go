package main

import (
	_ "embed"
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
	"github.com/tylergannon/go-gen-jsonschema/internal/common"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

//go:embed tmpl/config.go.tmpl
var configTmplContents string

func main() {

	if len(os.Args) == 1 {
		handleGen(1)
		return
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printGlobalHelp()
		return
	}

	// Extract the subcommand
	subcommand := os.Args[1]

	// Switch on the subcommand
	switch subcommand {
	case "gen":
		handleGen(2)
	case "new":
		handleNew()
	default:
		handleGen(1)
	}
}

// Prints global help for the script
func printGlobalHelp() {
	fmt.Println("Usage:")
	fmt.Println("  [subcommand] [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  gen      Generate output (default)")
	fmt.Println("  new      Create a new project")
	fmt.Println("\nRun '[subcommand] --help' for more details.")
}

func handleGen(firstArg int) {
	// Define the --pretty flag
	var (
		genCmd         = flag.NewFlagSet("gen", flag.ExitOnError)
		pretty         = genCmd.Bool("pretty", false, "Enable pretty output")
		target         = genCmd.String("target", "", "Path to target package (default to local wd)")
		numTestSamples = genCmd.Int("num-test-samples", 5, "Number of test samples to generate")
		noChanges      = genCmd.Bool("no-changes", false, "Fail if any schema changes are detected")
		force          = genCmd.Bool("force", false, "Force regeneration of schemas even if no changes are detected")
		err            error
	)

	if *target == "" {
		if *target, err = os.Getwd(); err != nil {
			log.Fatal(err)
		}
	} else if st, err := os.Stat(*target); err != nil {
		log.Fatal(err)
	} else if !st.IsDir() {
		log.Fatalf("%s is not a directory", *target)
	}

	// Check if --help was requested
	if len(os.Args) > 2 && os.Args[2] == "--help" {
		fmt.Println("Usage: gen [options]")
		fmt.Println("\nOptions:")
		genCmd.PrintDefaults()
		return
	}
	_ = genCmd.Parse(os.Args[firstArg:])

	// Check environment variable
	*noChanges = *noChanges || os.Getenv("JSONSCHEMA_NO_CHANGES") != ""

	if *force && *noChanges {
		log.Fatal("Cannot use --force and --no-changes together")
	}

	if err = builder.Run(builder.BuilderArgs{
		TargetDir:      *target,
		Pretty:         *pretty,
		NumTestSamples: *numTestSamples,
		NoChanges:      *noChanges,
		Force:          *force,
	}); err != nil {
		log.Fatal(err)
	}
}

func handleNew() {
	// Define the --out flag
	var (
		newCmd  = flag.NewFlagSet("new", flag.ExitOnError)
		out     = newCmd.String("out", "", "Path to output file.  Empty val or '--' means print to stdout")
		pkg     = newCmd.String("pkg", "", "Package for generated file. Default is current directory or using the package name for the package specified in --out")
		methods = newCmd.String("methods", "", "Comma-separated list of methods to generate in the form of TypeName=MethodName,TypeName2=MethodName2")
	)

	// Check if --help was requested
	if len(os.Args) > 2 && os.Args[2] == "--help" {
		fmt.Println("Usage: new [options] FILENAME")
		fmt.Println("\nOptions:")
		newCmd.PrintDefaults()
		return
	}

	// Parse flags for the "new" subcommand
	var (
		err       = newCmd.Parse(os.Args[2:])
		args      = newCmd.Args()
		pkgName   string
		useStdout = *out == "" || *out == "--"
	)
	if err != nil {
		log.Fatalln(err)
	}

	// Remaining args (after parsing)

	if useStdout {
		if *pkg != "" {
			pkgName = *pkg
		} else {
			var wd string
			if wd, err = os.Getwd(); err != nil {
				log.Fatalln(err)
			}
			if pkgName, err = getPackageName(wd); err != nil {
				log.Fatalln(err)
			}
		}
	} else if pkgName, err = getPackageName(filepath.Dir(*out)); err != nil {
		log.Fatalln(err)
	}

	var tmplArg = configArg{
		BuildTag: syntax.BuildTag,
		PkgName:  pkgName,
	}

	for methodArg := range strings.SplitSeq(*methods, ",") {
		if len(methodArg) == 0 {
			continue
		}
		parts := strings.SplitN(methodArg, "=", 2)
		if len(parts) != 2 {
			log.Fatalln("Invalid method argument.  Must be keyvalue in the form TypeName=MethodName. -- ", methodArg)
		}
		tmplArg.Methods = append(tmplArg.Methods, methodDef{
			TypeName:   parts[0],
			MethodName: parts[1],
		})
	}
	if len(tmplArg.Methods) == 0 {
		log.Fatalln("No methods to generate.")
	}

	fmt.Println("Package name:", pkgName)

	writer, err := getOutputWriter(*out)
	if err != nil {
		log.Fatalln(err)
	}
	defer common.LogClose(writer)
	data, err := builder.RenderTemplate(configTmplContents, tmplArg)
	if err != nil {
		log.Fatalln(err)
	}
	if formatted, err := builder.FormatCodeWithGoimports(data.Bytes()); err != nil {
		log.Fatalln(err)
	} else if _, err = writer.Write(formatted); err != nil {
		log.Fatalln(err)
	}

	// Perform the "new" command action
	if len(args) > 0 {
		fmt.Printf("Creating new project: %s\n", args[0])
	} else {
		fmt.Println("Creating a new project.")
	}
	fmt.Printf("Output will be written to: %s\n", *out)
}

func getPackageName(path string) (string, error) {
	// Use build.Import to analyze the directory
	pkg, err := build.ImportDir(path, 0)
	if err != nil {
		return "", err
	}

	return pkg.Name, nil
}

type methodDef struct {
	TypeName, MethodName string
}

type configArg struct {
	PkgName  string
	BuildTag string
	Methods  []methodDef
}

// getOutputWriter returns an io.WriteCloser for either a file or stdout based on the output path.
func getOutputWriter(outputPath string) (io.WriteCloser, error) {
	if outputPath == "" || outputPath == "--" {
		// Use stdout
		return os.Stdout, nil
	}

	// Open a file for writing
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	return file, nil
}
