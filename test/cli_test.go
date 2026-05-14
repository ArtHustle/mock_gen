package test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	. "github.com/ArtHustle/mock_gen"
)

const simpleServiceSource = `package service

type Service interface {
	Add()
}`

var _ = Describe("CLI", func() {
	Context("with no arguments", func() {
		It("exits with code 1", func() {
			command := exec.Command(mockGeneratorPath)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when the input file does not exist", func() {
		It("exits with code 1", func() {
			command := exec.Command(mockGeneratorPath, "~/tmp/invalidInputFile.go")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("with a valid input file", func() {
		var inputFilePath string
		var outputFilePath string

		BeforeEach(func() {
			tempDir, err := os.MkdirTemp(os.TempDir(), "mock_gen_test_")
			Expect(err).NotTo(HaveOccurred())

			goModFile := filepath.Join(tempDir, "go.mod")
			Expect(os.WriteFile(goModFile, []byte("module mock_gen_test\n\ngo 1.23\n"), 0644)).To(Succeed())

			inputFile, err := os.CreateTemp(tempDir, "service.go")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { os.RemoveAll(tempDir) })

			testDir := filepath.Join(tempDir, "test")
			Expect(os.MkdirAll(testDir, os.ModePerm)).To(Succeed())
			outputFilePath = filepath.Join(testDir, "service_mock.go")

			inputFilePath = inputFile.Name()

			_, err = inputFile.WriteString(simpleServiceSource)
			Expect(err).NotTo(HaveOccurred())

			Expect(inputFile.Close()).To(Succeed())
		})

		runWithParams := func(params ...string) {
			command := exec.Command(mockGeneratorPath, params...)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(outputFilePath).To(BeAnExistingFile())
		}

		It("writes an output file when the source path is passed directly", func() {
			runWithParams(inputFilePath)
		})

		It("writes an output file when the source path follows `--`", func() {
			runWithParams("--", inputFilePath)
		})
	})

	// Regression test: when a package has a struct field with the same name as
	// an interface, findTypeDef must return the TypeName (interface declaration),
	// not the struct field. Previously this was non-deterministic due to map
	// iteration order, causing mock_gen to silently produce an empty mock on
	// subsequent runs.
	Context("when an interface name collides with a struct field name in the same package", func() {
		It("generates a non-empty mock on every run", func() {
			tempDir, err := os.MkdirTemp(os.TempDir(), "mock_gen_collision_test_")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { os.RemoveAll(tempDir) })

			Expect(os.WriteFile(
				filepath.Join(tempDir, "go.mod"),
				[]byte("module collision_test\n\ngo 1.23\n"),
				0644,
			)).To(Succeed())

			// domain.go defines a struct with a field named "Provider"
			Expect(os.WriteFile(filepath.Join(tempDir, "domain.go"), []byte(`package service

type Config struct {
	Provider string
}
`), 0644)).To(Succeed())

			// provider.go defines the Provider interface
			providerFile := filepath.Join(tempDir, "provider.go")
			Expect(os.WriteFile(providerFile, []byte(`package service

import "context"

type Provider interface {
	Delete(ctx context.Context, id string) error
}
`), 0644)).To(Succeed())

			testDir := filepath.Join(tempDir, "test")
			Expect(os.MkdirAll(testDir, 0755)).To(Succeed())
			outputFile := filepath.Join(testDir, "provider_mock.go")

			// Run 10 times — every run must produce a non-empty mock with the
			// Delete method. Without the TypeName filter in findTypeDef, map
			// iteration non-determinism causes findTypeDef to return the
			// struct field "Provider string" instead of the interface
			// declaration, resulting in an empty mock. Running 10 times makes
			// this highly likely to be caught even though the failure is
			// non-deterministic.
			for i := 1; i <= 10; i++ {
				result, err := GenerateMockFileWithOptions(providerFile, Options{})
				Expect(err).NotTo(HaveOccurred(), "run %d failed", i)
				Expect(result).To(ContainSubstring("ProviderMock"), "run %d: mock struct missing", i)
				Expect(result).To(ContainSubstring("Delete_calls"), "run %d: Delete method missing", i)
				Expect(os.WriteFile(outputFile, []byte(result), 0644)).To(Succeed())
			}
		})
	})
})
