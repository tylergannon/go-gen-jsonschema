# Adversarial Review — Issue #53 / PR #54 (YAML through the JSON contract), round 01

**Target:** branch `codex/issue-53-yaml-options` @ `9affe7a`, diff vs `origin/main`
(36 files, +828 / −965). Closes issue #53 ("native YAML owners cannot decode
ordinary Optional fields and ignore KnownFields").

**Authoritative sources used:** repository `CLAUDE.md`, GitHub issue #53, PR #54
body, the caller's restated user context (opt-in yaml/v4 → JSON translation into
the single canonical JSON decoder; JSON-only default; `both` adds YAML; schema
validation is the validation contract).

**Caller constraints honored:** read-only except this artifact. No caller
instruction narrowed the defect space, predicted findings, or declared safe
areas, so no narrowing was ignored.

## Evidence inspected

- Full diff vs `origin/main` for `internal/builder/schemas.go.tmpl`,
  `gen_schema.go`, `builder.go`, `gen-jsonschema/main.go`,
  `gen-jsonschema/tmpl/config.go.tmpl`.
- Generated goldens: `internal/builder/testfixtures/v1_interfaces_options/jsonschema_gen.go.golden`,
  `test_run/test9-v1-interfaces-options/*`, `test5-interfaces` goldens.
- Tests: `optionality_test.go` (both copies), `unmarshal_formats_test.go`,
  `gen-jsonschema/main_test.go`.
- Docs: `README.md`, `llms.txt`, `skills/go-gen-jsonschema/SKILL.md` +
  `references/registration-api.md`, `website/src/content/docs/**`.
- Proof: `ephemeral/worklog/202608171135-issue-53-yaml-options.md`; re-ran
  `go build ./...` and `go test ./...` (all pass).
- **Independent proof:** built the CLI from this branch (`go build ./gen-jsonschema`)
  and generated a fresh throwaway consumer in `/tmp/yprobe` with
  `gen --validate --formats=both` over a plain struct
  (`Name string`, `Label jsonschema.Optional[string]`, `Tags []string`), then ran
  probe tests against the real generated output. Repros below are from that run.

The core direction is right and materially simpler than what it replaces:
`__jsonUnmarshal__`/`__yamlUnmarshal__` duplication, the hand-built
`__jsonschema__yamlMappingNode`, and the `yaml`-tag name-resolution path are all
gone, replaced by one translation shim feeding the canonical JSON decoder. Union
dispatch, interface slices, `Optional[I]`, custom `UnmarshalJSON`, and
transactional receiver assignment are genuinely preserved, and the fixture tests
cover absent / present-zero / null / rollback. Findings below are what survived.

---

## Findings

### 1. YAML decoding into a non-zero destination silently discards fields the document omits — JSON does not. (issue)

`internal/builder/schemas.go.tmpl:226-236` generates:

```go
func (n *Node) UnmarshalYAML(node *yaml.Node) error {
	data, err := __gen_jsonschema_yamlNodeToJSON(node)
	if err != nil { return err }
	var next Node                       // <-- zero value, not *n
	if err := json.Unmarshal(data, &next); err != nil { return err }
	*n = next
	return nil
}
```

`var next {{.Name}}` starts from the **zero value**, so any field absent from the
YAML document is reset rather than left alone. `encoding/json` merges into the
existing destination; the generated YAML path replaces it. Issue #53's Expected
section requires ordinary presence-aware fields to decode "with the same absent,
present-zero, and null semantics as JSON", and the PR body claims semantics are
"preserved … through the shared decoder".

Reproduction (`/tmp/yprobe`, real CLI output, `--formats=both --validate`):

```go
mk := func() Node { return Node{Name: "orig", Tags: []string{"a"}} }

j := mk(); json.Unmarshal([]byte(`{"name":"new"}`), &j)
// JSON: Node{Name:"new", Tags:[]string{"a"}}

y := mk(); yaml.Load([]byte("name: new\n"), &y, yaml.WithV4Defaults())
// YAML: Node{Name:"new", Tags:[]string(nil)}   <-- Tags wiped
```

The same divergence hits generated union owners harder: `UnmarshalJSON` contains
the deliberate `__next.IFaces = o.IFaces` fallback for an absent interface slice
(see `v1_interfaces_options/jsonschema_gen.go.golden`), but under YAML the
receiver `o` is always the zero `next`, so that fallback can never fire. The
branch's own fixture test never exercises this because every YAML case decodes
into a fresh `var got Owner` or supplies `ifs:` explicitly.

**Impact:** two decoders with the same advertised contract produce different
results for the same logical document whenever the destination carries defaults
or prior state — silently, with no error. **Fix is one line:** `next := *n`
before `json.Unmarshal`, which reproduces `json.Unmarshal(data, n)` exactly while
keeping assignment transactional. A regression test decoding into a pre-populated
destination should accompany it.

### 2. Issue #53 Problem 2 (`WithKnownFields`) is neither fixed nor recorded anywhere. (issue)

Reproduction against the freshly generated consumer:

```go
var n Node
err := yaml.Load([]byte("name: a\ntags: []\nunknown: true\n"), &n,
	yaml.WithV4Defaults(), yaml.WithKnownFields())
// err = <nil>, "unknown" silently discarded
```

This is verbatim the behavior issue #53 filed as Problem 2, whose Expected bullet
reads "Strict known-field loading is not silently weakened inside generated
custom unmarshalling." I accept the authoritative design answer — schema
validation, not decoder options, is the validation contract, and `ValidateYAML`
does reject `additionalProperties` (fixture test `TestGeneratedYAMLValidationUsesJSONSchema`
confirms it). The defect is that the *silence* was never addressed:

- `grep -rn KnownFields` over the whole repo returns **zero hits** — not in
  `README.md`, `llms.txt`, `SKILL.md`, `website/`, the template, or any test.
- The PR #54 body does not mention it; issue #53 has 0 comments.
- The branch's own worklog decided to "record that opt-in decoder behavior as a
  narrow limitation … in issue or PR context" — that step was not carried out.

The new design also *widens* the silence: previously only union owners overrode
`UnmarshalYAML`, now **every** registered type does
(`gen_schema.go:1124-1141` seeds `YAMLTypes` from all `SchemaMethods()` plus all
`SpecialTypes`), so `WithKnownFields` is now inert for every registered type in a
`--formats=both` package.

**Impact:** a user who deliberately opts into strict loading gets a no-op with no
error and no documentation. Required remedy is small: one documented limitation
sentence beside the existing "Go `yaml` struct tags are ignored" note in
`README.md` / `llms.txt` / `SKILL.md`, pointing at `ValidateYAML` as the
supported strictness mechanism.

### 3. Validation and decoding can interpret the same document under different YAML dialects. (nitpick)

`schemas.go.tmpl:51-57` hardcodes `yaml.WithV4Defaults()` inside
`__gen_jsonschema_yamlToJSON` (used by `ValidateYAML`), while
`__gen_jsonschema_yamlNodeToJSON` (used by `UnmarshalYAML`) inherits whatever
options the caller passed to `yaml.Load` / `yaml.Unmarshal`. Legacy
`yaml.Unmarshal` still works against generated types (verified: it decodes
`Node` correctly), so a caller can validate under v4 resolution and then decode
under v3 resolution. The validate-then-unmarshal workflow the README recommends
is exactly where this bites. Worth either documenting that YAML entry points
assume v4 defaults, or accepting options on the validation helper.

### 4. Translation failures surface raw `encoding/json` internals as YAML errors. (nitpick)

A YAML mapping with a non-string key produces:

```
yaml: construct errors: line 1: json: unsupported type: map[interface {}]interface {}
```

The documentation's phrasing — "YAML constructs that cannot be represented by
JSON are rejected" — is correct behavior, but the message names a Go type the
user never wrote and does not identify the offending key or the actual rule.
Wrapping the `json.Marshal` error in `__gen_jsonschema_yamlNodeToJSON`
(`schemas.go.tmpl:42-48`) with something like "YAML value cannot be represented
in the JSON data model" would make the documented rule discoverable.

### 5. Mixed naming when a registered type nests inside an unregistered struct is undocumented. (nitpick)

Because `UnmarshalYAML` is generated per registered type, a document decoded into
an *unregistered* outer struct resolves the outer keys by yaml/v4's normal rules
(Go field names / `yaml` tags) and the inner registered type's keys by JSON
Schema property names. The docs state the rule for registered types
("YAML uses the JSON Schema property names. Go `yaml` struct tags are ignored")
but never state that the rule stops at the registration boundary, so a partially
migrated consumer gets two naming conventions in one file with no diagnostic.

---

## Not findings (checked and cleared)

- Union dispatch, `Optional[I]`, pointer-vs-value impls, indexed slice errors,
  and rollback on failure all still work through the shared JSON decoder
  (fixture tests + independent generation).
- `--formats=json` (and the zero value) emits no `go.yaml.in/yaml/v4` import and
  no YAML symbols — `unmarshal_formats_test.go` asserts both directions.
- Removal of `--formats=yaml` is a coherent consequence of the design (a
  YAML-only mode cannot exist when YAML is defined as translation into the JSON
  decoder); the CLI rejects it with a clear message and docs/tests were updated
  consistently, with no stale `--formats=yaml` references left in the tree.
- Numeric fidelity survives the round trip (int64 max and uint64 values decode
  exactly); `!!timestamp` marshals to RFC3339, matching the `time.Time` contract.
- `new --formats=both --validate` correctly emits the `ValidateYAML` stub and is
  covered by `TestNewConfigValidationStubsFollowFormats`.
- `go build ./...` and `go test ./...` pass on this branch.

---

**Outcome:** material findings remain
