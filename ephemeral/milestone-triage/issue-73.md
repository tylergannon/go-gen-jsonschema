## Outcome

Add `Declare(T.Schema)` as the canonical typed fluent declaration API for 1.0. It should make the common authoring path readable and bind receiver-based providers to the declared root through Go's type checker.

```go
var _ = jsonschema.Declare(Example.Schema).
    Accessor(Example{}.A, Example.ASchema).
    Method(Example{}.B, Example.BSchema).
    Function(Example{}.C, BoolSchema).
    Enum(Example{}.Status).
    RenderProviders()
```

## 1.0 scope

- Provide the typed `Declaration[T]` chain for schema callables, accessor providers, field-taking methods, free provider functions, enums, string enums, refs, render-provider flags, and cohesive interface registration using the existing `Discriminator` and `Impl` options.
- Infer the root type from the schema callable. Let Go reject mismatched receiver/provider and field types where the language can express the relationship.
- Resolve supported chains through `go/types` symbol identity and normalize them into the existing registration model. Do not create a second schema or codec implementation.
- Preserve value/pointer root behavior and the existing containing-struct union/enum codec architecture.
- Update the scaffolder, doc extraction, examples, README, website, shipped skill, and plugin guidance to teach `Declare` as the normal syntax.
- Deprecate `NewJSONSchemaMethod`, `NewJSONSchemaFunc`, and their standalone declaration helpers in Go documentation. Remove legacy syntax from primary tutorials and examples; mention it only in concise migration/compatibility guidance.
- Keep legacy declarations source-compatible through 1.x so existing users can upgrade without a flag day. Compatibility fixtures should continue to prove them even though user-facing documentation no longer teaches them.

## Sequencing

#80 owns the focused existing bug where recognized legacy options can be silently ignored and lands first. This issue builds the new API on the corrected scanner path and must not duplicate that work.

## Acceptance

- Compile-time fixtures prove inference, valid value/pointer roots, and rejection of mismatched receiver providers and field types.
- Scanner and command tests cover every supported fluent method, import aliases, invalid chains, and propagation of source-positioned errors.
- Equivalent fluent and legacy declarations generate equivalent artifacts for representative structs, providers, refs, enums, and typed unions.
- `gen-jsonschema new` emits the fluent form and works through real generation.
- Primary documentation and agent guidance use only the fluent syntax; migration guidance clearly labels legacy forms deprecated.
- Root tests, relevant generated consumer tests, clean regeneration, and lint pass.

Structured machine diagnostics (#60), full preview infrastructure (#61), new codec semantics, generic streaming, and a new union descriptor system are outside this issue.
