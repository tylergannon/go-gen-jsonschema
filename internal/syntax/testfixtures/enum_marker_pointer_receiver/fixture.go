//go:build jsonschema

// Package enummarkerpointerreceiver is a standalone negative fixture: its
// enum marker method uses a pointer receiver, which must be a hard,
// source-positioned diagnostic naming the type rather than a silent miss.
// It lives in its own directory because loading a package eagerly scans
// every type in it, and this package is expected to fail that scan.
package enummarkerpointerreceiver

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

type Status string

func (*Status) enum() {}

const Ready Status = "ready"

type Owner struct {
	Status Status `json:"status"`
}

func (Owner) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.Declare(Owner.Schema)
