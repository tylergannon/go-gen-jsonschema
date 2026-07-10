package remotestruct

import "github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/traversal/remoteenum"

// RemoteStruct must render with concrete properties, not as an empty schema.
type RemoteStruct struct {
	Name   string                `json:"name"`
	Status remoteenum.RemoteEnum `json:"status"`
}
