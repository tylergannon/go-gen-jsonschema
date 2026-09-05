# Adoption transport proof

decision: Keep the proof to one ordinary shipment owner with one union field, one string-mode enum, one Optional string, and one Nullable timestamp; this demonstrates the requested adoption path without expanding into the issue #71 edge matrix.
decision: Pin the consumer library to published v1.0.0-rc.4 with no replace directive, while building the generator binary from the current repository head; record both sources in proof provenance.
decision: Execute TypeScript-produced JavaScript under Node so the evidence covers runtime JSON transport, not only declaration compilation.
friction: A bare `go mod download` recorded only the dependency's go.mod checksum, so the generator importer reported an invalid imported package name -> run `go mod tidy` before first generation to prime the external package checksum, matching a fresh consumer's normal setup.
proof: `./ephemeral/adoption-transport-proof/run-proof.sh` generated Go, schema, and TypeScript together; TypeScript 6.0.3 compiled; Node v26.8.1 and Go 1.27.1 completed the Go-to-Node-to-Go-to-Node semantic assertions; `--no-changes` confirmed deterministic regeneration.
proof: Proven generator source commit is 4374753202591d6b508ecb6f4fbe85f818d0f237; consumer dependency is github.com/tylergannon/go-gen-jsonschema@v1.0.0-rc.4 with no replace directive.
decision: Closeout branch is `codex/adoption-transport-proof`; this branch contains proof artifacts only, has not been pushed, and has not received a separate review because the root manager will integrate it with the release preparation.
