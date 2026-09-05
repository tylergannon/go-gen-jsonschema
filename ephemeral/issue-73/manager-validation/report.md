# Issue #73 consumer-side validation report

Role: independent validation using the real, built `gen-jsonschema` CLI from this worktree — not
`go run`, not the source-review agent's own internal `go test` suite. All work under
`ephemeral/issue-73/manager-validation/`. No product files were left modified as a result of this
pass (see "Working-tree hygiene" at the end). Baseline: worktree HEAD `5a074a1` plus the
uncommitted fluent-conversion working tree already present when this validation started.

CLI built once and reused for every step below:

```
$ go build -o ephemeral/issue-73/manager-validation/bin/gen-jsonschema ./gen-jsonschema
```

## 1. Scaffolding produces the fluent form and a working consumer end to end

Fixture: `ephemeral/issue-73/manager-validation/scaffold-demo/` (fresh module, `replace` to this
worktree).

```
$ gen-jsonschema new --methods Widget=Schema --out schema.go
Package name: scaffold_demo
Output written to: schema.go
```

Emitted `schema.go`:

```go
var (
    _ = jsonschema.Declare(Widget.Schema)
)
```

Confirms the acceptance criterion "`gen-jsonschema new` emits the fluent form" directly from the
built binary, not from reading `config.go.tmpl`.

```
$ gen-jsonschema gen --target .
```

produced real artifacts:

- `jsonschema/Widget.json`: `{"type":"object","properties":{"name":{"type":"string"},"count":{"type":"integer"}},"required":["name","count"],"additionalProperties":false}`
- `jsonschema_gen.go`: valid `!jsonschema`-tagged Go embedding the schema behind `Widget.Schema()`.

Added a throwaway `_test.go` (removed after) that called `Widget{}.Schema()`, unmarshaled it, and
asserted on `required`: `go build ./... && go vet ./... && go test ./...` all passed. This is a
full scaffold → generate → compile → runtime round trip driven by nothing but the fluent
declaration the CLI itself wrote.

## 2. Representative in-tree fluent examples: independent regeneration matches checked-in artifacts

The examples under review (`providers_rendering`, `interfaces_options`, `sealed_interface_slices`,
`ref_types`, `stringer_enums`) already carry fluent `Declare(...)` registrations in the working
tree (uncommitted docs/examples pass). I regenerated each **in place, with the exact flags each
directory's own `//go:generate` line specifies**, using my independently built CLI binary, then
diffed against the checked-in state and restored anything that didn't match (see hygiene note):

```
(cd examples/sealed_interface_slices && gen-jsonschema gen --pretty)
(cd examples/ref_types              && gen-jsonschema gen --pretty --validate)
(cd examples/providers_rendering    && gen-jsonschema gen)
(cd examples/interfaces_options     && gen-jsonschema gen)
(cd examples/stringer_enums         && gen-jsonschema gen)
```

Result: `git diff HEAD` after regeneration touched only `ref_types/jsonschema/{Shared,Container,NullableConfig}.json`
and `stringer_enums/jsonschema/ApplicationConfig.json`, and in every case the only content change
is the doc-comment wording (`"registered with AsRef()"` → `"registered with Ref()"`,
`"WithStringerEnum emits..."` → `"StringerEnum emits..."`) that the source-side doc-comment edits
flow through into JSON `description` fields. `sealed_interface_slices/jsonschema/Batch.json`,
`interfaces_options` (no generated JSON — inline-only interface), and `providers_rendering`
produced byte-identical output. This independently confirms (with my own binary, not the
implementer's) that the fluent registrations are pixel-for-pixel equivalent to what they replaced,
for providers (Accessor/Method/Function + RenderProviders), refs, enums, string enums, and typed
unions.

Then ran the real test suites against those freshly-regenerated artifacts:

```
$ go test ./examples/providers_rendering/... ./examples/interfaces_options/... \
    ./examples/sealed_interface_slices/... ./examples/ref_types/... ./examples/stringer_enums/... -v
```

All pass, notably:
- `providers_rendering`: `TestRenderedSchema` proves `Accessor`/`Method`/`Function` providers
  actually execute at runtime through `RenderedSchema()` (value-receiver root).
- `sealed_interface_slices`: full typed-union codec round trip — mixed value/pointer
  implementations, marshal → unmarshal → re-marshal byte equality, missing/null → nil, and
  transactional/indexed error paths on bad discriminators — all against a `.Interface(...)`
  fluent registration.
- `ref_types`: `$ref`/`$defs` structural checks, `ValidateJSON` through a ref, and
  `openai_strict_test.go`'s structural-output compatibility check, all against
  `Declare(Shared.Schema).Ref()`.
- `stringer_enums`: wire-format assertions (string names for `StringerEnum`, raw ints for `Enum`)
  round-tripped through `json.Marshal`/`Unmarshal`.

## 3. Pointer-root provider rendering — the one supported shape no existing example covers

No in-tree example uses a pointer-receiver `Declare` root with providers, and the acceptance bar
explicitly requires proving providers "including pointer roots," so I built a standalone fixture:
`ephemeral/issue-73/manager-validation/pointer-provider-fixture/` (own module + `replace`).

```go
func (*Thing) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.Declare((*Thing).Schema).
    Accessor(Thing{}.Name, (*Thing).NameSchema).
    Method(Thing{}.Count, (*Thing).CountSchema).
    RenderProviders()
```

```
$ gen-jsonschema gen
```

produced a `jsonschema_gen.go` whose generated `RenderedSchema()` calls `t.NameSchema()` and
`t.CountSchema(t.Count)` on the pointer receiver. A throwaway test (removed after) instantiated
`&Thing{Name: "widget", Count: 7}`, called `RenderedSchema()`, and asserted the two providers'
distinct `description` sentinel strings actually appeared in the rendered JSON:

```
--- PASS: TestPointerRootRenderedSchemaExecutesProviders (0.00s)
    rendered schema: {"type":"object","description":"Thing is registered from a pointer-receiver root...",
    "properties":{"name":{"type":"string","description":"pointer-root accessor provider ran"},
    "count":{"type":"integer","description":"pointer-root method provider ran"}},...}
```

This proves execution, not just compilation: the previous silent-drop bug the implementer's
worklog describes (pointer-root providers accepted by the scanner but never attached) would have
rendered a plain `{"type":"string"}`/`{"type":"integer"}` here instead of the sentinel
descriptions.

## 4. TypeScript codec output from a fluent declaration

`tests/typescript/fixture/schema.go` (the repo's actual TS-conformance fixture) still uses legacy
syntax and is source under active review, so I did not touch it. Instead I copied its `types.go`
unchanged into two independent consumer modules under
`ephemeral/issue-73/manager-validation/ts-fluent-fixture/{legacy,fluent}/` — `legacy/` keeps the
original `schema.go` verbatim, `fluent/` replaces every registration with the equivalent
`Declare(...)` chain:

```go
var _ = jsonschema.Declare(Detail.Schema).Ref()
var _ = jsonschema.Declare(Composition.Schema).
    Enum(Composition{}.D)
var _ = jsonschema.Declare(Envelope.Schema).
    Interface(Envelope{}.Event, jsonschema.Discriminator("!kind"), jsonschema.Impl("created", Created{}), jsonschema.Impl("deleted", (*Deleted)(nil))).
    Interface(Envelope{}.Other, jsonschema.Discriminator("other-key"), jsonschema.Impl("create\"雪", Created{})).
    Interface(Envelope{}.Maybe, jsonschema.Discriminator("!kind"), jsonschema.Impl("created", Created{}), jsonschema.Impl("deleted", (*Deleted)(nil))).
    Interface(Envelope{}.Events, jsonschema.Discriminator("!kind"), jsonschema.Impl("created", Created{}), jsonschema.Impl("deleted", (*Deleted)(nil))).
    Enum(Envelope{}.Status).
    Enum(Envelope{}.Priority).
    StringerEnum(Envelope{}.PriorityName)
```

Ran the real CLI on both, with TypeScript generation on:

```
(cd legacy && gen-jsonschema gen --typescript ts --typescript-barrel --pretty)
(cd fluent && gen-jsonschema gen --typescript ts --typescript-barrel --pretty)
```

Then compiled each `ts/` output with the repo's own pinned compiler
(`tests/typescript/node_modules/typescript`, strict mode, `ES2022`/`NodeNext`):

```
$ node tests/typescript/node_modules/typescript/bin/tsc --project .../ts/tsconfig.json --pretty false
```

Both compile clean (exit 0). Diffing the two runs' output:

- `jsonschema/Envelope.json` — **byte-identical**. This is the file carrying all four
  `.Interface(...)` union registrations (value/pointer impls, two discriminator key styles, the
  Unicode `create"雪` implementation name), plus the `Priority`/`PriorityName` enum pair. Proves
  fluent `.Interface`/`.Enum`/`.StringerEnum` reproduce the legacy split-option registration
  exactly for a non-trivial, multi-union, multi-enum struct.
- `jsonschema/Detail.json` — byte-identical (`.Ref()` parity).
- `ts/index.ts` (the barrel) — byte-identical.
- `jsonschema/Composition.json` and `ts/types.ts` — **differ**, isolated entirely to the shared
  `Status` type (see finding below); every other field (`a`/`b`/`e`/`f`/`g`/`h`/`i`, the `Optional`
  and `Nullable` wrapper handling) is identical.

### Finding: field-level `.Enum`/`.StringerEnum` is not a full substitute for package-level `NewEnumType[T]()` when one enum type is shared across independently-registered structs

The legacy fixture registers `Status` once, standalone: `jsonschema.NewEnumType[Status]()`. That
makes `Status` a shared, named enum recognized everywhere it appears — as `Envelope.Status`, as
`Composition.C` (`Optional[[]Status]`), and as `Composition.D` (`Nullable[Status]`) — and the
generated TypeScript emits one reusable `export type Status = "ready" | "wait\"ing" | "converted"`
that all three usages reference.

There is no fluent equivalent for that package-level registration: `Declare[T]` requires a
`func(T) json.RawMessage` schema entrypoint, which a bare enum type like `Status` doesn't have.
The only fluent tool is the field-scoped `.Enum(field)`/`.StringerEnum(field)`, attached to
whichever struct's `Declare` chain owns that field. This has two real consequences, confirmed by
running the CLI, not by inspection:

1. **One shape is flatly unsupported at the field level regardless of syntax.** `Composition.C`
   is `Optional[[]Status]`. Calling `.Enum(Composition{}.C)` fails with `field Composition.C:
   WithEnum/WithStringerEnum supports only a direct named enum, Optional[E], or Nullable[E]` — a
   scanner error, immediately, with the built CLI. This is a pre-existing restriction shared by
   legacy `WithEnum` (same error text, same code path) — **not** a fluent-specific regression —
   but it means there is no way, fluent or legacy-field-level, to mark a slice-of-enum field as an
   enum; only the package-level `NewEnumType[T]()` route covers that shape, by making the slice's
   *element type* recognized wherever it appears.
2. **Where field-level annotation is legal, it doesn't unify.** Marking only
   `Envelope.Status` and `Composition.D` with `.Enum(...)` (leaving `Composition.C` unmarked, since
   (1) forbids it) produces: `Composition.C` renders as a plain `{"type":"string"}` array item (no
   `enum` constraint at all — silently weaker validation, not an error) and its TypeScript element
   type widens to `string`; the shared `Status` TypeScript type alias itself degrades to
   `export type Status = string`; and `Envelope.status`/`Composition.d` each get their *own*
   independently-inlined `"ready" | "wait\"ing" | "converted"` literal union instead of a shared
   named reference. The JSON Schema and TypeScript are both still internally valid and each
   individually-marked field is still correctly constrained — this is a loss of sharing and of
   `Composition.C`'s validation, not a crash or a wrong-value bug — but it is a real,
   CLI-confirmed divergence from "equivalent artifacts," and it is silent: nothing in the fluent
   path warns that leaving one co-occurrence of a shared enum type unmarked will quietly drop that
   occurrence's constraint and flatten the shared type everywhere else.

This is the same category of gap the implementer's own worklog already flagged and deliberately
worked around for `examples/ref_types`'s `NullableConfig.Mode` (kept on legacy `NewEnumType`
rather than converizing, after finding a doc-comment-description loss). What this fixture adds is
a sharper, CLI-verified reproduction showing the gap is broader than a lost description: it's a
lost shared-type structure and a lost constraint on one field, whenever the same enum type is
reused across more than one independently-`Declare`d struct. Since 1.x explicitly keeps legacy
declarations "source-compatible... so existing users can upgrade without a flag day," the correct
takeaway is not that anything needs fixing in this pass, but that migration guidance for enums
should say explicitly: *a type-level `NewEnumType[T]()` registration whose type is shared across
more than one struct field has no fluent replacement — keep it on the legacy form.* Root manager's
call on whether/where to add that guidance; no fix attempted here.

## 5. Full-repo regression check (read-only)

```
$ go build ./...     # clean
$ go test ./...       # ok, every package, including all five example suites above
```

## Working-tree hygiene

Regenerating in `examples/*` in step 2 runs the CLI directly against the live product tree (the
only way to compare against the checked-in artifacts under the exact `go:generate` flags). Two
regenerations initially drifted from the checked-in state because I first ran `gen-jsonschema gen`
without each directory's specific flags (missed `--pretty` on `sealed_interface_slices`, missed
`--pretty --validate` on `ref_types`), producing spurious diffs in
`sealed_interface_slices/jsonschema/Batch.json(.sum)` and `ref_types/jsonschema_gen.go` that were
not present in the working tree before I ran anything. Caught immediately by diffing against
`git diff HEAD` and comparing to the original tracked-modification list, and reverted with
`git checkout -- <path>` before re-running with the correct flags. After the correct-flag rerun,
`git status --porcelain` for every touched example directory matched the pre-existing
uncommitted-docs-pass state exactly (same file set, same content) — nothing from this validation
pass is left behind in the product tree.

Unrelated to this validation: `examples_regenerate_test.go`, `internal/builder/typegrammar.go`,
and `json_schema_helpers_test.go` show small (2-4 line) uncommitted diffs in the shared worktree
that were not present at the start of this session and that I did not create — consistent with the
implementer's own worklog note that a `just lint`/`goimports` pass touches exactly these three
files as an incidental formatting side effect. Left untouched per "do not alter its files" (the
source-review agent's).

All new artifacts from this validation live under `ephemeral/issue-73/manager-validation/`:
`bin/` (built CLI), `scaffold-demo/`, `pointer-provider-fixture/`, `ts-fluent-fixture/{legacy,fluent}/`.
