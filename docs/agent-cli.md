# Machine-readable discovery and inspection

Use the installed binary to discover capabilities rather than inferring them
from the library version or this documentation:

```sh
gen-jsonschema version --json
gen-jsonschema inspect --json --target ./models Request Response
```

`version` reports executable build identity and available capabilities.
`inspect` examines registered schema roots; omit type names to inspect all
registered roots in the target package. Put flags before type names. An
unregistered requested name is reported as unknown with a registration remedy.
Inspection describes what this binary can generate, not whether generated
files already exist or are up to date.

For an existing installation, first read `gen-jsonschema --help` and check
that it lists these subcommands. Older releases can treat unrecognized
subcommands as generation requests. Upgrade the pinned tool to a release
providing discovery before calling these commands.

Inspection does not render artifacts, invoke provider functions or custom
value hooks, run `go generate`, or write source/generated/module files. Package
loading uses `-mod=readonly`; missing module requirements/checksums produce a
toolchain diagnostic. Normal external Go build/module caches may be populated.
Resolve dependencies explicitly, then inspect again.

## Output and exit status

With `--json`, stdout contains one JSON document and human diagnostics are not
mixed into it. Without `--json`, the human rendering goes to stderr. Both
commands support `--help`; `--json --help` returns a successful envelope with a
`usage` string. The machine guarantee applies to `version` and `inspect`.

| Exit | Result |
| --- | --- |
| 0 | Version/help succeeded, or inspection is fully supported. |
| 2 | Invalid request, target, malformed registration source, or Go source. |
| 3 | Inspection found unsupported or unknown capabilities. |
| 1 | Internal error or toolchain/dependency failure. |

Use structured diagnostic codes and classifications for automation. English
`message`, `detail`, and `remedy` strings are explanatory text.

## Envelope version 1

Required top-level fields are `schemaVersion: "1"`, `kind` (`version` or
`inspection`), `status`, `tool`, and `contractVersion: "v1"`. Optional collections
are omitted when empty. Clients must tolerate additional optional fields in
version 1; existing field meanings remain stable.

| Field | Meaning |
| --- | --- |
| `tool` | `name`, `version`, `versionState`, `revision`, `revisionState`, and `modified` from executable build metadata. Version state is `release`, `pseudo`, `development`, or `unknown`; revision state is `known` or `unknown`. No `latest` inference is made. |
| `capabilities[]` | Installed binary capabilities, each with `name`, `status`, and optional explanatory `detail`. Returned by version discovery. |
| `types[]` | Inspected roots with `typePath`, aggregate `status`, per-operation `capabilities`, and any `diagnostics`. |
| `diagnostics[]` | Package/request errors that cannot be assigned to one root. |
| `usage` | Help text when help was requested. |

Per-type capability names are `schema`, `json_encode`, `json_decode`,
`validation`, and `yaml_input`. Schema support does not imply codec support;
for example, runtime provider templates have separate limits. Unproved custom
wire formats are unknown, even if the renderer can emit a permissive schema.

Statuses are `supported`, `unsupported`, `unknown`, `invalid`, and `error`.
Aggregate precedence is `error > invalid > unsupported > unknown > supported`.
Type results sort by `typePath`; type diagnostics sort by `fieldPath`, then
`code`. Consumers should identify capabilities by name rather than array index.

A diagnostic has `code`, `classification`, and `message`, with optional
`remedy`, `typePath`, `fieldPath`, and `source`. Classification is one of
`invalid_request`, `unsupported`, `unknown`, `toolchain`, or `internal`.
`source` supplies `file`, `line`, and `column` when known; locations refer to Go
source, and field paths use Go field names. An unavailable location or remedy
is omitted rather than invented.

The accepted supported boundary is in [the v1 contract](spec/v1.md). Discovery
reports the current executable's capabilities, including features it does not
yet implement; it does not promote the complete release target to current
support.
