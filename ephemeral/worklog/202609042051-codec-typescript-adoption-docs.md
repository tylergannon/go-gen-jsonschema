# Codec and TypeScript adoption documentation

correction: The merged TypeScript output is structural only; do not describe a Go roundtrip plus `tsc` compilation as executed JavaScript transport. Adoption proof must run JavaScript; #71 retains the broader transport matrix.

decision: Document one generation run for schemas, generated Go owner codecs, validation, and TypeScript declarations. Owner codecs are selected by `WithInterface` and `WithStringerEnum` registrations and need no separate CLI flag.

decision: Reproducible consumers must pin the generator and imported marker/runtime package to the same module release. The combined surface requires v1.0.0-rc.5 or newer; v1.0.0-rc.4 includes declarations but predates generated owner codecs.

correction: Avoid time-sensitive claims about what `@latest` resolves to; state the durable minimum release and require an explicit version pin.

friction: A double-quoted shell `rg` pattern containing Markdown backticks triggered command substitution during the wording audit -> use single-quoted shell patterns when searching literal backticks.

doc_bug: The website CLI reference omits `--typescript` and `--typescript-barrel`, and its no-change descriptions mention schemas only -> align it with live CLI help and requested TypeScript artifact behavior.

doc_bug: The shipped hook guide says no-change mode never writes, but generated Go can still update when schema and requested TypeScript outputs are unchanged -> pair strict generation with a Git status check and avoid parallel readers.

skill_issue: /Users/tyler/.agents/skills/go-gen-jsonschema source=merged v1 contract severity=bug -> the installed skill says Go 1.24+ while the shipped skill and public docs correctly say Go 1.27+; do not regress the shipped copy, and refresh the installed skill through the normal release/install path.

correction: Root remains the manager; this worker owns the bounded documentation edits, verification, and checkpoint only. Runtime transport proof and release/tagging stay with their assigned owners.

proof: `npm ci && npm run check` in `website/` built 13 routes and verified all internal links across 17 HTML pages; final `go test ./...` and `git diff --check` passed. The test suite left no temporary `enum_json_string_*` fixture in this worktree.

state: Documentation was published for review on `codex/codec-adoption-docs`; no feature code, tag, or release was performed.

decision: Root accepted the bounded cross-runtime proof at artifact commit 74396ed for adoption readiness. It exercises Go to Node to Go with registered union and enum values plus Optional present-empty and Nullable null, while leaving #71's broader matrix open.

state: Cherry-picked the accepted proof artifact into the documentation PR. Publishing `v1.0.0-rc.5` remains a post-merge release action owned by root.
