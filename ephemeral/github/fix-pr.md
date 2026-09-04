Manual schema helpers could emit incorrect or invalid schemas: named integer values were labeled as strings, object keys were not JSON-escaped, and strict required ordering depended on map iteration. This change fixes those cases and makes strict/error behavior intentional.

- Named integer/string constants and enums get the correct schema type; object property names are JSON-escaped; map-backed strict required keys have stable order.
- Both strict object builders require every declared property and forbid extras, overriding explicit settings. Non-strict settings remain effective.
- Empty EnumSchema construction returns a node that fails descriptively at marshaling, preserving the existing signature.

Proof is in committed public-API regression tests: generated schema compilation and instance acceptance/rejection for named values, escaped-key roundtrips, repeated serialization, strict overrides and empty-enum errors.

Closes #59.
