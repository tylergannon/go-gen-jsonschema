package common

import (
	"strings"

	"github.com/tylergannon/structtag"
)

type JSONSchemaTag struct {
	Ref       string
	HasRef    bool
	ParamName string
	ParamIdx  int
	HasParam  bool
}

// ParseJSONSchemaTag parses a raw struct tag string (contents between backticks)
// and extracts jsonschema-specific options and values.
// Pass the full raw tag (including backticks) or just the inside; both work.
func ParseJSONSchemaTag(raw string) JSONSchemaTag {
	var res JSONSchemaTag
	if raw == "" {
		return res
	}
	trimmed := strings.Trim(raw, "`")
	tags, err := structtag.Parse(trimmed)
	if err != nil {
		return res
	}
	if t, err := tags.Get("jsonschema"); err == nil {
		// key=val
		if v, ok := t.GetOptValue("ref"); ok {
			res.Ref = v
			res.HasRef = true
		}
		if v, ok := t.GetOptValue("param"); ok {
			res.ParamName = v
			res.HasParam = v != ""
		}
		if v, ok := t.GetOptValue("idx"); ok {
			// ignore parsing errors; leave zero default
			// idx usage will be validated by callers when needed
			// convert only if numeric
			var n int
			for i := range len(v) {
				if v[i] < '0' || v[i] > '9' {
					n = -1
					break
				}
			}
			if n != -1 {
				// simple atoi
				n = 0
				for i := range len(v) {
					n = n*10 + int(v[i]-'0')
				}
				res.ParamIdx = n
			}
		}
	}
	return res
}
