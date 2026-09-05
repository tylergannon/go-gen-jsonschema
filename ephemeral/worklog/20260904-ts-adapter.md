# Type grammar source adapter worklog

decision: TypeDefinitions lowers directly from the scanner and field registrations, preserves named dependencies as Ref definitions, and validates the completed graph before returning it.
decision: Current source registration supports WithEnum and WithStringerEnum; WithStringerEnum becomes field-local EnumNames only for integer enums, while string enums retain exact underlying values.
decision: Provider fields, explicit schema refs, json string coercion, and custom JSON or text codec methods fail lowering because the accepted static grammar has no resolved wire shape for them.
decision: Ordinary omitempty and omitzero fields remain Required in the structural grammar; those tags constrain the valid Go value domain and do not express Optional.
friction: syntax.TypeID.Concrete does not clear indirection on its returned copy -> the adapter constructs typegrammar.Name explicitly and never uses Concrete for identity normalization.
decision: Promoted fields are selected with Go encoding/json depth and explicit-tag dominance; outer fields shadow promoted fields, ambiguous same-depth fields disappear, and surviving fields retain source index order.
decision: Enum members are enumerated from go/types package-scope constants of the exact named type in source order, avoiding the old renderer's iota/index and literal-trimming assumptions.
correction: A grouped declaration can contain unexported names before exported names; pair JSON names with filtered exported Go names rather than indexing the original declaration list.
proof: go test ./internal/builder -run '^TestTypeDefinitions' -count=1 passed with source lowering and admission tests.
proof: go vet ./internal/builder passed after the final adapter changes.
correction: Consensus round 01 reproduced contradictory enum sets because schema rendering used AST inheritance/literal heuristics while TypeScript used go/types -> enum membership and evaluated values now come from syntax.ResolveEnum for both artifacts.
decision: ResolveEnum includes package constants whose exact typed identity matches the defined enum, including conversion-typed and computed constants, and excludes untyped constants; aliases to defined enums resolve through types.Unalias while aliases directly to predeclared types fail because their membership cannot be distinguished from unrelated constants.
decision: Numeric JSON Schema enum values use json.Number built from exact Go integer constants, preserving full-width signed and unsigned literals without int or float64 narrowing; string-mode integer fields use constant names and reject duplicate underlying values.
proof: focused regression `go test ./internal/builder -run '^TestTypeDefinitions' -count=1` passed after shared enum resolution; it compares schema and grammar values for global, WithEnum and WithStringerEnum paths and checks a MaxUint64 schema literal.
proof: go vet ./internal/syntax ./internal/builder passed after the shared enum fix.
correction: Consensus round 02 found the shared renderer unintentionally added field and member descriptions to existing field-local WithEnum and WithStringerEnum schema properties -> field-local calls now explicitly suppress all descriptions while global NewEnumType rendering retains type and member descriptions.
proof: focused source/schema regression distinguishes description-bearing global enum output from description-free field-local output; the focused builder test, builder/syntax vet, and diff check passed.
