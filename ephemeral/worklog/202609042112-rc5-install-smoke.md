# RC5 installed-tool smoke

decision: Preserve the accepted RC4/current-head proof byte-for-byte and prepare a separate fresh-consumer smoke for v1.0.0-rc.5.
decision: The smoke must obtain both the generator tool and runtime library from the tagged module, invoke generation through `go tool gen-jsonschema`, reject replace directives, and retain evidence only after all Go/Node assertions pass.
decision: Do not execute the network/tag-dependent install until the release manager announces that v1.0.0-rc.5 is available.
proof: After the release signal, the public annotated v1.0.0-rc.5 tag resolved to object db33bee926d6fda42b447a1b28679400dc1c0342 and commit 8d83354421e2f37ea42bfaccdcbd1ff9211715a2.
proof: Fresh `go get -tool` resolved both the generator build and consumer runtime to github.com/tylergannon/go-gen-jsonschema@v1.0.0-rc.5 with no replace; `go tool gen-jsonschema` generated validation, schema, Go codecs, and TypeScript declarations.
proof: TypeScript 6.0.3 compiled, Node v26.8.1 and Go 1.27.1 completed the Go-to-Node-to-Go-to-Node semantic assertions, `--no-changes` passed, consumer Go tests passed, and the canonical fixture hash remained unchanged.
decision: Closeout remains on `codex/adoption-transport-proof`; the release manager explicitly authorized pushing this proof branch for an immutable public evidence link, with no PR, merge, or tag action.
