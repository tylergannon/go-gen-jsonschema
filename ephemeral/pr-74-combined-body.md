Generated union decoding previously lost discriminator information when the
containing value was marshaled again. Integer enum fields registered for string
mode also published constant-name strings in their schema while Go encoded the
underlying number and could not decode those names.

Generate one field-aware codec pair on each containing struct. Registered union
fields use closed implementation switches with their resolved discriminator;
integer string-mode enum fields use the exact typed constant membership shared
by JSON Schema and TypeScript generation. The owner wrapper shadows only fields
that need special wire handling and delegates ordinary fields to
`encoding/json`. It preserves Optional/Nullable state, custom concrete hooks,
transactional decoding, collision/promotion checks, nil rejection, and
deterministic repeated rendering.

Validation passed the uncached Go suite, full regeneration plus no-change
generation, and the native TypeScript conformance lane. An independent fresh
consumer generated Go, JSON Schema, and TypeScript from exact candidate
`5e1040a99ba7f51f126601bd84ad5e6b912847d2`; runtime/schema/transactional
negative tests and TypeScript 6.0.3 compilation passed, and the candidate and
canonical fixture remained unchanged. See the
[machine result](https://github.com/tylergannon/go-gen-jsonschema/blob/dbcc9f21cc6baefbc0ab8f71c35eea56e631e827/ephemeral/codec-ts-validation/runs/5e1040a99ba7f51f126601bd84ad5e6b912847d2/result.txt),
[Go transcript](https://github.com/tylergannon/go-gen-jsonschema/blob/dbcc9f21cc6baefbc0ab8f71c35eea56e631e827/ephemeral/codec-ts-validation/runs/5e1040a99ba7f51f126601bd84ad5e6b912847d2/go-test.stdout),
[generated Go codec](https://github.com/tylergannon/go-gen-jsonschema/blob/dbcc9f21cc6baefbc0ab8f71c35eea56e631e827/ephemeral/codec-ts-validation/runs/5e1040a99ba7f51f126601bd84ad5e6b912847d2/generated/jsonschema_gen.go),
and [generated TypeScript](https://github.com/tylergannon/go-gen-jsonschema/blob/dbcc9f21cc6baefbc0ab8f71c35eea56e631e827/ephemeral/codec-ts-validation/runs/5e1040a99ba7f51f126601bd84ad5e6b912847d2/generated/ts/types.ts).

Closes #57.
Closes #58.
