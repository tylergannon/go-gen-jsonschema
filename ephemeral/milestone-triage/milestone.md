The 1.0 goal is a dependable Go/JSON boundary that people can adopt through their coding agents. Stable means the shipped API and wire behavior work for the documented types, failures do not silently discard user declarations, the examples are truthful, and a fresh consumer can install and use the release.

## Priority by real user value

1. **Fix concrete correctness gaps:** #77 makes every advertised example runnable; #80 rejects declaration options that would otherwise be silently ignored.
2. **Improve the authoring API:** #73 adds the typed fluent `Declare` API, makes it the canonical documented syntax, and deprecates the legacy declaration syntax while preserving compatibility.
3. **Make agent adoption easy:** #64 packages the shipped guidance as a small installable Codex plugin using the current pinned tool workflow and the fluent API.
4. **Check the product boundary:** #62 runs a bounded fresh-consumer check across representative Go, YAML, TypeScript, union, enum, and presence workflows. Existing RC5 evidence counts; this is not an exhaustive conformance program.
5. **Ship:** #66 settles the license, reconciles migration/support docs, verifies the exact candidate, tags v1.0.0, and publishes release notes.

#77 can proceed independently. #80 lands before #73 because both touch declaration diagnostics. Plugin packaging can begin in parallel, then adopt #73's final syntax. #62 consumes the completed behavior; #66 is last.

## Deliberately after 1.0

- #60: a versioned machine-readable capability and diagnostic protocol. Useful agent infrastructure, but it creates a new public protocol and the current PR conflicts with main.
- #61: complete read-only diff/preview infrastructure. `--no-changes` plus normal generated diffs are adequate for initial adoption.
- #65: a paid/model-backed real-agent evaluation harness. Continue it as a quality program, not a library release gate.

Completed contract, codec, helper, cleanup, TypeScript, and RC proof issues remain attached below as release history. App-server assistance, streaming runtime APIs, broad new Go shapes, and MCP tool generation remain future work. No release date is assumed.
