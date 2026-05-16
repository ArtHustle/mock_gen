package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

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

	// Idempotent writes: if the generated content matches what's already on
	// disk byte-for-byte, the CLI should leave the file's mtime unchanged.
	// Preserves build caches and avoids spurious git noise when running
	// `go generate` in a loop. (#5)
	Context("when run twice with no source change", func() {
		var tempDir, srcPath, outPath string

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp(os.TempDir(), "mock_gen_idem_test_")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { os.RemoveAll(tempDir) })

			Expect(os.WriteFile(
				filepath.Join(tempDir, "go.mod"),
				[]byte("module idem_test\n\ngo 1.23\n"),
				0644,
			)).To(Succeed())

			srcPath = filepath.Join(tempDir, "svc.go")
			Expect(os.WriteFile(srcPath, []byte(`package svc

type Service interface {
	Do(x int) error
}
`), 0644)).To(Succeed())

			outPath = filepath.Join(tempDir, "test", "svc_mock.go")
		})

		runGen := func() {
			command := exec.Command(mockGeneratorPath, srcPath)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		}

		It("does not rewrite the output file when content is unchanged", func() {
			runGen()
			before, err := os.Stat(outPath)
			Expect(err).NotTo(HaveOccurred())

			// Sleep just long enough that any rewrite would show a different mtime.
			// File-system mtime resolution on macOS APFS is nanosecond, on some
			// Linux filesystems it's 1s. 1.1s is safe everywhere.
			time.Sleep(1100 * time.Millisecond)
			runGen()

			after, err := os.Stat(outPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(after.ModTime()).To(Equal(before.ModTime()),
				"second run with unchanged source must not rewrite the file")
		})

		It("does rewrite when the source changes", func() {
			runGen()
			before, err := os.Stat(outPath)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(1100 * time.Millisecond)
			Expect(os.WriteFile(srcPath, []byte(`package svc

type Service interface {
	Do(x int) error
	Another() string
}
`), 0644)).To(Succeed())
			runGen()

			after, err := os.Stat(outPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(after.ModTime()).NotTo(Equal(before.ModTime()),
				"changed source must trigger a rewrite")
		})
	})

	// Directory-mode input: when the positional arg is a directory, the CLI
	// scans every non-test `.go` file in it and generates a mock per source
	// file that declares at least one interface.
	Context("when the input is a directory", func() {
		var tempDir string

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp(os.TempDir(), "mock_gen_dir_test_")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { os.RemoveAll(tempDir) })

			Expect(os.WriteFile(
				filepath.Join(tempDir, "go.mod"),
				[]byte("module dir_test\n\ngo 1.23\n"),
				0644,
			)).To(Succeed())

			Expect(os.WriteFile(filepath.Join(tempDir, "a.go"), []byte(`package svc

type A interface {
	DoA()
}
`), 0644)).To(Succeed())

			Expect(os.WriteFile(filepath.Join(tempDir, "b.go"), []byte(`package svc

type B interface {
	DoB()
}
`), 0644)).To(Succeed())

			// types-only file: no interfaces; should not produce a mock.
			Expect(os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(`package svc

type Config struct {
	X int
}
`), 0644)).To(Succeed())

			// _test.go file: must be ignored even if it contains an interface.
			Expect(os.WriteFile(filepath.Join(tempDir, "helpers_test.go"), []byte(`package svc

type TestOnly interface {
	Ignored()
}
`), 0644)).To(Succeed())
		})

		It("generates one mock per source file that declares interfaces", func() {
			command := exec.Command(mockGeneratorPath, tempDir)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Expect(filepath.Join(tempDir, "test", "a_mock.go")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, "test", "b_mock.go")).To(BeAnExistingFile())

			Expect(filepath.Join(tempDir, "test", "types_mock.go")).NotTo(BeAnExistingFile(),
				"files without interfaces should not produce a mock")
			Expect(filepath.Join(tempDir, "test", "helpers_test_mock.go")).NotTo(BeAnExistingFile(),
				"_test.go files should be ignored")

			aContent, err := os.ReadFile(filepath.Join(tempDir, "test", "a_mock.go"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(aContent)).To(ContainSubstring("type AMock struct"))
			Expect(string(aContent)).NotTo(ContainSubstring("type BMock"),
				"per-file output must not include interfaces from other files")
		})

		It("rejects --output when input is a directory", func() {
			command := exec.Command(mockGeneratorPath, "--output", "/tmp/whatever.go", tempDir)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
		})

		It("exits with code 1 when the directory contains no interfaces", func() {
			emptyDir, err := os.MkdirTemp(os.TempDir(), "mock_gen_empty_test_")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { os.RemoveAll(emptyDir) })

			Expect(os.WriteFile(
				filepath.Join(emptyDir, "go.mod"),
				[]byte("module empty_test\n\ngo 1.23\n"),
				0644,
			)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(emptyDir, "x.go"), []byte(`package svc

type Config struct{}
`), 0644)).To(Succeed())

			command := exec.Command(mockGeneratorPath, emptyDir)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
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
				Expect(result).To(ContainSubstring("func (this *ProviderMock) Delete("), "run %d: Delete method missing", i)
				Expect(os.WriteFile(outputFile, []byte(result), 0644)).To(Succeed())
			}
		})
	})
})
