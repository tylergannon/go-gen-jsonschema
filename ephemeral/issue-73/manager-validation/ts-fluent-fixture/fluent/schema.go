//go:build jsonschema

package consumer

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Envelope) Schema() json.RawMessage    { panic("not implemented") }
func (Detail) Schema() json.RawMessage      { panic("not implemented") }
func (Composition) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.Declare(Detail.Schema).Ref()
var _ = jsonschema.Declare(Composition.Schema)

// Status is shared across Envelope.Status, Composition.C (Optional[[]Status]),
// and Composition.D (Nullable[Status]). Field-level .Enum has no fluent
// replacement for a slice-of-enum field (Composition.C), and per-field
// annotation on only some occurrences of a shared enum type silently
// degrades the others (loses the shared TypeScript alias and, for
// Composition.C, all enum validation). The package-level legacy form
// registers Status everywhere it's used, matching the equivalent legacy
// fixture's registration exactly.
var _ = jsonschema.NewEnumType[Status]()

var _ = jsonschema.Declare(Envelope.Schema).
	Interface(Envelope{}.Event,
		jsonschema.Discriminator("!kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
		// ADDED_VARIANT
	).
	Interface(Envelope{}.Other,
		jsonschema.Discriminator("other-key"),
		jsonschema.Impl("create\"雪", Created{}),
	).
	Interface(Envelope{}.Maybe,
		jsonschema.Discriminator("!kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
	).
	Interface(Envelope{}.Events,
		jsonschema.Discriminator("!kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
	).
	Enum(Envelope{}.Priority).
	StringerEnum(Envelope{}.PriorityName)
