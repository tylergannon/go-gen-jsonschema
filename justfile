test:
    go test ./...

watch focus:
    # no ginkgo; use `go test` with -run for focus
    go test ./... -run '{{focus}}'

testregistry:
    cd internal/typeregistry && go test ./...

lint:
    go mod tidy
    modernize -fix ./...
    go vet ./...
    staticcheck ./...
    govulncheck ./...
    golangci-lint run ./...
    find . -name '*.go' -exec goimports -w {} \;

update-deps:
    #!/usr/bin/env sh
    for f in $(fd go.mod); do
        pushd $(dirname $f)
        go mod tidy
        popd
    done

upgrade-module mod-path:
    #!/usr/bin/env sh
    for f in $(fd go.mod); do
        pushd $(dirname $f)
        go get -u $mod
        go mod tidy
        popd
    done
