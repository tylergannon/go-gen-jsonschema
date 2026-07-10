Run all Go unit tests before doing anything, to establish a baseline. Your job
CANNOT be considered complete if you do not run tests after completing the work.
Therefore it is imperative to verify that tests are working before you begin.

If there is an issue running tests at all (i.e. missing modules that can't be loaded),
STOP.
If there are broken tests, begin by fixing those broken tests.

DO NOT consider the job to be complete until unit tests have been updated
and `go test ./...` completes successfully.

If you have been asked to add new functionality, you must write unit tests
that verify the new functionality.
If you have been asked to perform a refactoring, you need only change unit
tests to the extent needed to ensure all functionality has good tests.

## Session Worklog protocol

All agents doing non-trivial repository work MUST follow the Session Worklog
protocol in `/Users/tyler/.agents/skills/session-worklog/SKILL.md`.

- Create or continue a task worktree before writing operating artifacts.
- Keep the tracked session worklog under `ephemeral/worklog/` unless a more
  specific repository policy supersedes that path.
- Keep review prompts, review results, scratch notes, source captures, generated
  packets, and all other temporary or raw session artifacts under `ephemeral/`.
- NEVER place ephemeral or raw session artifacts in `docs/`. Only polished,
  durable project documentation belongs there.
- Update the worklog with commands, decisions, corrections, proof, and final
  branch or review state before closeout. Do not delete an active worklog.
