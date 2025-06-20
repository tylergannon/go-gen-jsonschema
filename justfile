test:
    ginkgo ./...

watch focus:
    ginkgo watch --focus "{{focus}}" ./...

testregistry:
    cd internal/typeregistry && ginkgo

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
