## Outcome

An agent can discover the exact installed tool's supported behavior and diagnose an unsupported model without inferring state from English log messages or online documentation.

## Acceptance and proof

- Provide an explicit machine-readable mode with a documented, versioned result shape. Final command/flag names are part of this issue's design, not assumed here.
- Report executable version/revision, supported generation/codec modes, and the relevant capability/contract version.
- Inspect requested package types and report supported, unsupported, and unknown cases with stable diagnostic codes, source positions, Go type/field paths, concise explanations, and actionable remedies where known. Do not report unresolved external types/permissive fallback schemas, unproved custom hooks, base64 byte-like slices, or json:,string mappings as bidirectionally supported without matching schema/codec evidence.
- Machine stdout contains only the documented result; human diagnostics use a separate channel. Define exit status and malformed-request behavior.
- Preserve source context when errors propagate from scanner/builder to CLI; eliminate direct debug dumps using #34.
- A fresh agent can inspect a supported model, diagnose an unsupported shape, and distinguish invalid input from an internal/toolchain failure using structured fields alone.
- CLI and any future MCP adapter share the same domain operations/results; this issue does not require an MCP server.

Depends on #56 and #34. Capability reporting must match executable behavior and must not claim future features.