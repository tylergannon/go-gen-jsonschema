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
