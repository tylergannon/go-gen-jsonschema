test:
    ginkgo ./...

watch focus:
    ginkgo watch --focus "{{focus}}" ./...

testregistry:
    cd internal/typeregistry && ginkgo
