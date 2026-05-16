package test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// Behavior tests generate a mock, write a tiny Go test file alongside it
// that exercises the runtime API (Calls/On/Reset, At/Last/All/Count,
// Returns/Handler, Param/Return capture), and run `go test` on the tempdir.
//
// String-match tests can verify the generated *shape*; only running the
// mock proves the *semantics*. Each entry costs a `go test` (~1-3s). Keep
// the set focused on behaviors that aren't visible from the generated
// source alone.
var _ = Describe("Generated mock behavior", func() {

	// --- core: Calls() / On() / Reset() semantics ---

	It("Calls().M().Count() reflects the number of recorded calls", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("c")
if got := m.Calls().Send().Count(); got != 3 {
	t.Fatalf("Count() = %d, want 3", got)
}
`, Options{})
	})

	It("Calls().M().At(i) returns the i-th recorded call with params and returns captured", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().Returns("ID-1", nil)
m.Send("a")
m.On().Send().Returns("ID-2", nil)
m.Send("b")

c0 := m.Calls().Send().At(0)
if c0.Param.Email != "a" {
	t.Fatalf("At(0).Param.Email = %q, want %q", c0.Param.Email, "a")
}
if c0.Return.Id != "ID-1" {
	t.Fatalf("At(0).Return.Id = %q, want %q", c0.Return.Id, "ID-1")
}

c1 := m.Calls().Send().At(1)
if c1.Return.Id != "ID-2" {
	t.Fatalf("At(1).Return.Id = %q, want %q", c1.Return.Id, "ID-2")
}
`, Options{})
	})

	It("Calls().M().At(-1) and Last() return the most recent call", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("first")
m.Send("last")

if got := m.Calls().Send().At(-1).Param.Email; got != "last" {
	t.Fatalf("At(-1).Param.Email = %q, want %q", got, "last")
}
if got := m.Calls().Send().Last().Param.Email; got != "last" {
	t.Fatalf("Last().Param.Email = %q, want %q", got, "last")
}
`, Options{})
	})

	It("Calls().M().All() returns a defensive copy that the caller can mutate without affecting the mock", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")

calls := m.Calls().Send().All()
calls[0].Param.Email = "MUTATED"

if m.Calls().Send().At(0).Param.Email != "a" {
	t.Fatal("All() did not return a defensive copy")
}
`, Options{})
	})

	It("On().M().Returns(...) sets static return values used by interface calls", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (string, error)
}`, `
m := &ServiceMock{}
m.On().Get().Returns("hello", nil)

s, err := m.Get()
if s != "hello" || err != nil {
	t.Fatalf("Get() = (%q, %v), want (hello, nil)", s, err)
}
`, Options{})
	})

	It("On().M().ReturnsX(...) overrides individual return values without touching others", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, count int, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Returns("default", 42, nil)
m.On().Get().ReturnsErr(errorsNew("override"))

id, count, err := m.Get()
if id != "default" || count != 42 {
	t.Fatalf("got id=%q count=%d, expected unchanged", id, count)
}
if err == nil || err.Error() != "override" {
	t.Fatalf("err = %v, want override", err)
}
`, Options{}, `import errorsPkg "errors"`, `var errorsNew = errorsPkg.New`)
	})

	It("On().M().Handler(...) overrides static returns with a dynamic function", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().Returns("STATIC", nil)
m.On().Send().Handler(func(email string) (string, error) {
	return "HANDLER:" + email, nil
})

id, err := m.Send("hello")
if id != "HANDLER:hello" || err != nil {
	t.Fatalf("Send() = (%q, %v), want (HANDLER:hello, nil)", id, err)
}
`, Options{})
	})

	It("Each recorded call's Return reflects what the handler actually returned on that call", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
n := 0
m.On().Send().Handler(func(email string) (string, error) {
	n++
	if n == 1 {
		return "first", nil
	}
	return "second", nil
})

m.Send("a")
m.Send("b")

if got := m.Calls().Send().At(0).Return.Id; got != "first" {
	t.Fatalf("At(0).Return.Id = %q, want first", got)
}
if got := m.Calls().Send().At(1).Return.Id; got != "second" {
	t.Fatalf("At(1).Return.Id = %q, want second", got)
}
`, Options{})
	})

	// --- Reset semantics ---

	It("mock.Reset() clears every log and every cfg", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
	Close() error
}`, `
m := &ServiceMock{}
m.On().Send().ReturnsError(errorsNew("injected"))
m.Send("a")
m.Close()

m.Reset()

if m.Calls().Send().Count() != 0 || m.Calls().Close().Count() != 0 {
	t.Fatal("Reset did not clear logs")
}
// cfg cleared: Send returns the zero value of error (nil), not the injected error
if got := m.Send("b"); got != nil {
	t.Fatalf("after Reset, Send returned %v, want nil (cfg cleared)", got)
}
`, Options{}, `import errorsPkg "errors"`, `var errorsNew = errorsPkg.New`)
	})

	It("Calls().Reset() clears all logs but preserves On() configuration", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (string, error)
}`, `
m := &ServiceMock{}
m.On().Get().Returns("armed", nil)
m.Get()

m.Calls().Reset()

if m.Calls().Get().Count() != 0 {
	t.Fatal("Calls().Reset() did not clear log")
}
// Config should be preserved
s, _ := m.Get()
if s != "armed" {
	t.Fatalf("Get() = %q after Calls().Reset(), want armed (config preserved)", s)
}
`, Options{})
	})

	It("On().Reset() clears all cfg but preserves call logs", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (string, error)
}`, `
m := &ServiceMock{}
m.On().Get().Returns("armed", nil)
m.Get()

m.On().Reset()

if m.Calls().Get().Count() != 1 {
	t.Fatal("On().Reset() should preserve the call log")
}
// Config should be cleared — second call returns zero
s, _ := m.Get()
if s != "" {
	t.Fatalf("Get() = %q after On().Reset(), want empty (config cleared)", s)
}
`, Options{})
	})

	It("Calls().Send().Reset() clears just one method's log", func() {
		runMockBehavior(`package svc

type Service interface {
	A() error
	B() error
}`, `
m := &ServiceMock{}
m.A()
m.B()

m.Calls().A().Reset()

if m.Calls().A().Count() != 0 {
	t.Fatal("A's log should be empty")
}
if m.Calls().B().Count() != 1 {
	t.Fatal("B's log should be preserved")
}
`, Options{})
	})

	It("On().Send().Reset() clears just one method's cfg", func() {
		runMockBehavior(`package svc

type Service interface {
	A() (string, error)
	B() (string, error)
}`, `
m := &ServiceMock{}
m.On().A().Returns("a-val", nil)
m.On().B().Returns("b-val", nil)

m.On().A().Reset()

if got, _ := m.A(); got != "" {
	t.Fatalf("A() = %q after On().A().Reset(), want empty", got)
}
if got, _ := m.B(); got != "b-val" {
	t.Fatalf("B() = %q after On().A().Reset(), want b-val (preserved)", got)
}
`, Options{})
	})

	// --- At() panic semantics ---

	It("Calls().M().At(i) panics with a descriptive message on positive out-of-range", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(x int) error
}`, `
m := &ServiceMock{}
m.Send(1)

defer func() {
	r := recover()
	if r == nil { t.Fatal("expected panic") }
	msg, _ := r.(string)
	want := "ServiceMock call index 5 out of range, log has 1 entries"
	if msg != want { t.Fatalf("panic = %q, want %q", msg, want) }
}()
m.Calls().Send().At(5)
`, Options{})
	})

	It("Calls().M().At(-N) where N > len panics with both input and resolved index", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(x int) error
}`, `
m := &ServiceMock{}
m.Send(1)

defer func() {
	r := recover()
	if r == nil { t.Fatal("expected panic") }
	msg, _ := r.(string)
	want := "ServiceMock call index -5 (resolved to -4) out of range, log has 1 entries"
	if msg != want { t.Fatalf("panic = %q, want %q", msg, want) }
}()
m.Calls().Send().At(-5)
`, Options{})
	})

	// --- entry-point override behavior ---

	It("InspectEntry override produces accessors with the overridden name and the source-named method is recorded", func() {
		runMockBehavior(`package svc

type Service interface {
	Calls()
	DoThing()
}`, `
m := &ServiceMock{}
m.DoThing()
m.Calls()

if m.Inspect().DoThing().Count() != 1 {
	t.Fatal("Inspect().DoThing().Count() should be 1")
}
if m.Inspect().Calls().Count() != 1 {
	t.Fatal("Inspect().Calls().Count() should be 1 — the source-method named Calls was recorded")
}
`, Options{InspectEntry: "Inspect"})
	})

	It("ConfigEntry override produces config accessors with the overridden name", func() {
		runMockBehavior(`package svc

type Service interface {
	On() error
}`, `
m := &ServiceMock{}
m.Configure().On().ReturnsError(nil)
err := m.On()
if err != nil {
	t.Fatalf("On() = %v, want nil", err)
}
if m.Calls().On().Count() != 1 {
	t.Fatal("Calls().On() should reflect the source method named On")
}
`, Options{ConfigEntry: "Configure"})
	})

	// --- safety: append-before-handler so panicking handlers don't lose the call ---

	It("a panicking handler still leaves the call recorded (append precedes handler)", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.On().Send().Handler(func(email string) error {
	panic("handler exploded")
})

defer func() {
	_ = recover()
	if m.Calls().Send().Count() != 1 {
		t.Fatalf("expected 1 recorded call after panicking handler, got %d", m.Calls().Send().Count())
	}
	if got := m.Calls().Send().At(0).Param.Email; got != "boom@x.com" {
		t.Fatalf("recorded email = %q, want boom@x.com", got)
	}
}()

m.Send("boom@x.com")
`, Options{})
	})

	It("--thread-safe: a panicking handler still leaves the call recorded", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.On().Send().Handler(func(email string) error {
	panic("ts-handler exploded")
})

defer func() {
	_ = recover()
	if m.Calls().Send().Count() != 1 {
		t.Fatalf("expected 1 recorded call after panicking handler, got %d", m.Calls().Send().Count())
	}
}()

m.Send("boom@x.com")
`, Options{MakeThreadSafe: true})
	})

	// --- multi-mock isolation in the same source/package ---

	It("two mocks in the same package have independent Calls() / On() / logs", func() {
		runMockBehavior(`package svc

type A interface { DoA(x int) error }
type B interface { DoB(y string) error }`, `
a := &AMock{}
b := &BMock{}
a.DoA(1)
a.DoA(2)
b.DoB("x")

if a.Calls().DoA().Count() != 2 {
	t.Fatalf("a.Calls().DoA().Count() = %d, want 2", a.Calls().DoA().Count())
}
if b.Calls().DoB().Count() != 1 {
	t.Fatalf("b.Calls().DoB().Count() = %d, want 1", b.Calls().DoB().Count())
}

a.Reset()
if a.Calls().DoA().Count() != 0 {
	t.Fatal("a.Reset() should clear a's log")
}
if b.Calls().DoB().Count() != 1 {
	t.Fatal("a.Reset() must not affect b")
}
`, Options{})
	})

	// --- embedded interface methods are recorded under the embedding mock ---

	It("embedded interface methods appear on the mock's Calls()/On() namespaces", func() {
		runMockBehavior(`package svc

import "io"

type Service interface {
	io.Closer
	Greet() string
}`, `
m := &ServiceMock{}
m.On().Greet().Returns("hi")
m.On().Close().Returns(nil)

if got := m.Greet(); got != "hi" {
	t.Fatalf("Greet() = %q, want hi", got)
}
if err := m.Close(); err != nil {
	t.Fatalf("Close() = %v, want nil", err)
}
if m.Calls().Greet().Count() != 1 || m.Calls().Close().Count() != 1 {
	t.Fatal("embedded Close and own Greet should both be recorded")
}
`, Options{})
	})

	// --- thread-safe variant: same semantics, no race ---

	It("--thread-safe: parallel calls record all entries without dropping any", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(x int) error
}`, `
m := &ServiceMock{}
var wg syncPkg.WaitGroup
const N = 200
for i := 0; i < N; i++ {
	wg.Add(1)
	go func(x int) { defer wg.Done(); m.Send(x) }(i)
}
wg.Wait()

if got := m.Calls().Send().Count(); got != N {
	t.Fatalf("Count() = %d, want %d", got, N)
}
`, Options{MakeThreadSafe: true}, `import syncPkg "sync"`)
	})
})

// runMockBehavior generates a mock for the given source and runs the given
// harness code as the body of a `TestBehavior` function in the tempdir's
// test package. Fails the spec if `go test` errors.
//
// extras (optional, variadic) are extra lines pasted at file scope of the
// generated test file (e.g. package-level vars referenced by harness).
// Each extra is pasted on its own line above the TestBehavior function.
func runMockBehavior(source, harnessCode string, opts Options, extras ...string) {
	GinkgoHelper()

	tempDir, err := os.MkdirTemp(os.TempDir(), "mock_gen_behavior_")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { os.RemoveAll(tempDir) })

	Expect(os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module behavior\n\ngo 1.23\n"),
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

	harness := "package test\n\nimport \"testing\"\n"
	for _, e := range extras {
		harness += e + "\n"
	}
	harness += "\nfunc TestBehavior(t *testing.T) {\n" + harnessCode + "\n}\n"
	Expect(os.WriteFile(
		filepath.Join(testDir, "behavior_test.go"),
		[]byte(harness),
		0644,
	)).To(Succeed())

	cmd := exec.Command("go", "test", "./test/", "-run", "TestBehavior", "-count=1")
	cmd.Dir = tempDir
	output, runErr := cmd.CombinedOutput()
	Expect(runErr).NotTo(HaveOccurred(),
		"go test failed:\n%s\n\n--- mock source ---\n%s\n\n--- harness ---\n%s",
		string(output), mockContent, harness)
}
