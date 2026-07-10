Here is my goal:

Produce an implementation-ready plan for GitHub issue 29 in
`/Users/tyler/src/go-gen-jsonschema`: repair exported/unexported struct-field
type traversal without changing embedded-field behavior; include the coupled
registered-local-interface correction exposed by the visibility fix; ensure
exported cross-package fields load and render real schemas rather than `{}`;
avoid recursively loading renderer-owned leaves such as `time.Time`; cover
newly reachable remote registered-enum dependency behavior; and define
regression, generation, and end-to-end proof strong enough to implement
safely.

Review the revised plan at:

`docs/design/issue-29-plan.md`

This is review round 2 after material revisions. Review the whole plan against
the repository and issue rather than limiting yourself to the prior changes.
Look for correctness problems, hidden dependencies, wrong ownership
boundaries, over-engineered components, unrequested features or layers,
missing tests, migration risk, proof gaps, and simpler implementation paths
that still satisfy the work order. In particular, verify that the remote-name
classification design is correct, the remote-interface non-goal is supported
by the repository contract, the fixture topology can prove the stated failure
modes, and the candidate-commit generation gate is executable as written.

Relevant primary evidence includes GitHub issue 29 and both comments,
`internal/syntax/scan_result.go`, `internal/syntax/node_wrappers.go`,
`internal/builder/gen_schema.go`, syntax and builder fixtures/tests,
`.github/workflows/go.yml`, and the issue 28 prerequisite design documents.

Do not edit product code or the proposed plan. You may inspect the repository
and GitHub. Write the exact prompt you were given and your findings to:

`ephemeral/reviews/202607101200-issue-29-plan-sonnet.md`

Label every finding with exactly one of:

- **critical:** must fix before proceeding.
- **bug:** demonstrable incorrect behavior, broken contract, race, or
  regression.
- **design:** architecture, boundary, scope, maintainability, or proof issue
  that is materially likely to cause problems.
- **nit:** small cleanup that should not block progress.

Use file/line references and concrete evidence when possible. Prefer real
findings over style preferences; do not merely approve the plan. If no
critical, bug, or design findings remain, say so explicitly.
