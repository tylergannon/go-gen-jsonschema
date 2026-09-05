# TypeScript projection conformance

This lane directly constructs backend grammar values that cannot all be
expressed by the source fixture. It regenerates an empty module and an edge-case
module, checks atomic rejection of an inexact numeric literal, and compiles the
fresh output with the repository's pinned TypeScript compiler.

From the repository root:

```sh
go run ./tests/typescript/projection/generate \
  ./tests/typescript/projection/generated
node ./tests/typescript/node_modules/typescript/bin/tsc \
  --project ./tests/typescript/projection/tsconfig.json --pretty false
```

The committed consumer proves exact large-number membership, an empty
discriminator tag over a colliding optional payload property, shared-payload
nonmutation, escaped keys and comments, Unicode name collisions, an importable
empty graph, primitive and null exclusion for empty objects, and absence of an
`any` or `unknown` fallback.
