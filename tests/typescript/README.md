# TypeScript conformance lane

Run `npm ci --ignore-scripts --no-audit --no-fund` and `npm test` from this directory. This development-only lane pins the TypeScript compiler and exercises actual CLI output. Neither production generation nor `go test ./...` invokes Node or npm.

The harness first regenerates and compiles the backend projection edge cases, then builds the Go CLI, creates a fresh consumer, and invokes generation with only Go available on PATH. It checks output directories, the optional barrel, deterministic generation, stale-output detection, ownership conflicts, and strict TS compilation. Independent type obligations cover constructor compositions and positive/negative assignments; compiler diagnostics demonstrate missing-case and added-variant exhaustiveness failures. This establishes structural declaration behavior. Runtime bidirectional transport is tracked separately in issue #71.

Inspectable consumer sources, generated files, compiler diagnostics, and command results are retained beneath `ephemeral/typescript-generation/proof/` (override the root with `TS_PROOF_DIR`). Each run has its own directory. Failed runs retain their partial evidence too.
