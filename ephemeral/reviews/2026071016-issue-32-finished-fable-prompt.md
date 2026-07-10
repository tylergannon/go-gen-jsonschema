You are the independent finished-product reviewer for issue 32 in
go-gen-jsonschema.

Goal: review the complete current worktree implementation against the concrete
definition of done at `docs/design/issue-32-definition-of-done.md`. The product
is public `jsonschema.Optional[T]` and `jsonschema.Nullable[T]`, direct-field
generator support, generated decoder behavior, legacy optional-tag removal,
and an executable public consumer proof. OpenAI compatibility must follow the
published strict Structured Outputs rules in the artifact; no live API call or
credential is part of the proof.

Review the actual current worktree, including tracked and untracked files. Read
the definition of done, `git diff origin/main`, the runtime types, syntax and
builder implementation, generated code/template changes, fixtures, public
docs, `examples/optionality`, its negative fixtures, and its deterministic
proof transcript. Run relevant commands where useful.

Give a broad, adversarial review focused on exactly these four axes:

1. correctness;
2. absence of over-engineering;
3. completeness against the concrete product contract;
4. appropriate level of code factoring.

Do not give general approval. Find concrete issues. Pay particular attention to
presence/null runtime semantics, canonical wrapper recognition, unsupported
placement diagnostics, Optional requiredness and omitzero, Nullable schema
encoding, named/scalar/container/object/interface paths, transactional generated
decoding, legacy behavior removal, generator idempotence, and whether the proof
actually exercises the public behavior it claims.

Use these labels exactly:

- critical: must fix before proceeding.
- bug: demonstrable incorrect behavior, broken contract, race, or regression.
- design: architecture, boundary, scope, maintainability, or proof issue that is
  materially likely to cause problems.
- nit: small cleanup that should not block progress.

Write the exact prompt you received followed by your findings to
`ephemeral/reviews/2026071016-issue-32-finished-fable.md`. Include file and line
references, evidence, and the smallest appropriate fix. Do not edit product
code or any other artifact.
