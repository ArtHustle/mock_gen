package test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// Compile-check tests run the generator on adversarial source interfaces,
// write the result to a tempdir, and run `go build` on it. They catch
// classes of bugs that string-match tests cannot:
//
//   - identifier shadowing between user params and generated locals
//     (e.g. param `handler` shadowed `handler :=`, fixed in v0.8.0)
//   - missing imports in generated code (the source's type uses a package
//     the generator forgot to walk)
//   - inline type rendering that produces empty/broken Go (fixed in v0.8.1)
//   - cross-mock declaration collisions in the same package
//
// Each Entry costs a `go build` invocation (~0.3-1s). Keep the set focused
// on bug classes; don't duplicate shape assertions already covered by the
// literal-output tests.
var _ = Describe("Generated mock compiles", func() {
	DescribeTable("source interface -> generated mock builds cleanly",
		func(source string, opts Options) {
			compileGeneratedMock(source, opts)
		},

		// --- snapshot-local shadowing (regression for #14, #17) ---

		Entry("param named handler under --thread-safe --generate-return-handler", `
package svc

type Service interface {
	Run(handler func(int) error) error
}`, Options{Async: true}),

		Entry("param named _handler doesn't collide with __mockHandler", `
package svc

type Service interface {
	Run(_handler func(int) error) error
}`, Options{Async: true}),

		Entry("param named _r0 doesn't collide with __mockR0", `
package svc

type Service interface {
	Get(_r0 int) (string, error)
}`, Options{Async: true}),

		Entry("multiple snapshot-shadowing names at once", `
package svc

type Service interface {
	Do(_handler func() error, _r0 int, _r1 string) (int, string, error)
}`, Options{Async: true}),

		// --- callIndex-body local names as user params ---

		Entry("user params named i, length, resolved build cleanly", `
package svc

type Service interface {
	Lookup(i, length, resolved int) int
}`, Options{Async: true}),

		// --- inline types (regression for #18) ---

		Entry("chan struct{} return", `
package svc

type Service interface {
	Done() <-chan struct{}
}`, Options{Async: true}),

		Entry("empty interface{} param", `
package svc

type Service interface {
	Process(v interface{}) error
}`, Options{Async: true}),

		Entry("inline struct with fields", `
package svc

type Service interface {
	Configure(opts struct{ Host string; Port int }) error
}`, Options{Async: true}),

		// --- cross-mock package-level collisions ---

		Entry("two mocks in same source file build together", `
package svc

type A interface { DoA(x int) error }
type B interface { DoB(y string) error }`, Options{Async: true}),

		// --- realistic interface combining many features ---

		Entry("realistic interface: ctx, variadic, func param, multi-return", `
package svc

import (
	"context"
)

type Event struct {
	Topic   string
	Payload []byte
}

type Runtime interface {
	Activate(ctx context.Context, id string) error
	Fetch(ctx context.Context, ids ...string) (events []Event, total int, err error)
	Subscribe(ctx context.Context, topics []string, handler func(Event) error) (cancel func(), err error)
}`, Options{Async: true}),
	)
})

// compileGeneratedMock writes a minimal Go module to a tempdir, runs the
// generator against the given source, writes the output to test/, and runs
// `go build` on the test package. Fails the spec if any step errors.
func compileGeneratedMock(source string, opts Options) {
	GinkgoHelper()

	tempDir, err := os.MkdirTemp(os.TempDir(), "mock_gen_compile_")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { os.RemoveAll(tempDir) })

	Expect(os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module compile_check\n\ngo 1.23\n"),
		0644,
	)).To(Succeed())

	srcPath := filepath.Join(tempDir, "src.go")
	Expect(os.WriteFile(srcPath, []byte(source), 0644)).To(Succeed())

	mockContent, err := GenerateMockFileWithOptions(srcPath, opts)
	Expect(err).NotTo(HaveOccurred(), "generator failed")

	testDir := filepath.Join(tempDir, "test")
	Expect(os.MkdirAll(testDir, 0755)).To(Succeed())
	Expect(os.WriteFile(
		filepath.Join(testDir, "src_mock.go"),
		[]byte(mockContent),
		0644,
	)).To(Succeed())

	cmd := exec.Command("go", "build", "./test/")
	cmd.Dir = tempDir
	output, buildErr := cmd.CombinedOutput()
	Expect(buildErr).NotTo(HaveOccurred(),
		"generated mock failed to compile:\n%s\n\n--- mock source ---\n%s", string(output), mockContent)
}
