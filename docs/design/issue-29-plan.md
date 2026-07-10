# Issue 29: exported-field traversal repair

## Goal

Repair type discovery so schema generation traverses exported struct fields,
ignores unexported or deliberately skipped fields, loads the real schema for
exported cross-package field types, and continues to resolve registered local
interfaces. The change must not turn renderer-owned leaf types such as
`time.Time` into packages for the syntax scanner to recursively load.

Issue 29's visibility defect and its registered-interface follow-up are one
implementation unit. Fixing only `skipField` activates the adjacent inverted
identifier condition and changes a silent omission into a generation error.

## Evidence and current behavior

- `internal/syntax/scan_result.go:583` looks for a lowercase field name and
  skips the field when none exists. That skips ordinary exported fields and
  traverses ordinary unexported fields.
- `internal/syntax/scan_result.go:523` returns success for an unknown local
  identifier but errors when the identifier is a registered interface.
- Once exported traversal is enabled, every exported external identifier is
  added to `remoteTypes`. Without a leaf exception, `time.Time` is loaded and
  scanned even though `SchemaBuilder.renderSchema` already renders it directly
  as an RFC3339 string.
- `resolveTypes` classifies newly loaded remote names through
  `loadPackageInternal`, but its already-loaded dependency branch blindly
  queues `remote.LocalNamedTypes[typeName]`. Registered enums and interfaces
  deliberately do not live in that map, so a repeated remote enum dependency
  can enqueue a zero `TypeSpec` and panic in an order-dependent way.
- `StructField.Skip` already has correct named-field visibility and `json:"-"`
  behavior, but it is not identical to traversal policy: rendering skips an
  unexported embedded identifier, while issue 29 requires traversal's existing
  embedded-field behavior to remain unchanged; `jsonschema:"ref=..."` is also
  traversal-only because the renderer still emits the property.
- The existing `interfaces` builder fixture contains an exported registered-
  interface field, but its successful golden generation currently does not
  exercise syntax traversal because the field is skipped too early.
- Baseline proof was rerun after the requested fast-forward on 2026-07-10:
  `go test ./...` passes on `89962c4`.

## Design decisions

### 1. Share primitives, not the entire skip policy

Add package-private helpers in `internal/syntax` for the two rules that really
are common:

- a named field is visible when at least one declared name is exported;
- a field is ignored when its JSON tag names `-`.

Use `token.IsExported`, not the absence of a lowercase rune or byte indexing.
Have `skipField`, `StructField.Skip`, and `StructField.PropNames` use the named-
visibility helper. Have both skip callers use the JSON-ignore helper. Keep
these caller-specific rules where they are:

- `skipField` continues traversing embedded fields and additionally skips
  `jsonschema:"ref=..."` fields;
- `StructField.Skip` retains its current embedded-field rule and does not skip
  explicit refs, because refs must still become schema properties.

This removes the duplicated inverted logic without conflating two different
policies.

### 2. Make local identifier classification explicit

In `resolveTypeExpr`, classify a local identifier in this order:

1. basic type: leaf, return success;
2. local named type: cycle-check and recurse once;
3. registered enum/constant type: leaf, return success;
4. registered interface: known dependency boundary, return success;
5. anything else: return the existing position-bearing undeclared-local error.

This directly fixes both halves of the inverted condition and makes the
unknown-name contract testable without an invalid Go fixture.

### 3. Do not recursively scan renderer-owned external leaves

Before adding an external identifier to `remoteTypes`, recognize `time.Time`
as a renderer-owned leaf and return success. Apply the same predicate to the
`SelectorExpr` fallback path or assert the decorated-ident invariant there so
the boundary cannot be bypassed. Put the predicate in `internal/syntax` with
an explicit, narrow name; do not exempt an entire package or all unknown
external types. The predicate and the renderer's special case must use the
same canonical type identity: add `syntax.IsTimeType(path, name)` and call it
from both `resolveTypeExpr` and `SchemaBuilder.renderSchema`, so the scanner
and renderer do not maintain independent string literals.

Do not add speculative `Optional`/`Nullable` behavior in issue 29. Issue 28's
later `IndexExpr` classifier must recognize those exact wrapper types and
recurse into their type arguments before any ordinary external-type handling.

### 4. Classify requested remote names once

Extract one `ScanResult` helper that accepts a requested local type name and:

- queues a known `LocalNamedTypes` entry once;
- accepts a registered enum or registered interface as an already-resolved
  leaf;
- returns the existing undeclared-local error for anything else.

Use it both when `loadPackageInternal` handles `typesToMap` and when
`resolveTypes` finds the dependency already in `r.deps`. Never append a map
lookup to `resolveQueue` without checking its boolean result. Add a direct
regression for the already-loaded registered-enum path so correctness does not
depend on Go map iteration order.

The integration fixture should also contain a direct remote enum field plus a
remote struct whose traversal reaches the same enum package. This exercises
the shared-dependency topology that made the old panic possible, while the
focused helper test makes the old failure deterministic.

### 5. Keep type traversal honest

Remove the unreachable `*dst.SliceExpr` branch from `resolveTypeExpr`; slice
types are represented by `*dst.ArrayType`, which is already handled. Do not add
generic `IndexExpr`, map, channel, function, or interface traversal in this
issue. Map and interface declarations are non-recursive discovery boundaries:
maps remain explicitly rejected by the renderer, while named interfaces used
by the v1 builder are resolved through its interface options. This does not add
rendering support for either shape. Channels, functions, and other unsupported
expressions retain their existing explicit errors.

## Implementation sequence

1. Add focused failing syntax tests for named-field visibility and traversal
   skip behavior before changing implementation.
2. Add a registered-interface holder to the existing `typescanner` fixture and
   tests proving the registered identifier succeeds while a synthetic unknown
   local identifier returns the position-bearing error.
3. Extract the shared named-visibility and `json:"-"` helpers; update
   `skipField`, `StructField.Skip`, and `StructField.PropNames` while preserving
   the two skip callers' distinct embedded and ref rules.
4. Rewrite the local-identifier branch as the explicit classification above.
5. Extract the requested-remote-name classifier and use it in both the fresh-
   load and already-loaded dependency paths. Add a deterministic unit test for
   the registered-enum case that previously enqueued a zero `TypeSpec`.
6. Add `syntax.IsTimeType`, use it in the renderer and at every syntax remote-
   registration site, and remove the dead `SliceExpr` case.
7. Add a dedicated `internal/builder/testfixtures/traversal` fixture and a
   `test11-traversal` `TestBasic` case with this exact layout:
   - root `types.go`: registered `TraversalHolder` with fields
     `Remote remotestruct.RemoteStruct`,
     `Status remoteenum.RemoteEnum`, and `When time.Time`;
   - `remoteenum/types.go`: `RemoteEnum` and values;
   - `remoteenum/schema.go`: `NewEnumType[RemoteEnum]()` registration;
   - `remotestruct/types.go`: `RemoteStruct` with a scalar property and a
     `Status remoteenum.RemoteEnum` field.
   Give the fixture its own `go.mod`, generator entrypoint, registration file,
   and golden `TraversalHolder.json`. The root directly reaches both remote
   packages, while resolving `RemoteStruct` independently reaches
   `remoteenum`; whichever root package is visited first, the second path must
   classify an already-loaded remote enum rather than enqueue a zero
   `TypeSpec`.
8. Regenerate the affected fixture goldens and inspect the diff to verify the
   remote property contains the remote struct's actual properties rather than
   `{}`, the remote enum has its actual enum values, and the time property
   retains its RFC3339 string schema.
9. Run focused tests and the builder integration test. On a clean candidate
   commit, run the repository's full generate/clean-tree/no-change/test CI
   sequence.

## Test matrix

### Syntax unit tests

`skipField` table cases:

- one exported name: traverse;
- one unexported name: skip;
- grouped names with at least one exported name: traverse;
- grouped names with every name unexported: skip;
- exported field with `json:"-"`: skip;
- exported field with `jsonschema:"ref=..."`: skip traversal;
- embedded field: preserve current traversal behavior.

Identifier resolution cases:

- exported field whose type is the registered local `MarkerInterface`: no
  undeclared-type error;
- synthetic, genuinely unknown local identifier: undeclared-type error with
  useful expression/position context;
- local named types, enum types, and basic types remain successful;
- an already-loaded remote registered enum is accepted without queuing a zero
  `TypeSpec` or depending on map iteration order.

### Builder integration proof

Generate a schema for the fixture holder and assert through its golden file
that:

- the exported remote field is a concrete object schema containing a known
  remote property, not an empty `{}` fallback;
- the exported remote enum field contains the registered enum's actual values;
- the exported `time.Time` field is a string with the existing RFC3339
  description;
- the existing legacy registered-interface fixture still generates its union
  schema and generated unmarshaler.

The focused integration command should target the affected `TestBasic`
subtests, followed by the mandatory repository-wide test command.

## Proof commands

```sh
go test ./internal/syntax -run 'TestSkipField|TestRegisteredInterfaceIdentifierResolves|TestUnknownLocalIdentifierFails' -v
go test ./internal/builder -run 'TestBasic/(test5-interfaces|test11-traversal)' -v
go generate ./...
test -z "$(git status --porcelain)"
JSONSCHEMA_NO_CHANGES=1 go generate ./...
go test ./...
```

Use the actual fixture selected during implementation in the focused builder
pattern if it is not `test2-indirecttypes`. Review generated JSON and Go golden
diffs in addition to relying on exit status. The clean-tree assertion is a
candidate-commit gate: run it only after the intended source, fixture, and
generated changes have been committed in the implementation worktree.

## Non-goals

- Implementing issue 28's `Optional[T]` or `Nullable[T]` wrappers or generic
  `IndexExpr` traversal.
- Fixing issue 30's omitted JSON-tag name behavior.
- Expanding support for maps, channels, function types, inline interfaces, or
  recursive types.
- Adding support for a field whose type is an interface declared in another
  package and registered only by the consuming package. The legacy scanner's
  interface registry is keyed by local bare names and its own contract limits
  supported interface references to the local package; issue 29 repairs the
  proven local-interface activation path without redesigning that registry.
- Changing embedded-field rendering semantics, ref rendering, schema
  requiredness, or public APIs.

## Consensus reconciliation

The initial independent Fable review found no critical or bug findings. Its
material findings were accepted as follows:

- add the repository's generate, clean-tree, no-change generation, and full-
  test CI sequence to closeout proof;
- repair and directly test the fresh-load/already-loaded remote-name
  classification asymmetry, including a remote registered enum;
- explicitly retain remote registered-interface fields as a non-goal because
  supporting them requires redesigning the legacy registry beyond issue 29.

Its nits were also accepted: cover the `SelectorExpr` fallback in the leaf
boundary, use `token.IsExported`, route `PropNames` through the shared
visibility helper, and back-fill the final focused test names during
implementation.

The second-round Sonnet review found no critical or bug findings. Its one
design finding was accepted by replacing the conceptual integration fixture
with the concrete `test11-traversal` root/`remotestruct`/`remoteenum` package
layout above. Its remaining nit was also accepted by making
`syntax.IsTimeType` the shared canonical-identity helper used by both scanner
and renderer.

The third and final Sonnet review independently traced both possible remote-
package visitation orders and confirmed that `test11-traversal` reaches the
old zero-`TypeSpec` panic in either order. It also confirmed that the shared
time helper follows the repository's existing `builder` to `syntax`
dependency direction. No critical, bug, or design findings remain, and there
is no unresolved dissent.

## Review boundary

Before implementation, obtain an independent goal-level plan review. Revise
for every supported `critical`, `bug`, or material `design` finding, then repeat
with a lighter reviewer only if material findings remain. Implementation is
ready to begin when remaining findings are nits or are rejected with concrete
repository evidence.
