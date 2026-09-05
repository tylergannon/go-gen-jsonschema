module example.com/codec-ts-consumer

go 1.27

require (
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
	github.com/tylergannon/go-gen-jsonschema v0.0.0
)

require golang.org/x/text v0.14.0 // indirect

replace github.com/tylergannon/go-gen-jsonschema => /Users/tyler/src/.worktrees/go-gen-jsonschema/v1-codec-integration
