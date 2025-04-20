test:
    ginkgo ./...

watch focus:
    ginkgo watch --focus "{{focus}}" ./...

testregistry:
    cd internal/typeregistry && ginkgo


lint:
    golangci-lint run ./...


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
