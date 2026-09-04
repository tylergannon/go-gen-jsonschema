# Issue 33 enum string-mode coverage

decision: Treat GitHub issue 33 as the authoritative contract: register the existing v1 enum string-mode fixture in TestBasic and preserve its generated schema, template, and Go-output coverage.

discovery: The mandated pre-change `go test ./...` baseline passed; the reported defect is a coverage omission rather than an existing test failure.

discovery: Activating the skipped fixture exposed stale tracked output: the current generator emits `Paint.json`, while the old untested artifact was `Paint.json.tmpl`; the fixture had also lost its former string-mode option during the API migration.

decision: Use the supported `WithStringerEnum` registration so this specifically named string-mode fixture retains its intended constant-name schema, and golden-check both `Paint.json` and `jsonschema_gen.go`.

proof: `go test ./internal/builder -run '^TestBasic$/^test10-v1-enums-stringmode$' -count=1` exercises the real fixture-copy, `go generate`, golden comparison, generated-code build, and nested-module test path and passes.

proof: `go generate ./...`, `JSONSCHEMA_NO_CHANGES=1 go generate ./...`, `go build ./...`, `go test ./...`, `go vet ./...`, `golangci-lint run`, and `git diff --check` pass on the implementation worktree.

state: Implementation is complete on `codex/issue-33-enum-stringmode-test`; independent review is pending.
