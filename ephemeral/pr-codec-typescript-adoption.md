The merged TypeScript declarations and field-aware Go codecs were documented separately, leaving consumers without one correct setup path and leaving the website and shipped skill CLI references stale. This change documents a single pinned-release generation flow for validation, TypeScript declarations, and automatically selected owner codecs, then shows validation before Go decoding and the TypeScript runtime-validation boundary.

The combined surface requires `v1.0.0-rc.5` or newer; `v1.0.0-rc.4` contains TypeScript declarations but predates generated owner codecs. The guide pins that version explicitly, uses one `--validate --typescript DIR` invocation with an optional barrel, and makes clear that TypeScript output is structural rather than a runtime validator.

A fresh external-style consumer [executed the everyday Go → Node → Go path](https://github.com/tylergannon/go-gen-jsonschema/blob/74396ede51e9e47b21d260db6ce2a56eee7663dc/ephemeral/adoption-transport-proof/results/result.txt) against generator source `4374753202591d6b508ecb6f4fbe85f818d0f237`. Node observed Go-authored union, enum, Optional, and Nullable values; compiled TypeScript authored a new union and enum value; Go validated, decoded, changed, and re-encoded it; Node observed the returned meaning, including present-empty and null. The consumer module used the published `v1.0.0-rc.4` marker package without a replacement while the generator binary came from the merged source candidate, because the `v1.0.0-rc.5` release does not exist until this readiness change merges.

This bounded proof demonstrates the adoption path. It does not close #71 or claim that issue's exhaustive numeric, nesting, and error-boundary matrix.

Validation:

- `npm ci && npm run check` in `website/` (production build; all links across 17 HTML pages passed)
- `go test ./...`
- `git diff --check`
- `./ephemeral/adoption-transport-proof/run-proof.sh` (generation idempotent, TypeScript 6.0.3 compiled, Go and Node runtime assertions passed)
