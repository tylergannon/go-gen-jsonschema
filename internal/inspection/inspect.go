package inspection

import (
	"errors"
	"fmt"
	"go/token"
	"slices"
	"strings"

	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

type InspectRequest struct {
	TargetDir string
	TypeNames []string
}

func Inspect(request InspectRequest) Result {
	result := NewResult("inspection")
	inspected, err := builder.Inspect(builder.InspectArgs{
		TargetDir: request.TargetDir,
		TypeNames: request.TypeNames,
	})
	if err != nil {
		classification := ClassificationToolchain
		code := "package_load_failed"
		remedy := "fix the reported Go package or module error and run inspection again"
		var position token.Position
		var loadErr *syntax.PackageLoadError
		var scanErr *syntax.ScanError
		var inspectionErr *builder.InspectionError
		switch {
		case errors.As(err, &scanErr):
			classification = classificationFromCertainty(scanErr.Certainty)
			code = scanErr.Code
			position = scanErr.Position
			remedy = scanErr.Remedy
		case errors.As(err, &loadErr) && loadErr.HasToolchainError():
			position = loadErr.Position()
		case errors.As(err, &loadErr) && loadErr.HasSourceError():
			classification = ClassificationInvalidRequest
			code = "invalid_go_package"
			position = loadErr.Position()
		case errors.As(err, &loadErr):
			position = loadErr.Position()
		case errors.As(err, &inspectionErr):
			classification = classificationFromCertainty(inspectionErr.Certainty)
			code = inspectionErr.Code
			position = inspectionErr.Position
			remedy = inspectionErr.Remedy
		}
		result.Diagnostics = []Diagnostic{{
			Code:           code,
			Classification: classification,
			Message:        err.Error(),
			Remedy:         remedy,
			Source:         sourceFromPosition(position),
		}}
		result.Status = AggregateStatus(result.Diagnostics)
		return result
	}

	for _, name := range inspected.Unregistered {
		typePath := inspected.PackagePath + "." + name
		result.Types = append(result.Types, TypeResult{
			TypePath: typePath,
			Status:   StatusUnknown,
			Diagnostics: []Diagnostic{{
				Code:           "type_not_registered",
				Classification: ClassificationUnknown,
				Message:        fmt.Sprintf("%s is not registered as a schema root", typePath),
				Remedy:         "register the type with NewJSONSchemaMethod, NewJSONSchemaFunc, or NewJSONSchemaBuilder",
				TypePath:       typePath,
			}},
		})
	}
	for _, root := range inspected.Roots {
		result.Types = append(result.Types, typeResult(root))
	}
	if len(result.Types) == 0 {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Code:           "no_registered_types",
			Classification: ClassificationInvalidRequest,
			Message:        "the target package has no registered schema roots",
			Remedy:         "register a schema root or pass a target package containing registrations",
		})
	}
	slices.SortFunc(result.Types, func(a, b TypeResult) int {
		return strings.Compare(a.TypePath, b.TypePath)
	})

	allDiagnostics := slices.Clone(result.Diagnostics)
	for _, inspectedType := range result.Types {
		allDiagnostics = append(allDiagnostics, inspectedType.Diagnostics...)
	}
	result.Status = AggregateStatus(allDiagnostics)
	for _, inspectedType := range result.Types {
		result.Status = CombineStatus(result.Status, inspectedType.Status)
	}
	return result
}

func classificationFromCertainty(certainty string) Classification {
	switch certainty {
	case "unsupported":
		return ClassificationUnsupported
	case "invalid":
		return ClassificationInvalidRequest
	case "internal":
		return ClassificationInternal
	case "toolchain":
		return ClassificationToolchain
	default:
		return ClassificationUnknown
	}
}

func typeResult(root builder.RootInspection) TypeResult {
	result := TypeResult{
		TypePath: root.TypePath,
		Capabilities: []Capability{
			{Name: "schema", Status: StatusSupported},
			{Name: "json_encode", Status: StatusSupported},
			{Name: "json_decode", Status: StatusSupported},
			{Name: "validation", Status: StatusSupported},
			{Name: "yaml_input", Status: StatusSupported},
		},
	}
	for _, finding := range root.Findings {
		classification := ClassificationUnknown
		if finding.Certainty == "unsupported" {
			classification = ClassificationUnsupported
		}
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Code:           finding.Code,
			Classification: classification,
			Message:        finding.Message,
			Remedy:         finding.Remedy,
			TypePath:       root.TypePath,
			FieldPath:      finding.FieldPath,
			Source:         sourceFromPosition(finding.Position),
		})
	}
	if root.Err != nil {
		result.Diagnostics = append(result.Diagnostics, diagnosticFromBuilderError(root))
	}

	if root.HasProviders {
		setCapability(&result, "json_encode", StatusUnsupported, "provider-rendered schemas have no general provider-derived codec")
		setCapability(&result, "json_decode", StatusUnsupported, "provider-rendered schemas have no general provider-derived codec")
		setCapability(&result, "validation", StatusUnsupported, "runtime-dependent provider schemas have no static generated validation")
		setCapability(&result, "yaml_input", StatusUnsupported, "provider-rendered schemas have no general generated decoder")
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Code:           "provider_codec_unavailable",
			Classification: ClassificationUnsupported,
			Message:        "provider-configured roots support templated/runtime schema rendering but have no general generated codec or static validator",
			Remedy:         "use the retained provider API for schema rendering and provide separately proven codecs where needed",
			TypePath:       root.TypePath,
			Source:         sourceFromPosition(root.Position),
		})
	}
	if root.RequiresUnionCodec {
		setCapability(&result, "json_encode", StatusUnsupported, "installed binary does not generate union encoding")
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Code:           "union_encode_unavailable",
			Classification: ClassificationUnsupported,
			Message:        "this root reaches a registered union but the installed binary does not generate union encoding",
			Remedy:         "install a release whose capabilities report union.json_encode as supported",
			TypePath:       root.TypePath,
			Source:         sourceFromPosition(root.Position),
		})
	}
	if root.RequiresStringEnumCodec {
		setCapability(&result, "json_encode", StatusUnsupported, "installed binary does not generate string-mode enum encoding")
		setCapability(&result, "json_decode", StatusUnsupported, "installed binary does not generate string-mode enum decoding")
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Code:           "string_enum_codec_unavailable",
			Classification: ClassificationUnsupported,
			Message:        "this root reaches a string-mode enum but the installed binary does not generate its JSON codec",
			Remedy:         "install a release whose capabilities report string-mode enum encoding and decoding as supported",
			TypePath:       root.TypePath,
			Source:         sourceFromPosition(root.Position),
		})
	}

	for _, diagnostic := range result.Diagnostics {
		status := StatusUnknown
		if diagnostic.Classification == ClassificationUnsupported {
			status = StatusUnsupported
		}
		for _, capabilityName := range affectedCapabilities(diagnostic.Code) {
			setCapabilityIfSupported(&result, capabilityName, status, "shape is outside the proven v1 boundary; see diagnostics")
		}
	}
	slices.SortFunc(result.Diagnostics, func(a, b Diagnostic) int {
		if c := strings.Compare(a.FieldPath, b.FieldPath); c != 0 {
			return c
		}
		return strings.Compare(a.Code, b.Code)
	})
	result.Status = AggregateStatus(result.Diagnostics)
	if result.Status == StatusSupported {
		for _, capability := range result.Capabilities {
			switch capability.Status {
			case StatusUnsupported:
				result.Status = StatusUnsupported
			case StatusUnknown:
				if result.Status == StatusSupported {
					result.Status = StatusUnknown
				}
			}
		}
	}
	return result
}

func setCapability(result *TypeResult, name string, status Status, detail string) {
	for index := range result.Capabilities {
		if result.Capabilities[index].Name == name {
			result.Capabilities[index].Status = status
			result.Capabilities[index].Detail = detail
			return
		}
	}
}

func setCapabilityIfSupported(result *TypeResult, name string, status Status, detail string) {
	for index := range result.Capabilities {
		if result.Capabilities[index].Name == name && result.Capabilities[index].Status == StatusSupported {
			result.Capabilities[index].Status = status
			result.Capabilities[index].Detail = detail
		}
	}
}

func affectedCapabilities(code string) []string {
	switch code {
	case "unsupported_required_omission":
		return []string{"json_encode"}
	case "union_encode_unavailable":
		return []string{"json_encode"}
	case "string_enum_codec_unavailable":
		return []string{"json_encode", "json_decode"}
	case "provider_codec_unavailable":
		return []string{"json_encode", "json_decode", "validation", "yaml_input"}
	default:
		return []string{"schema", "json_encode", "json_decode", "validation", "yaml_input"}
	}
}

func diagnosticFromBuilderError(root builder.RootInspection) Diagnostic {
	var typed *builder.InspectionError
	if !errors.As(root.Err, &typed) {
		return Diagnostic{
			Code:           "internal_untyped_builder_error",
			Classification: ClassificationInternal,
			Message:        root.Err.Error(),
			Remedy:         "report this diagnostic and the installed tool revision",
			TypePath:       root.TypePath,
			Source:         sourceFromPosition(root.Position),
		}
	}
	diagnostic := Diagnostic{
		Code:           typed.Code,
		Classification: classificationFromCertainty(typed.Certainty),
		Message:        typed.Message,
		Remedy:         typed.Remedy,
		TypePath:       typed.TypePath,
		FieldPath:      typed.FieldPath,
		Source:         sourceFromPosition(typed.Position),
	}
	if diagnostic.TypePath == "" {
		diagnostic.TypePath = root.TypePath
	}
	if diagnostic.Source == nil {
		diagnostic.Source = sourceFromPosition(root.Position)
	}
	return diagnostic
}

func sourceFromPosition(position token.Position) *Source {
	if position.Filename == "" {
		return nil
	}
	return &Source{File: position.Filename, Line: position.Line, Column: position.Column}
}
