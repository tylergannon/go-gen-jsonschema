# Adversarial review: V1 codec integration

Review target: `5e1040a99ba7f51f126601bd84ad5e6b912847d2` against `origin/main` (`b2fd7ee189527c2cee0314614b75528e771bd15b`), GitHub issues #57 and #58, `docs/spec/v1.md`, repository instructions, and the caller-restated codec/TypeScript constraints.

Evidence inspected:

- the complete `origin/main...5e1040a` diff and candidate history;
- `AGENTS.md`, issues #57/#58, and the complete accepted V1 contract;
- codec planning, rendering, collision discovery, field resolution, enum membership, TypeScript lowering, generated fixtures, runtime tests, and retained proof artifacts;
- required candidate baseline `go test ./...` (pass);
- a fresh generated consumer at `ephemeral/review-probes/promoted-ambiguous` using the exact candidate CLI.

## Findings

### issue — ambiguous sibling promoted union fields generate uncompilable code

The V1 contract requires supported promoted fields to follow Go field-selection and shadowing rules (`docs/spec/v1.md:74`). `resolveLocalInterfaceProps` updates `seenProps` inside each recursive call, but does not return that state to the caller; the two sibling calls at `internal/builder/gen_schema.go:2131-2143` therefore each accept the same promoted JSON property. The owner plan then contains both fields even though two same-depth, equally tagged promoted fields are ambiguous to `encoding/json` and should not be selected.

The fresh probe defines `Owner` by embedding `Left` and `Right`, each with `Value Event `json:"value"`` (`ephemeral/review-probes/promoted-ambiguous/types.go:11-22`). The candidate generator exits successfully, but emits two identically named `Value json.RawMessage` fields in each wrapper (`ephemeral/review-probes/promoted-ambiguous/jsonschema_gen.go:35-39` and `:58-62`) and duplicate `value` schema properties. `go test ./...` in that generated consumer fails:

```text
./jsonschema_gen.go:38:3: Value redeclared
	./jsonschema_gen.go:37:3: other declaration of Value
./jsonschema_gen.go:61:3: Value redeclared
	./jsonschema_gen.go:60:3: other declaration of Value
```

This is a supported-embedding correctness failure on issue #57's owner-codec path: generation reports success and writes artifacts that cannot compile. Resolve promoted candidates with the same depth/tag dominance rules already used by the native TypeScript lowerer, or reject this exceptional ambiguous composition before writing output. Add a generator/compile regression test for sibling embeddings; the current test covers only a direct outer field shadowing one embedded field.

Outcome: material findings remain
