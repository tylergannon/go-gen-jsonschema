## Outcome

Run a bounded consumer-level check that the primary 1.0 portability claims work before the stable tag.

## 1.0 acceptance

- Exercise representative fresh-consumer workflows for ordinary structs with Optional/Nullable fields, a registered typed union, a string-mode integer enum, YAML through the documented JSON semantics, and generated TypeScript declarations.
- For the union and enum cases, encode, validate, decode, and compare semantic meaning. Include one expected rejection for an unsupported or invalid input.
- Run the generated consumer code explicitly, including any nested module tests, and verify clean regeneration with `--no-changes`.
- Reuse existing repository fixtures and the published RC5 Go → TypeScript → Go smoke evidence from #71/#78/#79. Add proof only for a material advertised workflow that is not already covered.
- Turn any reproducible product failure into a focused bug. Do not expand this issue into an exhaustive Cartesian matrix of Go shapes, platforms, and agent hosts.

This is a confidence check on the supported product, not a generalized conformance framework. Exhaustive combination coverage, artifact publication infrastructure, and broad platform matrices can continue after 1.0.

Depends on the codec work already completed in #57/#58, the focused correctness work in #77/#80, and the canonical authoring API in #73.
