package test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/ArtHustle/mock_gen"
)

// Timeline behavior: m.Calls().Timeline(), HappenedBefore, and AssertOrder API.
// Phase 1: all 14 tests are expected to FAIL because Timeline/HappenedBefore/
// AssertOrder do not exist on the generated mock yet. They compile fine here
// (outer binary); the inner `go test` spawned by runMockBehavior fails with
// undefined identifiers.
var _ = Describe("Timeline", func() {

	// --- Timeline shape ---

	It("Empty timeline on fresh mock", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}

if got := len(m.Calls().Timeline()); got != 0 {
	t.Fatalf("fresh mock: len(Timeline()) = %d, want 0", got)
}
`, Options{})
	})

	It("Single-method timeline records Method, Index, and GlobalIndex in order", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("c")

tl := m.Calls().Timeline()
if len(tl) != 3 {
	t.Fatalf("len(Timeline()) = %d, want 3", len(tl))
}
if tl[0].Method != "Send" || tl[0].Index != 0 || tl[0].GlobalIndex != 0 {
	t.Fatalf("tl[0] = {Method:%q Index:%d GlobalIndex:%d}, want {Send 0 0}", tl[0].Method, tl[0].Index, tl[0].GlobalIndex)
}
if tl[1].Index != 1 || tl[1].GlobalIndex != 1 {
	t.Fatalf("tl[1] = {Index:%d GlobalIndex:%d}, want {1 1}", tl[1].Index, tl[1].GlobalIndex)
}
if tl[2].Index != 2 || tl[2].GlobalIndex != 2 {
	t.Fatalf("tl[2] = {Index:%d GlobalIndex:%d}, want {2 2}", tl[2].Index, tl[2].GlobalIndex)
}
`, Options{})
	})

	It("Cross-method timeline preserves interleaved GlobalIndex ordering", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
	Close() error
}`, `
m := &ServiceMock{}
m.Send("a")
m.Close()
m.Send("b")

tl := m.Calls().Timeline()
if len(tl) != 3 {
	t.Fatalf("len(Timeline()) = %d, want 3", len(tl))
}
if tl[0].Method != "Send" || tl[0].Index != 0 || tl[0].GlobalIndex != 0 {
	t.Fatalf("tl[0] = {Method:%q Index:%d GlobalIndex:%d}, want {Send 0 0}", tl[0].Method, tl[0].Index, tl[0].GlobalIndex)
}
if tl[1].Method != "Close" || tl[1].Index != 0 || tl[1].GlobalIndex != 1 {
	t.Fatalf("tl[1] = {Method:%q Index:%d GlobalIndex:%d}, want {Close 0 1}", tl[1].Method, tl[1].Index, tl[1].GlobalIndex)
}
if tl[2].Method != "Send" || tl[2].Index != 1 || tl[2].GlobalIndex != 2 {
	t.Fatalf("tl[2] = {Method:%q Index:%d GlobalIndex:%d}, want {Send 1 2}", tl[2].Method, tl[2].Index, tl[2].GlobalIndex)
}
`, Options{})
	})

	It("Timeline() returns a defensive copy; mutating the returned slice does not affect subsequent calls", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("a")

tl := m.Calls().Timeline()
tl[0].Method = "X"

tl2 := m.Calls().Timeline()
if tl2[0].Method != "Send" {
	t.Fatalf("Timeline defensive copy broken: tl2[0].Method = %q, want Send", tl2[0].Method)
}
`, Options{})
	})

	It("m.Reset() clears the timeline", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("c")
m.Reset()

if got := len(m.Calls().Timeline()); got != 0 {
	t.Fatalf("after m.Reset(): len(Timeline()) = %d, want 0", got)
}
`, Options{})
	})

	It("m.Calls().Reset() clears the timeline", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("c")
m.Calls().Reset()

if got := len(m.Calls().Timeline()); got != 0 {
	t.Fatalf("after m.Calls().Reset(): len(Timeline()) = %d, want 0", got)
}
`, Options{})
	})

	It("m.Calls().Send().Reset() clears per-method log but does NOT clear the timeline", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Calls().Send().Reset()

if got := len(m.Calls().Timeline()); got != 2 {
	t.Fatalf("after Send().Reset(): len(Timeline()) = %d, want 2 (timeline preserved)", got)
}
if got := m.Calls().Send().Count(); got != 0 {
	t.Fatalf("after Send().Reset(): Count() = %d, want 0 (per-method log cleared)", got)
}
`, Options{})
	})

	It("--async: 100 concurrent calls produce 100 timeline entries with strictly monotonic GlobalIndex 0..99", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}

var wg syncPkg.WaitGroup
for i := 0; i < 100; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.Send("x")
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
	t.Fatal("timed out waiting for goroutines")
}

tl := m.Calls().Timeline()
if len(tl) != 100 {
	t.Fatalf("len(Timeline()) = %d, want 100", len(tl))
}

indices := make([]int, 100)
for i, e := range tl {
	indices[i] = e.GlobalIndex
}
sortPkg.Ints(indices)
for i, v := range indices {
	if v != i {
		t.Fatalf("GlobalIndex[%d] = %d after sort, want %d — not strictly monotonic 0..99", i, v, i)
	}
}
`, Options{Async: true}, `import syncPkg "sync"`, `import timePkg "time"`, `import sortPkg "sort"`)
	})

	It("No-returns method body also records to the timeline", func() {
		runMockBehavior(`package svc

type Service interface {
	Tick()
}`, `
m := &ServiceMock{}
m.Tick()
m.Tick()
m.Tick()

tl := m.Calls().Timeline()
if len(tl) != 3 {
	t.Fatalf("len(Timeline()) = %d, want 3", len(tl))
}
for i, e := range tl {
	if e.Method != "Tick" || e.Index != i || e.GlobalIndex != i {
		t.Fatalf("tl[%d] = {Method=%q, Index=%d, GlobalIndex=%d}, want {Tick, %d, %d}",
			i, e.Method, e.Index, e.GlobalIndex, i, i)
	}
}
`, Options{})
	})

	// --- HappenedBefore / AssertOrder ---

	It("HappenedBefore is true when first call was recorded before second", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")

first := m.Calls().Send().At(0)
last := m.Calls().Send().Last()
if !first.HappenedBefore(last) {
	t.Fatalf("HappenedBefore(first, last) = false, want true")
}
`, Options{})
	})

	It("HappenedBefore is false when comparing a call to itself (strict less-than)", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
m := &ServiceMock{}
m.Send("a")

c := m.Calls().Send().Last()
if c.HappenedBefore(c) {
	t.Fatalf("HappenedBefore(c, c) = true, want false (strict less-than semantics)")
}
`, Options{})
	})

	It("HappenedBefore works across different methods", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
	Close() error
}`, `
m := &ServiceMock{}
m.Send("a")
m.Close()

sendCall := m.Calls().Send().Last()
closeCall := m.Calls().Close().Last()

if !sendCall.HappenedBefore(closeCall) {
	t.Fatalf("Send.HappenedBefore(Close) = false, want true")
}
if closeCall.HappenedBefore(sendCall) {
	t.Fatalf("Close.HappenedBefore(Send) = true, want false")
}
`, Options{})
	})

	It("AssertOrder is a no-op when entries are in order", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
type mockT struct {
	errs []string
}
func (mt *mockT) Helper() {}
func (mt *mockT) Errorf(format string, args ...any) {
	mt.errs = append(mt.errs, fmtSprintf(format, args...))
}

m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("c")

mt := &mockT{}
m.Calls().AssertOrder(mt, m.Calls().Send().At(0), m.Calls().Send().At(1), m.Calls().Send().At(2))
if len(mt.errs) != 0 {
	t.Fatalf("AssertOrder in-order: got %d error(s), want 0: %v", len(mt.errs), mt.errs)
}
`, Options{}, `import fmtPkg "fmt"`, `var fmtSprintf = fmtPkg.Sprintf`)
	})

	It("AssertOrder reports a violation when entries are out of order", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
type mockT struct {
	errs []string
}
func (mt *mockT) Helper() {}
func (mt *mockT) Errorf(format string, args ...any) {
	mt.errs = append(mt.errs, fmtSprintf(format, args...))
}

m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("c")

mt := &mockT{}
m.Calls().AssertOrder(mt, m.Calls().Send().At(2), m.Calls().Send().At(0))
if len(mt.errs) < 1 {
	t.Fatalf("AssertOrder out-of-order: expected at least 1 error, got 0")
}
found := false
for _, e := range mt.errs {
	if stringsPkg.Contains(e, "AssertOrder violated") {
		found = true
		break
	}
}
if !found {
	t.Fatalf("AssertOrder error %v does not contain \"AssertOrder violated\"", mt.errs)
}
`, Options{}, `import fmtPkg "fmt"`, `import stringsPkg "strings"`, `var fmtSprintf = fmtPkg.Sprintf`)
	})

	It("AssertOrder is a no-op with 0 or 1 entry", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) error
}`, `
type mockT struct {
	errs []string
}
func (mt *mockT) Helper() {}
func (mt *mockT) Errorf(format string, args ...any) {
	mt.errs = append(mt.errs, fmtSprintf(format, args...))
}

m := &ServiceMock{}
m.Send("a")

// 0 entries: no-op
mt0 := &mockT{}
m.Calls().AssertOrder(mt0)
if len(mt0.errs) != 0 {
	t.Fatalf("AssertOrder with 0 entries: got %d error(s), want 0", len(mt0.errs))
}

// 1 entry: no-op
mt1 := &mockT{}
m.Calls().AssertOrder(mt1, m.Calls().Send().Last())
if len(mt1.errs) != 0 {
	t.Fatalf("AssertOrder with 1 entry: got %d error(s), want 0", len(mt1.errs))
}
`, Options{}, `import fmtPkg "fmt"`, `var fmtSprintf = fmtPkg.Sprintf`)
	})

	It("AssertOrder treats identical (same GlobalIndex) entries as in-order", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
type mockT struct {
	errs []string
}
func (mt *mockT) Helper() {}
func (mt *mockT) Errorf(format string, args ...any) {
	mt.errs = append(mt.errs, fmtSprintf(format, args...))
}

m := &ServiceMock{}
m.Send("a")

// Same call passed twice — cur == prev branch, treated as in-order.
mt := &mockT{}
c := m.Calls().Send().Last()
m.Calls().AssertOrder(mt, c, c)
if len(mt.errs) != 0 {
	t.Fatalf("AssertOrder with identical entries: got %d error(s), want 0 (ties are in-order)", len(mt.errs))
}
`, Options{}, `import fmtPkg "fmt"`, `var fmtSprintf = fmtPkg.Sprintf`)
	})

	It("AssertOrder stops at first violation; reports exactly one error", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
type mockT struct {
	errs []string
}
func (mt *mockT) Helper() {}
func (mt *mockT) Errorf(format string, args ...any) {
	mt.errs = append(mt.errs, fmtSprintf(format, args...))
}

m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("c")
m.Send("d")

mt := &mockT{}
// Two distinct violations would be reachable if AssertOrder kept walking:
// [3, 0, 2, 1] — 0<3 violates AND 1<2 violates. Code returns after the
// first, so we expect exactly one Errorf.
m.Calls().AssertOrder(mt,
	m.Calls().Send().At(3),
	m.Calls().Send().At(0),
	m.Calls().Send().At(2),
	m.Calls().Send().At(1),
)
if len(mt.errs) != 1 {
	t.Fatalf("AssertOrder: got %d error(s), want exactly 1 (stops at first violation)", len(mt.errs))
}
`, Options{}, `import fmtPkg "fmt"`, `var fmtSprintf = fmtPkg.Sprintf`)
	})

	It("m.Reset() zeroes the GlobalIndex counter; next call gets GlobalIndex 0", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("c")

m.Reset()

m.Send("d")

c := m.Calls().Send().Last()
if c.GlobalIndex != 0 {
	t.Fatalf("after Reset + 1 call: GlobalIndex = %d, want 0 (counter must reset)", c.GlobalIndex)
}

tl := m.Calls().Timeline()
if len(tl) != 1 || tl[0].GlobalIndex != 0 {
	t.Fatalf("timeline after Reset + 1 call: len=%d, tl[0].GlobalIndex=%d, want len=1, GlobalIndex=0", len(tl), tl[0].GlobalIndex)
}
`, Options{})
	})

})
