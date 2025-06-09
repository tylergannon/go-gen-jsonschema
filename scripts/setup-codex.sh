go install github.com/tylergannon/go-gen-jsonschema/gen-jsonschema@latest

go mod download all
go mod tidy
go test ./...
