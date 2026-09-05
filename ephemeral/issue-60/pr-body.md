Agents need to discover the installed generator's behavior and diagnose model boundaries before generating code. Add `version` and read-only `inspect` commands with a versioned JSON envelope, executable identity, per-operation capabilities, stable diagnostic codes, source locations, and defined exit statuses.

Inspection shares scanner/builder operations with generation, preserves the effective Go build context, and reports unproved hooks and wire shapes conservatively. Documentation and the shipped skill explain the contract and compatibility with older binaries.

Validation: full Go suite and website checks passed. A fresh-module proof distinguishes supported, unsupported, unknown, invalid, and dependency-error cases; every machine response is parseable with empty stderr and unchanged consumer files. Independent review is in progress.

Closes #60.
