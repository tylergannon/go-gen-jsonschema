# Issue 32 definition-of-done review, round 2 (Fable, 2026-07-10)

## Exact prompt received

> Re-review the revised `docs/design/issue-32-definition-of-done.md` after your
> first findings in `ephemeral/reviews/2026071015-issue-32-dod-fable.md`.
>
> The goal remains to crystallize a concrete, appropriately scoped completion
> contract and genuinely conclusive proof for `Optional[T]` and `Nullable[T]`,
> without adding implementation ceremony.
>
> Check whether both design findings and the four nits from round one are now
> resolved without introducing new ambiguity or scope. Do not review product code
> and do not modify the definition-of-done artifact.
>
> Write the exact prompt you received and findings to:
>
> `ephemeral/reviews/2026071015-issue-32-dod-fable-round2.md`
>
> Use these labels:
>
> - **critical:** must fix before proceeding.
> - **bug:** demonstrable incorrect behavior, broken contract, race, or regression.
> - **design:** architecture, boundary, scope, maintainability, or proof issue
>   materially likely to cause problems.
> - **nit:** small cleanup that should not block progress.
>
> Explicitly state whether consensus has been reached on direction and proof.

## What I checked

The revised DoD in full against the round-one findings, GitHub issue #32
(re-fetched, full body), and the repository at `efab995`. The DoD file is a
new untracked document, so each round-one item was verified directly against
the current text rather than via a diff.

## Resolution of round-one findings

### design 1 — OpenAI proof round-trip: RESOLVED

The Behavioral proof section now reads: "Feed the returned JSON through the
example's generated `ValidateJSON` method and unmarshal it into the
wrapper-bearing Go type; the evidence records both successful steps." This
restores the issue's "returned JSON passes generated validation" leg and adds
the wrapper-decode step round one recommended. The accompanying "Do not
substitute a local schema validator for the live schema-acceptance claim"
sentence hardens the anti-fake-proof stance without adding scope. The
Optional-bearing schema is correctly excluded from strict submission and
proves its contract by visible omission from `required`, matching the issue's
proof bullet exactly.

### design 2 — leftover legacy-tag semantics: RESOLVED

The DoD now states: "A leftover legacy tag is inert: it has no optionality
effect and is not itself a generation error. Existing fixtures and examples
stop relying on it." This is the decided contract from issue decision 5,
verbatim in substance, and it covers the live usages in `examples/test_options`
and the builder fixtures flagged in round one. Unambiguous: the
silently-ignore implementation is the only one that satisfies it.

### nit 1 — tracked session worklog: RESOLVED

Hygiene now includes "The tracked session worklog records decisions, review,
proof, and final state." Matches the issue's completion contract; the added
detail describes existing repo practice, not new ceremony.

### nit 2 — exact-head CI: RESOLVED

Hygiene now includes "Required CI is green for the exact pull-request head
SHA" as an explicit bullet, no longer buried in review checkpoint 2.

### nit 3 — decision 14 integer/float lockdown: RESOLVED (one residual nit below)

Hygiene now includes: "Tests cover every signed and unsigned integer width
plus `byte`, `rune`, and `uintptr` as JSON Schema integers, and both float
widths as numbers with ordinary JSON rejection of NaN and infinities." The
core doubt from round one — whether the exotic integer variants are locked
down at all — is removed.

### nit 4 — decision 16 documentation enumeration and decision 13 manual API: RESOLVED

The documentation bullet now enumerates decision 16's full agreement list:
three contracts, selection guidance, `omitzero`, the missing-Nullable-key
validation caveat, supported and rejected shapes, interface behavior,
strict-schema implications, and usable examples — across README, `llms.txt`,
examples, and the checked-in skill. "The manual schema-building API remains
unchanged" now appears as its own line in Generator acceptance.

## New findings in the revised document

### critical

No findings.

### bug

No findings.

### design

No findings. The revisions are tight, one-sentence amendments; none introduces
new ambiguity, new deliverables, or implementation ceremony. The proof section
remains self-checking and the review-checkpoint section is unchanged and
self-limiting.

### nit

1. The new integer-width test bullet omits decision 14's qualifier "for
   ordinary and wrapped fields." Read in isolation, the bullet is satisfiable
   by ordinary-field tests alone, leaving e.g. `Nullable[uintptr]` →
   `["integer", "null"]` formally unpinned. Decision 13's
   normalize-once-then-ordinary-rendering design makes a per-width wrapper
   divergence unlikely, so this does not block; adding the four words "for
   ordinary and wrapped fields" would complete the match with the issue.

## Consensus

**Consensus is reached on both direction and proof.** Both round-one design
findings and all four nits are resolved without new ambiguity or scope
expansion. The definition of done is now a concrete, standalone completion
contract faithful to issue #32's decisions, and the proof — real-CLI
generation with validation, a compiled proof command that exits non-zero, a
committed deterministic transcript, generator-level failure demonstrations,
and live OpenAI schema acceptance with the returned JSON round-tripped through
generated validation and wrapper decoding — is genuinely conclusive for the
public generator and generated consumer behavior. The single residual nit is
a four-word wording completion and should not gate implementation.
