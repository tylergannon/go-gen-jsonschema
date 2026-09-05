# Grammar test evidence

The external-package tests construct definition graphs through the exported Go API. They establish admission and rejection at `Definitions.Validate`, without claiming that the current CLI already lowers source into this grammar.

Coverage includes finite shared graphs versus named-reference, union-implementation and in-memory cycles; all scalar kinds; required/Optional/Nullable forms; field-only unions and named ownership; exact tags (including explicit empty tags), discriminator compatibility, pointer/value registration identity; enum modes, exact integer bounds and contextual adaptation; missing references, nil/foreign constructors, invalid names, and source-aware diagnostics. The executable example constructs a discriminated envelope through this API.

Source review corrected two draft assumptions: explicit union tags may be empty, and numeric-value enums compose in ordinary containers while constant-name string adaptation remains restricted to direct fields of named owners. Tests preserve these distinctions, including named owners nested under collection references versus anonymous inline objects. Byte-like slices are rejected even when their uint8 element is reached through references or a numeric enum registration.

The tests assert accepted/rejected behavior and actionable diagnostic structure. They do not compare validator internals, promise full JSON validation, or establish a TypeScript or codec transport proof.
