# Adversarial Review — Issue #53 / PR #54 (YAML through the JSON contract), round 03

**Target:** branch `codex/issue-53-yaml-options` @ `9affe7a` **plus the uncommitted
working tree** (15 modified files, 4 new untracked fixture artifacts), diff vs
`origin/main`: 37 files, +1121 / −953. Closes issue #53.

**Authoritative sources used:** repository `CLAUDE.md`, GitHub issue #53, PR #54
body, and the caller's restated user context from round 01 (opt-in yaml/v4 → JSON
translation into the single canonical JSON decoder; JSON-only default; `both`
adds YAML; schema validation is the validation contract).

**Caller constraints honored:** read-only except this artifact; rounds 01 and 02
left intact. No caller instruction narrowed the defect space, predicted findings,
declared safe areas, or requested a verdict, so no narrowing was ignored.

## Evidence inspected

- Working-tree delta since round 02: `internal/builder/schemas.go.tmpl`
  (reverted `next := *r` → `var next T`), new `Plain`/`PlainInner` fixture types
  and registrations in both `testfixtures/v1_interfaces_options` and
  `test_run/test9-v1-interfaces-options` (`types.go`, `schema.go`), new
  `TestGeneratedYAMLUnmarshalSimple`, regenerated `jsonschema_gen.go` +
  goldens, new `jsonschema/Plain.json{,.golden,.sum}`, `basic_test.go` file
  list, and doc updates in `README.md`, `llms.txt`, `SKILL.md`,
  `website/src/content/docs/guides/validation-ci.md`, plus the worklog.
- Full committed diff vs `origin/main` re-read (`gen_schema.go`, `builder.go`,
  `gen-jsonschema/main.go`, `tmpl/config.go.tmpl`, `unmarshal_formats_test.go`,
  `main_test.go`, `website/**`, `skills/**`).
- `internal/syntax/node_wrappers.go` (tag accessors) for orphaned code.
- `go build ./...` and `go test ./...` — all pass; `Plain.json` byte-identical to
  its golden.
- **Independent proof:** rebuilt the CLI from the working tree and generated a
  fresh throwaway consumer (`/tmp/yp`) with `gen --validate --formats=both` over
  a plain registered struct (`Tags []string`, `Inner *Inner`, `Count int`) plus
  an **unregistered** outer struct, then re-ran the round-02 reproductions and
  new probes (null document, nested null, multi-document, mixed naming,
  `WithKnownFields`, v3-vs-v4 resolution drift).

## Prior findings: status

| Round/# | Finding | Status |
|---|---|---|
| 01/1 | YAML decode discarded omitted fields (merge divergence) | **Resolved by decision, not by code.** The branch reverted to `var next T` and now *documents* replacement semantics explicitly (`README.md:320-326`, `llms.txt:290-295`, `SKILL.md:185-190`). Verified: `yaml.Load("count: 1", &prepopulated)` yields `{Tags:nil, Inner:nil, Count:1}`. This is a defensible reading of the authoritative contract ("single canonical JSON decoder", schema validation as the contract) and is no longer silent. I accept it. |
| 01/2 | `WithKnownFields` silent no-op, undocumented | **Fixed** — documented in README, llms.txt, SKILL.md, and now the website. Claim re-verified against yaml/v4 source. |
| 02/1 | `next := *r` aliasing let a failed decode mutate caller state | **Fixed and independently re-verified.** Both round-02 reproductions (slice backing array, pointer target) now leave the destination and all external references untouched on a failed decode. |
| 02/2 | Regression test only covered a type immune to the tested path | **Fixed.** `Plain`/`PlainInner` are plain registered types with no interface fields, and `TestGeneratedYAMLUnmarshalSimple` (`optionality_test.go:13-52`, both copies) covers successful decode, rollback of *both* slice and pointer state after a mid-document type error, and omitted-field replacement — exactly the uncovered class. |
| 02/3 | Validate/decode YAML dialect mismatch | **Withdrawn.** I probed for an actual divergence (`count: 010` and friends through `yaml.Unmarshal` v3 defaults vs `yaml.Load` + `WithV4Defaults` vs `ValidateYAML`) and could not produce one — the node-level decode resolves identically on both paths. The docs now also advise `WithV4Defaults()`. My round-02 claim was speculative and is retracted. |
| 02/4 | Raw `encoding/json` error leakage | Not addressed — see finding 3. |
| 02/5 | Naming boundary / website doc gap | Partially addressed (website got the `WithKnownFields` note) — see findings 2 and 4. |

The implementation is now in good shape. The decoder is a genuinely thin
translation shim, the guarantees it advertises (transactional replacement,
JSON-canonical names, schema-backed strictness) match what the generated code
actually does, and the fixture suite finally exercises the plain-struct path.
**No material findings remain.** What follows are genuine nitpicks.

---

## Findings

### 1. Dead code left behind by the refactor: `StructField.YAMLTag()` now has no callers. (nitpick)

`internal/syntax/node_wrappers.go:771-773`:

```go
func (f StructField) YAMLTag() *structtag.Tag {
	return f.structTag("yaml")
}
```

It was introduced by PR #52 (`ccb055d feat!: add native YAML union decoding`)
and its only consumer, `InterfaceProp.YAMLName()`, was deleted by this branch
(`internal/builder/gen_schema.go`, removed alongside `YAMLUnmarshalerFunc`). A
repo-wide grep for `YAMLTag` over `*.go` and `*.tmpl` returns exactly one hit —
the definition. Nothing in production code, tests, or templates reads Go `yaml`
tags any more, which is the stated intent.

**Impact:** cosmetic, but this accessor is the last remaining hook of the
abandoned yaml-tag naming path; leaving it in `internal/syntax` invites someone
to wire it back up and quietly reintroduce dual naming. Deleting it makes the
"JSON tags are canonical" decision structurally enforced rather than merely
documented.

### 2. The naming rule silently stops at the registration boundary, and that is still undocumented. (nitpick)

Because `UnmarshalYAML` is generated per registered type, a document decoded into
an **unregistered** outer struct resolves outer keys by yaml/v4's normal rules
and inner registered-type keys by JSON Schema property names. Verified on the
fresh consumer (`Outer` unregistered, `Node` registered):

```go
var o Outer   // Outer.Node has `yaml:"node_yaml"`, Node's fields use json tags only
yaml.Load([]byte("node_yaml:\n  tags: [x]\n  inner: {a: p, b: q}\n  count: 1\nextra: hi\n"),
	&o, yaml.WithV4Defaults())
// err = <nil>; both conventions apply in one document, no diagnostic
```

The docs state the rule for registered types ("YAML uses the JSON Schema property
names. Go `yaml` struct tags are ignored") but never say the rule applies only
below a registration boundary. A partially migrated consumer therefore gets two
naming conventions in one file with nothing to signal it. One clause in the
existing YAML paragraph would close this.

### 3. Translation failures still surface raw `encoding/json` internals. (nitpick — carried, unaddressed)

A non-string YAML mapping key still yields
`yaml: construct errors: line 1: json: unsupported type: map[interface {}]interface {}`.
The behavior matches the documented rule ("YAML constructs that cannot be
represented by JSON are rejected"), but the message names a Go type the user
never wrote and does not identify the offending key. Wrapping the `json.Marshal`
error in `__gen_jsonschema_yamlNodeToJSON` (`schemas.go.tmpl:42-48`) would make
the documented rule discoverable from the error alone.

### 4. The website is the only doc surface missing the replacement-semantics sentence. (nitpick)

`website/src/content/docs/guides/validation-ci.md:50-56` received the
`WithKnownFields` limitation and the `WithV4Defaults()` advice, but not the
third sentence added everywhere else: "Decoding is transactional replacement, so
omitted YAML fields do not retain receiver values"
(`README.md:323-325`, `llms.txt:293-294`, `SKILL.md:188-190`). Replacement vs.
`encoding/json`'s merge behavior is the single most surprising consequence of
this design, and the website is the published surface, so it is the one place
that most needs the sentence.

### 5. PR #54 as published does not contain any of the three rounds of remediation. (nitpick — process)

`git rev-parse HEAD`, `origin/codex/issue-53-yaml-options`, and
`gh pr view 54 --json headRefOid` all return `9affe7a`. Every fix reviewed above
— the transactionality revert, the `Plain` fixture and its test, and all four
doc surfaces — exists only in the uncommitted working tree. Anyone reading PR #54
today sees the round-01 state, and the PR's "Proof" section describes a
`UnmarshalYAML` implementation that has since changed twice. Committing and
pushing before requesting merge review, and refreshing the PR Proof section to
match the final shape (replacement semantics, the plain-struct regression test),
would keep the published proof honest.

---

## Not findings (checked on the current tree)

- Round-02's aliasing reproductions are dead: failed decodes leave the
  destination, a caller-held slice, and a caller-held pointee all untouched.
- `Plain.json` is generated correctly (`inner` inlined, all three fields
  required, `additionalProperties: false`) and matches its golden byte for byte;
  `basic_test.go` was updated to expect it; generation remains idempotent.
- A whole-document `null` leaves the receiver unchanged with no error — but
  `encoding/json` behaves identically for the same input, so YAML still tracks
  the canonical decoder, and `ValidateYAML` rejects it (`got null, want object`).
  Consistent with "schema validation is the validation contract".
- Multi-document YAML is rejected identically by `UnmarshalYAML` and
  `ValidateYAML`.
- v3-default (`yaml.Unmarshal`) and v4-default (`yaml.Load` + `WithV4Defaults`)
  decoding produced identical results in every case I probed; the documented
  `WithV4Defaults()` advice is sufficient.
- `--formats=json` and the zero value still emit no yaml/v4 import and no YAML
  symbols; `--formats=yaml` is cleanly rejected; no stale references remain.
- `ValidateYAML` is correctly gated behind `--validate`; `new --formats=both
  --validate` emits the matching stub and is covered by
  `TestNewConfigValidationStubsFollowFormats`.
- The worklog records rounds 01 and 02, the remediations, and — notably — the
  reversal of the round-01 fix with its reasoning. The proof trail is honest
  about what changed and why.

---

**Outcome:** only nitpicks remain
