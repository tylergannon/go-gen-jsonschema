# v1.0.0-rc.5 installed-tool smoke

This proof is intentionally separate from `ephemeral/adoption-transport-proof`
so the accepted RC4/current-head evidence stays unchanged.

After `v1.0.0-rc.5` is published, run from the repository root:

```sh
./ephemeral/rc5-install-smoke/run-smoke.sh
```

The runner copies `fixture/` to a fresh temporary module, then:

1. installs `github.com/tylergannon/go-gen-jsonschema/gen-jsonschema@v1.0.0-rc.5`
   with `go get -tool`;
2. verifies the runtime library resolves to that same version with no `replace`;
3. invokes the installed tool as `go tool gen-jsonschema` to generate validation,
   JSON Schema, Go codecs, and TypeScript declarations;
4. compiles TypeScript and executes the real Go-to-Node-to-Go-to-Node transport
   assertions; and
5. checks `--no-changes` regeneration and consumer Go tests.

Successful evidence is copied to `evidence/` once. Failed temporary runs are
left on disk and their path is printed for diagnosis. Use `--prepare-only` to
validate the pre-tag fixture without installing anything.

This is a release provenance smoke for one everyday transport case. It does not
expand the issue #71 test matrix.
