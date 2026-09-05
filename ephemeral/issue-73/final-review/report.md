# Issue #73 Final Review — Independent Verification

Scope: full working-tree delta (staged + unstaged + untracked product files) against
commit `5a074a1`, on branch `codex/issue-73-fluent-declarations`. Verified against
`ephemeral/issue-73/manager-review/authority.json` and the user's explicit
AGENTS.md/CLAUDE.md requirement. No previous review verdicts were read before
reaching these conclusions.

## Checks actually performed

- Read `declare.go`, `declare_test.go`, `internal/syntax/fluent_expr.go`,
  the `scan_expr.go`/`scan_result.go` diffs, all four `internal/compiletest`
  negative fixtures, `internal/builder/fluent_declaration_test.go` in full,
  `internal/syntax/scanner_test.go`'s new fluent tests, `gen-jsonschema/main_test.go`,
  `internal/cmd/doc-gen/main.go`/`main_test.go`, `gen-jsonschema/tmpl/config.go.tmpl`,
  `union_type.go`, `AGENTS.md`/`Agents.md`/`CLAUDE.md`, README.md, examples/README.md,
  skills/go-gen-jsonschema/SKILL.md and references/*, and spot-checked the website docs.
- Made an isolated rsync copy of the whole worktree under
  `ephemeral/issue-73/final-review/testdata/repo-copy` (no `.git`) and, only there:
  ran `go build ./...`, `go vet ./...`, `go test ./...` (all packages, all green),
  ran the negative-compile fixture test directly and confirmed all four fixtures
  genuinely fail `go build`, ran `go generate ./...` and diffed the regenerated
  `examples/` tree against a pristine second copy (zero diff — full byte-for-byte
  regeneration determinism from the fluent-only `schema.go` sources).
- Built the `gen-jsonschema` binary from the isolated copy and, in a separate
  `/tmp` scratch module (via a `replace` directive at the isolated copy, never the
  shared worktree), ran `new` end-to-end, then `gen -validate`, confirming the
  scaffold emits `jsonschema.Declare(...)`, that generation succeeds, and that the
  generated `Schema()`/`ValidateJSON` methods actually work at runtime (wrote and
  ran a throwaway runtime test in scratch).
- Confirmed a genuine Go 1.27 language feature — methods may declare their own
  type parameters distinct from the receiver's — with a minimal isolated repro,
  which is what makes `Declaration[T].Method[F any]`/`.Function[F any]` able to
  jointly infer field/provider type agreement at compile time as the issue asks.
  (An IDE/LSP diagnostic flagged this as illegal Go; the real `go1.27.1` toolchain
  disagrees and is authoritative — false alarm, not a finding.)
- Verified `CLAUDE.md` is a real working-tree relative symlink (`lrwxr-xr-x
  CLAUDE.md -> AGENTS.md`, git blob mode `120000`, index target string `AGENTS.md`),
  `AGENTS.md` is the sole canonical file in the git index (`git ls-files -s` shows
  only `AGENTS.md`; `Agents.md`/`AGENTS.md` collide to one inode because the
  filesystem is case-insensitive, and git's index agrees there is one file), and
  that its content is an accurate merge of the old `Agents.md` operating rules and
  the old `CLAUDE.md` architecture doc, updated for `Declare(...)`.
- Re-read `union_type.go` and `website/src/content/docs/api/index.md` after the
  concurrent docs-only pass reported complete (see finding below).

## Result: SHIP

The fluent `Declare[T]` API matches the authority issue's scope precisely: typed
chain for schema callables/accessors/methods/functions/enums/refs/render-providers/
interfaces, root type inferred from the schema callable, mismatched receiver and
field/provider types rejected at compile time via genuine Go generics (proven by
four independently-compiled negative fixtures, not just asserted), value and
pointer roots both supported, and the legacy registration model reused verbatim —
no second schema/codec path was created. Every legacy marker function
(`NewJSONSchemaMethod`, `NewJSONSchemaFunc`, `NewInterfaceImpl`, `NewEnumType`,
and every `With*`/`As*` option) carries a correct `Deprecated:` godoc pointing at
its fluent equivalent, remains source-compatible, and is proven so by dedicated
parity tests that render both forms through the real builder and assert
byte-identical JSON output — including a pointer-root parity test that documents
and locks in a real bug the fluent work found and fixed (`providerRef` originally
dropped the provider for a `(*T).Method` pointer-root chain). Primary docs
(README, examples/README, SKILL.md, registration-api.md, examples.md,
website pages, scaffolder template) all teach `Declare(...)` as the only
supported form in tutorial code; every remaining legacy mention is explicitly a
migration/compatibility note, matching the issue's requirement to remove legacy
syntax from primary teaching material without deleting the option. The scaffolder
(`gen-jsonschema new`) emits the fluent form and was proven to work through a real
`new` → `gen -validate` → compile → runtime-validate pipeline in scratch. Full
regeneration determinism against every checked-in example was reproduced
independently, and `go test ./...` is green. The user's AGENTS.md/CLAUDE.md
requirement is met exactly: `AGENTS.md` is canonical, `CLAUDE.md` is a relative
symlink to it, and the merged content is accurate.

No functional, correctness, or acceptance-criterion gaps were found. One doc
nit below; it does not block shipping.

## Findings

### Nit: `NewEnumType`'s Deprecated comment still omits the nested-shape caveat the fuller docs state

- `union_type.go:167-170` (godoc on `NewEnumType`), mirrored verbatim in the
  regenerated `website/src/content/docs/api/index.md:219`
- The Deprecated comment says `.Enum`/`NewEnumType` compatibility is only about
  a shared enum type across multiple fields: *"NewEnumType has no fluent
  replacement and must be retained when the enum type is shared across more
  than one struct field."* README.md:264-270 and
  skills/go-gen-jsonschema/references/registration-api.md:70-76 state a second,
  independent reason to keep `NewEnumType`: `.Enum`/`.StringerEnum` only support
  a direct named-enum, `Optional[E]`, or `Nullable[E]` field shape — *not*, e.g.,
  `Optional[[]E]` — regardless of whether the field is shared. A reader who only
  sees the terse godoc (the first thing golangci/godoc/gopls surfaces) could
  reasonably conclude a single unshared `Optional[[]Status]` field is a
  `.Enum(...)`-compatible case, try it, and hit a generation-time rejection the
  docstring gave them no reason to expect.
- I was told a bounded docs-only pass had just landed specifically to correct
  this comment to match the fuller caveat; re-reading `union_type.go` and the
  regenerated `api/index.md` after that pass reported complete shows byte-identical
  text to what was there beforehand — the nested-shape sentence was not added to
  either file. Flagging so the manager can confirm whether that pass's diff
  actually landed as intended.
- Not a functional defect (the correct, complete guidance exists in README and
  the shipped skill, which is what a real user/agent following the Quick Start
  or skill will read) and not a new problem introduced by the fluent API itself
  — it's a pre-existing terse-vs-full-doc gap made slightly more visible by this
  issue's doc pass touching the surrounding paragraph. Does not block shipping.

## Explicitly out of scope / not flagged

- Untracked `internal/builder/testfixtures/{asref_collision_*,unsupported_interface_container_*}`
  directories found in the shared worktree at review start: these are debris
  from some earlier interrupted `go test` run (the relevant test files —
  `asref_collision_test.go`, `unsupported_interface_containers_test.go`,
  `owner_codec_test.go`, `enum_codec_test.go` — are unmodified by this issue and
  already used this `os.MkdirTemp(cwd/testfixtures)` + `t.Cleanup` pattern before
  5a074a1). The one new file using this pattern, `fluent_declaration_test.go`,
  registers `t.Cleanup` correctly and left no debris in the isolated run. Not a
  product or test-hygiene regression from this change; not touched or cleaned up
  by this review per the read-only mandate.
- `.github/workflows/website-pages.yml`'s `go-version: '1.24.x'` → `go-version-file:
  go.mod` is a real, necessary fix (go.mod already requires `go 1.27`, which this
  issue's code needs for method-level type parameters; the other two workflows
  already used `go-version-file`), not scope creep.
