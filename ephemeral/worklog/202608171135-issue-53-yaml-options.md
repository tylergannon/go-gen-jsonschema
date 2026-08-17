# Issue 53 YAML decoding

decision: Keep Issue 53 scoped to native decoding of ordinary presence wrappers. Do not adopt yaml/v4's legacy callback unmarshaler solely to propagate `WithKnownFields`; record that opt-in decoder behavior as a narrow limitation.

correction: `WithKnownFields` is an opt-in yaml decoder feature, not schema validation and not an existing go-gen-jsonschema promise. Preserve its current limitation in issue or PR context rather than expanding the public contract or general documentation.

friction: A naive generated `ValidateYAML` decoded YAML into `any` and passed it to santhosh-tekuri/jsonschema, but valid fixtures with distinct `yaml` and `json` tag names failed because the generated schema uses JSON property names. Correct YAML validation requires a YAML-named schema or recursive name normalization; do not claim the upstream validator makes this integration free.

decision: Reuse `resolveLocalInterfaceProps` to collect ordinary Optional/Nullable metadata during the existing owner-field scan. Do not add a second AST traversal.

## Proof claims

- A YAML-enabled generated owner decodes ordinary `Optional` and `Nullable` fields with absent, zero, non-zero, and null semantics matching their JSON contract.
- A failed YAML decode leaves the destination receiver unchanged.
- JSON-only generation neither imports yaml/v4 nor emits YAML decoding support.

## Proof

- `go generate ./...` was idempotent against the pre-generation status.
- `go build ./...` and uncached `go test -count=1 ./...` passed.
- In the regenerated nested consumer fixture, `go generate ./... && go test -run TestGeneratedYAMLUnmarshalComprehensive -count=1 -v` passed. That test exercises yaml/v4 `Load` with V4 defaults, present-zero Optional/Nullable values, absent Optional, null Nullable, union decoding, indexed union failure, null-Optional failure, and receiver rollback.

friction: The feature branch pushed successfully, but GitHub returned HTTP 503 from both GraphQL `gh pr create` and the REST pulls endpoint; the in-app browser had no available backend. PR creation and the Issue 53 limitation comment remain blocked on GitHub service recovery.
