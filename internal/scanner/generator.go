package scanner

import "golang.org/x/tools/go/packages"

type Generator struct {
	TypeNames   []string
	LocalPkg    *packages.Package
	RemoteTypes map[string]Generator
}
