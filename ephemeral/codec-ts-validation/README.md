# Fresh codec and TypeScript consumer proof

This fixture is executed only against an exact clean candidate commit. It builds
that candidate's CLI, generates one mixed owner in a disposable consumer module,
tests the real `encoding/json` boundary and generated schema, and compiles the
generated TypeScript with positive and expected-negative assignments.

Run from the validation worktree:

```sh
ephemeral/codec-ts-validation/run-proof.sh \
  /absolute/path/to/candidate \
  FULL_CANDIDATE_SHA \
  ephemeral/codec-ts-validation/runs/FULL_CANDIDATE_SHA
```

The output directory retains candidate/tool provenance, exact stdout and stderr,
the generated Go codec/schema and TypeScript, and a machine-readable result. The
runner rejects a SHA mismatch or tracked candidate changes and verifies that the
candidate and canonical fixture remain unchanged.
