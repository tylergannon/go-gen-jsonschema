# Issue #73 closeout: bounded corrections — result

Authority: `ephemeral/issue-73/manager-review/authority.json` +
`ephemeral/issue-73/finish/adjudication.md`. Scope: only the corrections the
adjudication accepted. No commits/pushes. Prior worklogs and evidence
directories preserved and built on, not replaced.

## 1. Finding 1 — fluent provider-shape mismatch silently produced a nil-panic

`internal/syntax/fluent_expr.go`'s `parseFluentChainOptions` (Accessor/Method/
Function case) now enforces, for the fluent path only:

- `.Accessor`/`.Method` require a receiver method expression; a free function
  is a hard, source-positioned error (`... provider must be a %s method
  expression, not a free function, at %s`).
- `.Function` requires a free function; a method expression is a hard error
  (`... .Function provider must be a free function, not a %s method
  expression, at %s`).
- `providerRef`'s `matched=false` case (a method-expression selector whose
  receiver doesn't match what the chain expects — e.g.
  `Function(fieldOfOtherType, OtherType.Method)`) is now also a hard error
  instead of the silent `continue` it fell through to before. This is the
  same finding surfacing through a second code path (Go's type system can't
  catch it here either, since `.Function`'s provider type is `func(F)
  json.Marshaler` where `F` is the field's own type, not the Declare root),
  not new-shape support.

`providerRef`'s doc comment, which incorrectly asserted the mismatch was
"unreachable in practice," is corrected to explain why it's reachable and
that the fluent caller (not `providerRef` itself) is responsible for
rejecting it.

**Explicitly not touched:** `providerRef` itself, and the legacy
`WithStructAccessorMethod`/`WithStructFunctionMethod`/`WithFunction`
variadic-option parser — both still silently skip the analogous mismatch,
exactly as before. Verified live: re-running the review's
`ephemeral/issue-73/manager-review/scratch/repro_legacy_accessor_free_func`
fixture through `go run .../gen-jsonschema/` still exits 0 (unchanged legacy
behavior). Re-running `repro_accessor_free_func` (the fluent repro) now
fails fast with `jsonschema.Declare: .Accessor provider must be a Example
method expression, not a free function, at .../schema.go:13:9`.

## 2. Finding 2 — no scanner/CLI test exercised a fluent error path

Three new tests, one exact concrete case each (no combinatorial suite):

- **Scanner** (`internal/syntax/scanner_test.go`,
  `TestFluentChainFieldSelectorMismatchFailsToLoad`): a new standalone
  fixture package, `internal/syntax/testfixtures/fluent_field_mismatch/`
  (kept out of the shared `typescanner` happy-path fixture, since
  `LoadPackage` eagerly parses every marker in a package and would break
  every other fluent scanner test if this lived there). Its `.Enum(...)`
  chain link names a field on a type other than the `Declare(...)` root;
  `Load` + `LoadPackage` return a source-positioned error naming
  `fixture.go`.
- **Command** (`internal/builder/fluent_declaration_test.go`, via `New(...)`
  — the same entry point `gen-jsonschema gen` calls):
  `TestFluentAccessorRejectsFreeFunctionProvider` (the review's reproduced
  case) and `TestFluentFunctionRejectsUnrelatedMethodExpressionProvider` (the
  `matched=false` path). Both assert the error text and that it names
  `schema.go`.
- **Actual CLI subprocess** (`gen-jsonschema/main_test.go`,
  `TestGenCommandRejectsInvalidFluentFieldAssociationWithSourcePosition`):
  runs the real built command (`go run .`, via the existing
  `testutils.RunCommand` helper already used elsewhere in the repo for
  generate/build round-trips) against the same
  `fluent_field_mismatch` fixture and asserts a non-zero exit and stderr
  naming the file.

All four pass; full `go test ./...` is green.

## 3. Manager findings missed by the prior docs pass

- **llms.txt** still taught legacy syntax throughout (scaffold snippet,
  enums, discriminated interfaces, `AsRef`, provider-rendered schemas, and
  the full registration table). Converted every primary example to
  `Declare(...)` chains with one "Migration:" note per section, following
  the same pattern already used in README.md/website docs; kept a compact
  legacy-marker → fluent-equivalent table for completeness (llms.txt is a
  comprehensive reference, unlike README's trimmed table).
- **`website/src/content/docs/api/index.md`** had no `Declare`/`Declaration[T]`
  docs. Root cause was not "gomarkdoc can't parse `declare.go`'s generic
  methods" (the prior pass's conclusion) but the website CI pinning Go
  1.24 for the gomarkdoc prebuild step, one release short of the go1.27
  generic-methods grammar gomarkdoc's installed `go/parser`-based dependency
  needs. Reproduced locally (`GOTOOLCHAIN=go1.27.1 gomarkdoc -e -o
  './src/content/docs/api/index.md' ../` from `website/`) — clean run,
  full `Declare`/`Declaration[T]` docs and every `Deprecated:` comment now
  present. Fixed `.github/workflows/website-pages.yml`'s "Set up Go" step to
  `go-version-file: go.mod` (matching `go.yml`/`typescript.yml`'s existing
  convention) instead of a hardcoded `1.24.x`, then regenerated
  `api/index.md` for real with the local toolchain and confirmed
  `npm run check` (full `astro build` + internal-link check, 17 pages) passes
  end to end from a clean `prebuild` (gomarkdoc + llms.txt copy) through
  `check-links`.

## 4. Enum migration accuracy (validator's caveat)

Field-level `.Enum`/`.StringerEnum` is not a full replacement for
package-level `NewEnumType[T]()` when the enum type is shared across more
than one struct field (confirmed by the validator's `ts-fluent-fixture`
finding: `Optional[[]Status]` can't take a field-level enum option at all,
and partially annotating a shared type degrades the unmarked occurrences).
Corrected the migration text to say so — no new API or codec behavior — in:
`README.md`, `llms.txt`, `website/src/content/docs/features/enums.md`,
`skills/go-gen-jsonschema/references/registration-api.md`, and
`CLAUDE.md`/`AGENTS.md` (see §6). Left the actual `examples/ref_types`
`NullableConfig.Mode` registration on the legacy form, matching the prior
pass's own decision there (an unrelated description-loss finding, not the
sharing issue).

### `ts-fluent-fixture` corrected to retain the global registration on both sides

`ephemeral/issue-73/manager-validation/ts-fluent-fixture/fluent/schema.go`
had dropped `NewEnumType[Status]()` and tried to compensate with field-level
`.Enum(Composition{}.D)`/`.Enum(Envelope{}.Status)` — exactly the
non-equivalent conversion the validator's own report flagged. Fixed to keep
`jsonschema.NewEnumType[Status]()` (matching the legacy fixture) and drop the
now-unnecessary field-level `.Enum` calls for `Status`. Regenerated both
fixtures with the rebuilt CLI
(`ephemeral/issue-73/manager-validation/bin/gen-jsonschema`) and diffed:

```
diff -u legacy/jsonschema/Composition.json fluent/jsonschema/Composition.json   # empty
diff -u legacy/jsonschema/Envelope.json    fluent/jsonschema/Envelope.json      # empty
diff -u legacy/jsonschema/Detail.json      fluent/jsonschema/Detail.json        # empty
diff -u legacy/ts/types.ts                 fluent/ts/types.ts                   # empty
diff -u legacy/ts/index.ts                 fluent/ts/index.ts                   # empty
```

All five now byte-identical (previously `Composition.json`/`types.ts`
differed). Both `ts/` outputs still compile clean with the repo's pinned
`tests/typescript/node_modules/typescript` compiler (`tsc --project ...
--pretty false`, exit 0 both).

## 5. Recovered discarded runtime smoke tests

Per the adjudication's proof-gap note, recovered the two throwaway tests the
validator ran and then `rm`'d, verbatim from
`ephemeral/issue-73/manager-validation/run/events/000001-validate.jsonl`
(events 62 and 87), and committed them as permanent files under the
validator's existing fixture directories (no new framework):

- `ephemeral/issue-73/manager-validation/scaffold-demo/{main_check.go,demo_test.go}`
  — `TestSchemaRuntime`, proving the scaffolded `Declare(Widget.Schema)`
  produces a working runtime `Widget{}.Schema()` with the expected
  `required` set.
- `ephemeral/issue-73/manager-validation/pointer-provider-fixture/runtime_test.go`
  — `TestPointerRootRenderedSchemaExecutesProviders`, proving a pointer-root
  `Declare((*Thing).Schema).Accessor(...).Method(...)` chain's providers
  actually execute at runtime.

Both are separate nested Go modules (their own `go.mod` + `replace` back to
this worktree), so they're outside this repo's own `go test ./...` and don't
affect the main suite. Both pass standalone (`go build ./... && go vet
./... && go test ./... -v`, confirmed this pass).

## 6. AGENTS.md / CLAUDE.md unification

Per an in-turn correction, `AGENTS.md`/`CLAUDE.md` were discovered to already
collide on this case-insensitive filesystem — the tracked file was actually
`Agents.md` (mixed case; git is case-sensitive even though the filesystem
isn't), separate from the plain `CLAUDE.md` regular file. Unified them:

- Merged content into one canonical file: `Agents.md`'s existing
  baseline-testing and Session Worklog protocol rules first (verbatim,
  unmodified), followed by the project-specific guidance previously only in
  `CLAUDE.md` (Project Overview, Commands, Architecture incl. the
  fluent-`Declare` Registration System update from §0, Test Structure) — no
  duplicated rules.
- Two corrections to that merged content, since two of the old `CLAUDE.md`
  claims were already stale before this pass and shouldn't propagate as
  "current guidance": Optional fields are `jsonschema.Optional[T]` with
  `json:",omitzero"` (not `json:",omitempty"`, which only affects Go
  marshaling — `AGENTS.md` had this right from the skill's own docs);
  `ValidateJSON` is opt-in via `--validate` with a generated-file build-tag
  stub, not automatic with no stub.
- Renamed the tracked file to the exact uppercase `AGENTS.md` (required for
  case-sensitive Linux) via a two-hop `git mv` through a temporary name
  (macOS's case-insensitive filesystem can't rename directly to a
  same-spelled-different-case target in one hop), then `git add` to make
  sure the staged blob reflects the merged content rather than a stale
  index entry from the first `git mv` hop (verified via `git show
  :AGENTS.md`, 108 lines, before proceeding).
- Replaced the regular `CLAUDE.md` file with a relative symlink,
  `CLAUDE.md -> AGENTS.md` (`git diff --cached` shows `new file mode
  120000`).
- Staged (not committed) per instruction: `git status --porcelain --
  AGENTS.md Agents.md CLAUDE.md` shows `A  AGENTS.md`, `D  Agents.md`,
  `T  CLAUDE.md`.
- Confirmed no other tracked file references the old `Agents.md` spelling
  outside historical worklogs/reviews (left untouched, per "preserve
  existing worklogs").
- `go build ./...` and `go test ./...` unaffected (neither file is a build
  input).

## 7. Remaining primary-example syntax migration

The original adjudication's "complete straightforward primary-example syntax
migration where fluent equivalents exist" bullet was still incomplete after
§3-§4: `examples/basictypes`, `examples/structs`, and `examples/indirecttypes`
still registered every type with `NewJSONSchemaMethod(T.Schema)` (plain
roots, no options — the simplest possible case), `examples/enums` used it for
its four plain roots, and `examples/enums_stringmode` still taught
`WithStringerEnum` directly. Converted all of these to `Declare(...)`/
`.StringerEnum(...)`.

`examples/uniontypes` also still used the legacy package-level
`NewInterfaceImpl[Shape](...)`/`NewInterfaceImpl[PaymentMethod](...)` for two
interfaces each used at exactly one field (`Drawing.Shapes`,
`Payment.Method`) — the same shape `examples/sealed_interface_slices`
already converted in the prior pass. Converted both to per-field
`.Interface(field, Impl(value, impl), ...)`, supplying the exact
Go-type-name discriminator values the legacy split form derives
automatically (read directly off the checked-in `Circle.json`/`Rectangle.json`/etc.,
e.g. `Impl("Circle", Circle{})`) so the union schemas stay byte-identical.

Kept on the legacy form, with a concise inline comment explaining why:
`examples/enums`'s `NewEnumType[Status]()` — `Status` is used both as a
struct field (`Task.Status`) and as a bare slice-element type
(`SliceOfStatus`), and field-level `.Enum` has nothing to attach to for the
latter — plus `NewEnumType[Priority]()`, kept alongside it for consistency
even though `Priority` (single-use) would also work as
`.Enum(Task{}.Priority)`.

Left untouched, as internal compatibility/limitation-demonstration fixtures
rather than teaching examples (each carries its own "PROBLEM"/"DESIRED"/
"CURRENT REALITY"/"THIS WILL PANIC" comments documenting a specific legacy
edge case, not a style choice): `examples/iota_global`,
`examples/template_rendering`, `examples/self_contained`,
`examples/test_options`.

Verification: `go generate ./...` in each converted directory, then
`git status --porcelain` on that directory's `jsonschema/`,
`jsonschema_gen.go`. Every JSON artifact is byte-identical for all six
directories. `examples/uniontypes/jsonschema_gen.go` shows a cosmetic diff
only — the same five `Schema()` method bodies in a different declaration
order, because the registration order in `schema.go` changed (verified by
reading the full diff; no content or behavior difference). `go build
./...`, `go vet ./...`, `go test ./...` (full repo, including
`TestAdvertisedExamplesHaveGenerateDirectiveAndArtifacts`), and `go run
./internal/cmd/doc-gen -check` are all clean after this pass (none of these
six directories are doc-gen example sources, so no regeneration was needed
there).

## What passed

```
go build ./...                              # clean
go vet ./...                                # clean
gofmt -l <all changed/new .go files>        # clean
go test ./...                               # ok, full repo (incl. all new tests below)
go run ./internal/cmd/doc-gen -check        # references/examples.md matches checked-in file
cd website && npm run check                 # prebuild (gomarkdoc regen + llms.txt copy) →
                                             # astro build (17 pages) → check-links: all pass
just lint                                   # clean end to end (see below)
```

New tests, all passing:

- `internal/syntax` — `TestFluentChainFieldSelectorMismatchFailsToLoad`
- `internal/builder` — `TestFluentAccessorRejectsFreeFunctionProvider`,
  `TestFluentFunctionRejectsUnrelatedMethodExpressionProvider`
- `gen-jsonschema` — `TestGenCommandRejectsInvalidFluentFieldAssociationWithSourcePosition`
- `ephemeral/issue-73/manager-validation/scaffold-demo` —
  `TestSchemaRuntime` (recovered, standalone module)
- `ephemeral/issue-73/manager-validation/pointer-provider-fixture` —
  `TestPointerRootRenderedSchemaExecutesProviders` (recovered, standalone
  module)

`just lint`: `go mod tidy` clean, `modernize -fix ./...` rewrote three
unrelated files with Go-1.27 idiom changes
(`examples_regenerate_test.go`, `internal/builder/typegrammar.go`,
`json_schema_helpers_test.go`) — same three files the prior implementer and
reviewer both already flagged; reverted with `git checkout --` (confirmed
`git status --porcelain` back to the pre-lint state for those three).

## Lint blocker — resolved by the manager, not a product fix

`staticcheck ./...` initially failed on two unused-func findings in the
independent reviewer's pre-existing scratch fixtures
(`ephemeral/issue-73/manager-review/scratch/repro_*_accessor_free_func`),
since `staticcheck ./...` scans the whole module tree including `ephemeral/`.
Not product code and not touched by this pass. The manager relocated those
fixtures to `ephemeral/issue-73/manager-review/testdata/scratch/` (Go
tooling conventionally skips `testdata/`), recorded in the worklog. Re-ran
`just lint` after that move: `go mod tidy`, `go vet`, `staticcheck`,
`govulncheck`, and `golangci-lint` are all clean now (0 issues). `goimports
-w` (the last step) again reformatted the same historical/unrelated files
three earlier passes have already flagged
(`ephemeral/codec-integration-consumer/*`,
`ephemeral/typescript-generation/proof/*/consumer/*`,
`examples_regenerate_test.go`, `internal/builder/typegrammar.go`,
`json_schema_helpers_test.go`, `tests/typescript/fixture/*`); reverted all
of them with `git checkout --` and confirmed `git status --porcelain`
matches the pre-lint state exactly. `go.mod`/`go.sum` untouched.
`just lint` now passes clean end to end.

## Files changed this pass (beyond what was already staged/untracked)

- `internal/syntax/fluent_expr.go` — the finding-1 fix + corrected doc
  comment.
- `internal/syntax/scanner_test.go`,
  `internal/syntax/testfixtures/fluent_field_mismatch/fixture.go` (new) —
  scanner-level negative test.
- `internal/builder/fluent_declaration_test.go` — two command-level
  negative tests.
- `gen-jsonschema/main_test.go` — CLI subprocess negative test.
- `.github/workflows/website-pages.yml` — Go toolchain selection fix.
- `website/src/content/docs/api/index.md` — regenerated (not hand-edited).
- `llms.txt`, `README.md`,
  `website/src/content/docs/features/enums.md`,
  `skills/go-gen-jsonschema/references/registration-api.md` — legacy syntax
  conversion (llms.txt) and enum-migration-accuracy corrections.
- `CLAUDE.md` → symlink to `AGENTS.md`; `AGENTS.md` — merged, corrected
  canonical guidance file (staged rename from `Agents.md`).
- `ephemeral/issue-73/manager-validation/ts-fluent-fixture/fluent/schema.go`
  and regenerated `jsonschema/*.json`, `ts/*.ts` under both `fluent/` and
  `legacy/` — global-enum-registration equivalence fix.
- `ephemeral/issue-73/manager-validation/scaffold-demo/{main_check.go,demo_test.go}`,
  `.../pointer-provider-fixture/runtime_test.go` (new) — recovered smoke
  tests.
- `examples/{basictypes,structs,indirecttypes,enums,enums_stringmode,uniontypes}/schema.go`,
  `examples/uniontypes/jsonschema_gen.go` (regenerated) — remaining
  primary-example syntax migration.
- `ephemeral/issue-73/finish/result.md` (this file).
