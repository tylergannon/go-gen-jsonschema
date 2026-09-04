package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/tylergannon/go-gen-jsonschema/internal/inspection"
)

func runAgentCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return writeAgentResult(invalidCommandResult("missing command"), false, stdout, stderr)
	}
	switch args[0] {
	case "version":
		return runVersionCommand(args[1:], stdout, stderr)
	case "inspect":
		return runInspectCommand(args[1:], stdout, stderr)
	default:
		return writeAgentResult(invalidCommandResult("unknown agent command "+args[0]), hasJSONFlag(args), stdout, stderr)
	}
}

func runVersionCommand(args []string, stdout, stderr io.Writer) int {
	machine := hasJSONFlag(args)
	flags := flag.NewFlagSet("version", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	jsonOutput := flags.Bool("json", false, "emit the versioned machine result")
	if err := flags.Parse(args); err != nil {
		return writeAgentResult(invalidCommandResult(err.Error()), machine, stdout, stderr)
	}
	if flags.NArg() != 0 {
		return writeAgentResult(invalidCommandResult("version accepts no positional arguments"), *jsonOutput, stdout, stderr)
	}
	return writeAgentResult(inspection.Version(), *jsonOutput, stdout, stderr)
}

func runInspectCommand(args []string, stdout, stderr io.Writer) int {
	machine := hasJSONFlag(args)
	flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	jsonOutput := flags.Bool("json", false, "emit the versioned machine result")
	target := flags.String("target", "", "path to the target package (defaults to the working directory)")
	if err := flags.Parse(args); err != nil {
		return writeAgentResult(invalidCommandResult(err.Error()), machine, stdout, stderr)
	}
	if *target == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			return writeAgentResult(errorCommandResult("working_directory_failed", err.Error()), *jsonOutput, stdout, stderr)
		}
		*target = workingDir
	}
	if info, err := os.Stat(*target); err != nil {
		return writeAgentResult(invalidTargetResult(*target, err.Error()), *jsonOutput, stdout, stderr)
	} else if !info.IsDir() {
		return writeAgentResult(invalidTargetResult(*target, "target is not a directory"), *jsonOutput, stdout, stderr)
	}
	result := inspection.Inspect(inspection.InspectRequest{
		TargetDir: *target,
		TypeNames: flags.Args(),
	})
	return writeAgentResult(result, *jsonOutput, stdout, stderr)
}

func writeAgentResult(result inspection.Result, machine bool, stdout, stderr io.Writer) int {
	if machine {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(result); err != nil {
			_, _ = fmt.Fprintf(stderr, "gen-jsonschema: encode machine result: %v\n", err)
			return 1
		}
		return inspection.ExitCode(result.Status)
	}
	if err := writeHumanResult(stderr, result); err != nil {
		return 1
	}
	return inspection.ExitCode(result.Status)
}

func writeHumanResult(writer io.Writer, result inspection.Result) error {
	var writeErr error
	write := func(format string, args ...any) {
		if writeErr == nil {
			_, writeErr = fmt.Fprintf(writer, format, args...)
		}
	}
	write("%s %s (%s, revision %s)\n", result.Tool.Name, result.Tool.Version, result.Status, result.Tool.Revision)
	for _, capability := range result.Capabilities {
		write("%s: %s", capability.Name, capability.Status)
		if capability.Detail != "" {
			write(" - %s", capability.Detail)
		}
		write("\n")
	}
	for _, inspectedType := range result.Types {
		write("%s: %s\n", inspectedType.TypePath, inspectedType.Status)
		for _, capability := range inspectedType.Capabilities {
			write("  %s: %s", capability.Name, capability.Status)
			if capability.Detail != "" {
				write(" - %s", capability.Detail)
			}
			write("\n")
		}
		for _, diagnostic := range inspectedType.Diagnostics {
			writeHumanDiagnostic(write, diagnostic)
		}
	}
	for _, diagnostic := range result.Diagnostics {
		writeHumanDiagnostic(write, diagnostic)
	}
	return writeErr
}

func writeHumanDiagnostic(write func(string, ...any), diagnostic inspection.Diagnostic) {
	location := ""
	if diagnostic.Source != nil {
		location = fmt.Sprintf(" %s:%d:%d", diagnostic.Source.File, diagnostic.Source.Line, diagnostic.Source.Column)
	}
	write("  %s [%s]%s: %s\n", diagnostic.Code, diagnostic.Classification, location, diagnostic.Message)
	if diagnostic.Remedy != "" {
		write("    remedy: %s\n", diagnostic.Remedy)
	}
}

func invalidCommandResult(message string) inspection.Result {
	result := inspection.NewResult("inspection")
	result.Status = inspection.StatusInvalid
	result.Diagnostics = []inspection.Diagnostic{{
		Code:           "invalid_request",
		Classification: inspection.ClassificationInvalidRequest,
		Message:        message,
		Remedy:         "run gen-jsonschema version --help or gen-jsonschema inspect --help",
	}}
	return result
}

func invalidTargetResult(target, message string) inspection.Result {
	result := invalidCommandResult(fmt.Sprintf("invalid target %q: %s", target, message))
	result.Diagnostics[0].Code = "invalid_target"
	result.Diagnostics[0].Remedy = "pass --target with an existing Go package directory"
	return result
}

func errorCommandResult(code, message string) inspection.Result {
	result := inspection.NewResult("inspection")
	result.Status = inspection.StatusError
	result.Diagnostics = []inspection.Diagnostic{{
		Code:           code,
		Classification: inspection.ClassificationInternal,
		Message:        message,
		Remedy:         "report this diagnostic and the installed tool revision",
	}}
	return result
}

func hasJSONFlag(args []string) bool {
	return slices.ContainsFunc(args, func(arg string) bool {
		return arg == "--json" || strings.HasPrefix(arg, "--json=")
	})
}
