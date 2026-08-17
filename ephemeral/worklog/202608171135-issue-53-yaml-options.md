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

correction: All consumer code is expected to be agent-authored; preserve one canonical JSON/schema field contract instead of YAML-tag flexibility.

decision: Replace native YAML union decoding with a YAML-to-JSON translator feeding the existing JSON decoder. Keep json as the default, both as the YAML opt-in, and remove the misleading yaml-only mode.

proof: A fresh temporary consumer generated through the real CLI with `--formats=both --validate` passed the comprehensive YAML decode and schema-validation tests. The generated output contained `ValidateYAML` and the thin `UnmarshalYAML` adapter and contained no native `__yamlUnmarshal__` dispatch.

proof: Generation was idempotent; `go build ./...`, uncached `go test -count=1 ./...`, `golangci-lint run`, and the website build/link check passed.

review: Fable round 01 found that generated YAML decoding started its transactional temporary from zero, unlike `json.Unmarshal` into an existing value, and that the `WithKnownFields` limitation had not been made visible. Preserve omitted receiver fields by copying the receiver before JSON decoding, add a regression case to the existing comprehensive YAML test, and document `ValidateYAML` as the strictness boundary.

review_decision: The round-01 receiver finding was directionally correct but overstated for ordinary fields because generated owners already implement custom `UnmarshalJSON`. The material divergence was omitted interface-slice fallback. The regression test therefore compares YAML and JSON decoding from identical non-zero receivers and asserts the preserved interface slice, rather than inventing broader merge semantics.

review: Fable round 02 demonstrated that a shallow receiver copy aliases slices and pointers, allowing a failed plain-struct decode to mutate caller-visible state before assignment. Preserve the explicit transactional guarantee and the thin adapter by decoding into a fresh value; document replacement rather than merge semantics. Add the simple plain-struct YAML test requested by the user alongside the existing comprehensive union test, covering successful decode, rollback of slice and pointer state, and omitted-field replacement.

review: Fable round 03 independently reproduced the plain-struct rollback fix and accepted transactional replacement as the documented contract. Consensus outcome: only nitpicks remain. Review session `9bed6e64-6ab0-4c4f-babc-1472a1c1f2dd`; artifacts are under `ephemeral/reviews/202608171244-issue-53-yaml-options-round-0{1,2,3}.md`.

proof: After the consensus changes, `go generate ./...`, `git diff --check`, `go build ./...`, uncached `go test -count=1 ./...`, and `golangci-lint run` passed. `npm ci && npm run check` rebuilt the website and verified all internal links.
