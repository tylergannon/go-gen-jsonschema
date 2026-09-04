## Outcome

Publish v1.0.0 only when the stable product contract, installable distribution, and primary agent workflows have inspectable evidence.

## Acceptance and proof

- Obtain the maintainer's intended license choice and add the license and required notices; the assessed tree contains no tracked LICENSE/COPYING file.
- Publish concise compatibility/support and migration guidance covering Go 1.27+, removal of the old optional tag, Optional/Nullable semantics, generated union/enum codecs, custom-method ownership, and YAML dependency/support policy.
- Document the supported Go/platform matrix and validate the chosen matrix in CI; do not silently equate one green platform with universal portability.
- Cold-cache prerequisite #36 is satisfied: PR #38 fixed the stderr assertion; an initially empty module-cache `go test -count=1 -v ./...` run passed on 2026-09-04. Retain clean-cache coverage in release verification.
- Produce a release candidate after milestone implementation; install it in fresh consumer modules without local replace directives, generate JSON/YAML support, execute roundtrip/validation proof, and demonstrate clean regeneration/no-change checks.
- Record successful product conformance and real-agent evaluation results for the exact candidate commit. Reconcile every milestone issue with evidence before tagging stable.
- Publish the stable tag and GitHub release notes from the proved commit; verify installation of the tag and plugin version alignment.

Depends on #62, #63, and #65; #33/#34 are prerequisites through those dependencies. #36 is already satisfied. Other issues in this milestone are release prerequisites through these dependencies.

App-server assistance, streaming runtime APIs, and broad new type support are deferred beyond 1.0; their implementation is not a release criterion.