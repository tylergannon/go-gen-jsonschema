# Issue 30: omitted JSON tag name plan

Issue: [#30 — JSON tag with omitted name renders an empty schema property](https://github.com/tylergannon/go-gen-jsonschema/issues/30)

## Goal

Make generated schema property names follow `encoding/json` when a JSON tag
specifies options but omits its name. For example, both `json:",omitzero"` and
`json:",omitempty"` must use the exported Go field name rather than `""`.

This is a prerequisite for issue 28's planned `Optional[T]` idiom, but the fix
must remain independent of the Optional/Nullable implementation and of issue
29.

## Current behavior and ownership

`internal/syntax.StructField.PropNames` owns the mapping from a Go struct field
to the names consumed by schema generation. For a single named field,
`PropNames` currently returns `JSONTag().Options[0]` whenever a JSON tag exists.
The struct-tag parser correctly represents an omitted name as an empty first
option, so `json:",omitzero"` becomes a schema property named `""`.

The downstream builder consistently consumes `PropNames` when collecting seen
properties and rendering object properties. The correction therefore belongs
in `PropNames`, not in the builder's several call sites or in the tag parser.

`StructField.Skip` separately owns field exclusion. It already treats
`json:"-"` as skipped, so this issue does not require changing skip semantics.

## Implementation plan

0. From the isolated issue-30 worktree, run the repository-required baseline
   before editing production code:

   ```text
   go test ./...
   ```

   Stop if the suite cannot run; repair any pre-existing failure before working
   on issue 30. The planning baseline passed at `13421b6`.

1. Add a focused, table-driven `TestStructFieldPropNames` unit test in
   `internal/syntax/node_wrappers_test.go` before changing production code.
   Cover:

   - `json:",omitzero"` falling back to `MaxRetries`;
   - `json:",omitempty"` falling back to `MaxRetries`;
   - `json:"max_retries,omitzero"` preserving `max_retries`;
   - an untagged exported field using `MaxRetries`;
   - `json:""` continuing to use `MaxRetries`;
   - `json:"-"` remaining excluded through `StructField.Skip`.

   Add small grouped and embedded-field characterization cases only if needed
   to make their unchanged behavior explicit.

2. Change `StructField.PropNames` in
   `internal/syntax/node_wrappers.go`. For a single named field, return the
   parsed JSON name only when it is non-empty. When it is empty, continue to
   the existing exported-name fallback. Do not special-case individual options
   such as `omitzero` or `omitempty`, because the JSON name omission rule is
   independent of the option list.

3. Add generated-schema coverage to the existing structs code-generation
   fixture under `internal/builder/testfixtures/structs`:

   - define a dedicated fixture struct containing omitted-name `omitzero` and
     `omitempty` fields, an explicitly named field, an untagged exported field,
     and a `json:"-"` field;
   - register its schema entry point in `schema.go`;
   - add the generated JSON schema to the `TestBasic` golden-file list and
     commit its `.golden` file.

   This statically proves the generator emits the Go fallback names, preserves
   the explicit name, and excludes the skipped field. Do not change the generic
   nested-module harness for this issue.

4. Add the issue's behavioral comparison to `examples/structs`, which is part
   of the root module and already tests generated `Schema()` methods:

   - add a small, legitimate example struct demonstrating omitted-name
     `omitzero` and `omitempty`, an explicit name, an untagged exported field,
     and a skipped field;
   - register its schema method in `examples/structs/schema.go` and run the
     existing generator so the schema, its `.json.sum` checksum, and generated
     Go code are updated;
   - add `TestStructSchemaPropertyNamesMatchJSON` to
     `examples/structs/schema_test.go`; marshal a populated value with
     `encoding/json`, parse the generated `Schema()` document, and compare the
     two property-key sets directly.

   Keep this assertion stdlib-only and focused on property names. This package
   runs under the repository's normal `go test ./...`, so no fixture-harness or
   dependency changes are needed.

5. Prove the change in increasing scope:

   ```text
   go test -run '^TestStructFieldPropNames$' -count=1 -v ./internal/syntax
   go test -run 'TestBasic/test4-structs' -count=1 -v ./internal/builder
   go test -run '^TestStructSchemaPropertyNamesMatchJSON$' -count=1 -v ./examples/structs
   go test ./...
   ```

   Confirm the generated schema has no empty property key, preserves the
   explicit name, omits the ignored field, and exposes the same property keys
   as `encoding/json`.

## Non-goals

- Implementing Optional or Nullable wrappers from issue 28.
- Fixing the separate traversal and registered-interface work in issue 29.
- Replacing or changing the struct-tag parser.
- Changing `jsonschema:"optional"` or required-property semantics.
- Refactoring grouped or embedded field handling.
- Correcting the pre-existing grouped-field divergence where a tag on
  `A, B int` is ignored by `PropNames`.
- Correcting the pre-existing `json:"-,omitempty"` divergence where this
  repository skips the field but `encoding/json` treats `-` as a literal name
  when options are present.
- Defining or enforcing same-level JSON property-name collision policy. An
  omitted-name fallback can collide with a sibling's explicit JSON name, just
  as an untagged field already can; that broader behavior is not changed here.

## Risks and controls

- **Accidental skip regression:** `json:"-"` must continue to be tested through
  `Skip`; the production change should not alter that method.
- **Overfitting to `omitzero`:** the fallback must depend only on an empty tag
  name, with both `omitzero` and `omitempty` proving the general rule.
- **A green unit test but broken generation:** retain the builder-level fixture
  and behavioral schema/JSON key comparison.
- **Scope leakage from issue 28:** keep all new types and behavior specific to
  JSON-name resolution; no Optional API or validation policy belongs here.

## Completion criteria

- The regression test fails before the production change and passes after it.
- Omitted JSON names use the exported Go name in both `PropNames` and generated
  schemas.
- Explicit names, skipped fields, and untagged exported fields retain their
  current behavior.
- Generated schema property keys match `encoding/json` for the behavioral
  example.
- `go test ./...` passes from the issue-30 worktree.

## Consensus review decision

Decision: fix the empty JSON-name fallback only in `StructField.PropNames`,
with a syntax unit table, a builder golden fixture, and a root-module generated
schema versus `encoding/json` behavioral comparison.

Evidence: issue 30 and its comment via `gh`; the `PropNames` and `Skip` source;
all builder `PropNames` call sites; the `TestBasic` fixture harness; the
`examples/structs` generated-schema tests; a green baseline at `13421b6`; and
the independent Fable reproduction of the current failure.

Accepted findings: make the baseline an explicit step; keep the generic fixture
harness unchanged; place runtime proof in `examples/structs`; characterize
`json:""`; pin proof test names; name the generated checksum; and state adjacent
JSON-tag divergences as non-goals.

Rejected or deferred findings: same-level JSON-name collision detection remains
deferred because it is a broader pre-existing policy gap, already reachable by
untagged fields, and is outside issue 30's acceptance criteria. No unresolved
dissent remains.

Review artifacts:

- `ephemeral/reviews/20260710-issue-30-plan-round-1-fable.md`
- `ephemeral/reviews/20260710-issue-30-plan-round-2-sonnet.md`

Proof still required during implementation: demonstrate the regression test is
red before the fix, then run the three focused commands and final `go test ./...`
listed in step 5.
