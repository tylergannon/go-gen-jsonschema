Here is my goal:

Produce an implementation-ready plan for GitHub issue 29 in
`/Users/tyler/src/go-gen-jsonschema`: repair exported/unexported struct-field
type traversal without changing embedded-field behavior; include the coupled
registered-local-interface correction exposed by the visibility fix; ensure
exported cross-package fields load and render their real schema rather than
`{}`; avoid recursively loading renderer-owned leaves such as `time.Time`; and
define regression and end-to-end proof strong enough to implement safely.

The proposed plan is:

`docs/design/issue-29-plan.md`

Relevant primary evidence includes:

- GitHub issue 29 and both comments;
- `internal/syntax/scan_result.go`, especially `resolveTypeExpr`, `skipField`,
  and `resolveTypes`;
- `internal/syntax/node_wrappers.go`, especially `StructField.Skip`;
- `internal/builder/gen_schema.go`, especially `renderSchema`;
- existing syntax and builder fixtures/tests;
- `docs/design/issue-28-optional-plan.md` prerequisite section;
- `docs/design/issue-28-fable-review.md` sections 6.1, 6.2, 6.4, 8, and 9.

Review this implementation plan before coding. Look broadly for correctness
problems, hidden dependencies, wrong ownership boundaries, over-engineered
components, unrequested features or layers, missing tests, migration risk,
proof gaps, and simpler implementation paths that still satisfy the work
order. Pay particular attention to whether the proposed skip-policy sharing,
external-leaf boundary, fixture choice, and proof commands match the actual
repository.

Do not edit product code or the proposed plan. You may inspect the repository
and GitHub. Write the exact prompt you were given and your findings to:

`ephemeral/reviews/202607101143-issue-29-plan-fable.md`

Label every finding with exactly one of:

- **critical:** must fix before proceeding.
- **bug:** demonstrable incorrect behavior, broken contract, race, or
  regression.
- **design:** architecture, boundary, scope, maintainability, or proof issue
  that is materially likely to cause problems.
- **nit:** small cleanup that should not block progress.

Use file/line references and concrete evidence when possible. Prefer real
findings over style preferences; do not merely approve the plan.
