# Adversarial review: rename to polytype, round 01

## Review target

The uncommitted working-tree change based on `HEAD`, described as a behavior-preserving rename and package split:

- module `github.com/tylergannon/go-gen-jsonschema` to `github.com/tylergannon/polytype`;
- CLI directory/binary `gen-jsonschema` to `polytype`;
- declaration/registration and codec APIs retained in the renamed root package `polytype`;
- the manual JSON Schema construction DSL moved unchanged to `github.com/tylergannon/polytype/jsonschema`;
- mechanical updates to consumers, fixtures, generated output, scaffolding, and documentation.

The deliberately retained hosted-docs domain `https://go-gen-jsonschema.tylergannon.com` was excluded from findings as requested.

## Evidence inspected

- Full `git status`, `git diff HEAD --stat`, `git diff HEAD --summary`, `git diff HEAD --name-status`, and focused full diffs for the root API, moved DSL, CLI/scaffold, `internal/builder`, `internal/syntax`, generated fixtures, examples, and documentation.
- The relevant surrounding source in `declare.go`, `union_type.go`, `optionality.go`, `jsonschema/json_schema.go`, `polytype/main.go`, `polytype/tmpl/config.go.tmpl`, `internal/syntax`, and `internal/builder`.
- Exported API comparison using `go doc -all` against a temporary `git archive HEAD`: after normalizing the module/package paths and sorting declarations, the old root's exported declarations exactly equal the union of the new root and new `jsonschema` subpackage, including signatures and methods.
- Rename detection reports `json_schema.go` to `jsonschema/json_schema.go` at 100% similarity and the CLI implementation at 97% similarity; its only implementation changes are internal import paths. The scaffold emitted by old and new CLIs is identical after the intended import/package-name substitution.
- Searches for the old module and CLI names outside historical `ephemeral/`/tool worktrees found only the three explicitly retained hosted-domain references.
- Baseline `go test ./...` passed. `go test ./...` also passed independently in all 13 `internal/builder/test_run` modules and all 13 `internal/builder/testfixtures` modules. `go run ./internal/cmd/doc-gen --check` passed, but that checker only owns the skill examples and does not validate the hosted Go API reference.

## Findings

### 1. Issue: the public API documentation still assigns the moved DSL to the root package

`website/src/content/docs/api/index.md:10-14` labels the documented package `jsonschema` while telling readers to import `github.com/tylergannon/polytype`. Its single index then interleaves root declarations such as `Declare` (`:19-28`) with DSL declarations such as `JSONSchema`, `ObjectSchema`, and `ParentSchema` (`:37-55`) and the schema helper functions (`:74-82`). That package does not exist as documented: the root import declares package `polytype` (`declare.go:1`), while those DSL symbols now exist only under package/import `github.com/tylergannon/polytype/jsonschema` (`jsonschema/json_schema.go:1,12-26`). A consumer following the API page cannot compile `polytype.JSONSchema`, and a consumer importing the displayed path without an alias cannot refer to it as `jsonschema`.

The same missed boundary is visible in the README's manual-construction entry point: `README.md:576-588` links to the now-nonexistent root path `json_schema.go` and uses `jsonschema.JSONSchema`/helpers without identifying the new subpackage import. This is not equivalent to the stated “everything else ... was updated to match” restructure; the primary API reference still advertises the old monolithic surface under a contradictory name/import pair.

Impact: users are directed to nonexistent root exports for every moved DSL type and helper, so the documented public surface is materially wrong even though the Go declarations themselves were preserved. Regenerate or split the API reference so the root `polytype` API and `polytype/jsonschema` DSL are represented under their actual import paths, and update the README link/import guidance.

### 2. Nitpick: the default-prefix helper still returns the old package identifier

`internal/syntax/import_map.go:11-12` updates `SchemaPackagePath` to the new root path but leaves `schemaPackagePrefix = "jsonschema"`. Consequently, `GetGenJSONPrefix` returns `jsonschema` for an unaliased `import "github.com/tylergannon/polytype"` (`:33-43`), although Go binds that import as `polytype`. Repository search found no current caller, so this does not affect the present scanner path-based resolution or the passing tests, but the helper's documented behavior is now false and it is rename residue that would break any future caller.

## Outcome

**material findings remain**

The implementation and exported Go declaration set preserve the intended functionality/package assignment in the areas inspected, but the hosted API reference materially misrepresents the new package boundary. The stale internal default-prefix helper is a separate low-impact residue.
