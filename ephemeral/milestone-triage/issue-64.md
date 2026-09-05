## Outcome

Users can install a small Codex plugin and ask their coding agent to adopt or maintain go-gen-jsonschema in a Go repository using the product that exists today.

## 1.0 acceptance

- Package the existing shipped skill and references with a valid `.codex-plugin/plugin.json` and a documented repository install path.
- Teach the pinned `go get -tool` / `go tool gen-jsonschema` workflow, schema and TypeScript generation, generated union/enum codecs, validation, roundtrip checks, and `--no-changes`.
- Resolve commands through the consumer module's pinned tool dependency so guidance does not silently use an unrelated PATH binary.
- In a fresh Codex session, install the repository plugin and complete one representative consumer adoption task using only published repository content.
- Keep the reusable skill files usable by other coding-agent hosts.

## Explicitly deferred

- A new machine-readable capabilities protocol (#60)
- Full read-only diff/preview infrastructure (#61)
- MCP or app-server runtime
- A model matrix or paid real-agent evaluation harness (#65)
- Public plugin-directory submission

The plugin may describe only behavior supported by the pinned release. It can use the current CLI and `--no-changes`; #60 and #61 are not prerequisites.

Related foundations: #63, #71. Plugin packaging can begin in parallel, but final guidance follows the canonical fluent syntax from #73 and the bounded release proof in #62/#66.
