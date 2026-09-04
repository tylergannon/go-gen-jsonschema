# Typed-union MarshalJSON worklog

decision: Implement issue #57 against the accepted v1 contract at commit 991202c; field-specific registration is the single source for schema, encoding, and decoding.
decision: Compose all generated field adapters through one owner-codec model so each owner has exactly one value-receiver MarshalJSON and one pointer-receiver UnmarshalJSON, with room for issue #58 enum fields.
decision: Concrete union hooks retain ownership of their payload bytes; generated encoding only verifies an object payload and adds or validates the registered discriminator.
decision: Generated owner decode success replaces omitted fields from a fresh value; any decode error leaves the receiver unchanged.
decision: Reject an owner that embeds another type requiring generated owner codecs; embedding would promote the inner MarshalJSON through the method-free alias wrapper and bypass field overrides. Named nested owner fields remain supported and invoke their generated codec once.
decision: Production collision discovery evaluates default production files plus custom GOFLAGS tags, ignores generation-only stubs and generator-owned output, and combines direct syntax discovery with effective go/types method sets for promoted methods.

friction: An idempotence check initially ran before the fixture module was tidied after validation imports were generated, so Go requested go.mod updates -> run the second generation after the existing post-generation go mod tidy step.

proof: Baseline `GOFLAGS=-p=2 go test ./...` passed at 71ceca8 before edits.
proof: Dedicated generated consumer tests pass through internal/builder TestBasic, including encode, schema validation, decode, semantic checks, pointer/value calls, field-specific discriminator variants, escaping, custom hooks, nil/error cases, and a byte-identical second generation.
proof: Focused owner collision and duplicate legacy discriminator tests pass and preserve pre-existing sentinel artifacts on failure.
proof: After merging origin/main 7991124, `GOFLAGS=-p=2 go test ./...` passed at 616a2eb, including the updated optionality transcript and sealed-interface slice example.

final_state: Branch `codex/issue-57-union-marshal` contains the selected-root constructor dependency aff9a36, production JSON method discovery ebb76d7, typed-union codec implementation a0f9f9b, and main integration 616a2eb. No push, PR, merge to main, or issue mutation was performed.

parent integration: aligned README, llms, tutorial, website, registration reference and shipped skill with owner-based union encoding; removed obsolete global discriminator-marshaler guidance. Updated accepted-contract evidence and documented promoted generated-owner rejection. Full Go suite passed after initial documentation edits; website npm ci and npm run check passed. Website API regeneration also refreshes the already-merged helper semantics. Final website interface guide correction is included in the next check.

review convergence: independent review found ordinary pointer-only MarshalJSON hooks lost through a non-addressable value wrapper, helper collisions for same-named interfaces from same-named packages, and JSON null accepted for a registered empty discriminator. The owner now marshals an addressable wrapper, helper reuse keys the exact resolved interface/configuration and emits a collision-safe identity hash, import aliases are emitted consistently, and discriminator decoding distinguishes null from a valid empty string.
proof: Runtime fixture verifies pointer-only ordinary hook bytes and one call for value/pointer owner encoding, valid empty-string discriminator encode/schema/decode, null rejection on encode/decode, transactional replacement after schema-valid manual input, and generated artifact idempotence. A generated temporary consumer with two package-path-distinct `events.Event` types round-trips both legacy fields and compiles with distinct helper/import identities. `GOFLAGS=-p=2 go test ./...` passed after the review fixes.

parent proof: rebuilt candidate at 6db8115 and ran three independent consumer modules against published rc.3 markers. General union roundtrip, ordinary pointer-hook custom bytes/semantic equality, and rejection of null with empty discriminator all passed. Repeated generation preserved every file byte-identically in all three modules. Website build/link check passed after final guide alignment. Independent review round 2 remains required before merge.
