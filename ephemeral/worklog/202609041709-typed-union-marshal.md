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
