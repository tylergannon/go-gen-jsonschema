The old v1 draft contradicted the current release direction and mixed schema generation with codec support. This replaces it with an explicit Go/JSON capability matrix and the contract needed to implement union and enum roundtrips.

It defines semantic equality and valid-value limits; shared field-owned generated methods; nil and discriminator behavior; enum alias/collision rules; custom-hook failure stages; canonical YAML input; helper strict/error policy; and compatibility boundaries. Existing evidence, linked implementation requirements, unknown cases and exclusions are distinguished.

The draft has received an independent source review. Its findings were addressed by specifying legacy discriminator collision handling, concrete custom-hook examples, and the closed supported interface forms. All local document links resolve.

**Approved provider decision:** retain the current provider API for 1.0 and defer additive typed props beyond 1.0. The maintainer approved this direction on 2026-09-04.

The milestone now records the parallel execution plan, and native issue dependencies record completion gates. The helper-specific decisions have been resolved, so #59 can proceed independently. #33/#34 were implemented and merged separately by other agents in #68/#67.

Closes #56. This is a target contract, not a claim that the remaining milestone implementation is complete.
