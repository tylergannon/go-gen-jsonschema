# Native yaml/v4 unmarshaling plan

## Objective

Generate native `go.yaml.in/yaml/v4` `UnmarshalYAML` methods for structs whose
registered interface fields already require generated `UnmarshalJSON` methods.
Keep YAML decoding node-native; do not convert YAML through JSON.

## Proof claims

1. A generated type with one registered interface field can be decoded through
   `yaml.Unmarshal`, and the discriminator selects and populates the concrete Go
   implementation.
2. The same generated path handles the difficult subset without duplicating
   the JSON suite: YAML-specific field names, merge keys, interface slices,
   optional interfaces, nested YAML-specific decoding, indexed errors, and
   transactional receiver assignment.
3. Existing JSON generation and decoding behavior remains unchanged.

## Sequential gates

1. Add one simple YAML test to the existing interface fixture. Run it and retain
   the expected failure. Implement only enough native node dispatch to pass it.
2. Add one comprehensive YAML test to the v1 interface-options fixture. Run it
   and retain the expected failure. Extend the implementation for the hard
   semantics and make it pass.
3. Regenerate checked-in outputs and goldens, format, run focused fixture tests,
   then run `go test ./...` and review the final diff.

All three gates completed successfully.

## Design boundary

- Reuse the existing scanner output and registered-interface model.
- Decode mappings into `map[string]yaml.Node`, remove the polymorphic fields,
  reconstruct the ordinary-field mapping, and decode that mapping into a local
  alias before dispatching interface nodes.
- Commit the receiver only after all ordinary and polymorphic fields succeed.
- Do not add YAML methods to the generic `Optional` or `Nullable` wrappers in
  this change; generated optional-interface handling remains owner-level.
