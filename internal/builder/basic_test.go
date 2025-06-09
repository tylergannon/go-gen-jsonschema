package builder_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/testutils"
)

const cleanUp = true

type testCase struct {
	inputDir  string
	testName  string
	runGinkgo bool
	files     []string
}

func TestBasic(t *testing.T) {
	CmdSuccessAssertions := func(t *testing.T, stdout, stderr string, exitCode int) {
		require.Empty(t, stderr)
		require.Equal(t, 0, exitCode)
	}

	CodegenTest := func(tc testCase) {
		cwd, err := os.Getwd()
		require.NoError(t, err)
		tempDir := filepath.Join(cwd, "test_run", tc.testName)
		if cleanUp {
			defer os.RemoveAll(tempDir)
		}
		require.NoError(t, os.RemoveAll(tempDir))
		require.NoError(t, os.MkdirAll(tempDir, 0755))
		inputPathFull := filepath.Clean(filepath.Join(cwd, "..", tc.inputDir))
		fi, err := os.Stat(inputPathFull)
		require.NoError(t, err)
		require.True(t, fi.IsDir())
		require.NoError(t, testutils.CopyDir(inputPathFull, tempDir))

		exitCode, stdout, stderr, err := testutils.RunCommand("go", tempDir, "generate", "./...")
		require.NoError(t, err)
		CmdSuccessAssertions(t, stdout, stderr, exitCode)

		for _, fname := range tc.files {
			fpath := filepath.Clean(filepath.Join(tempDir, fname))
			info, err := os.Stat(fpath)
			require.NoError(t, err, fmt.Sprintf("Expected file %s to be created in %s", fname, tempDir))
			require.False(t, info.IsDir())
			testutils.AssertGoldenFile(t, fpath, ".golden")
		}
		if tc.runGinkgo {
			exitCode, stdout, stderr, err = testutils.RunCommand("ginkgo", tempDir, "./...")
			require.NoError(t, err)
			CmdSuccessAssertions(t, stdout, stderr, exitCode)
		}
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
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			CodegenTest(tc)
		})
	}
}
