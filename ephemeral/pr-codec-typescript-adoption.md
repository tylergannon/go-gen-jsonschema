The merged TypeScript declarations and field-aware Go codecs were documented separately, leaving consumers without one correct setup path and leaving the website and shipped skill CLI references stale. This change documents a single pinned-release generation flow for validation, TypeScript declarations, and automatically selected owner codecs, then shows validation before Go decoding and the TypeScript runtime-validation boundary.

It also states that the combined published surface begins with planned `v1.0.0-rc.5`: current `v1.0.0-rc.4` has TypeScript declarations but predates the owner codec merge. Issue #71 remains the broader executed Go/JavaScript transport proof; it is not presented as a generated TypeScript runtime codec or validator.

Validation:

- `npm ci && npm run check` in `website/` (production build; all links across 17 HTML pages passed)
- `go test ./...`
- `git diff --check`
