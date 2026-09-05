# Issue 73: typed fluent schema declarations

baseline: `main` at `5a074a1`; `go test ./...` passed before starting.
decision: Work begins only after issue #80 merged and its worktree was removed.
decision: Root manages product scope, review, integration, and merge. Claude owns codebase research and implementation heavy lifting through Tractor.
scope: Deliver the useful typed `Declare` authoring API and make it the documented default while preserving legacy compatibility. Reject speculative validation, support for unrequested Go shapes, defensive handling of deliberately malformed chains, and generalized parser machinery that is not required by the supported fluent surface.

## Functional core implementation (this pass)

manager contract diverged from ephemeral/issue-73/research.md in two ways I followed as authoritative:
- `Declare[T]`'s root argument supports BOTH a method expression and a free function (not method-expression-only as research recommended for 1.0). Both existing legacy roots (`NewJSONSchemaMethod`, `NewJSONSchemaFunc`) are being deprecated, so both forms had to have a fluent replacement.
- `Accessor` carries a method type parameter `Accessor[F]` per the contract's literal signature list, even though research argued F is unused/unneeded there. Implemented as `Accessor(field any, provider func(T) json.Marshaler)` — no runtime relationship between F and anything else, harmless, matches contract text exactly (the contract text uses `Accessor[F](F, func(T) json.Marshaler)`; I kept `field` as plain `any` since F never appears elsewhere in the signature and Go doesn't need a declared type param when it isn't referenced by the receiver-bound method - the *type-parameter-free* form is behaviorally identical to `Accessor[F]` and is what actually shipped. This is a values-first deviation from the literal text, not from the intent: goal was "the field argument stays untyped like a bookmark," which both spellings achieve equally, so I chose the simpler one).

### Root-cause finding that shaped the scanner design

The decorator (`dst`) collapses ANY package-qualified call target — not just method-
expression receivers as the existing `unwrapSchemaMethodReceiver` comment already
documented, but also a plain `pkg.Func(...)` call's `Fun` itself — into a `*dst.Ident`
with `.Path` set, rather than a `*dst.SelectorExpr`. This means the base
`jsonschema.Declare(...)` call at the bottom of a fluent chain has `Fun` as a bare
`*dst.Ident{Path: SchemaPackagePath, Name: "Declare"}`, NOT a `SelectorExpr` — so the
chain-walk in `parseFluentChain` must try `CallExpr.IdentifyFunc()` (which already
handles this Ident/Path shape) at the START of every loop iteration, not only as a
special-cased base case after `Fun.(*dst.SelectorExpr)` fails. Got this wrong on the
first pass (walked assuming the base call's `Fun` would be a `SelectorExpr`), which
silently produced zero fluent hits for any chain with a `.` after `Declare(...)`;
caught immediately by the scanner unit tests (empty-slice assertion failures), fixed
by moving the `IdentifyFunc()` check to the top of the loop. Left the fix as a code
comment in `internal/syntax/fluent_expr.go` since this collapse behavior is
non-obvious and easy to get wrong again.

### Design (internal/syntax)

- Added `MarkerFuncDeclare = "Declare"` to `markerFunctions`. `ParseValueExprForMarkerFunctionCall`
  now falls back to `parseFluentChain` whenever a var-decl's call expression doesn't identify
  directly (i.e. its `Fun` is a `SelectorExpr` chained off another `CallExpr`) — a bare
  `Declare(fn)` with no chained options still hits the ordinary direct-identify path unchanged.
- `parseFluentChain` walks a `CallExpr` down through `.Method(...)` links until it bottoms out
  at a `Declare(...)` marker call, returning the base call plus the ordered chain links
  (`fluentChainLink{methodName, call}`), or `ok=false` for any non-Declare-rooted chain (so
  unrelated chained calls in the same var block are still silently ignored, matching existing
  behavior for anything that isn't a jsonschema marker).
- `MarkerFunctionCall.ParseFluentDeclaration(localFuncs []FuncDecl)` resolves the root
  (method-expression via the existing `unwrapSchemaMethodReceiver`, or free function by looking
  up the identifier in the package's own `FuncDecl`s and reading its sole parameter's type off
  the AST — deliberately LOCAL-ONLY; an imported free-function root is out of scope, matching
  research's "don't add unrequested Go-shape support," and legacy `NewJSONSchemaFunc[T]` remains
  available for that case) and produces the *same* `SchemaMethod`/`SchemaFunction` shape the
  legacy markers produce, dispatched in `scan_result.go`'s `loadPackageInternal` exactly like the
  existing `MarkerFuncNewJSONSchemaMethod`/`MarkerFuncNewJSONSchemaFunc` arms.
- Extracted three helpers out of the legacy `parseSchemaMethodOptions` (`fieldNameForReceiver`,
  `providerRef`, `parseInterfaceNestedOptions`) so both the legacy option-list parser and the new
  `parseFluentChainOptions` share identical field-selector/provider/interface-option parsing —
  no parallel logic, per the contract's "reuse existing... paths" instruction. Legacy behavior is
  byte-for-byte unchanged (all existing scanner/builder tests still pass unmodified).
- Deliberate asymmetry: a fluent chain link with a field-selector that names the wrong receiver
  type (e.g. `Declare(A.Schema).Enum(B{}.X)`) is a HARD scanner error in the fluent path, whereas
  the legacy option-list parser silently skips the analogous case. Justification: every legacy
  option lives in a variadic list that generically tolerates unrelated entries; every fluent
  chain link, by construction, belongs to exactly one `Declare[T]` root, so a receiver mismatch
  can only be a typo, and silently dropping a `.Enum()`/`.Interface()` call would be exactly the
  "silently mis-normalized artifact" the task explicitly wants avoided. A provider-reference
  mismatch (`.Accessor`/`.Method`/`.Function`'s second arg) stays a silent skip in both paths,
  since Go's generics already make that case impossible to compile in the fluent form.
- An unrecognized chain-link method name mid-chain (impossible today since the public API only
  exposes the 8 documented methods, but future-proofing against a stale scanner after a public
  API addition) is also a hard error rather than silent skip, per research §6 point 5.

### Public API (declare.go, root package)

`Declaration[T any]` + `Declare[T any](fn func(T) json.RawMessage) *Declaration[T]` plus the 8
contract-specified chain methods, all pure markers (never called, `//go:build`-free since they're
useful in normal builds too — unlike `union_type.go`'s markers, `Declare`'s call sites live in
`//go:build jsonschema` files same as always, but the type/function *declarations* themselves have
no reason to be gated). Verified against real scratch compiles (not just this repo's own tests):
value root, pointer root, free-function root, and the mismatched-receiver/mismatched-provider
rejections research promised all behave exactly as documented in `research.md` §1.

### Files touched

- `declare.go` (new) — public API.
- `declare_test.go` (new) — positive compile coverage: value root, pointer root, free-function
  root, every chain method, Method/Function joint field↔provider type inference.
- `internal/compiletest/negative_test.go` + `internal/compiletest/testdata/{mismatched_receiver,mismatched_accessor,mismatched_method_field,mismatched_function_field}/pkg.go`
  (new) — negative compile fixtures. Fixtures live under `testdata/` so `go build ./...`/`go test
  ./...` skip them automatically (Go's own convention); the test shells `go build
  ./testdata/<name>/` per fixture and asserts non-zero exit, no compiler-string matching.
- `internal/syntax/fluent_expr.go` (new) — chain walk, root resolution, option parsing, all the
  design above.
- `internal/syntax/scan_expr.go` — `MarkerFuncDeclare` const/list entry, fluent-chain fallback in
  `ParseValueExprForMarkerFunctionCall`, `fluentLinks` field on `MarkerFunctionCall`, extraction
  of the three shared helpers out of `parseSchemaMethodOptions` (behavior-preserving refactor).
- `internal/syntax/scan_result.go` — one new `case MarkerFuncDeclare:` arm in
  `loadPackageInternal`, mirroring the legacy arms.
- `internal/syntax/testfixtures/typescanner/fluent_calls.go` (new) + one added `Schema()` method
  on `scannersubpkg.TypeForSchemaMethod` (`remote_func_defs.go`) — fixture exercising every
  supported chain method in one chain, a free-function root, a bare `Declare` with no chain, and
  an import-alias-qualified method-expression root.
- `internal/syntax/scanner_test.go` — `TestFluentDeclareParser`, four subtests over that fixture,
  asserting parsed `SchemaMethod`/`Options` equal the expected normalized shape (existing
  `TestFuncCallParser` untouched, still passes as-is).
- `internal/builder/fluent_declaration_test.go` (new) — in-process builder parity: writes a
  legacy-syntax and a fluent-syntax variant of the same fixture package to temp dirs via
  `syntax.Load` + `builder.New` (no `go generate`/temp modules needed, reusing the pattern from
  `interface_options_test.go`/`asref_collision_test.go`), and asserts the rendered JSON schema
  bytes are byte-identical for: provider (Accessor+Method+Function+RenderProviders), Enum +
  StringerEnum, Ref (`$ref` collapsing), and sealed-interface (`Interface` + inline
  `Discriminator`/`Impl`). Also ports `TestInlineInterfaceRegistrationDiagnostics`'s
  "implementation does not satisfy interface" case through the fluent spelling, proving identical
  error-path reuse (not a fluent-specific diagnostic).
- `gen-jsonschema/tmpl/config.go.tmpl` — scaffolder now emits `jsonschema.Declare(Type.Method)`
  instead of `jsonschema.NewJSONSchemaMethod(Type.Method)`.
- `gen-jsonschema/main_test.go` — `TestNewConfigUsesFluentDeclareForm` asserts the rendered
  template contains the `Declare(...)` form and not `NewJSONSchemaMethod`.

### Explicitly not done (per contract, next pass)

Docs/examples/skill conversion (README, SKILL.md, website content, `examples/*/schema.go`) —
contract says that's the next bounded pass after the core lands, not this one.

### Proof

```
gofmt -l .                         # clean (excluding testfixtures/test_run golden dirs and the
                                    # non-Go .tmpl scaffolder template, which gofmt can't parse)
go build ./...                     # clean
go vet ./...                       # clean
go test ./internal/syntax/...      # ok, incl. new TestFluentDeclareParser
go test ./internal/builder/...     # ok, incl. new TestFluent*ParityWithLegacy (4) +
                                    # TestFluentInterfaceRegistrationDiagnosticsParity
go test ./gen-jsonschema/...       # ok, incl. new TestNewConfigUsesFluentDeclareForm
go test ./internal/compiletest/... # ok, all four negative fixtures fail to compile as expected
go test ./...                      # ok, full repo, no regressions in any pre-existing package
```

### Known environment limitation (not a regression)

`just lint`'s `modernize`/`staticcheck`/`golangci-lint` binaries in this sandbox were built
against an older `golang.org/x/tools` that predates Go 1.27's generic-methods parser support.
Since `Method[F]`/`Function[F]` are the first code in this repo to actually use a generic method
(a real feature gated on the `go 1.27` this repo's `go.mod` already declared, per research.md's
module-graph analysis), `modernize -fix ./...` now fails on files it was already failing to
parse correctly before — including files this change never touched (e.g.
`internal/syntax/node_wrappers_test.go`, `examples/structs/schema_test.go`) and even the Go 1.27
standard library itself (`math/rand/v2/rand.go`) when analyzed by the same toolchain-mismatched
binary. `go build`, `go vet`, and `go test ./...` (the real go1.27.1 toolchain) all pass clean.
This is an environment tooling-version gap, not something fixable in this codebase; flagging for
whoever owns the lint toolchain upgrade rather than silently working around it.

## Correction: Method/Function weren't jointly inferring F (this pass)

The original signatures were `Method[F any](field any, provider func(T, F) json.Marshaler)` and
`Function[F any](field any, provider func(F) json.Marshaler)`. Because `field` was `any`, F was
inferred solely from `provider`, so a same-root field/provider type mismatch (e.g. a `string`
field paired with an `int`-taking provider) still compiled — the "F never appears elsewhere" false
equivalence claimed above for `Accessor` (line 12) does not hold for `Method`/`Function`, where F
*is* referenced twice and both references must agree. Fixed by binding `field F` directly on both
methods, so F is now inferred jointly from field and provider, matching the contract. Added
`mismatched_method_field`/`mismatched_function_field` negative fixtures alongside the existing
`mismatched_receiver`/`mismatched_accessor` ones to prove it. `declare_test.go`'s positive coverage
and `internal/builder/fluent_declaration_test.go`'s parity tests required no changes — both already
passed field selectors matching their providers' types.

## Correction: pointer-root providers silently dropped (independent review finding)

Independent review of this pass found that `providerRef` (`internal/syntax/fluent_expr.go`) only
matched the parenthesized-value-receiver method-expression shape `(Example).ASchema` inside its
`*dst.ParenExpr` case — it type-asserted `x.X` straight to `*dst.Ident` and returned
`matched=false` for anything else, silently dropping the provider option (via
`parseFluentChainOptions`'s existing "unmatched provider" skip) rather than erroring. But
`Declare((*Thing).Schema)` is the only working spelling for a pointer-root receiver: `Method`/
`Function`/`Accessor`'s field-parameter binding (see the correction above) requires `provider`'s
first argument to be exactly `T`, and when `T` is `*Thing` the corresponding method expression is
necessarily `(*Thing).FieldSchema`, whose receiver position is a `*dst.StarExpr` wrapping the
`Ident`, not a bare `Ident`. So the one supported pointer-root shape hit the `matched=false` path
every time, and `.Accessor(...)`/`.Method(...)` chained after a pointer-root `Declare(...)` were
accepted by the scanner but produced schemas with no provider attached at all — a correct-looking
declaration that silently generated the wrong schema.

Fixed by adding a `*dst.StarExpr` case inside the existing `*dst.ParenExpr` switch in
`providerRef`, matching `(*ReceiverType).Method` the same way the `Ident` case already matches
`(ReceiverType).Method` — reusing the same `receiver.TypeName` comparison, no new expression
handling. Added `internal/syntax/testfixtures/typescanner/fluent_calls.go`'s pointer-receiver
`FluentStruct` methods (`PtrSchema`/`PtrASchema`/`PtrBSchema`) plus a `pointer-root chain` subtest
in `TestFluentDeclareParser` (`internal/syntax/scanner_test.go`) proving the scanner now retains
both `WithStructAccessorMethod`/`WithStructFunctionMethod` options, and
`TestFluentPointerRootProviderParityWithLegacy` in
`internal/builder/fluent_declaration_test.go` proving the full builder path renders the same
template-hole schema for a pointer-root fluent chain as the equivalent pointer-receiver legacy
`NewJSONSchemaMethod` registration (template holes, rather than inline descriptions, are what
prove the providers survived — the previous silent-drop bug would have rendered plain
string/integer property schemas instead).

Proof: `gofmt -l .` clean; `go test ./internal/syntax/... -run TestFluentDeclareParser` and
`go test ./internal/builder/... -run TestFluentPointerRootProviderParityWithLegacy` both green;
`go test ./...` full repo green, no regressions.

## Docs/examples conversion pass (this pass)

Prior pass (above) explicitly deferred docs/examples/skill conversion to "next pass, once the
core lands." This pass does that: makes `Declare(...)` the documented default everywhere a user
or coding agent is directed, while keeping every legacy entry point source-compatible with a
`Deprecated:` godoc comment.

Note: on entry, `git status` already showed README.md, `examples/optionality/schema.go`, and
`examples/stringer_enums/schema.go` converted, plus `union_type.go`'s `Deprecated:` comments in
place — a prior, interrupted attempt at this same pass (`ephemeral/issue-73/convert-docs-run/`,
Tractor pipeline `issue-73-convert-docs-and-examples`, timeline shows `"turn was interrupted"` at
2026-09-05T12:32Z). I resumed from that partial state rather than redoing it; verified the
existing README/union_type.go changes were correct before building on them.

### `Deprecated:` godoc (union_type.go)

Added a one-line `Deprecated: use Declare(T.Schema)....` comment to every legacy marker:
`WithRenderProviders`, `AsRef`, `WithFunction`, `WithStructFunctionMethod`,
`WithStructAccessorMethod`, `WithInterface`, `WithInterfaceImpls`, `WithDiscriminator`, `WithEnum`,
`WithStringerEnum`, `NewJSONSchemaMethod`, `NewJSONSchemaFunc`, `NewInterfaceImpl`, `NewEnumType`.
All stay callable and behaviorally unchanged (scanner still parses them the same way); this is
comment-only. `NewJSONSchemaBuilder` (no-arg schema-accessor stub) has no fluent replacement and
was deliberately left undeprecated, per the contract's explicit "unaffected" carve-out.

### README.md / examples/README.md / SKILL.md / references/*.md / website docs

Converted every registration example these files show to `Declare(...)` chains, using the actual
chain names/signatures from `declare.go` (`Accessor`, `Method`, `Function`, `Enum`, `StringerEnum`,
`Ref`, `RenderProviders`, `Interface`). Each doc keeps exactly one "Migration: `<old form>` is now
`<new form>`" note per feature area (enums, interfaces, providers, refs) rather than re-explaining
the legacy surface at every mention — per the contract's "do not repeatedly document legacy
syntax" instruction. Untouched by design: `examples/self_contained`, `template_rendering`,
`test_options` and other non-advertised example dirs (their `WithEnum` mentions are pre-existing,
not doc-linked, and out of scope — "avoid... unrelated cleanup").

### Example packages converted (representative, per category)

- **plain root**: README's `Person` example (`Declare(Person.Schema)`).
- **enums/string enums**: `examples/stringer_enums` (`.Enum`/`.StringerEnum`); `types.go`'s doc
  comment ("WithStringerEnum emits...") updated to "StringerEnum emits..." since Go doc comments
  become JSON Schema `description` text verbatim — the generated
  `ApplicationConfig.json`'s `description` field changed to match (intentional; not a behavior
  regression, just updated prose flowing through).
- **sealed interfaces**: `examples/interfaces_options` and `examples/sealed_interface_slices`
  (`.Interface(field, Discriminator(...), Impl(...), ...)`). For `sealed_interface_slices`, the
  legacy split form (`WithInterface`+`WithInterfaceImpls` with no explicit `Impl` values) derives
  wire discriminators from Go type names, so the fluent replacement supplies
  `Impl("Created", Created{})`/`Impl("Deleted", (*Deleted)(nil))` explicitly — verified
  byte-identical `Batch.json` after regeneration, confirming the derived names match.
- **refs**: `examples/ref_types`. `Shared.Schema` converted to `Declare(Shared.Schema).Ref()`
  (byte-identical `Shared.json`/`Container.json` after regen, modulo the intentional doc-comment
  wording change on `Shared.json`'s `description` — see below). `NullableConfig`'s `Mode` enum
  registration was **left on the legacy `NewEnumType[Mode]()`** after finding a real behavior
  difference: converting it to `Declare(NullableConfig.Schema).Enum(NullableConfig{}.Mode)` (field-
  level `Enum` targeting a `Nullable[Mode]` field) silently dropped `Mode`'s type-doc-comment
  `description` from the generated schema (package-level `NewEnumType` and field-level `Enum` are
  not equivalent in that one respect). Since the task requires confirming fluent/legacy output
  stays equivalent and this case doesn't, I kept the legacy form here rather than "fixing" the
  description gap — that's a separate bug, out of scope for a docs-conversion pass. Also updated
  `ref_types/types.go`'s two "AsRef()" doc comments to "Ref()" (again flows into generated JSON
  `description` fields; confirmed via regen diff that only that text changed).
- **providers**: `examples/providers_rendering` (`.Accessor`/`.Method`/`.Function`/
  `.RenderProviders()`). Moved the trailing `// v1: generate RenderedSchema()...` line comment to
  a standalone comment above the `Declare(...)` call — a trailing comment on the chain's last line
  produced an extra blank line inside doc-gen's extracted code fence (cosmetic quirk of
  `go/format.Node` on a decl whose last line carries an end-of-line comment); no behavior change,
  confirmed via regen that `Example.json` is unchanged.
- Added a `providers` entry to `internal/cmd/doc-gen/skill-examples.json` (new; the manifest had
  no provider-example entry before this pass) so `references/examples.md` gets a
  "Provider-rendered fields" section sourced from real, compiling `examples/providers_rendering`
  code, consistent with how the other four categories are already sourced.

Verification method for every conversion above: ran `go generate ./...` in the changed example
directory and diffed the checked-in `jsonschema/*.json`/`*.json.sum` against the pre-conversion
versions. Only `ref_types` and `stringer_enums` show any diff, and in both cases it's exactly the
one doc-comment wording change described above — everything else (schema shape, `$defs`,
discriminators, provider template holes) is byte-identical, proving the fluent registrations are
behaviorally equivalent to what they replaced.

### `internal/cmd/doc-gen` fluent-chain support

`references/examples.md` carries a `<!-- Code generated by internal/cmd/doc-gen; DO NOT EDIT -->`
header and is covered by `TestCheckedInReferenceIsCurrent`, so hand-editing it was not an option.
The tool's `registrationNames`/`isRegistrationCall` (`internal/cmd/doc-gen/main.go`) only recognized
a *direct* `NewJSONSchemaMethod(...)`/`NewJSONSchemaFunc(...)` call as a `ValueSpec`'s value — a
fluent `Declare(Foo.Schema).Enum(...)` chain's outer call has `Fun` as a `*ast.SelectorExpr` whose
`X` is the previous link's `*ast.CallExpr`, not `NewJSONSchemaMethod`/`NewJSONSchemaFunc` directly,
so it silently matched nothing. Replaced `isRegistrationCall` with `baseRegistrationCall(expr
ast.Expr) *ast.CallExpr`, which recurses down through chain links (`fun.Sel.Name` not one of the
three base-call names → recurse into `fun.X`) until it finds the base call, then reads `Args[0]`
from that base call exactly as before. Legacy non-chained calls still match on the first
iteration (no recursion needed). Added `TestExtractSelectsFluentChainRegistration` proving a
`Declare(Task.Schema).Enum(Task{}.Status)` var decl is now selectable by `Registrations:
["Task.Schema"]`, alongside the pre-existing legacy-call test (unchanged, still passes).

### Known gap: `gomarkdoc` cannot regenerate `website/src/content/docs/api/index.md`

`website/package.json`'s `prebuild` script runs `gomarkdoc -e -o
./src/content/docs/api/index.md ../` (pinned to `v1.1.0` in
`.github/workflows/website-pages.yml`, and that's also the version installed locally) to generate
this file from root-package godoc — same generation contract as `references/examples.md` (a
"Code generated by gomarkdoc. DO NOT EDIT" header), so it should pick up `declare.go`'s new
`Declare`/`Declaration` docs and `union_type.go`'s new `Deprecated:` comments automatically once
regenerated. It doesn't: `gomarkdoc -e -o ... ../` fails with `gomarkdoc: failed to parse package
file declare.go` (no further detail even at `-v -v -v`). Root cause matches the modernize/
staticcheck gap already flagged in the prior pass's proof section: `declare.go`'s `Method[F
any](...)`/`Function[F any](...)` are real go1.27 generic methods (`go build`/`go vet` with the
installed `go1.27.1` toolchain accept them fine — confirmed directly), but `gomarkdoc` v1.1.0
(latest available; checked `go list -m -versions`) bundles an older `go/parser`-based dependency
that predates go1.27's generic-method grammar and can't parse the file at all, so it produces no
output rather than a partial one.

This isn't a sandbox-only inconvenience this time: `.github/workflows/website-pages.yml` installs
the identical `gomarkdoc@v1.1.0` pin, so the real website-pages CI job's `npm run check` step
(`prebuild` → `gomarkdoc` → `astro build`) will fail the same way once this branch's `declare.go`
lands on `main`, blocking the website deploy until whoever owns that pipeline upgrades or replaces
gomarkdoc (no newer gomarkdoc release exists to bump to — this needs an upstream fix or a
different tool). `website/src/content/docs/api/index.md` is left unregenerated/stale in this pass
as a result; I did not hand-edit the "DO NOT EDIT" file to work around it, since that would drift
further at the next real generation. Flagging prominently for whoever owns the website deploy
pipeline, separate from (but same root cause as) the previously-flagged lint-toolchain gap.

### Files touched (this pass)

- `union_type.go` — `Deprecated:` godoc comments only, no behavior change.
- `README.md`, `examples/README.md` — fluent conversions + migration notes (README.md's
  enum/interface conversions were already done entering this pass; verified them, didn't redo).
- `skills/go-gen-jsonschema/SKILL.md`,
  `skills/go-gen-jsonschema/references/registration-api.md` — fluent conversions + migration
  notes; `registration-api.md`'s "Full registration surface" section rewritten around
  `Declare(fn)` + its chain methods instead of the flat legacy option list.
- `skills/go-gen-jsonschema/references/examples.md` — regenerated via `go run
  ./internal/cmd/doc-gen` (not hand-edited).
- `internal/cmd/doc-gen/main.go`, `internal/cmd/doc-gen/main_test.go`,
  `internal/cmd/doc-gen/skill-examples.json` — fluent-chain scanning support + new `providers`
  manifest entry, described above.
- `website/src/content/docs/features/{enums,interfaces,optionality,providers,ref-types}.md`,
  `website/src/content/docs/{getting-started,examples,spec}.mdx`,
  `website/src/content/docs/reference/cli.md` — fluent conversions + migration notes; nav-card
  titles referencing "AsRef" renamed to "Ref".
- `website/src/content/docs/api/index.md` — **not regenerated**, see gap above.
- `examples/interfaces_options/schema.go`, `examples/sealed_interface_slices/schema.go`,
  `examples/ref_types/schema.go`, `examples/ref_types/types.go`,
  `examples/providers_rendering/schema.go`, `examples/stringer_enums/types.go` — fluent
  conversions (schema.go) and doc-comment wording updates (types.go), described above.
- `examples/ref_types/jsonschema/{Shared,Container,NullableConfig}.json(.sum)`,
  `examples/stringer_enums/jsonschema/ApplicationConfig.json(.sum)` — regenerated via `go generate
  ./...` in each example dir; diffs are exactly the doc-comment wording changes described above.

### Proof (this pass)

```
go build ./...                       # clean
gofmt -l .                           # clean (excluding testfixtures/test_run golden dirs)
go run ./internal/cmd/doc-gen -check # references/examples.md matches checked-in file
go test ./...                        # ok, full repo, no regressions (incl. new
                                      # TestExtractSelectsFluentChainRegistration)
just lint                            # clean this run (modernize/staticcheck/golangci-lint all
                                      # passed — unlike the prior pass's note, this run's installed
                                      # toolchain apparently handles it); `find . -exec goimports
                                      # -w` reformatted several unrelated historical files
                                      # (ephemeral/codec-integration-consumer/*,
                                      # ephemeral/typescript-generation/proof/*/consumer/*,
                                      # examples_regenerate_test.go, internal/builder/typegrammar.go,
                                      # json_schema_helpers_test.go, tests/typescript/fixture/*) —
                                      # reverted all of those via `git checkout --`, keeping only
                                      # this pass's issue-73 edits.
```

Regeneration verification for every converted example: `go generate ./...` per changed dir, then
`git diff`/`git status` on `examples/*/jsonschema/` — see "Verification method" above.
