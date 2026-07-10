Here is my goal:

Implement GitHub issue 29 in `/Users/tyler/src/go-gen-jsonschema`: repair
exported/unexported struct-field type discovery, preserve deliberate skip and
embedded-field behavior, resolve registered local interfaces, load concrete
cross-package schemas, keep `time.Time` a renderer-owned leaf, and safely
classify already-loaded remote registered enums.

Perform a final implementation review of the current working tree against:

- `docs/design/issue-29-plan.md`
- the current product-code and test diff (`git diff`)
- untracked `internal/syntax/scan_result_test.go`
- untracked `internal/builder/testfixtures/traversal/`
- generated untracked `internal/builder/test_run/test11-traversal/`

The reviewed plan had already reached consensus. During implementation, the
full test suite exposed two compatibility paths that were dormant behind the
visibility bug: map expressions in syntax-only flattening and named interfaces
registered through v1 builder options. The implementation now treats
`*dst.MapType` and `*dst.InterfaceType` as non-recursive discovery boundaries;
maps remain renderer-rejected, and v1 interface implementations remain owned
by the builder. Review that adjustment especially carefully for hidden
acceptance of unsupported schemas, missed remote dependencies, or a simpler
correct boundary.

Also verify the complete implementation for correctness, regression risk,
scope drift, generated-fixture fidelity, proof gaps, and order-dependent
behavior. The current proof state is:

- focused syntax tests pass;
- `TestBasic/(test5-interfaces|test9-v1-interfaces-options|test11-traversal)`
  passes;
- `go generate ./...` passes;
- `JSONSCHEMA_NO_CHANGES=1 go generate ./...` passes;
- `go test ./...` passes after the compatibility adjustment.

Do not edit product code, tests, fixtures, or the plan. You may inspect the
repository and GitHub. Write the exact prompt you were given and your findings
to:

`ephemeral/reviews/202607101300-issue-29-implementation-sonnet.md`

Label every finding with exactly one of:

- **critical:** must fix before proceeding.
- **bug:** demonstrable incorrect behavior, broken contract, race, or
  regression.
- **design:** architecture, boundary, scope, maintainability, or proof issue
  that is materially likely to cause problems.
- **nit:** small cleanup that should not block progress.

Use file/line references and concrete evidence. Prefer real findings over
style preferences. If you identify a real bug outside Issue 29, place it in a
separate `Out-of-scope bugs` section so it can be classified independently.
If no critical, bug, or design findings remain for Issue 29, say so explicitly.
