# Issue #73 Independent Review — Fluent `Declare(T.Schema)` API

Reviewed against `ephemeral/issue-73/manager-review/authority.json` (issue body +
retained-case comment). Scope: tracked diff against HEAD (`5a074a1`) plus all
untracked new source/tests for the fluent declaration feature. Prior worklogs
and handoff notes were not consulted as authority; findings below are based on
reading the diff, the surrounding source, and running the actual code.

## Blockers

### 1. `.Accessor(field, provider)` silently accepts a free function and generates a nil-panic at runtime

`internal/syntax/fluent_expr.go:176-204` (`providerRef`) cannot distinguish "a
method expression on the declared root" from "a free function that happens to
have the same `func(T) json.Marshaler` shape" — both are valid arguments to
`Declaration[T].Accessor(field any, provider func(T) json.Marshaler)`
(`declare.go:25`), and Go's type system cannot reject the free-function case
(the whole point of the fluent API per the issue body is "let Go reject
mismatched receiver/provider ... where the language can express the
relationship" — this is a case where it can't).

When a free function is passed to `.Accessor(...)`, `providerRef` returns
`isMethod=false, matched=true` (fluent_expr.go:200-201). This produces a
`FieldProvider{Kind: "WithStructAccessorMethod", ProviderIsMethod: false}`
(`gen_schema.go:91-96`). The template that emits `RenderedSchema()`
(`internal/builder/schemas.go.tmpl:525-536`) only has branches for
`(ProviderIsMethod=true, Kind=WithStructAccessorMethod|WithStructFunctionMethod)`
and `(ProviderIsMethod=false, Kind=WithFunction)`. The
`(ProviderIsMethod=false, Kind=WithStructAccessorMethod)` combination matches
no branch, so `__prov` is declared but never assigned, and the generated code
calls `__prov.MarshalJSON()` on a nil interface.

**Reproduced live** (scratch fixture at
`ephemeral/issue-73/manager-review/scratch/repro_accessor_free_func/`):
`go generate` on
```go
func aSchemaFreeFunc(Example) json.Marshaler { return json.RawMessage(`{"type":"string"}`) }
var _ = jsonschema.Declare(Example.Schema).Accessor(Example{}.A, aSchemaFreeFunc).RenderProviders()
```
compiles and generates without any diagnostic. Calling
`Example{}.RenderedSchema()` at runtime panics:
```
panic: runtime error: invalid memory address or nil pointer dereference
.../jsonschema_gen.go:43 (in the generated RenderedSchema body)
```
The symmetric case (a method expression passed to `.Function()`, or generally
any shape/chain-link mismatch that Go's type checker permits) produces the
same class of silent gap for the same reason.

**Scope note (lowers this from "new regression" to "confirmed, but
pre-existing"):** `providerRef` in `fluent_expr.go` is a byte-for-byte lift of
logic that already existed inline in `internal/syntax/scan_expr.go` before
this diff (confirmed via `git diff HEAD -- internal/syntax/scan_expr.go`,
which shows a pure extract-to-shared-helper refactor). I reproduced the
identical panic through the **legacy** API
(`ephemeral/issue-73/manager-review/scratch/repro_legacy_accessor_free_func/`,
`jsonschema.WithStructAccessorMethod(Example{}.A, aSchemaFreeFunc)`) — same
generated code, same nil dereference. So this is not something #73
introduced; it inherits a shared, pre-existing defect, consistent with the
issue's explicit charter to "resolve supported chains ... into the existing
registration model" rather than build new validation.

**Why it's still a blocker for this issue specifically:** the fluent
`fluent_expr.go:169-175` doc comment for `providerRef` asserts "for a fluent
chain this is unreachable in practice because the provider's receiver type is
already pinned to T by Go's type system" — my repro shows this claim is
false: the free-function/method-expression ambiguity is fully reachable
through the fluent chain and Go's type system does not resolve it. The issue
body's acceptance criteria explicitly require "invalid chains ... propagation
of source-positioned errors" for the new surface, and the whole marketing
premise of `Declare(...)` is that the compiler catches this class of mistake.
Shipping the fluent API with a code comment that misstates its own safety
property, and no diagnostic for the one case Go genuinely cannot catch, means
the "let Go reject it" pitch doesn't hold for `.Accessor`/`.Method`/
`.Function` provider-shape mixups. At minimum, the scanner should reject a
free-function argument to `.Accessor`/`.Method` (and a method-expression
argument to `.Function`) with a source-positioned error instead of silently
producing a template-unmatched `FieldProvider`. Whether to also backport the
fix to the legacy option parser is a separate call, but leaving the new
fluent surface exposed to the same silent-panic footgun, with an incorrect
"unreachable" comment guarding it, does not meet this issue's own bar.

### 2. Zero scanner/command tests exercise a fluent-chain error path

The acceptance criteria require: "Scanner and command tests cover every
supported fluent method, import aliases, invalid chains, and propagation of
source-positioned errors." Every fluent-related test in the diff is
happy-path:

- `internal/syntax/scanner_test.go` (`TestFluentDeclareParser`): 5 subtests,
  all `require.NoError`.
- `internal/builder/fluent_declaration_test.go`: parity tests (all
  `require.NoError`) plus one genuine negative case,
  `TestFluentInterfaceRegistrationDiagnosticsParity`, which is a
  **builder-level** semantic error (`Impl` type doesn't satisfy the
  interface) reusing the shared error path — not a scanner-level AST/chain
  error.
- `gen-jsonschema/main_test.go`: no invalid-chain or error-path test at all
  (only scaffolder-emits-fluent-form and format-flag tests).

Concretely untested, reachable-through-valid-Go error paths in
`internal/syntax/fluent_expr.go`:
- `ParseFluentDeclaration`'s free-function branch, "could not resolve free
  function %q" (fluent_expr.go:97-102) — reachable whenever `Declare(x)` names
  a package-level `var` or anything else that isn't a top-level func decl in
  the same package.
- The `default:` "unsupported schema func expression" case
  (fluent_expr.go:110-112) — reachable via e.g. `Declare(func(t T) json.RawMessage {...})`
  (an inline func literal) or any other expression shape satisfying the
  `func(T) json.RawMessage` parameter type.
- `parseFluentChainOptions`'s arg-count and field-selector-mismatch errors
  (fluent_expr.go:288-330).

None of these have a test proving the scanner actually returns a
source-positioned error for them, despite the acceptance criterion naming
this explicitly. This is a real, checkable gap, not a stylistic nit.

## Verified as solid (no material finding)

- **Compile-time rejection**: `internal/compiletest/negative_test.go` +
  4 fixtures under `testdata/` genuinely fail `go build` today (ran
  `TestNegativeFixturesFailToCompile` directly — all 4 subtests pass),
  covering mismatched receiver, mismatched accessor receiver, and joint
  field/provider type mismatches for `.Method`/`.Function`. This is real
  compiler-enforced rejection, not just an assertion.
- **Retained pointer-enum acceptance case** (authority.json comment): fluent
  `.Enum()` on a pointer field hard-fails identically to the legacy path,
  fixed for both by the shared `resolveEnumFieldPlan` change in `5a074a1`
  (already on HEAD, not part of this diff, but I re-verified the fluent path
  inherits it — reproduced live in
  `ephemeral/issue-73/manager-review/scratch/repro_ptr_enum_fluent/`: `go generate`
  exits 1 with `field Thing.PL: WithEnum/WithStringerEnum supports only a
  direct named enum, Optional[E], or Nullable[E]`).
- **Fluent/legacy artifact equivalence**: regenerated all six touched example
  packages (`interfaces_options`, `optionality`, `sealed_interface_slices`,
  `stringer_enums`, `ref_types`, `providers_rendering`) with `go generate`;
  produced zero additional diffs beyond what was already staged — the
  fluent-form `schema.go` files reproduce byte-identical (or, for `ref_types`/
  `stringer_enums`, description-text-only-different) generated JSON/Go vs.
  what's committed. `internal/builder/fluent_declaration_test.go` also proves
  this at the builder level for providers (value + pointer root), enums, ref,
  and interfaces via direct JSON-string comparison.
- **Pointer-root support**: value and pointer receivers both work for
  `Declare`, `.Accessor`, `.Method` (compile fixtures + parity tests +
  `TestFluentPointerRootProviderParityWithLegacy`, whose comment documents
  that this was caught and fixed mid-development — the `*dst.StarExpr` case
  inside `(*Example).ASchema` was originally dropped silently and is now
  handled in `providerRef`).
- **Import aliases**: `TestFluentDeclareParser/import alias root` proves a
  fluent chain resolves an aliased-import root type identically to the
  existing `IdentifyFunc`/package-resolution machinery.
- **Scaffolder**: `gen-jsonschema new` emits `Declare(...)` (confirmed by
  `TestNewConfigUsesFluentDeclareForm` and by building the actual CLI and
  running `new --methods=Example=Schema` + `gen` end-to-end in an isolated
  module — produced a working `Declare(Example.Schema)` stub and a correct
  generated schema through real generation, not just template assertions).
- **Documentation**: README, website feature pages, and the shipped skill
  (`skills/go-gen-jsonschema/`) teach `Declare(...)` as the primary syntax;
  every remaining legacy mention I found is explicitly under a "Migration:"
  heading naming the fluent equivalent, matching the issue's requirement to
  "remove legacy syntax from primary tutorials ... mention it only in concise
  migration/compatibility guidance." `skill-examples.json` and the derived
  skill docs regenerate byte-identical via `go run ./internal/cmd/doc-gen`
  (clean regeneration, no drift).
- **Deprecation**: every legacy entry point (`NewJSONSchemaMethod`,
  `NewJSONSchemaFunc`, `WithEnum`, `WithStringerEnum`, `WithFunction`,
  `WithStructAccessorMethod`, `WithStructFunctionMethod`, `WithInterface`,
  `WithInterfaceImpls`, `WithDiscriminator`, `NewInterfaceImpl`,
  `NewEnumType`, `WithRenderProviders`, `AsRef`) carries a `Deprecated:` godoc
  comment naming its fluent replacement (`union_type.go` diff), and all
  remain callable/compiled (source-compatible), satisfying "keep legacy
  declarations source-compatible through 1.x."
- **doc-gen's own AST walk** (`internal/cmd/doc-gen/main.go`,
  `baseRegistrationCall`): correctly walks down a chained
  `Declare(...).A(...).B(...)` call to the base call carrying the schema
  method argument, verified by tracing the recursion by hand and by the
  passing `TestExtractSelectsFluentChainRegistration` test.
- **No scope creep**: `declare.go`'s chain surface (`Accessor`, `Method`,
  `Function`, `Enum`, `StringerEnum`, `Ref`, `RenderProviders`, `Interface`)
  matches exactly the outcome example in the issue body; no extra methods,
  no speculative generality, no second codec/schema implementation
  introduced — `internal/builder/gen_schema.go` normalizes everything into
  the pre-existing `SchemaMethodOptionInfo`/`FieldProvider` shapes.
- **Build/vet/lint/vulncheck**: `go build ./...`, `go vet ./...`, `gofmt -l`
  on all changed files, `staticcheck ./...` (excluding my own scratch dir),
  and `govulncheck ./...` are all clean. `go test ./...` passes cleanly on a
  clean test cache (one transient build-cache-related failure on
  `examples/ref_types` was observed once during a busy multi-`go generate`
  loop and did not reproduce after `go clean -testcache`; ruled out as a
  local artifact of my own test loop, not a product defect).

## Process note

Running `just lint` (which shells out to `modernize -fix ./...` and `go mod
tidy`) auto-rewrote three unrelated product files
(`examples_regenerate_test.go`, `internal/builder/typegrammar.go`,
`json_schema_helpers_test.go`) with Go-1.27 idiom modernizations unrelated to
this issue. I reverted these immediately (`git checkout --`) since I'm not
authorized to modify product files; the working tree is back to the original
40-file baseline. Flagging in case `just lint` is run un-reviewed elsewhere —
it will make unrelated changes as a side effect on this toolchain version.

## Verdict

**One blocker, one acceptance-criteria gap; not clean to ship as-is, but the
fix surface is narrow.**

1. (Blocker, but scoped) `.Accessor`/`.Method`/`.Function` provider-shape
   confusion silently produces a nil-panic at runtime instead of a
   source-positioned scanner error. Confirmed pre-existing in the shared
   `providerRef` helper (not new to #73), but the fluent API's own design
   comment incorrectly claims this is unreachable, and the issue's explicit
   "let Go reject it" premise and "no silent option loss" requirement both
   argue for closing this specific gap as part of shipping `Declare(...)`
   rather than deferring it silently.
2. (Acceptance gap) No test anywhere exercises a fluent scanner-level error
   path (malformed `Declare(...)` argument, unresolvable free function,
   arg-count/selector mismatches in chain options), despite this being named
   explicitly in the issue's acceptance criteria.

Everything else checked — compile-time inference/rejection, value/pointer
root support, fluent/legacy artifact equivalence (verified by actual
regeneration, not just by reading tests), the retained pointer-enum
acceptance case, scaffolder-to-real-generation, documentation/skill
migration framing, deprecation coverage, and clean build/vet/lint/vulncheck/
test — is solid and matches the issue's scope without over-engineering.

## Checks actually performed

- Read `authority.json` (issue body + retained-case comment) as the sole
  acceptance authority.
- Read `declare.go`, `declare_test.go`, `internal/syntax/fluent_expr.go` in
  full; read the relevant diff hunks of `scan_expr.go`, `scan_result.go`,
  `gen_schema.go`, `schemas.go.tmpl`, `union_type.go`,
  `internal/cmd/doc-gen/main.go`.
- Traced `FieldProvider`/`ProviderIsMethod` end-to-end from scanner through
  `gen_schema.go` into the code-generation template to find the nil-panic
  gap, then reproduced it live for both the fluent and legacy APIs in
  isolated scratch fixtures (`go generate` + a throwaway test calling
  `RenderedSchema()`, panic captured).
- Reproduced the authority.json retained pointer-enum case through the
  fluent path directly.
- Regenerated every touched example package with `go generate` and diffed
  against committed artifacts (no drift).
- Ran `go build ./...`, `go vet ./...`, `gofmt -l` on all changed Go files,
  `staticcheck`/`govulncheck` over the product package list, and
  `go test ./...` (twice, once with a clean test cache).
- Ran `internal/compiletest`'s negative fixtures directly to confirm real
  `go build` failures, not just assertions.
- Built the actual `gen-jsonschema` CLI binary and ran `new` +`gen` in an
  isolated module end-to-end to confirm the scaffolder's fluent output
  actually generates.
- Regenerated `internal/cmd/doc-gen`'s output and diffed (no drift).
- Grepped all touched website/skill docs for stray non-migration-labeled
  legacy syntax.
- Confirmed no product files were left modified after tooling side effects
  (`just lint`'s `modernize -fix`) by reverting and re-diffing against HEAD.
