// Package inspection defines read-only, transport-independent operations for
// discovering an installed gen-jsonschema binary and inspecting Go models.
package inspection

const (
	SchemaVersion   = "1"
	ContractVersion = "v1"
	ToolName        = "gen-jsonschema"
)

type Status string

const (
	StatusSupported   Status = "supported"
	StatusUnsupported Status = "unsupported"
	StatusUnknown     Status = "unknown"
	StatusInvalid     Status = "invalid"
	StatusError       Status = "error"
)

type Classification string

const (
	ClassificationInvalidRequest Classification = "invalid_request"
	ClassificationUnsupported    Classification = "unsupported"
	ClassificationUnknown        Classification = "unknown"
	ClassificationToolchain      Classification = "toolchain"
	ClassificationInternal       Classification = "internal"
)

type Result struct {
	SchemaVersion   string       `json:"schemaVersion"`
	Kind            string       `json:"kind"`
	Status          Status       `json:"status"`
	Tool            Tool         `json:"tool"`
	ContractVersion string       `json:"contractVersion"`
	Usage           string       `json:"usage,omitempty"`
	Capabilities    []Capability `json:"capabilities,omitempty"`
	Types           []TypeResult `json:"types,omitempty"`
	Diagnostics     []Diagnostic `json:"diagnostics,omitempty"`
}

type Tool struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	VersionState  string `json:"versionState"`
	Revision      string `json:"revision"`
	RevisionState string `json:"revisionState"`
	Modified      bool   `json:"modified"`
}

type Capability struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type TypeResult struct {
	TypePath     string       `json:"typePath"`
	Status       Status       `json:"status"`
	Capabilities []Capability `json:"capabilities,omitempty"`
	Diagnostics  []Diagnostic `json:"diagnostics,omitempty"`
}

type Diagnostic struct {
	Code           string         `json:"code"`
	Classification Classification `json:"classification"`
	Message        string         `json:"message"`
	Remedy         string         `json:"remedy,omitempty"`
	TypePath       string         `json:"typePath,omitempty"`
	FieldPath      string         `json:"fieldPath,omitempty"`
	Source         *Source        `json:"source,omitempty"`
}

type Source struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

func NewResult(kind string) Result {
	return Result{
		SchemaVersion:   SchemaVersion,
		Kind:            kind,
		Status:          StatusSupported,
		Tool:            BuildIdentity(),
		ContractVersion: ContractVersion,
	}
}

func ExitCode(status Status) int {
	switch status {
	case StatusSupported:
		return 0
	case StatusInvalid:
		return 2
	case StatusUnsupported, StatusUnknown:
		return 3
	default:
		return 1
	}
}

func AggregateStatus(diagnostics []Diagnostic) Status {
	status := StatusSupported
	for _, diagnostic := range diagnostics {
		switch diagnostic.Classification {
		case ClassificationInternal, ClassificationToolchain:
			return StatusError
		case ClassificationInvalidRequest:
			if status != StatusError {
				status = StatusInvalid
			}
		case ClassificationUnsupported:
			if status == StatusSupported || status == StatusUnknown {
				status = StatusUnsupported
			}
		case ClassificationUnknown:
			if status == StatusSupported {
				status = StatusUnknown
			}
		}
	}
	return status
}

func CombineStatus(current, candidate Status) Status {
	priority := map[Status]int{
		StatusSupported:   0,
		StatusUnknown:     1,
		StatusUnsupported: 2,
		StatusInvalid:     3,
		StatusError:       4,
	}
	if priority[candidate] > priority[current] {
		return candidate
	}
	return current
}
