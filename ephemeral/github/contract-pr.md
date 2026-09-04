The old v1 draft contradicted the current release direction and mixed schema generation with codec support. This replaces it with an explicit Go/JSON capability matrix and the contract needed to implement union and enum roundtrips.

It defines semantic equality and valid-value limits; shared field-owned generated methods; nil and discriminator behavior; enum alias/collision rules; custom-hook failure stages; canonical YAML input; helper strict/error policy; and compatibility boundaries. Existing evidence, linked implementation requirements, unknown cases and exclusions are distinguished.

The draft has received an independent source review. Its findings were addressed by specifying legacy discriminator collision handling, concrete custom-hook examples, and the closed supported interface forms. All local document links resolve.

**Open maintainer decision:** retain the current provider API for 1.0, replace it with #27's typed props, or remove parameterized schemas before 1.0. The recommendation is to retain it and defer additive props. This PR stays draft and #56 stays open until that decision is made.

The milestone now records the parallel execution plan, and native issue dependencies record completion gates. The helper-specific decisions have been resolved, so #59 can proceed independently. #33/#34 were implemented and merged separately by other agents in #68/#67.

Related to #56. This is a target contract, not a claim that the remaining milestone implementation is complete.
