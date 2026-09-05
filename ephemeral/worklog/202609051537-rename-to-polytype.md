# Rename go-gen-jsonschema -> protract

## Context
decision: user is renaming this project from go-gen-jsonschema to "protract".
Reason (user, 2026-09-05): scope grew beyond JSON Schema generation to bidi
Go<->JSON codegen with possible projection to other contract types. Pun:
"why make a contract if you have protract" (shared "tract" root with
contract/extract/retract). User wants a name that "isn't a lie" and is durable.

decision: confirmed via AskUserQuestion — new module path
github.com/tylergannon/protract (includes renaming the GitHub repo itself),
and CLI binary renamed from gen-jsonschema to protract.

Checked web search for collisions before proceeding: no existing Go package/
module named "protract" found (pkg.go.dev, GitHub). Only distant cousin is
"Protractor", the deprecated Angular/Selenium e2e test tool — different word,
not judged a real collision.

## Baseline
`go test ./...` passed clean before starting (all packages ok or no test files).

## Naming, continued
decision: after further back-and-forth (protract, manifold, goty, tygo, typeconv,
go-contracts, contractsmith, contractor, ontrack all rejected — see reasons in
memory `project_rename_go_gen_jsonschema`), user picked **polytype**: poly (many)
+ type, one Go type projected into many typed representations (JSON Schema, TS,
JSON codec). Unclaimed, no collision found.

decision: single `polytype` CLI binary (not separate `gen-jsonschema`/`gen-types`
binaries as earlier floated). User: "This module REALLY isn't famous... I think I
have one or two tools that use it and we'll just update those guys next time we
have to deploy them" — i.e. accept the breaking change, low external blast radius.

decision (mid-session correction from user): root package should NOT just get its
clause renamed from `jsonschema` to `polytype` wholesale. Only the pieces that are
"literally shorthand for building a JSON schema" deserve the `jsonschema` name;
everything else belongs in the renamed root `polytype` package. This required
inspecting the actual API surface, not just doing a mechanical rename.

## Execution

Actual package split, decided after reading declare.go/union_type.go/
optionality.go/json_schema.go in full:
- New real subpackage `jsonschema/` (import path
  `github.com/tylergannon/polytype/jsonschema`, package clause unchanged) holds
  the JSON-Schema-construction DSL: `JSONSchema`, `ObjectSchema`, `ParentSchema`,
  `StringSchema`/`BoolSchema`/`IntSchema`/`ArraySchema`/`ConstSchema`/
  `EnumSchema`, `SchemaNode`, `DataType`, `JSONUnionType`, `UnionSchemaEl`,
  `RefSchemaEl`, plus its own test files. Confirmed via repo-wide grep that
  *zero* consumers (examples, internal fixtures, tests) used these symbols
  outside the root package's own tests — safe, low-risk move.
- Root package `polytype` keeps: `Declare`/`.Interface`/`.Enum`/`.Accessor`/
  `.Method`/`.Function`/`.Ref`/`.RenderProviders` (declare.go), the legacy
  `NewJSONSchemaMethod`/`NewJSONSchemaFunc`/`NewInterfaceImpl`/`NewEnumType`/
  `With*` markers plus `Discriminator`/`Impl` (union_type.go), and the runtime
  codec wrapper types `Optional[T]`/`Nullable[T]` (optionality.go) — these are
  the declaration/registration API and bidi JSON codec types, not JSON-Schema-
  specific.

friction: internal/syntax resolves markers by import PATH
(`SchemaPackagePath` constant in internal/syntax/import_map.go), not by the
locally-aliased identifier name — confirmed by reading node_wrappers.go's
`wrapperExpr`/`Wrapper()` logic before assuming the split was risky. This meant
the scanner needed zero logic changes for the package split; only cosmetic
error-message/doc-comment strings ("jsonschema.Declare" -> "polytype.Declare",
"jsonschema.Optional" -> "polytype.Optional" in `WrapperKind.String()`) needed
updating, plus the test assertions checking those exact error strings
(scanner_test.go, fluent_declaration_test.go).

friction: found a real, pre-existing naming collision risk before it caused a
correctness bug: generated code with `--validate` imports the *third-party*
`github.com/santhosh-tekuri/jsonschema/v6` library, unaliased, which also
resolves to local identifier `jsonschema` (e.g. `jsonschema.Schema`,
`jsonschema.NewCompiler`). A blind repo-wide "jsonschema." -> "polytype."
substitution would have corrupted these third-party references. Fix: scoped
every substitution to the exact whitelist of our own root-package identifiers
(Declare, Optional, Nullable, Discriminator, Impl, AsRef, NewEnumType,
NewInterfaceImpl, NewJSONSchemaBuilder, NewJSONSchemaFunc, NewJSONSchemaMethod,
WithDescription, WithDiscriminator, WithEnum, WithFunction, WithInterface,
WithInterfaceImpls, WithRenderProviders, WithStringerEnum,
WithStructAccessorMethod, WithStructFunctionMethod) via a regex alternation,
never a bare "jsonschema." replace. Verified zero overlap between this
whitelist and the third-party library's real exports before running it.

friction: BSD/macOS `sed -E` silently no-ops on `\b` word-boundary regex (no
error, just doesn't match) — a `sed -i '' -E "s/\bjsonschema\.(...)\b/.../"`
pass appeared to succeed (exit 0) but changed nothing. Caught it only by
re-grepping after the "successful" run. Switched to `perl -pi -e` for every
word-boundary-dependent substitution from that point on. Lesson: on macOS,
never trust a `sed -E` regex using `\b`; use perl instead.

friction: found genuinely pre-existing (not rename-caused) breakage under
`go build -tags jsonschema ./...` — the build mode the registration scanner
actually type-checks against, which the project's own `go test ./...` never
exercises directly: `examples/test_options/schema.go` calls
`jsonschema.WithDescription(...)`, a function that does not exist anywhere in
the codebase; `examples/iota_global/schema.go` calls `NewEnumType[Priority]`
where `Priority` doesn't satisfy `~string`; `examples/indirecttypes/schema.go`
declares methods on invalid pointer receiver types. Confirmed pre-existing (pure
identifier-text sed cannot cause type/constraint errors) and out of scope for
this rename; flagged via spawn_task rather than fixed inline.

friction: golden-file fixtures in `internal/builder/testfixtures/{interfaces,
union_codec,v1_interfaces_options}/jsonschema_gen.go.golden` embed a
content-derived hex hash in generated marshal/unmarshal function names that
changes when the module import path changes. Not caught by the top-level
`go generate ./...` (fixtures are separate Go modules via local `replace`
directives, outside `./...` from repo root). Fix: `go generate ./...` inside
each fixture module directly, diff against its `.golden`, copy over once
confirmed the only diff was the expected hash change.

## Final state
- Module: `github.com/tylergannon/polytype`. CLI: single `polytype` binary
  (was `gen-jsonschema/`). Root package: `polytype`. New subpackage:
  `jsonschema` (DSL only). GitHub repo not yet renamed; nothing committed —
  awaiting explicit go-ahead per repo's own risk-confirmation norms (repo
  rename + push are shared/hard-to-reverse actions).
- `go test ./...` green repo-wide (root module + every fixture submodule under
  internal/builder/testfixtures and internal/builder/test_run), gofmt clean.
- Docs updated: AGENTS.md (source of the CLAUDE.md symlink) architecture
  section, README.md, llms.txt, examples/README.md, prompts/description.md,
  skills/polytype/*, website docs pages, docs/*.md, tests/typescript/check.mjs
  fixture snippet.
- Not updated: `ENUM_OPTIONS_TODO.md` speculative/aspirational sections using
  APIs that don't exist yet (`WithEnumMode`, `WithEnumName`) — left as-is,
  out of scope.
- Not updated (deliberate, per user 2026-09-05): hosted docs website domain
  `https://go-gen-jsonschema.tylergannon.com` in README.md/llms.txt/
  website/astro.config.mjs — the Cloudflare Pages site itself isn't being
  migrated yet; reverted after an earlier blanket-sed pass changed it.

## Adversarial review (round 1)

decision: launched an Opus review via `agent -model opus <repo> "/adversarial-
review ..."` (see [[feedback_naming_brainstorm_style]]-adjacent norm: this was
a structural/correctness check, not a taste call, so the request-adversarial-
review skill applied normally). Framed explicitly as a functional-equivalence
check ("is this a pure rename/move"), not a code-quality review, per user
instruction. Review written to
ephemeral/reviews/202609051650-rename-to-polytype-round-01.md.

Verdict: "material findings remain" — but the core equivalence check passed:
reviewer diffed exported `go doc -all` output between old and new package
layout and confirmed the old root's exported declarations exactly equal the
union of the new root + new `jsonschema` subpackage, byte-for-byte after path
normalization. Two findings, both fixed:

1. (real) `website/src/content/docs/api/index.md` — the gomarkdoc-generated
   API reference page — was never regenerated after the package split, so it
   still showed a single "# jsonschema" section interleaving root API
   (Declare, etc.) with the moved DSL under the new import path. Root cause:
   `website/package.json`'s `prebuild` script only ever pointed gomarkdoc at
   `../` (repo root); it needed `../jsonschema` too now that there are two
   packages. Fixed both the generated file and the prebuild script so future
   builds don't go stale again. Also fixed README.md's manual-construction
   section, which linked to the now-nonexistent root `json_schema.go` and
   didn't name the new subpackage import.
2. (nitpick, zero blast radius) `internal/syntax/import_map.go`'s
   `schemaPackagePrefix` constant (feeding the unused `GetGenJSONPrefix`
   helper — confirmed zero callers repo-wide) still said `"jsonschema"`
   instead of `"polytype"`. Fixed for correctness even though nothing calls
   it today.

`go test ./...` re-confirmed green after both fixes; `go run
./internal/cmd/doc-gen --check` passes.

