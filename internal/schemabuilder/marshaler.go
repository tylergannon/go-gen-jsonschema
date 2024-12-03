package schemabuilder

import (
	"bytes"
	"encoding/json"
)

type namedProp struct {
	Name string
	Prop json.Marshaler
}

type namedProps []namedProp

func (p *namedProps) Add(name string, marshaler json.Marshaler) {
	// It should append a new namedProp to the slice.
	*p = append(*p, namedProp{name, marshaler})
}

type StrictSchema struct {
	Description string
	Props       namedProps
	Definitions namedProps
}

func (s *StrictSchema) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	// Start the JSON object
	buf.WriteString(`{"type":"object",`)

	// Add description if present
	if s.Description != "" {
		buf.WriteString(`"description":`)
		descJSON, err := json.Marshal(s.Description)
		if err != nil {
			return nil, err
		}
		buf.Write(descJSON)
		buf.WriteString(`,`)
	}

	// Add properties
	buf.WriteString(`"properties":{`)
	for i, prop := range s.Props {
		propJSON, err := prop.Prop.MarshalJSON()
		if err != nil {
			return nil, err
		}
		buf.WriteString(`"`)
		buf.WriteString(prop.Name)
		buf.WriteString(`":`)
		buf.Write(propJSON)
		if i < len(s.Props)-1 {
			buf.WriteString(`,`)
		}
	}
	buf.WriteString(`},`)

	if len(s.Definitions) > 0 {
		// Add definitions
		buf.WriteString(`"definitions":{`)
		for i, def := range s.Definitions {
			defJSON, err := def.Prop.MarshalJSON()
			if err != nil {
				return nil, err
			}
			buf.WriteString(`"`)
			buf.WriteString(def.Name)
			buf.WriteString(`":`)
			buf.Write(defJSON)
			if i < len(s.Definitions)-1 {
				buf.WriteString(`,`)
			}
		}
		buf.WriteString(`},`)
	}

	// Add required keys
	buf.WriteString(`"required":[`)
	for i, prop := range s.Props {
		buf.WriteString(`"`)
		buf.WriteString(prop.Name)
		buf.WriteString(`"`)
		if i < len(s.Props)-1 {
			buf.WriteString(`,`)
		}
	}
	buf.WriteString(`],`)

	// Add additionalProperties: false
	buf.WriteString(`"additionalProperties":false}`)

	return buf.Bytes(), nil
}
