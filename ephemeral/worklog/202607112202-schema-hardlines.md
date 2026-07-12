# Schema-aware hardline formatting

goal: Replace the v0.11.2 zero-indent pretty printer with a compact, schema-aware hardline renderer, then merge and publish a corrective patch release.
worktree: `/Users/tyler/src/.worktrees/go-gen-jsonschema/schema-hardlines`
branch: `codex/schema-hardlines`
base: `a2b41c62a94be0d48dc617bec90fcb619f113dc5` (`origin/main`, tag `v0.11.2`)

## Session record

- Baseline in root checkout: `go test ./...` passed before task work began.
- Root checkout intentionally left untouched; it contains pre-existing `.claude/` and `.codex/` untracked paths.
- skill_use: go-gen-jsonschema source=pagerguild/core-tools -> used for generator behavior and generation proof requirements.
- skill_use: session-worklog source=pagerguild/core-tools -> required by repository policy for non-trivial work.
- skill_use: ship source=pagerguild/core-tools -> used for automatic commit, push, and PR creation.
- skill_use: caveman-commit source=pagerguild/core-tools -> staging auto-triggered concise Conventional Commit guidance.
- skill_issue: ship source=pagerguild/core-tools severity=design -> ship requires AI attribution while caveman-commit forbids it; use the terse project-compatible commit without attribution.
- correction: v0.11.2 used `json.Indent` with an empty indentation string. That expanded every object and array, retained one space after every colon, and did not implement the requested semantic break policy.
- decision: Preserve deterministic schema/struct-field order; do not sort keys.
- decision: Use unconditional schema-owned hardlines rather than width-based or fixed-depth reflow.
- intended contract: compact leaves and scalar arrays; each object schema breaks after `"type":"object",`; each `properties` member starts on its own line; nested object schemas recurse; no indentation or optional spaces.
- Regression-first run failed to compile because `marshalSchemaHardlines` did not yet exist, establishing the red state.
- implementation: Added a typed hardline renderer over root, object, array, nullable, union, ref, scalar, and template-hole schema nodes; ordinary leaf marshalers remain compact.
- implementation: Non-pretty regular and template schemas now use the hardline renderer; pretty regular schemas retain two-space `encoding/json` output and pretty templates retain their existing raw path.
- Targeted exact-output tests and the full `internal/builder` suite pass, including JSON validity, nested objects, scalar arrays, unions, nullable objects, sorted `$defs`, refs, and template holes.

## Proof and closeout

- Regeneration: `go generate ./...` passed and updated non-pretty schemas/checksums, including the non-pretty provider template.
- Semantic equivalence: every changed `.json` canonicalized with `jq -S -c` matches `v0.11.2`.
- Organization measurement: 4,731 bytes, 39 lines, 344-byte longest line, zero indentation; this is +38 bytes/+0.81% versus original compact and 265 bytes/150 lines less than v0.11.2.
- Corpus measurement: 46 schemas total 34,104 bytes and 347 lines; +301 bytes/+0.89% versus original one-line compact, and -2,149 bytes/-1,245 lines versus v0.11.2.
- Drift proof: `JSONSCHEMA_NO_CHANGES=1 go generate ./...` passed.
- Formatting proof: `git diff --check` passed.
- Build proof: `go build ./...` passed.
- Full unit proof: `go test ./...` passed after implementation and regeneration, including pretty golden fixtures and provider-rendering runtime tests.
- Pending final review, PR, merge, patch tag, merged-head proof, and cleanup.
