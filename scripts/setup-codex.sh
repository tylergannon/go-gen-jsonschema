set -euo pipefail

apt update && apt install -y curl just

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6

function install_staticcheck() {
    URL="https://github.com/dominikh/go-tools/releases/download/2025.1.1/staticcheck_linux_386.tar.gz"
    DEST="/usr/local/bin"

    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT

    curl -L "$URL" -o "$tmpdir/staticcheck.tar.gz"
    tar -C "$tmpdir" -xzf "$tmpdir/staticcheck.tar.gz"

    # binary is inside the extracted directory (named “staticcheck”)
    install -m 0755 "$tmpdir/staticcheck/staticcheck" "$DEST"
}

install_staticcheck
go install github.com/tylergannon/go-gen-jsonschema/gen-jsonschema@latest
# fmt / imports
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest
go install github.com/tylergannon/go-gen-jsonschema/gen-jsonschema@latest

go mod download all
go mod tidy
go test ./...
