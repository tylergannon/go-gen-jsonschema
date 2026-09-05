# Generated TypeScript transport adoption proof

This is a small external-style consumer for the everyday JSON transport path.
Its `go.mod` pins the published `github.com/tylergannon/go-gen-jsonschema`
`v1.0.0-rc.4` module and contains no `replace` directive. The runner builds the
generator from this repository, then generates Go codecs, JSON Schema, and
TypeScript declarations in one invocation.

The running proof demonstrates three observable claims:

1. Go marshals a shipment containing a discriminated union, a string-mode enum,
   a present Optional value, and a present Nullable timestamp; JavaScript parses
   it and observes the registered discriminator and enum constant name.
2. Compiled TypeScript changes that typed value, `JSON.stringify` writes it, and
   Go validation plus unmarshaling observes the selected union variant, enum,
   present empty Optional value, and null Nullable value.
3. Go changes and re-encodes the TypeScript-authored value; JavaScript parses the
   result and observes the same transport meaning plus the Go changes.

Run from the repository root:

```sh
./ephemeral/adoption-transport-proof/run-proof.sh
```

Inspectable generated files and the latest command results remain in this
directory. This proof deliberately uses ordinary safe integers and timestamp
strings. It does not claim exhaustive numeric, nesting, or error-boundary
coverage from issue #71.
