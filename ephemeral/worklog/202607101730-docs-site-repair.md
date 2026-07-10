# Session worklog: docs site repair

- Started: 2026-07-10 17:30 CST
- Worktree: `/Users/tyler/src/.worktrees/go-gen-jsonschema/docs-site-repair`
- Branch: `codex/docs-site-repair` from current `origin/main` (`08c6cf2`)
- Goal: Crawl `https://go-gen-jsonschema.tylergannon.com/` in Chrome, repair broken links and stale feature documentation, verify the site locally, and deliver the fix.
- Constraint: Keep crawl notes and other raw session material under `ephemeral/`, never `docs/`.

skill_use: chrome:control-chrome source=openai-bundled -> Inspect the deployed documentation site's visible and interactive behavior as explicitly requested.
skill_use: diagnose source=pagerguild/core-tools -> Build a repeatable crawl/link feedback loop before changing documentation.
skill_use: session-worklog source=pagerguild/core-tools -> Track deployed symptoms, source findings, fixes, and proof outside durable docs.

## Baseline

- Root checkout baseline before work: `go test ./...` passed.
- Refreshed `origin/main` and created a new isolated worktree at merge commit `08c6cf2`.
- Root checkout and its existing untracked `.codex/` were left untouched.

## Progress

- Chrome reproduced obsolete setup commands, placeholder guide/reference pages,
  a broken `/spec/v1/` link, legacy example dumps, and a stale generated API.
- Root cause: starter content remained public, routes drifted independently, and
  the deployment workflow did not run when root public Go API files changed.
- First `npm run check` exposed that gomarkdoc v1.1.0 cannot resolve the module
  import path from `website/`; using the explicit local package path `../` works
  and keeps generation bound to the checkout being deployed.
- A clean locked install exposed a high-severity advisory in the inherited Astro
  line. The safe supported pair is Astro 7.0.7 plus Starlight 0.41.3 on Node 22;
  upgrading them together avoids an incompatible peer combination.

## Implementation

- Replaced the starter splash with an agent-first homepage modeled on the useful
  hierarchy at `plainterms-docs.guilde.ai`: human docs, `llms.txt`, one prominent
  agent install command, and task-oriented entry cards.
- Added a working `Click to copy` control for
  `npx skills add tylergannon/go-gen-jsonschema`.
- Rewrote Getting Started for the Go tool directive, mutually exclusive build
  tags, scaffolding, validation, generated outputs, and CI drift checking.
- Replaced legacy fixture dumps and placeholder pages with focused optionality,
  enums, interfaces, providers, validation/CI, CLI, and specification routes.
- Added redirects from `/spec/v1/`, `/guides/example/`, `/reference/example/`,
  and `/implementation/` to current useful destinations.
- Regenerated the checked-in Go API reference from the current checkout; it now
  includes Optional, Nullable, and WithStringerEnum and omits removed APIs.
- Added a deterministic built-site link and fragment checker.
- Added and committed the npm lockfile; pinned gomarkdoc v1.1.0, Astro 7.0.7,
  Starlight 0.41.3, and Node 22 in CI. Public root Go API changes now trigger the
  docs build/deploy workflow.

decision: Generated godoc remains available as a secondary reference; the public homepage and navigation lead with user and agent tasks.
decision: Preserve old public URLs with static redirects rather than allowing removed starter pages to become 404s.
doc_bug: Production homepage and getting-started page taught obsolete non-tool CLI invocation -> replaced with pinned `go tool gen-jsonschema` workflow.
doc_bug: Production examples linked to missing `/spec/v1/` and exposed legacy registration APIs -> replaced with current task pages and redirects.
doc_bug: Website deploy ignored public root Go API changes -> workflow paths now include root Go files and module metadata.

## Proof before commit

- Chrome production crawl reproduced the stale commands, placeholder pages,
  broken `/spec/v1/`, and stale API reference.
- `npm ci --prefix website` passed from a clean `node_modules` directory.
- `npm audit --omit=dev --prefix website` reports zero vulnerabilities.
- `npm run check --prefix website` built 12 content pages plus four redirects;
  the permanent checker validated all 16 HTML outputs and their fragments.
- A second build produced the same API-reference SHA-256.
- Chrome against Astro 7.0.7 preview verified the homepage hierarchy, current
  setup and API signals, all four redirects, and clipboard value
  `npx skills add tylergannon/go-gen-jsonschema` with visible `Copied` feedback.
- `go generate ./...`, `git diff --check`, and final `go test ./...` passed.
- Branch: `codex/docs-site-repair`; commit/PR/preview/merge pending.
