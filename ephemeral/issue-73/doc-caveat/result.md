# Issue 73 — NewEnumType Deprecated comment accuracy

## Scope
Only two files touched, per instructions:

- `union_type.go` — `NewEnumType`'s `Deprecated:` godoc comment
- `website/src/content/docs/api/index.md` — regenerated via gomarkdoc

## Change

Old comment:

    Deprecated: use Declare(T.Schema).Enum(field) on the field referencing the
    enum type instead.

New comment:

    Deprecated: use Declare(T.Schema).Enum(field) on the field referencing the
    enum type for direct field registrations. NewEnumType has no fluent
    replacement and must be retained when the enum type is shared across more
    than one struct field.

This matches the caveat already documented in README.md (lines 264-270):
field-level `.Enum`/`.StringerEnum` only supports a direct named enum,
`Optional[E]`, or `Nullable[E]` field, and cannot cover an enum type shared
across multiple struct fields — that case must stay on package-level
`NewEnumType[T]()`, which has no fluent equivalent.

## Regeneration

Ran the same command used by CI (`.github/workflows/website-pages.yml`) and
declared in `website/package.json`'s `prebuild` script:

    cd website && gomarkdoc -e -o './src/content/docs/api/index.md' ../

Confirmed `gomarkdoc` binary present at `/Users/tyler/go/bin/gomarkdoc`.

## Diff verification

`git diff --stat` touched only `union_type.go` and
`website/src/content/docs/api/index.md`. The website diff is a single-line
change reflecting the updated doc comment — no other content shifted.

## Not done (out of scope)

- No API, runtime, or test changes.
- No commit or push — working tree left with the two edits staged as
  unstaged changes for review.
- Did not touch any other file in the concurrent product-code review's
  scope.
