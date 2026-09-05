package main

import (
	_ "embed"
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"os"
	"os/exec"
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

type genOptions struct {
	pretty           bool
	target           string
	noChanges        bool
	force            bool
	validate         bool
	formats          string
	typeScriptDir    string
	typeScriptBarrel bool
}

func newGenFlagSet(errorHandling flag.ErrorHandling) (*flag.FlagSet, *genOptions) {
	options := &genOptions{}
	genCmd := flag.NewFlagSet("gen", errorHandling)
	genCmd.BoolVar(&options.pretty, "pretty", false, "Enable pretty output")
	genCmd.StringVar(&options.target, "target", "", "Path to target package (default to local wd)")
	genCmd.BoolVar(&options.noChanges, "no-changes", false, "Fail if any schema or requested TypeScript output changes are detected")
	genCmd.BoolVar(&options.force, "force", false, "Force regeneration of schemas and requested TypeScript output even if no changes are detected")
	genCmd.BoolVar(&options.validate, "validate", false, "Generate schema validation methods for the selected formats")
	genCmd.StringVar(&options.formats, "formats", "json", "Generated decoding and validation formats: json or both")
	genCmd.StringVar(&options.typeScriptDir, "typescript", "", "Generate structural TypeScript declarations in this directory")
	genCmd.BoolVar(&options.typeScriptBarrel, "typescript-barrel", false, "Generate an index.ts type-only export (requires --typescript)")
	return genCmd, options
}

func handleGen(firstArg int) {
	genCmd, options := newGenFlagSet(flag.ExitOnError)

	// Check if --help was requested
	if len(os.Args) > 2 && os.Args[2] == "--help" {
		fmt.Println("Usage: gen [options]")
		fmt.Println("\nOptions:")
		genCmd.PrintDefaults()
		return
	}
	_ = genCmd.Parse(os.Args[firstArg:])
	if options.target == "" {
		var err error
		if options.target, err = os.Getwd(); err != nil {
			log.Fatal(err)
		}
	} else if st, err := os.Stat(options.target); err != nil {
		log.Fatal(err)
	} else if !st.IsDir() {
		log.Fatalf("%s is not a directory", options.target)
	}

	unmarshalFormats, err := parseUnmarshalFormats(options.formats)
	if err != nil {
		log.Fatal(err)
	}

	// Check environment variable
	options.noChanges = options.noChanges || os.Getenv("JSONSCHEMA_NO_CHANGES") != ""

	if options.force && options.noChanges {
		log.Fatal("Cannot use --force and --no-changes together")
	}

	if err = builder.Run(builder.BuilderArgs{
		TargetDir:        options.target,
		Pretty:           options.pretty,
		NoChanges:        options.noChanges,
		Force:            options.force,
		Validate:         options.validate,
		TypeScriptDir:    options.typeScriptDir,
		TypeScriptBarrel: options.typeScriptBarrel,
		UnmarshalFormats: unmarshalFormats,
	}); err != nil {
		log.Fatal(err)
	}
}

func parseUnmarshalFormats(value string) (builder.UnmarshalFormats, error) {
	formats := builder.UnmarshalFormats(value)
	switch formats {
	case builder.UnmarshalFormatsJSON, builder.UnmarshalFormatsBoth:
		return formats, nil
	default:
		return "", fmt.Errorf("invalid --formats value %q: expected json or both", value)
	}
}

func handleNew() {
	// Define the --out flag
	var (
		newCmd      = flag.NewFlagSet("new", flag.ExitOnError)
		out         = newCmd.String("out", "", "Path to output file.  Empty val or '--' means print to stdout")
		pkg         = newCmd.String("pkg", "", "Package for generated file. Default is current directory or using the package name for the package specified in --out")
		methods     = newCmd.String("methods", "", "Comma-separated list of methods to generate in the form of TypeName=MethodName,TypeName2=MethodName2")
		runGenerate = newCmd.Bool("generate", false, "Run go generate in the target package after creating the stub file")
		newValidate = newCmd.Bool("validate", false, "Include validation stubs for the selected formats")
		newFormats  = newCmd.String("formats", "json", "Generated decoding and validation formats: json or both")
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
		pkgName   string
		useStdout = *out == "" || *out == "--"
	)
	if err != nil {
		log.Fatalln(err)
	}
	unmarshalFormats, err := parseUnmarshalFormats(*newFormats)
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
		Validate: *newValidate,
		YAML:     unmarshalFormats == builder.UnmarshalFormatsBoth,
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

	fmt.Printf("Output written to: %s\n", *out)

	// Optionally run go generate in the target package
	if *runGenerate && !useStdout {
		targetDir := filepath.Dir(*out)
		if targetDir == "" {
			targetDir = "."
		}
		fmt.Printf("Running go generate in %s...\n", targetDir)
		cmd := exec.Command("go", "generate", "./...")
		cmd.Dir = targetDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("go generate failed: %v", err)
		}
		fmt.Println("Code generation complete.")
	}
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
	Validate bool
	YAML     bool
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
