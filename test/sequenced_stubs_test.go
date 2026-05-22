package test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/ArtHustle/mock_gen"
)

// Sequenced-stub behavior: Once/Always/PendingStubs/Verify and the chainable
// per-slot setters that compose with them. Sticky-vs-queued mixing rules and
// Reset interactions are exercised here too. Generic sticky-path behavior
// (Returns/ReturnAll/ReturnCustom) lives in behavior_test.go.
var _ = Describe("Sequenced stubs", func() {

	It("Once() arms a single-call return; second call falls back to zero values", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("first").Err(nil)

id1, err1 := m.Get()
if id1 != "first" || err1 != nil {
	t.Fatalf("first call: Get() = (%q, %v), want (first, nil)", id1, err1)
}

id2, err2 := m.Get()
if id2 != "" || err2 != nil {
	t.Fatalf("second call: Get() = (%q, %v), want (\"\", nil) — should fall back to zero values", id2, err2)
}
`, Options{})
	})

	It("Once() chain — three Once() calls produce three queued returns consumed in order", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("a").Err(nil)
m.On().Get().Once().Return().Id("b").Err(nil)
m.On().Get().Once().Return().Id("c").Err(nil)

for i, want := range []string{"a", "b", "c"} {
	got, err := m.Get()
	if got != want || err != nil {
		t.Fatalf("call %d: Get() = (%q, %v), want (%q, nil)", i, got, err, want)
	}
}
`, Options{})
	})

	It("Always() arms a fallback consumed by every call", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Always().Return().Id("steady").Err(nil)

for i := 0; i < 3; i++ {
	got, err := m.Get()
	if got != "steady" || err != nil {
		t.Fatalf("call %d: Get() = (%q, %v), want (steady, nil)", i, got, err)
	}
}
`, Options{})
	})

	It("Mixed Once() then Always(): queue first, then fallback", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("a").Err(nil)
m.On().Get().Once().Return().Id("b").Err(nil)
m.On().Get().Always().Return().Id("steady").Err(nil)

wants := []string{"a", "b", "steady", "steady"}
for i, want := range wants {
	got, err := m.Get()
	if got != want || err != nil {
		t.Fatalf("call %d: Get() = (%q, %v), want (%q, nil)", i, got, err, want)
	}
}
`, Options{})
	})

	It("Per-slot chaining works in one statement", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, count int, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("a").Count(7).Err(nil)

id, count, err := m.Get()
if id != "a" || count != 7 || err != nil {
	t.Fatalf("Get() = (%q, %d, %v), want (a, 7, nil)", id, count, err)
}
`, Options{})
	})

	It("Per-slot last-write-wins within a single entry", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("first").Id("second")

id, err := m.Get()
if id != "second" || err != nil {
	t.Fatalf("Get() = (%q, %v), want (second, nil)", id, err)
}
`, Options{})
	})

	It("Unmentioned slot is the zero value (not inherited from sticky)", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, count int)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Count(42)

id, count := m.Get()
if id != "" || count != 42 {
	t.Fatalf("Get() = (%q, %d), want (\"\", 42) — unmentioned id should be zero", id, count)
}
`, Options{})
	})

	It("PendingStubs() reflects unconsumed queue entries", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("a").Err(nil)
m.On().Get().Once().Return().Id("b").Err(nil)
m.On().Get().Once().Return().Id("c").Err(nil)

if got := m.Calls().Get().PendingStubs(); got != 3 {
	t.Fatalf("before any calls: PendingStubs() = %d, want 3", got)
}

m.Get()
if got := m.Calls().Get().PendingStubs(); got != 2 {
	t.Fatalf("after 1 call: PendingStubs() = %d, want 2", got)
}

m.Get()
m.Get()
if got := m.Calls().Get().PendingStubs(); got != 0 {
	t.Fatalf("after 3 calls: PendingStubs() = %d, want 0", got)
}
`, Options{})
	})

	It("Always() does NOT contribute to PendingStubs()", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Always().Return().Id("steady").Err(nil)

if got := m.Calls().Get().PendingStubs(); got != 0 {
	t.Fatalf("before any calls: PendingStubs() = %d, want 0 (Always is not queued)", got)
}

for i := 0; i < 5; i++ {
	m.Get()
}
if got := m.Calls().Get().PendingStubs(); got != 0 {
	t.Fatalf("after 5 calls: PendingStubs() = %d, want 0", got)
}
`, Options{})
	})

	It("m.Calls().PendingStubs() sums across methods", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
	Close() error
}`, `
m := &ServiceMock{}
m.On().Send().Once().Return().Error(nil)
m.On().Send().Once().Return().Error(nil)
m.On().Close().Once().Return().Error(nil)
m.On().Close().Once().Return().Error(nil)
m.On().Close().Once().Return().Error(nil)

if got := m.Calls().PendingStubs(); got != 5 {
	t.Fatalf("PendingStubs() = %d, want 5 (2 Send + 3 Close)", got)
}

m.Send("x")
if got := m.Calls().PendingStubs(); got != 4 {
	t.Fatalf("after one Send: PendingStubs() = %d, want 4", got)
}
`, Options{})
	})

	It("m.Calls().Verify(t) is a no-op when all queues consumed", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("x").Err(nil)

m.Get()

// All stubs consumed — Verify should not call t.Errorf.
m.Calls().Verify(t)
`, Options{})
	})

	It("m.Calls().Verify(t) calls t.Errorf when queue non-empty; error includes count and method name", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
type mockT struct {
	errs []string
}
func (mt *mockT) Helper() {}
func (mt *mockT) Errorf(format string, args ...any) {
	mt.errs = append(mt.errs, fmtSprintf(format, args...))
}

mt := &mockT{}
m := &ServiceMock{}
m.On().Get().Once().Return().Id("a").Err(nil)
m.On().Get().Once().Return().Id("b").Err(nil)

m.Calls().Verify(mt)

if len(mt.errs) == 0 {
	t.Fatal("expected Verify to call Errorf, but it did not")
}
found := false
for _, e := range mt.errs {
	if stringsPkg.Contains(e, "2 unconsumed stubs") && stringsPkg.Contains(e, "Get") {
		found = true
		break
	}
}
if !found {
	t.Fatalf("Verify error message %v does not contain \"2 unconsumed stubs\" and method name \"Get\"", mt.errs)
}
`, Options{}, `import fmtPkg "fmt"`, `import stringsPkg "strings"`, `var fmtSprintf = fmtPkg.Sprintf`)
	})

	It("Reset() (mock-level) clears queue and always; subsequent call returns zero values", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("x").Err(nil)
m.On().Get().Always().Return().Id("fallback").Err(nil)

m.Reset()

if got := m.Calls().Get().PendingStubs(); got != 0 {
	t.Fatalf("after Reset: PendingStubs() = %d, want 0", got)
}

id, err := m.Get()
if id != "" || err != nil {
	t.Fatalf("after Reset: Get() = (%q, %v), want (\"\", nil) — zero values", id, err)
}
`, Options{})
	})

	It("On().Reset() clears queue and always but preserves call logs", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("first").Err(nil)

m.Get() // consumes the Once entry; logged

m.On().Get().Once().Return().Id("second").Err(nil)

m.On().Reset()

if got := m.Calls().Get().Count(); got != 1 {
	t.Fatalf("after On().Reset(): Count() = %d, want 1 (log preserved)", got)
}
if got := m.Calls().Get().PendingStubs(); got != 0 {
	t.Fatalf("after On().Reset(): PendingStubs() = %d, want 0 (queue cleared)", got)
}
`, Options{})
	})

	It("Per-method On().Get().Reset() clears queue for just that method", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
	Close() error
}`, `
m := &ServiceMock{}
m.On().Send().Once().Return().Error(nil)
m.On().Close().Once().Return().Error(nil)

m.On().Send().Reset()

if got := m.Calls().Send().PendingStubs(); got != 0 {
	t.Fatalf("Send PendingStubs() = %d, want 0 after Send reset", got)
}
if got := m.Calls().Close().PendingStubs(); got != 1 {
	t.Fatalf("Close PendingStubs() = %d, want 1 (unaffected)", got)
}
`, Options{})
	})

	It("Mixing panic: sticky-then-Once panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Return().Id("sticky")

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when adding Once() after sticky stub")
	}
}()
m.On().Get().Once().Return().Id("once")
`, Options{})
	})

	It("Mixing panic: Once-then-sticky panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("once")

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling sticky Return() after Once()")
	}
}()
m.On().Get().Return().Id("sticky")
`, Options{})
	})

	It("Mixing panic: Once-then-ReturnAll panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("once")

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling ReturnAll() after Once()")
	}
}()
m.On().Get().ReturnAll("x", nil)
`, Options{})
	})

	It("Mixing panic: Once-then-ReturnCustom panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("once")

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling ReturnCustom() after Once()")
	}
}()
m.On().Get().ReturnCustom(func() (string, error) { return "x", nil })
`, Options{})
	})

	It("Mixing panic: Always-then-sticky panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Always().Return().Id("always")

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling sticky Return() after Always()")
	}
}()
m.On().Get().Return().Id("sticky")
`, Options{})
	})

	It("Sticky-path chaining is allowed in one statement", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, count int, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Return().Id("x").Count(7).Err(nil)

id, count, err := m.Get()
if id != "x" || count != 7 || err != nil {
	t.Fatalf("Get() = (%q, %d, %v), want (x, 7, nil)", id, count, err)
}
`, Options{})
	})

	It("--async: 100 goroutines + 100-entry queue, every entry consumed exactly once", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(input int) (value int, err error)
}`, `
m := &ServiceMock{}
const N = 100
for i := 0; i < N; i++ {
	i := i
	m.On().Send().Once().Return().Value(i).Err(nil)
}

var wg syncPkg.WaitGroup
for i := 0; i < N; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.Send(0)
	}()
}

done := make(chan struct{})
go func() {
	wg.Wait()
	close(done)
}()
select {
case <-done:
case <-timePkg.After(5 * timePkg.Second):
	t.Fatal("timed out waiting for 100 goroutines")
}

calls := m.Calls().Send().All()
if len(calls) != N {
	t.Fatalf("recorded %d calls, want %d", len(calls), N)
}

values := make([]int, N)
for i, c := range calls {
	values[i] = c.Return.Value
}
sortPkg.Ints(values)
for i, v := range values {
	if v != i {
		t.Fatalf("values[%d] = %d, want %d — not all entries consumed exactly once", i, v, i)
	}
}
`, Options{Async: true}, `import syncPkg "sync"`, `import timePkg "time"`, `import sortPkg "sort"`)
	})

	It("Mixing panic: Always-then-ReturnAll panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Always().Return().Id("always")

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling ReturnAll() after Always()")
	}
}()
m.On().Get().ReturnAll("x", nil)
`, Options{})
	})

	It("Mixing panic: Always-then-ReturnCustom panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Always().Return().Id("always")

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling ReturnCustom() after Always()")
	}
}()
m.On().Get().ReturnCustom(func() (string, error) { return "x", nil })
`, Options{})
	})

	It("Multiple Always() — last-wins", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Always().Return().Id("first")
m.On().Get().Always().Return().Id("second")

for i := 0; i < 3; i++ {
	got, _ := m.Get()
	if got != "second" {
		t.Fatalf("call %d: Get() = %q, want second (last Always wins)", i, got)
	}
}
`, Options{})
	})

	It("After Reset, sticky-then-queue is allowed", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Return().Id("sticky")
m.On().Get().Reset()
// After Reset, arming a Once() stub must not panic.
m.On().Get().Once().Return().Id("once")

id, _ := m.Get()
if id != "once" {
	t.Fatalf("Get() = %q, want once", id)
}
`, Options{})
	})

	It("After Reset, queue-then-sticky is allowed", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Once().Return().Id("once")
m.On().Get().Reset()
// After Reset, arming a sticky stub must not panic.
m.On().Get().Return().Id("sticky")

id, _ := m.Get()
if id != "sticky" {
	t.Fatalf("Get() = %q, want sticky", id)
}
`, Options{})
	})

	// --- new gap-filling tests ---

	It("Mixing panic: sticky-then-Always panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().Return().Id("sticky")

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling Always() after sticky stub")
	}
}()
m.On().Get().Always().Return().Id("always")
`, Options{})
	})

	It("Mixing panic: ReturnAll-then-Once panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().ReturnAll("sticky", nil)

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling Once() after ReturnAll()")
	}
}()
m.On().Get().Once().Return().Id("once")
`, Options{})
	})

	It("Mixing panic: ReturnAll-then-Always panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().ReturnAll("sticky", nil)

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling Always() after ReturnAll()")
	}
}()
m.On().Get().Always().Return().Id("always")
`, Options{})
	})

	It("Mixing panic: ReturnCustom-then-Once panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().ReturnCustom(func() (string, error) { return "x", nil })

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling Once() after ReturnCustom()")
	}
}()
m.On().Get().Once().Return().Id("once")
`, Options{})
	})

	It("Mixing panic: ReturnCustom-then-Always panics", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Get().ReturnCustom(func() (string, error) { return "x", nil })

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when calling Always() after ReturnCustom()")
	}
}()
m.On().Get().Always().Return().Id("always")
`, Options{})
	})

	It("Verify breakdown names each method with unconsumed stubs", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
	Close() error
}`, `
type mockT struct {
	errs []string
}
func (mt *mockT) Helper() {}
func (mt *mockT) Errorf(format string, args ...any) {
	mt.errs = append(mt.errs, fmtSprintf(format, args...))
}

mt := &mockT{}
m := &ServiceMock{}
m.On().Send().Once().Return().Error(nil)
m.On().Send().Once().Return().Error(nil)
m.On().Close().Once().Return().Error(nil)

m.Calls().Verify(mt)

if len(mt.errs) == 0 {
	t.Fatal("expected Verify to call Errorf, but it did not")
}
// The breakdown must mention both method names.
found := false
for _, e := range mt.errs {
	if stringsPkg.Contains(e, "Send") && stringsPkg.Contains(e, "Close") {
		found = true
		break
	}
}
if !found {
	t.Fatalf("Verify message %v does not name both Send and Close", mt.errs)
}
`, Options{}, `import fmtPkg "fmt"`, `import stringsPkg "strings"`, `var fmtSprintf = fmtPkg.Sprintf`)
	})
})
