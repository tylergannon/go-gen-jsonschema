# Authoritative type grammar: adversarial review, round 1

Outcome: **material findings remain**.

## Target and authority

Reviewed the uncommitted `internal/typegrammar` implementation on `codex/type-definition-grammar`, based on `604fabb`, in `/Users/tyler/src/.worktrees/go-gen-jsonschema/type-definition-grammar`.

The authoritative user instruction is to first define Go types for the accepted type-definition grammar, including the no-cycle and sealed-union restrictions, as the foundation for native Go TypeScript generation. The accepted contract is `docs/spec/v1.md`. The prior request permits declaration work independently from the already requested #57 codecs and the later #71 bidirectional proof. Those boundaries justify finishing this grammar step without a scanner adapter, TypeScript emitter, or codec implementation; they do not exempt grammar admission or ownership mistakes from review.

Read `AGENTS.md`, `agent-protocol`, `session-worklog`, and `adversarial-review`; the artifact is under the review skill's required `ephemeral/reviews/` directory, as confirmed by the caller. No narrowing instructions were ignored. No product files were edited by this review. A memory-registry quick pass found only older documentation/release workflow material, which was not used to establish any product facts or findings.

## Evidence inspected

- Complete current `internal/typegrammar/grammar.go`, `validate.go`, and `grammar_test.go`.
- Complete accepted `docs/spec/v1.md` and supporting `ephemeral/type-grammar/source-review.md`.
- Existing `internal/builder/gen_schema.go` registration, source resolution, enum handling, struct-field rendering, and union implementation checks.
- Current working tree and task worklog. The new grammar remains internal and is explicitly not wired into generation yet.
- Baseline `go test ./...` passed before review. The updated suite passed again after inspection; the caller remains responsible for final checks after addressing the finding and settling concurrent test edits.

## Finding 1 — issue: anonymous objects admit field-local string enum adapters without a supported owner

**Evidence:** `internal/typegrammar/validate.go:186-198` passes `directEnum=true` for every Required, Optional, and Nullable field operand, even when `namedOwner=false`. The enum guard at `:126-130` checks only `directEnum`. Consequently this graph is admitted:

```text
Definition Owner = Object {
    Inline: Required(Object {
        State: Required(EnumNames(State, Ready = 0))
    })
}
```

The inline Object is visited with `namedOwner=false`, but its State field restores `directEnum=true`. Wrapping that anonymous object in an ordinary pointer, slice, or array likewise still admits the registration.

This is an unsupported ownership context, rather than ordinary nested data. `docs/spec/v1.md:57-59` assigns field-specific enum and union adaptations to composed containing-struct methods, and `:53` explicitly declines to promise all constructor combinations. Existing enum registration is looked up by named owner and direct Go field in `internal/builder/gen_schema.go:1286-1290`; there is no resolved named owner for the anonymous object's State field. The new model already enforces the corresponding named-owner restriction for registered unions, but its enum guard omits that requirement. Since EnumNames represents a resolved Go-field wire adaptation, admitting this graph expands the purported accepted grammar to a registration/codec context that is neither established nor represented with enough owner identity.

**Impact:** A future backend can rely on successful grammar validation and generate a string property for this case even though the accepted owner adaptation cannot be supplied through the established source registration path. It also weakens the induction base: successful admission no longer establishes that every field-local adapter has a supported owner context.

**Requested correction:** Permit EnumNames operands only in direct fields of a supported named object owner (including the accepted direct Optional/Nullable wrappers). Continue allowing a separately named Object reached through a Ref to own its own adapted fields when nested in an ordinary container. Preserve EnumValues composition and reusable named Enum definitions. Add paired positive/negative tests for anonymous-owner rejection and named nested-owner acceptance, including shared nodes so memoization cannot bypass the context check.

No other material findings were identified in this review. In particular, the finite-DAG checks account for named references, actual pointer-linked node cycles, and union implementation edges; acyclic sharing is allowed. Exact constants and field-local union tags preserve the information needed for later target projections, and the comments distinguish static structure from source admissibility and runtime transport proof.

## Follow-up verification — finding resolved

The caller accepted Finding 1 and changed ordinary FieldValue traversal to use `namedOwner` for the direct-enum context. The error and package comments now name the supported named-owner field requirement. I re-read the fix and source registration parser: `internal/syntax/scan_expr.go:350-364` requires a named receiver composite literal followed by a field selector; `:438-441` routes the enum options through that same path.

The revised traversal rejects anonymous enum adapters without blocking enum fields owned by a named definition reached through a Ref. The memoization key retains both `namedOwner` and `directEnum`, so previously validating the same Object or Enum pointer in a legal context cannot short-circuit a later illegal context. `TestValidateEnumOwnerContext/shared_owner_becomes_inline` now demonstrates that distinction for Required, Optional, and Nullable. Paired tests retain anonymous EnumValues and named-owner-in-collection support.

Verification command completed successfully:

```sh
go test ./internal/typegrammar -count=1 -run 'TestValidate(EnumContext|EnumOwnerContext|SharedAcyclicGrammar|RejectsCycles)' -v
```

The full Go suite also passed at the start of this follow-up. The caller owns the final serial full-suite check after all edits; parallel full suites should be avoided because existing builder tests share `test_run` paths.

**Updated outcome: no findings.** The original finding remains above as review history; its admission defect is resolved by the inspected code and regression tests. This follow-up changed only this review artifact.
