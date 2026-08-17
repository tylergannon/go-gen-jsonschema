# YAML v4 unmarshaling worklog

decision: Implement in two test-driven gates: one simple discriminator decode, then one comprehensive hard-semantics test.
decision: Keep the implementation native to yaml/v4 nodes and preserve the YAML-to-JSON bridge only as a rejected fallback unless the direct prototype exposes a fundamental blocker.
decision: Reuse existing interface discovery and AST traversal; this change extends generated decoding and YAML field-name metadata only.
correction: The repository root was initially six commits behind origin/main; fast-forwarded to 671aa9c before creating this worktree and re-established a green full-suite baseline.
proof: Gate 1 red state is a compile-time failure: `*FancyStruct does not implement libyaml.Unmarshaler (missing method UnmarshalYAML)`.
proof: Gate 1 green state directly decoded YAML into `FancyStruct.IFace` as `TestInterface1{Field1: "one"}` through the generated yaml/v4 method.
proof: Gate 2 red state reached the expected hard boundary: generated YAML lookup used JSON names, leaving `yaml_if` and `yaml_ifs` in the ordinary mapping and causing yaml/v4 to attempt generic-map assignment into the interface field.
decision: Resolve YAML field names from `yaml` tags in the existing generated interface-property metadata; do not add another scanner or traversal pass.
proof: Gate 2 green state decoded a YAML merge key, an explicit discriminator value, default discriminator values, an interface slice, an optional interface, and a nested `UnmarshalYAML`; its failure case reported `yaml_ifs[1]` and left the destination unchanged.
proof: `go generate ./...` regenerated all checked-in packages with registered interfaces and `go mod tidy` recorded `go.yaml.in/yaml/v4 v4.0.0-rc.6` as a direct dependency.
proof: `go test ./internal/builder` passed after updating the two fixture goldens.
proof: Final `go test ./...` passed across the repository.
proof: Final `go vet ./...` passed with no diagnostics.
review: `git diff --check` passed; generated methods are present in all checked-in special-type packages, and the branch started at the same commit as `origin/main` (`671aa9c`).
checkpoint: Completed on branch `codex/yaml-v4-unmarshal` with commit subject `feat: generate yaml v4 unmarshalers`; no merge or pull request was requested.
decision: Follow-up changed the global default discriminator from `!type` to `type` for both JSON and YAML; retain custom discriminator configuration as the explicit escape hatch and do not add dual-name compatibility behavior.
proof: Follow-up baseline `go test -count=1 ./...` passed before edits.
proof: Changing the simple authored YAML fixture to `type` produced the expected red runtime result while generated code still required `!type`.
proof: After changing `DefaultDiscriminatorPropName`, the focused generated YAML test passed with an unquoted `type` key.
generation: Ran `go generate ./...` at the repository root, ran each fixture module that owns a generated-code golden, copied the generated outputs into their `.golden` counterparts, and removed temporary generated fixture files.
proof: Focused builder golden/schema tests passed after regeneration; a repository search found no remaining literal `!type` contract outside historical session artifacts.
proof: Final uncached `go test -count=1 ./...`, `go vet ./...`, and `git diff --check` passed after the default and generated outputs changed.
proof: The orphaned tracked `test10-v1-enums-stringmode` generated fixture was regenerated separately, its unrelated output-mode drift was discarded, and its nested `go test ./...` passed.
checkpoint: Follow-up is ready on `codex/yaml-v4-unmarshal` with commit subject `feat!: default union discriminator to type`.
docs: Updated `README.md`, `llms.txt`, and the repo skill definition to describe native `go.yaml.in/yaml/v4` union decoding, the shared default `type` discriminator, YAML tag behavior, supported interface containers, and the required `go mod tidy` dependency step.
docs: Changed the primary union examples to exercise the default `type` contract instead of overriding it with `!kind`; retained the explicit custom-discriminator API in the reference material.
skill: Applied the write-prompts skill guidance by keeping the routing addition short, updating the complete generated-file mental model, and placing operational detail in the existing interface section rather than adding a parallel workflow.
review: The evaluate-skills pass found one stale routed reference that still described interface decoding as JSON-only; updated `references/registration-api.md` so the main skill and its delegated detail remain coherent.
proof: After the documentation and skill edits, `go test -count=1 ./...`, `go vet ./...`, `git diff --check`, and consistency searches for yaml/v4 coverage and stale `!type` wording all passed.
checkpoint: Documentation follow-up is ready with commit subject `docs: document native YAML union decoding`.
pull-request: Opened https://github.com/tylergannon/go-gen-jsonschema/pull/52 from `codex/yaml-v4-unmarshal` to `main`; the PR body calls out the `!type` to `type` wire-format break and the native yaml/v4 proof surface.
