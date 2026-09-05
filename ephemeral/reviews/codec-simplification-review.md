# Codec simplification review

Review target: `489851ce988441980cb2c49be14e33fddf14cd33` on `codex/issue-58-enum-codec`.

## Contract reviewed

The accepted architecture keeps one plain defined owner alias inside one wrapper, shadows only registered union and adapted enum fields with `json.RawMessage`, delegates ordinary fields to `encoding/json`, emits sealed-union helpers as closed type switches, and rejects competing direct or promoted production JSON methods. The bounded simplification in `ephemeral/reviews/codec-simplicity-audit.md` requires fresh render-only template projections and removal of metadata with no remaining consumer, while preserving emitted behavior.

## Evidence inspected

- Full commit diff for `489851c` and the pre-change implementation at `489851c^`.
- `internal/builder/gen_schema.go:476-490`, where `schemaTemplateData` now owns the render-only imports, owner codecs, YAML types, and interface helpers.
- `internal/builder/gen_schema.go:1400-1498`, where each `RenderGoCode` call constructs and renders a new `schemaTemplateData` instead of appending projections to `SchemaBuilder`.
- `internal/builder/schemas.go.tmpl:164-336`, where the parent alias and wrapper preserve ordinary `encoding/json` handling and shadow only special fields; `internal/builder/schemas.go.tmpl:397-464`, where registered implementations remain closed marshal/unmarshal switches; and `internal/builder/schemas.go.tmpl:468-492`, where the helper remains limited to discriminator validation and insertion.
- `internal/builder/gen_schema.go:492-661` and `internal/builder/owner_codec_test.go:14-153`, which retain direct, promoted, embedded generated-owner, and foreign generated-owner collision rejection.
- Repository-wide searches for every removed field and helper (`FuncNameAlias`, enum `StringValue` and `Underlying`, interface option/type package metadata, `IsSpecialType`, and `sortedCustomTypeNames`) found no surviving production consumer. Remaining similarly named values have distinct live uses.
- `internal/builder/gen_schema_determinism_test.go:13-46`. It builds the mixed union/enum fixture with YAML enabled, renders twice through the same builder, requires byte-identical source, and checks that the owner marshal, unmarshal, and YAML methods each occur once. This is a meaningful regression for the former append-on-every-render defect: the old implementation accumulated owner codecs, interface helpers, and YAML types between those two calls.
- The commit changes no generated source, golden, or checksum artifact. The implementation worklog records two full regenerations with an unchanged generated-corpus hash and a final uncached full suite. This review independently ran `GOFLAGS=-p=2 go test ./...` successfully and reran the repeated-render, enum rejection, and owner collision cases uncached successfully.

## Findings

No material or minor findings. The removed data is dead, the new projection lifetime is confined to one render, the emitted codec architecture is unchanged, and the regression exercises the state reuse that previously produced duplicate declarations.

## Outcome

no findings
