# Modern build tags

goal: Produce only `//go:build` constraints, remove every checked-in legacy `// +build` constraint, prove generated examples stay modern, and auto-merge the resulting PR.

constraints:
- Never add files under `./docs` without express permission.
- Include hand-written files, generated files, and code-generation sources.
- Run repository tests before and after the change.
- Work in an isolated worktree and leave the root checkout untouched.

worktree: `/Users/tyler/src/.worktrees/go-gen-jsonschema/modern-build-tags`

branch: `codex/modern-build-tags`

skill_use: session-worklog source=pagerguild/core-tools -> record inventory, decisions, proof, and PR/merge state.

skill_use: ship source=pagerguild/core-tools -> commit, push, and create the requested PR without another confirmation step.

baseline: `go test ./...` passed in the root checkout before task changes.

inventory:
- 55 tracked legacy constraints were present.
- Generator sources: `gen-jsonschema/tmpl/config.go.tmpl` and `internal/builder/schemas.go.tmpl`.
- Checked-in surfaces: generated examples, builder fixtures/goldens/test-run snapshots, builder messages output, and prompt code samples.

decision: Keep the modern constraint first in generated implementation files, followed by a blank line and the generated-code marker. This makes the output stable without relying on goimports to reorder paired legacy/modern constraints.

tests_added:
- `gen-jsonschema/main_test.go` renders and formats the `new` command template, asserts the modern constraint header, and rejects the legacy form.
- `internal/builder/basic_test.go` inspects every generated `jsonschema_gen.go`, asserts the modern constraint header, and rejects the legacy form.

red_proof: `go test ./gen-jsonschema -run TestNewConfigUsesOnlyGoBuildConstraint -count=1` failed against the original template.

correction_during_work: The first builder regression run showed the generated-code marker preceding the lone modern constraint. The builder template was reordered explicitly; rerunning `go test ./internal/builder -run '^TestBasic$' -count=1` passed.

proof:
- `go test ./gen-jsonschema -run TestNewConfigUsesOnlyGoBuildConstraint -count=1` passed.
- `go test ./internal/builder -run '^TestBasic$' -count=1` passed.
- `go generate ./...` passed.
- `JSONSCHEMA_NO_CHANGES=1 go generate ./...` passed.
- `go test ./...` passed after regeneration.
- `go vet ./...` passed.
- `go build ./...` passed.
- `git diff --check` passed.
- Repository-wide fixed-string search found no legacy build constraints outside the two negative regression assertions and this worklog.

files_intentionally_untouched: No files were added under `./docs`.

commit: `5198ff9 build: remove legacy build constraints` (pre-amend SHA).

pr_state: PR #43 created at `https://github.com/tylergannon/go-gen-jsonschema/pull/43`; auto-merge will be enabled after this worklog closeout is pushed.
