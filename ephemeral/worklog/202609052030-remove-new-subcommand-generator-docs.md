# Remove `new` subcommand; document driving polytype from another generator

decision: Delete the `new` scaffold subcommand (user, 2026-09-05). The tagged
file is three lines per type; agents and other generators hand-write it.
decision: Enum markers and sealing methods are always author-written in
untagged source. Foreign generators emit only the tagged file and run the
CLI; they never edit type source.
decision: One tagged file per package, one owner. No mixed-ownership story.
correction: Keep assessments short and do not invent edge cases (two tagged
files in one package) that nobody asked about.
friction: `just lint` runs `goimports -w` over every .go file, including
ephemeral/ proof consumers, and dirtied 20+ unrelated files -> scope the find
to tracked source or exclude ephemeral/.
doc_bug: website getting-started.mdx and docs/tutorial.mdx still referenced
`.Interface` and `NewJSONSchemaMethod` after #93 -> fixed in passing.
proof: go test ./... ok; just build-tagged ok; just lint 0 issues; grep for
`polytype new` / `-methods` / config.go.tmpl returns nothing outside ephemeral/.
