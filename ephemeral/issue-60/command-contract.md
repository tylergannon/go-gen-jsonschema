# Agent inspection command contract

`gen-jsonschema version [--json]` reports the installed executable build identity, v1 contract identity, and conservative capability declarations.

`gen-jsonschema inspect [--json] [--target DIR] [Type...]` inspects named registered roots, or every registered root when no type names are given. It never renders or writes generated artifacts.

Machine mode emits one JSON document to stdout. Human mode emits its rendering to stderr. Exit status is 0 for version and fully supported inspection, 2 for an invalid request, 3 for a valid model with unsupported or unknown shapes, and 1 for internal or toolchain failures.

The result envelope has `schemaVersion`, `kind`, `status`, `tool`, `contractVersion`, `capabilities`, `types`, and `diagnostics`. Build identity includes explicit release/development/unknown version state and known/unknown revision state. Each inspected type reports `schema`, `json_encode`, `json_decode`, `validation`, and `yaml_input` independently; this describes installed generator behavior and does not assert that generated files already exist. Diagnostics carry a stable code and classification plus available Go type, Go field, and source context. Optional fields may be added within schema version 1; existing field meaning is stable.

Types sort by `typePath`. Diagnostics sort by `fieldPath` and then `code`. Capability order is the fixed surface order above. Aggregate status uses `error > invalid > unsupported > unknown > supported`, including both top-level diagnostics and every per-type status. Unproved combinations are `unknown`; an unsupported codec surface does not downgrade an independently supported schema surface.
