## Problem

Some declaration options that look valid can be silently ignored during schema generation. A confirmed example is `WithEnum(Thing{}.PL)` where `PL` is `*Level`: schema-only generation exits successfully and emits an ordinary integer field without the requested enum, while TypeScript generation rejects the same registration.

Silent omission is dangerous at a serialization boundary because the generated artifact looks authoritative while losing an explicit user instruction.

## 1.0 outcome

Reject recognized declaration options that cannot be applied consistently. Generation must fail clearly before replacing generated artifacts.

## Acceptance

- The confirmed pointer-field `WithEnum` case either gains coherent supported semantics across generated outputs or fails with an actionable source-positioned diagnostic.
- Invalid field-owner and provider associations already recognized by the declaration scanner fail instead of disappearing from generation.
- A failed declaration scan does not replace or partially update generated artifacts.
- Focused regressions exercise the CLI failure path and the confirmed pointer-enum case.
- This issue does not add a new declaration API, structured diagnostic protocol, or support for new Go shapes.

Extracted from #73 so the correctness fix can ship without making the proposed fluent API part of the 1.0 compatibility surface.
