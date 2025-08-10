# v1 Roadmap (Working)

This file tracks execution steps to bring v1 to parity across spec, tests, and implementation.

Status key: [ ] pending  [~] in progress  [x] done

## Spec
- [x] Draft v1 contract (docs/spec/v1.md)
- [ ] Add canonical error message examples for top lints
- [ ] Migration guide v0.5.x → v1

## Tests (end-to-end fixtures)
- [ ] NewJSONSchemaFunc fixture + goldens
- [ ] NewJSONSchemaBuilder fixture + goldens
- [ ] Provider rendering fixture exercising RenderedSchema() deterministically
- [ ] iota enums (numeric + string modes) fixture

## Implementation
- [ ] Scanner: NewJSONSchemaFunc (free func) → infer T from func param
- [ ] Scanner: NewJSONSchemaBuilder (builder func) → infer T from registration site
- [ ] Options: WithEnum/WithEnumMode/WithEnumName
- [ ] Options: WithInterface/WithInterfaceImpls/WithDiscriminator
- [ ] Options: WithRenderProviders
- [ ] Builder/codegen: generate RenderedSchema() alongside Schema() when requested
- [ ] Enums: iota support and string mode (with String() and overrides)

## Nice-to-haves (post-v1)
- [ ] gen-jsonschema init/check/doctor commands (cobra)
- [ ] VSCode snippets and quickstart template
- [ ] Public API doc polishing and website page
