Here is my goal:

Read GitHub issue 30 in this repository and produce a correct, minimal,
implementation-ready written plan for fixing omitted JSON tag names so generated
schema property names agree with Go's encoding/json behavior. The work is
planning only: do not implement the fix. Keep issue 30 independent of issues 28
and 29 except for documenting relevant constraints. Respect the repository's
AGENTS.md test requirements and the user's requirement to work only in this
separate worktree.

Review the implementation plan at:
docs/design/issue-30-json-tag-name-plan.md

Use `gh` to read issue 30 and its comments. Inspect the relevant source, tests,
fixtures, generator harness, and repository conventions yourself. Review at the
level of the actual goal, not merely formatting. Look for correctness problems,
hidden dependencies, wrong ownership boundaries, over-engineering, unrequested
features or layers, missing test cases, proof gaps, and simpler implementation
paths that still satisfy the issue.

Do not edit product code or the plan. Write the exact prompt you were given and
your findings to:
ephemeral/reviews/20260710-issue-30-plan-round-1-fable.md

Label every finding using these definitions:

- critical: must fix before proceeding.
- bug: demonstrable incorrect behavior, broken contract, race, or regression.
- design: architecture, boundary, scope, maintainability, or proof issue that is
  materially likely to cause problems.
- nit: small cleanup that should not block progress.

Give file and line references where useful. End with a concise recommended plan
shape and list any evidence you could not verify.
