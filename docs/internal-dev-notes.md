---
applyTo: '**'
---

# go-gen-jsonschema Internal Developer Notes

Owner: opencode

Last updated: 2025-08-09

Changelog
- Added internal/common/tags.go: centralized JSONSchema tag parsing (optional, ref, param plan groundwork).
- Added builder pipeline hook (Transform) with RegisterTransform and applyTransforms in Run (no-op default). Tests remain green.


This is an internal, living document for deep understanding, navigation, and refactoring of go-gen-jsonschema. It is deliberately exhaustive. Keep this up to date.

## 0) Quick facts
- Purpose: Generate JSON Schemas from Go types for LLM tool/function calls.
- Outputs: jsonschema/*.json (+ *.sum) and jsonschema_gen.go with method impls + custom Unmarshalers.
- Build tag flow: `//go:build jsonschema` gates stub methods/markers; generator loads package with `-tags=jsonschema`.
- Entry points:
  - CLI: gen-jsonschema/gen
  - Core: internal/syntax (scan) → internal/builder (render) → files + gen code.
- Supported features: structs, primitives, arrays, enums, union types via interfaces, refs via struct tag, optional fields, description tags/comments.

## 1) Architecture map
- gen-jsonschema (CLI)
  - main.go: subcommands gen (default), new. Calls builder.Run(...)
  - tmpl/config.go.tmpl: used by `new` to scaffold.
- internal/syntax (package loader, AST scanner, type resolver)
  - loader.go: load packages with `-tags=jsonschema`.
  - scan_result.go: central scan pipeline → collects markers, local types, interfaces, enums, resolves types, tracks remote deps.
  - scan_expr.go: parse marker calls and schema method descriptors from AST.
  - node_wrappers.go: rich wrappers over dst AST (TypeSpec, StructType, StructField, etc.). Tag parsing, Required/Skip logic, PropNames.
- internal/builder (schema nodes + file/code generation)
  - gen_schema.go: SchemaBuilder: map types → internal schema nodes → write json files → render jsonschema_gen.go.
  - model.go: internal schema JSON model: ObjectNode, PropertyNode[T], ArrayNode, UnionTypeNode, RefNode. Custom MarshalJSON.
  - schemas.go.tmpl: generated code template; embeds jsonschema dir; emits method impls and interface unmarshaler helpers.
  - import_map.go, printer.go: template/goimports helpers.
- Public helpers (api) in root
  - json_schema.go: separate JSONSchema type and helper builders (StringSchema, ArraySchema, EnumSchema, ParentSchema...). Intended for manual schema construction.
  - union_type.go: marker functions: NewJSONSchemaMethod, NewEnumType, NewInterfaceImpl, NewJSONSchemaBuilder.

## 2) End-to-end flow
1. User writes:
   - schema.go (under build tag) + method stubs returning json.RawMessage
   - marker calls: NewJSONSchemaMethod, NewEnumType, NewInterfaceImpl, etc.
   - go:generate directive to run the generator.
2. gen-jsonschema gen: loads package with jsonschema tag; internal/syntax finds markers, types, enums, interfaces; resolves types recursively (local + remote) and enforces invariants.
3. internal/builder maps each registered type into internal schema nodes.
   - Primitives → PropertyNode
   - Arrays → ArrayNode
   - Structs → ObjectNode with Properties and Required (default required unless jsonschema:"optional")
   - Interfaces → UnionTypeNode(anyOf). Discriminator property injected when serializing union.
   - Ref via tag jsonschema:"ref=..." → RefNode
4. Writes jsonschema/<Type>.json (+ <Type>.json.sum checksum).
5. Writes jsonschema_gen.go (excluded from jsonschema build tag) with:
   - func (T) SchemaMethodName() json.RawMessage { read embed }
   - custom json.Unmarshaler for structs with interface-typed fields → reads discriminator and dispatches to impl type.

## 3) Key types (internal schema model)
- ObjectNode: Desc, Properties(ObjectPropSet = []ObjectProp{Name, Schema, Optional}), Discriminator (string), TypeID_. MarshalJSON: emits type:object, description, properties, required (computed), additionalProperties:false.
- PropertyNode[T]: Desc, Enum, Const, Typ (string), TypeID_. MarshalJSON: type, description, const (if set), enum (if set).
- ArrayNode: Desc, Items(JSONSchema), TypeID_.
- UnionTypeNode: Options []ObjectNode (each an object schema). MarshalJSON: { anyOf: [ object-with-discriminator, ...] }, discriminator property name defaults to `!type`.
- RefNode: emits {"$ref": "..."}.

## 4) Tag semantics
- json: standard behavior for names, skipping ("-").
- jsonschema:"optional": mark property as not required.
- jsonschema:"ref=...": replace field schema with $ref (field skipped from traversal).
- description:"...": overrides comment-sourced description for the field.

## 5) Interface/union semantics
- Register with NewInterfaceImpl[YourInterface](Impl1{}, Impl2{}, (*Impl3)(nil))
- Scanning records the interface and option types (pointer/value).
- JSON Schema for interface is anyOf of option object schemas with required discriminator `!type` const equal to the type name.
- Generated code: per-interface unmarshal helper switching on discriminator; per-struct UnmarshalJSON for fields that are interface-typed in local structs.

## 6) Strengths
- Clear split: scanning (syntax) vs building (builder) vs CLI.
- Deterministic generation with checksum guard and --no-changes.
- Build-tag strategy isolates stubs/markers from normal builds.
- Practical interface union handling with discriminator and generated Unmarshalers.
- Good unit/integration coverage using test fixtures and golden files.

## 7) Weaknesses / Over-complications
- Two schema models: public (json_schema.go) vs internal (internal/builder/model.go). Divergent behavior and duplication increase cognitive load.
- Lots of custom string-based MarshalJSON code; manual JSON assembly increases maintenance risk vs using structs + encoding/json consistently.
- AST wrappers add indirection and learning curve; some logic (e.g., skipping, tags) is split across syntax and builder.
- Interface field handling restrictions (must be at field type position) are implicit; errors occur deep in traversal when violated.
- Discriminator injection happens at union serialization time (hidden coupling).
- CLI option NumTestSamples currently unused in code path.
- AdditionalProperties hard-coded to false in internal ObjectNode; less flexible than public helper API.

## 8) Targeted refactor suggestions
1) Unify schema model
- Option A: make builder/model use the public JSONSchema type from json_schema.go and delete the internal model. Extend public type with bits we need (e.g., discriminator helper/util).
- Option B: remove public helpers and expose a thin adapter; pick one canonical model.
- Benefit: single mental model; easier extension (e.g., parameterization feature).

2) Reduce custom string-building
- Replace manual MarshalJSON with plain struct forms and encoding/json where practical. Keep a few custom cases (discriminator prepend) minimal.

3) Centralize tag parsing
- Move all struct tag evaluation to a single utility and pass resolved attributes (optional, ref, description, param) forward. Avoid duplicating parsing between syntax and builder.

4) Make union discriminator explicit in schema nodes
- Add DiscriminatorPropName to UnionTypeNode or a SchemaOptions context passed during render to make the behavior less magical.

5) Improve method signature capture
- Extend scanning to capture schema method signatures now (needed for parameterization); use it to validate they return json.RawMessage and to reproduce signature in generated code.

6) Prune unused features
- Remove or implement NumTestSamples; keep surface area minimal.

7) Testing granularity
- Add unit tests for small pieces (renderStructField rules, tag parsing). Fewer golden surprises.

## 9) Parameterized fields plan (preview)
- Tag: `jsonschema:"param=Name[,idx=N][,optional]"`
- Stub method: can accept parameters (typed as json.Marshaler) in a defined order (by idx else by discovery order).
- Generation:
  - If any parameterized fields exist for a type, write jsonschema/<Type>.json as a text/template with `{{.Name}}` placeholders where field schemas go.
  - Generated method will read template from embed, render it with provided args (MarshalJSON → string), and return json.RawMessage.
- Scanner: relax NewJSONSchemaMethod(fn any), capture the actual signature from FuncDecl, enforce return type json.RawMessage.
- Back-compat: unchanged for non-parameterized types.

## 10) Open questions
- Allow parameterization of interface fields (overriding union)? Initially no; keep union machinery.
- Allow combining `ref` and `param`? No; conflict. Emit friendly error.
- Pretty-print template output? Keep as-is; template preserves formatting.

## 11) Working notes
- Discriminator default name = `!type` (DefaultDiscriminatorPropName). Template has access to this in generated code (schemas.go.tmpl uses `{{$discriminatorProp}}`).
- Interface unmarshaler function name format: `__jsonUnmarshal__<pkgName>__<TypeName>`.

## 12) TODO (engineering roadmap)
- [ ] Decide on schema model unification (public vs internal) and implement.
- [ ] Implement parameterized fields (tags, scanner, template emission, method signature capture, codegen runtime rendering).
- [ ] Add focused unit tests for tag parsing and field rendering logic.
- [ ] Review and address CLI option drift (NumTestSamples).
- [ ] Document parameterized fields in README.
