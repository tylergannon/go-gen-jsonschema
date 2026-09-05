//go:build jsonschema

package boundary

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Pair) Schema() json.RawMessage   { panic("not implemented") }
func (Pair) ValidateJSON([]byte) error { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(Pair.Schema,
	jsonschema.WithInterface(Pair{}.Left, jsonschema.Discriminator("kind"), jsonschema.Impl("born", Created{}), jsonschema.Impl("gone", (*Deleted)(nil))),
	jsonschema.WithInterface(Pair{}.Right, jsonschema.Discriminator("!\"kind"), jsonschema.Impl("new", Created{}), jsonschema.Impl("removed", (*Deleted)(nil))),
	jsonschema.WithInterface(Pair{}.Events, jsonschema.Impl("created", Created{}), jsonschema.Impl("deleted", (*Deleted)(nil))),
	jsonschema.WithInterface(Pair{}.Extra, jsonschema.Impl("created", Created{}), jsonschema.Impl("deleted", (*Deleted)(nil))),
)

func (Config) Schema() json.RawMessage   { panic("stub") }
func (Config) ValidateJSON([]byte) error { panic("stub") }

var _ = jsonschema.NewJSONSchemaMethod(Config.Schema,
	jsonschema.WithStringerEnum(Config{}.Direct),
	jsonschema.WithEnum(Config{}.Numeric),
	jsonschema.WithStringerEnum(Config{}.Optional),
	jsonschema.WithStringerEnum(Config{}.Nullable),
)
