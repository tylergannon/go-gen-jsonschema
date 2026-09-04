The CLI accepted `--num-test-samples` without producing samples, and shipped references mixed supported behavior with unfinished experiments. Remove that flag and its unused internal fields, delete the disabled Tool/BuildTool experiment, and reconcile the README, agent references, tutorial, provider documentation, and historical plans with the accepted v1 contract.

The existing provider API is retained. The documentation separates schema generation from codecs, and the builder entrypoint test now checks its schema golden. Older optionality proposals #17 and #28 are superseded by #32; #10 stays open because its remaining acceptance has not yet been demonstrated.

Validation: `GOFLAGS=-p=2 go test ./...`, focused entrypoint fixture test, built CLI help/removed-flag exit checks, and website `npm ci && npm run check` all passed. The support inventory and proof are under `ephemeral/issue-63-cleanup/`.

Closes #63.
