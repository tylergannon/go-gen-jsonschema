package typescanner

import "github.com/tylergannon/go-gen-jsonschema/internal/scanner/testfixtures/typescanner/scannersubpkg"

var (
	_ = LocalFuncOneTypeArg(1)

	_ = LocalFuncOneTypeArg[scannersubpkg.Type001](scannersubpkg.Type001{})

	_ = LocalFuncTwoTypeArg[
		scannersubpkg.Type001, scannersubpkg.Type002
	]()

	_ = LocalFuncThreeTypeArg(1, 2, "")

	_ = LocalFuncThreeTypeArg[
		scannersubpkg.Type001, scannersubpkg.Type002, scannersubpkg.Type003
	](
		scannersubpkg.Type001{},
		scannersubpkg.Type002{},
		scannersubpkg.Type003{},
	)

	_ = scannersubpkg.RemoteFuncOneTypeArg("")

	_ = scannersubpkg.RemoteFuncOneTypeArg[
		scannersubpkg.Type001
	](
		scannersubpkg.Type001{},
	)
)

var (
	_ = scannersubpkg.RemoteFuncTwoTypeArg[
		scannersubpkg.Type001,
		scannersubpkg.Type002,
	]()

	_ = scannersubpkg.RemoteFuncThreeTypeArg[
		scannersubpkg.Type001,
		scannersubpkg.Type002,
		scannersubpkg.Type003,
	](
		scannersubpkg.Type001{},
		scannersubpkg.Type002{},
		scannersubpkg.Type003{},
	)

	_ = scannersubpkg.RemoteFuncThreeTypeArg(
		scannersubpkg.Type001{},
		scannersubpkg.Type002{},
		scannersubpkg.Type003{},
	)

	_ = scannersubpkg.RemoteFuncThreeTypeArg[
		*scannersubpkg.Type001,
		*scannersubpkg.Type002,
		*scannersubpkg.Type003,
	](
		&scannersubpkg.Type001{},
		(*scannersubpkg.Type002)(nil),
		(*scannersubpkg.Type003)(nil),
	)

	_ = scannersubpkg.RemoteFuncThreeTypeArg(
		&scannersubpkg.Type001{},
		(*scannersubpkg.Type002)(nil),
		(*scannersubpkg.Type003)(nil),
	)
)
