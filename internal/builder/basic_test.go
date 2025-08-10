package builder_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/testutils"
)

type testCase struct {
	inputDir  string
	testName  string
	runGinkgo bool
	files     []string
}

func TestBasic(t *testing.T) {
	t.Parallel()
	CmdSuccessAssertions := func(t *testing.T, stdout, stderr string, exitCode int) {
		require.Empty(t, stderr)
		require.Equal(t, 0, exitCode)
	}

	CodegenTest := func(tc testCase) {
		cwd, err := os.Getwd()
		require.NoError(t, err)
		tempDir := filepath.Join(cwd, "test_run", tc.testName)

		require.NoError(t, os.RemoveAll(tempDir))
		require.NoError(t, os.MkdirAll(tempDir, 0755))
		inputPathFull := filepath.Clean(filepath.Join(cwd, "..", tc.inputDir))
		fi, err := os.Stat(inputPathFull)
		require.NoError(t, err)
		require.True(t, fi.IsDir())
		require.NoError(t, testutils.CopyDir(inputPathFull, tempDir))

		// Ensure the temp module's dependencies are tidy before generation.
		preExit, preStdout, preStderr, err := testutils.RunCommand("go", tempDir, "mod", "tidy")
		require.NoError(t, err)
		CmdSuccessAssertions(t, preStdout, preStderr, preExit)

		exitCode, stdout, stderr, err := testutils.RunCommand("go", tempDir, "generate", "./...")
		require.NoError(t, err)
		// Some Go versions emit a hint to run 'go mod tidy' on fresh copies; handle that gracefully.
		if exitCode == 0 && strings.Contains(stderr, "go: updates to go.mod needed") {
			_, _, _, _ = testutils.RunCommand("go", tempDir, "mod", "tidy")
			// Re-run generate to get clean stderr
			exitCode, stdout, stderr, err = testutils.RunCommand("go", tempDir, "generate", "./...")
			require.NoError(t, err)
		}
		CmdSuccessAssertions(t, stdout, stderr, exitCode)

		for _, fname := range tc.files {
			fpath := filepath.Clean(filepath.Join(tempDir, fname))
			info, err := os.Stat(fpath)
			require.NoError(t, err, fmt.Sprintf("Expected file %s to be created in %s", fname, tempDir))
			require.False(t, info.IsDir())
			testutils.AssertGoldenFile(t, fpath, ".golden")
		}

		// Ensure generated code compiles in the temp module.
		buildExit, buildStdout, buildStderr, err := testutils.RunCommand("go", tempDir, "build", "./...")
		require.NoError(t, err)
		CmdSuccessAssertions(t, buildStdout, buildStderr, buildExit)

		// No ginkgo; all tests run via `go test ./...` at repo root.
	}

	cases := []testCase{
		{
			inputDir:  "builder/testfixtures/basictypes",
			testName:  "test1-basictypes",
			runGinkgo: false,
			files: []string{
				"jsonschema/TypeInItsOwnDecl.json",
				"jsonschema/TypeInNestedDecl.json",
				"jsonschema/TypeInSharedDecl.json",
			},
		},
		{
			inputDir:  "builder/testfixtures/indirecttypes",
			testName:  "test2-indirecttypes",
			runGinkgo: false,
			files: []string{
				"jsonschema/DefinedAsNamedType.json",
				"jsonschema/DefinedAsPointerToRemoteSliceType.json",
				"jsonschema/DefinedAsRemoteSliceType.json",
				"jsonschema/DefinedAsRemoteType.json",
				"jsonschema/DefinedAsSliceOfRemoteSliceType.json",
				"jsonschema/IntType.json",
				"jsonschema/NamedNamedSliceType.json",
				"jsonschema/NamedSliceType.json",
				"jsonschema/PointerToIntType.json",
				"jsonschema/PointerToNamedType.json",
				"jsonschema/PointerToRemoteType.json",
				"jsonschema/SliceOfNamedNamedSliceType.json",
				"jsonschema/SliceOfNamedType.json",
				"jsonschema/SliceOfPointerToInt.json",
				"jsonschema/SliceOfPointerToNamedType.json",
			},
		},
		{
			inputDir:  "builder/testfixtures/enums",
			testName:  "test3-enums",
			runGinkgo: false,
			files: []string{
				"jsonschema/EnumType.json",
				"jsonschema/SliceOfEnumType.json",
				"jsonschema/SliceOfPointerToRemoteEnum.json",
				"jsonschema/SliceOfRemoteEnumType.json",
			},
		},
		{
			inputDir:  "builder/testfixtures/structs",
			testName:  "test4-structs",
			runGinkgo: false,
			files: []string{
				"jsonschema/StructType1.json",
				"jsonschema/StructType2.json",
				"jsonschema/StructWithRefs.json",
			},
		},
		{
			inputDir:  "builder/testfixtures/interfaces",
			testName:  "test5-interfaces",
			runGinkgo: false,
			files: []string{
				"jsonschema/FancyStruct.json",
				"jsonschema_gen.go",
			},
		},
		{
			inputDir:  "builder/testfixtures/providers",
			testName:  "test6-providers",
			runGinkgo: false,
			files: []string{
				"jsonschema/Example.json.tmpl",
				"jsonschema_gen.go",
			},
		},
		{
			inputDir:  "builder/testfixtures/entrypoints",
			testName:  "test7-entrypoints",
			runGinkgo: false,
			files: []string{
				"jsonschema/MethodType.json",
				"jsonschema/FuncType.json",
				"jsonschema_gen.go",
			},
		},
		{
			inputDir:  "builder/testfixtures/providers_builder",
			testName:  "test8-providers-builder",
			runGinkgo: false,
			files: []string{
				"jsonschema/Example.json.tmpl",
				"jsonschema_gen.go",
			},
		},
		{
			inputDir:  "builder/testfixtures/v1_interfaces_options",
			testName:  "test9-v1-interfaces-options",
			runGinkgo: false,
			files: []string{
				"jsonschema/Owner.json",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()
			CodegenTest(tc)
		})
	}
}
