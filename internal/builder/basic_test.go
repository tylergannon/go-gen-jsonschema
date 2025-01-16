package builder_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tylergannon/go-gen-jsonschema/internal/testutils"
	"os"
	"path/filepath"
)

const cleanUp = true

var _ = Describe("Basic", func() {
	// Function to assert successful command execution
	var CmdSuccessAssertions = func(stdout, stderr string, exitCode int) {
		Expect(stderr).To(BeEmpty())
		//fmt.Println(stdout)
		//Expect(stdout).NotTo(BeEmpty())
		Expect(exitCode).To(Equal(0))
	}

	// Parameterized test function
	var CodegenTest = func(inputDir, testName string, runGinkgo bool, files ...string) {
		// Create tempDir and clean up automatically
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		tempDir := filepath.Join(cwd, "test_run", testName)
		defer func() {
			if cleanUp {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			}
		}()

		// Setup input directory
		Expect(os.RemoveAll(tempDir)).To(Succeed())
		Expect(os.MkdirAll(tempDir, 0755)).To(Succeed())
		inputPathFull := filepath.Clean(filepath.Join(cwd, "..", inputDir))
		Expect(inputPathFull).To(BeADirectory())
		Expect(testutils.CopyDir(inputPathFull, tempDir)).To(Succeed())

		// Run the "gen" command
		exitCode, stdout, stderr, err := testutils.RunCommand("go", tempDir, "generate", "./...")
		Expect(err).NotTo(HaveOccurred())
		CmdSuccessAssertions(stdout, stderr, exitCode)

		{
			fname := filepath.Join(tempDir, "jsonschema_gen.go")
			Expect(fname).To(
				testutils.MatchGoldenFile(".golden"),
			)
		}

		// Assertions on generated files
		//Expect(filepath.Join(tempDir, "tasks.go")).To(BeARegularFile())
		for _, fname := range files {
			fpath := filepath.Clean(filepath.Join(tempDir, fname))
			Expect(fpath).To(BeARegularFile(), fmt.Sprintf("Expected file %s to be created in %s", fname, tempDir))
			Expect(fpath).To(
				testutils.MatchGoldenFile(".golden"),
			)
		}
		if runGinkgo {
			exitCode, stdout, stderr, err = testutils.RunCommand("ginkgo", tempDir, "./...")
			Expect(err).NotTo(HaveOccurred())
			CmdSuccessAssertions(stdout, stderr, exitCode)
		}
	}

	DescribeTable("Run", CodegenTest)

})
