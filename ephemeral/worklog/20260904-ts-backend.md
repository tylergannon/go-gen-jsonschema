# TypeScript backend worklog

decision: Project field-local union branches as Omit of the shared payload discriminator intersected with a required exact literal property; this preserves shared declarations while narrowing compatible existing discriminator fields.

decision: Project the empty object constructor as TypeScript object because braces admit primitives; object satisfies the explicit primitive and null exclusion without any or unknown while retaining the documented structural-only limit.

decision: Reserve Array and Omit in exported name allocation because generated expressions depend on those global helper types; colliding Go definitions receive safe renamed declarations.

decision: Reject integer enum literals unless go constant conversion proves exact float64 representation, then print the original exact decimal constant rather than a rounded float spelling.

friction: Initial integer test constants used token ILLEGAL and panicked before exercising generation -> construct test integers with token.INT so failures reach the backend.

proof: Worktree-wide pre-change go test ./... was reported successful by the coordinator; focused post-change go test ./internal/typescript passes.

proof: A retained direct-Generate probe under ephemeral/typescript-generation/backend-compiler passes the pinned tests/typescript TypeScript compiler with strict, exactOptionalPropertyTypes, and noUncheckedIndexedAccess; it covers an empty graph and barrel, exact versus rejected inexact large integers, empty-tag narrowing over a colliding payload property, encoded Unicode name collisions, hostile comments and property keys, and primitive/null rejection for empty objects.

decision: Promote backend-only edge coverage into tests/typescript/projection with a Go executable that accepts its output directory and a separately committed strict TypeScript consumer; CI can now regenerate these grammar-only cases instead of relying on the retained snapshot.

proof: go run ./tests/typescript/projection/generate ./tests/typescript/projection/generated and the pinned tsc project tests/typescript/projection/tsconfig.json both pass; the generated directory is ignored while the original successful ephemeral output remains retained.
