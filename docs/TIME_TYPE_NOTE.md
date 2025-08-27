# time.Time Special Case Requirement

## Problem
The `time.Time` type from Go's standard library cannot be handled generically due to its internal implementation with unexported fields. When the schema generator encounters `time.Time`, it fails with:
```
panic: resolveLocalInterfaceProps could not resolve package for type Time at time
```

## Why This Happens
- `time.Time` has complex internal structure with unexported fields
- The type is from an external package (standard library)
- Generic type resolution fails for this special case

## Proposed Solution
Add special case handling for `time.Time` in the schema generator:
- Detect `time.Time` type specifically
- Generate appropriate JSON schema (likely `{"type": "string", "format": "date-time"}`)
- This matches how `time.Time` is typically serialized to JSON (RFC3339 format)

## Implementation Location
The special case should be added in `internal/builder/gen_schema.go` in the type mapping logic.