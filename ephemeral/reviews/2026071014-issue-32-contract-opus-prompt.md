Here is my goal:

Re-research GitHub issue #32 as an untrusted proposal and define a safe,
reviewable implementation sequence for explicit absent and nullable Go field
semantics. The result must culminate in real behavioral proof through the public
generator CLI, compiled generated consumer code, JSON marshal/unmarshal,
generated validation, and—if available—the live OpenAI Structured Outputs API.
The existing issue body's large checklist and implementation prescriptions are
not authority.

Review `docs/design/issue-32-checkpoint-plan.md` before implementation. Inspect
issue #32, current source, tests, existing issue-28 design artifacts, and git
history as useful. Look for incorrect semantics, unnecessary API or scope,
missing compatibility constraints, wrong ownership boundaries, hidden decoder
problems, unsafe checkpoint ordering, and proof that would still be fake or
insufficient. Prefer a smaller truthful vertical slice over accepting the issue
body wholesale.

Do not modify product code or the plan. Write the exact prompt you were given
and your findings to:

`ephemeral/reviews/2026071014-issue-32-contract-opus.md`

Label every finding using exactly these definitions:

- **critical:** must fix before proceeding.
- **bug:** demonstrable incorrect behavior, broken contract, race, or regression.
- **design:** architecture, boundary, scope, maintainability, or proof issue
  that is materially likely to cause problems.
- **nit:** small cleanup that should not block progress.

Do not approve the plan. Produce concrete findings, or explicitly state that
you found none at a given severity.
