Here is my goal:

Build the public `jsonschema.Optional[T]` and `jsonschema.Nullable[T]` value
types and teach go-gen-jsonschema to generate schemas and validation behavior
that exactly match their absent-versus-null contracts. Issue #32 records design
decisions made by the maintainer; it is not an instruction to create additional
architecture or ceremony.

Independently review `docs/design/issue-32-definition-of-done.md` against GitHub
issue #32 and the current repository. Do not use previous review artifacts as a
starting point. Determine whether the direction is concrete and appropriately
scoped, whether it faithfully captures the decided feature, and whether the
proposed proof would actually demonstrate the public generator and generated
consumer behavior. Look especially for missing acceptance behavior, accidental
scope expansion, ambiguous completion criteria, fake proof, or a simpler proof
that is equally conclusive.

This is a definition-of-done and proof review, not an implementation code
review. Do not modify product code or the definition-of-done artifact.

Write the exact prompt you received and your findings to:

`ephemeral/reviews/2026071015-issue-32-dod-fable.md`

Label every finding using exactly these definitions:

- **critical:** must fix before proceeding.
- **bug:** demonstrable incorrect behavior, broken contract, race, or regression.
- **design:** architecture, boundary, scope, maintainability, or proof issue
  materially likely to cause problems.
- **nit:** small cleanup that should not block progress.

Do not merely approve the document. Produce concrete findings, or explicitly
state that no findings exist at a severity.
