package test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/ArtHustle/mock_gen"
)

// Conditional-stub behavior: When(pred) typed-predicate argument-matcher API.
// Phase 1: all 23 tests are expected to FAIL because When() does not exist on
// the generated mock yet. They compile fine here (outer binary); the inner
// `go test` spawned by runMockBehavior fails with "undefined: When".
var _ = Describe("Conditional stubs", func() {

	// --- Dispatch: basic ---

	It("When(p).Once() fires on matching call, then falls through to zero values", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("vip").Err(nil)

id1, err1 := m.Send("vip@x.com")
if id1 != "vip" || err1 != nil {
	t.Fatalf("first matching call: Send() = (%q, %v), want (vip, nil)", id1, err1)
}

id2, err2 := m.Send("vip@x.com")
if id2 != "" || err2 != nil {
	t.Fatalf("second call (Once consumed): Send() = (%q, %v), want (\"\", nil) — zero values", id2, err2)
}
`, Options{})
	})

	It("When(p).Once() does NOT fire on non-matching call; conditional stays pending", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("vip").Err(nil)

id, err := m.Send("other@x.com")
if id != "" || err != nil {
	t.Fatalf("non-matching call: Send() = (%q, %v), want (\"\", nil)", id, err)
}
if got := m.Calls().Send().PendingStubs(); got != 1 {
	t.Fatalf("PendingStubs() = %d, want 1 (conditional not consumed)", got)
}
`, Options{})
	})

	It("When(p).Always() fires on every matching call; non-matching call hits zero", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Always().Return().Id("special").Err(nil)

for i := 0; i < 3; i++ {
	got, err := m.Send("vip@x.com")
	if got != "special" || err != nil {
		t.Fatalf("matching call %d: Send() = (%q, %v), want (special, nil)", i, got, err)
	}
}

id, err := m.Send("other@x.com")
if id != "" || err != nil {
	t.Fatalf("non-matching call: Send() = (%q, %v), want (\"\", nil)", id, err)
}
`, Options{})
	})

	It("Captured conditional: Once then Always on one predicate — Once consumed first, Always fires after", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
cond := m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" })
cond.Once().Return().Id("first").Err(nil)
cond.Always().Return().Id("steady").Err(nil)

id1, _ := m.Send("vip@x.com")
if id1 != "first" {
	t.Fatalf("first matching call: Send() = %q, want first", id1)
}

for i := 0; i < 2; i++ {
	got, _ := m.Send("vip@x.com")
	if got != "steady" {
		t.Fatalf("matching call %d after Once consumed: Send() = %q, want steady", i+2, got)
	}
}
`, Options{})
	})

	It("Two conditionals with disjoint predicates route correctly; unmatched call returns zero", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "a@x.com" }).Always().Return().Id("A").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "b@x.com" }).Always().Return().Id("B").Err(nil)

idA, _ := m.Send("a@x.com")
if idA != "A" {
	t.Fatalf("a@x.com: Send() = %q, want A", idA)
}

idB, _ := m.Send("b@x.com")
if idB != "B" {
	t.Fatalf("b@x.com: Send() = %q, want B", idB)
}

idC, _ := m.Send("c@x.com")
if idC != "" {
	t.Fatalf("c@x.com: Send() = %q, want \"\" (no matching predicate)", idC)
}
`, Options{})
	})

	It("Conditional Once exhausted falls through to unconditional Always", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("once-special").Err(nil)
m.On().Send().Always().Return().Id("default").Err(nil)

id1, _ := m.Send("vip@x.com")
if id1 != "once-special" {
	t.Fatalf("first vip call: Send() = %q, want once-special", id1)
}

id2, _ := m.Send("vip@x.com")
if id2 != "default" {
	t.Fatalf("second vip call (Once consumed): Send() = %q, want default", id2)
}
`, Options{})
	})

	It("Conditional Once exhausted falls through to unconditional sticky (Return)", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("once-special").Err(nil)
m.On().Send().Return().Id("sticky").Err(nil)

id1, _ := m.Send("vip@x.com")
if id1 != "once-special" {
	t.Fatalf("first vip call: Send() = %q, want once-special", id1)
}

id2, _ := m.Send("vip@x.com")
if id2 != "sticky" {
	t.Fatalf("second vip call (Once consumed): Send() = %q, want sticky", id2)
}
`, Options{})
	})

	It("Conditional Once exhausted with no unconditional registered returns zero values", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("once").Err(nil)

m.Send("vip@x.com") // consumes the Once

id, err := m.Send("vip@x.com")
if id != "" || err != nil {
	t.Fatalf("matching call after Once consumed: Send() = (%q, %v), want (\"\", nil)", id, err)
}
`, Options{})
	})

	It("Multiple Once entries on one captured conditional are consumed FIFO", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
cond := m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" })
cond.Once().Return().Id("a").Err(nil)
cond.Once().Return().Id("b").Err(nil)
cond.Once().Return().Id("c").Err(nil)

for i, want := range []string{"a", "b", "c"} {
	got, _ := m.Send("vip@x.com")
	if got != want {
		t.Fatalf("call %d: Send() = %q, want %q (FIFO order)", i, got, want)
	}
}
`, Options{})
	})

	It("Multiple Always armings on one conditional — last-wins", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
cond := m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" })
cond.Always().Return().Id("first").Err(nil)
cond.Always().Return().Id("second").Err(nil)

for i := 0; i < 3; i++ {
	got, _ := m.Send("vip@x.com")
	if got != "second" {
		t.Fatalf("call %d: Send() = %q, want second (last Always wins)", i, got)
	}
}
`, Options{})
	})

	It("Captured conditional namespace stays valid across subsequent When() registrations", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
condA := m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "a@x.com" })
// Register two more conditionals AFTER capturing condA — these grow the
// backing slice and would invalidate any pointer-based handle.
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "b@x.com" }).Always().Return().Id("B").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "c@x.com" }).Always().Return().Id("C").Err(nil)
condA.Once().Return().Id("A").Err(nil)

idA, _ := m.Send("a@x.com")
if idA != "A" {
	t.Fatalf("a@x.com: Send() = %q, want A (captured conditional must still address its original slot)", idA)
}
idB, _ := m.Send("b@x.com")
if idB != "B" {
	t.Fatalf("b@x.com: Send() = %q, want B", idB)
}
`, Options{})
	})

	It("Call log records Param and Return when a conditional fires", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("vip").Err(nil)

m.Send("vip@x.com")

c := m.Calls().Send().Last()
if c.Param.Email != "vip@x.com" {
	t.Fatalf("call log Param.Email = %q, want vip@x.com", c.Param.Email)
}
if c.Return.Id != "vip" {
	t.Fatalf("call log Return.Id = %q, want vip", c.Return.Id)
}
`, Options{})
	})

	It("When() on a zero-param method: predicate receives an empty Param struct", func() {
		runMockBehavior(`package svc

type Service interface {
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
// Predicate sees the empty Param struct — only useful as a registration that
// always-matches (or never-matches) for a no-args method. Guards against
// template regressions in __mockParam construction for empty param structs.
m.On().Get().When(func(p ServiceMockGetParam) bool { return true }).Always().Return().Id("matched").Err(nil)

id, _ := m.Get()
if id != "matched" {
	t.Fatalf("zero-param When().Always(): Get() = %q, want matched", id)
}
`, Options{})
	})

	It("Conditional Always fires before unconditional Once on matching call", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Always().Return().Id("special").Err(nil)
m.On().Send().Once().Return().Id("uncond-once").Err(nil)

idVip, _ := m.Send("vip@x.com")
if idVip != "special" {
	t.Fatalf("matching call: Send() = %q, want special (conditional wins)", idVip)
}
`, Options{})
	})

	It("Conditional registered but predicate never matches: unconditional path fires", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "never@x.com" }).Always().Return().Id("conditional").Err(nil)
m.On().Send().Once().Return().Id("uncond-once").Err(nil)

id, _ := m.Send("x@x.com")
if id != "uncond-once" {
	t.Fatalf("non-matching call with conditional registered: Send() = %q, want uncond-once (the conditional walk must fall through to unconditional)", id)
}
`, Options{})
	})

	// --- Overlap panic ---

	It("Two overlapping When predicates panic on matching call; message names indices 0 and 1", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x@x.com" }).Always().Return().Id("first").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x@x.com" }).Always().Return().Id("second").Err(nil)

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when two When predicates both match")
	}
	msg, _ := r.(string)
	if !stringsPkg.Contains(msg, "matched 2 When predicates") || !stringsPkg.Contains(msg, "indices 0 and 1") {
		t.Fatalf("panic message %q missing expected substrings", msg)
	}
}()
m.Send("x@x.com")
`, Options{}, `import stringsPkg "strings"`)
	})

	It("Overlap panic does NOT fire when no predicate matches", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x@x.com" }).Always().Return().Id("first").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x@x.com" }).Always().Return().Id("second").Err(nil)

panicked := false
func() {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	m.Send("nomatch@x.com")
}()

if panicked {
	t.Fatal("unexpected panic on non-matching call")
}
`, Options{})
	})

	It("Three-way overlap panic names indices 0 and 1 and reports 3 matched predicates", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x@x.com" }).Always().Return().Id("a").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x@x.com" }).Always().Return().Id("b").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x@x.com" }).Always().Return().Id("c").Err(nil)

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic on three-way overlap")
	}
	msg, _ := r.(string)
	if !stringsPkg.Contains(msg, "matched 3 When predicates") || !stringsPkg.Contains(msg, "indices 0 and 1") {
		t.Fatalf("panic message %q missing expected substrings", msg)
	}
}()
m.Send("x@x.com")
`, Options{}, `import stringsPkg "strings"`)
	})

	It("Overlap panic reports the actual matching indices, not always 0 and 1", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
// Predicate 0 does not match; predicates 1 and 2 do.
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "miss@x.com" }).Always().Return().Id("zero").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x@x.com" }).Always().Return().Id("one").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x@x.com" }).Always().Return().Id("two").Err(nil)

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when two non-zero-indexed predicates overlap")
	}
	msg, _ := r.(string)
	if !stringsPkg.Contains(msg, "matched 2 When predicates") || !stringsPkg.Contains(msg, "indices 1 and 2") {
		t.Fatalf("panic message %q must report \"indices 1 and 2\"", msg)
	}
}()
m.Send("x@x.com")
`, Options{}, `import stringsPkg "strings"`)
	})

	// --- PendingStubs / Verify ---

	It("Conditional Once counts in PendingStubs alongside unconditional Once entries", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().Once().Return().Id("a").Err(nil)
m.On().Send().Once().Return().Id("b").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("vip").Err(nil)

if got := m.Calls().Send().PendingStubs(); got != 3 {
	t.Fatalf("per-method PendingStubs() = %d, want 3", got)
}
if got := m.Calls().PendingStubs(); got != 3 {
	t.Fatalf("mock-level PendingStubs() = %d, want 3", got)
}
`, Options{})
	})

	It("Conditional Always does NOT count in PendingStubs", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Always().Return().Id("vip").Err(nil)

if got := m.Calls().Send().PendingStubs(); got != 0 {
	t.Fatalf("per-method PendingStubs() = %d, want 0 (Always is not queued)", got)
}
if got := m.Calls().PendingStubs(); got != 0 {
	t.Fatalf("mock-level PendingStubs() = %d, want 0", got)
}
`, Options{})
	})

	It("Consumed conditional Once decrements PendingStubs", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("vip").Err(nil)

if got := m.Calls().Send().PendingStubs(); got != 1 {
	t.Fatalf("before call: PendingStubs() = %d, want 1", got)
}

m.Send("vip@x.com")

if got := m.Calls().Send().PendingStubs(); got != 0 {
	t.Fatalf("after matching call: PendingStubs() = %d, want 0", got)
}
`, Options{})
	})

	It("Verify(t) reports conditional-pending stubs as unconsumed", func() {
		// harness imports `fmt` under an alias to avoid colliding with names
		// from the generated mock — `var fmtSprintf = fmtPkg.Sprintf` exposes
		// it to the test body, which can't write `fmt.Sprintf` directly.
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

mt := &mockT{}
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("vip").Err(nil)

m.Calls().Verify(mt)

if len(mt.errs) == 0 {
	t.Fatal("expected Verify to call Errorf for pending conditional stub")
}
found := false
for _, e := range mt.errs {
	if stringsPkg.Contains(e, "1 unconsumed stubs") && stringsPkg.Contains(e, "Send") {
		found = true
		break
	}
}
if !found {
	t.Fatalf("Verify error %v does not contain \"1 unconsumed stubs\" and method name \"Send\"", mt.errs)
}
`, Options{}, `import fmtPkg "fmt"`, `import stringsPkg "strings"`, `var fmtSprintf = fmtPkg.Sprintf`)
	})

	// --- Reset ---

	It("On().Send().Reset() drains conditional stubs for that method", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "a@x.com" }).Once().Return().Id("A").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "b@x.com" }).Always().Return().Id("B").Err(nil)

m.On().Send().Reset()

if got := m.Calls().Send().PendingStubs(); got != 0 {
	t.Fatalf("after Reset: PendingStubs() = %d, want 0", got)
}

id, _ := m.Send("a@x.com")
if id != "" {
	t.Fatalf("after Reset: Send(a) = %q, want \"\" (conditionals cleared)", id)
}
`, Options{})
	})

	It("Captured When() handle used after Reset() panics with a project-specific message", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
cond := m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" })
m.On().Send().Reset()

defer func() {
	r := recover()
	if r == nil {
		t.Fatal("expected panic when stale When() handle is used after Reset")
	}
	msg, _ := r.(string)
	if !stringsPkg.Contains(msg, "stale When() handle") {
		t.Fatalf("panic message %q missing \"stale When() handle\"", msg)
	}
}()
cond.Once().Return().Id("doomed").Err(nil)
`, Options{}, `import stringsPkg "strings"`)
	})

	// The When() handle shares the per-method config type, so it also
	// exposes the whole-method operations; they must not silently rewrite
	// the unconditional config.
	It("Whole-method operations on a When() handle panic", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
cond := m.On().Send().When(func(p ServiceMockSendParam) bool { return true })
for name, op := range map[string]func(){
	"ReturnAll()":    func() { cond.ReturnAll("x", nil) },
	"ReturnCustom()": func() { cond.ReturnCustom(func(string) (string, error) { return "", nil }) },
	"Return()":       func() { cond.Return() },
	"When()":         func() { cond.When(func(ServiceMockSendParam) bool { return true }) },
	"Reset()":        func() { cond.Reset() },
} {
	func() {
		defer func() {
			msg, _ := recover().(string)
			if msg != "ServiceMock.Send: "+name+" is not available on a When() handle; arm it with Once() or Always()" {
				t.Errorf("%s: panic = %q", name, msg)
			}
		}()
		op()
	}()
}
if got := m.Calls().Send().PendingStubs(); got != 0 {
	t.Fatalf("PendingStubs() = %d, want 0", got)
}
`, Options{})
	})

	It("m.On().Reset() drains conditionals across all methods (config-wide reset)", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return true }).Once().Return().Id("S").Err(nil)
m.On().Get().When(func(p ServiceMockGetParam) bool { return true }).Once().Return().Id("G").Err(nil)

m.On().Reset()

if got := m.Calls().Send().PendingStubs(); got != 0 {
	t.Fatalf("Send PendingStubs() = %d, want 0", got)
}
if got := m.Calls().Get().PendingStubs(); got != 0 {
	t.Fatalf("Get PendingStubs() = %d, want 0", got)
}
`, Options{})
	})

	It("m.Reset() drains conditionals across all methods", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
	Get() (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("S").Err(nil)
m.On().Get().When(func(p ServiceMockGetParam) bool { return true }).Always().Return().Id("G").Err(nil)

m.Reset()

if got := m.Calls().Send().PendingStubs(); got != 0 {
	t.Fatalf("after Reset: Send PendingStubs() = %d, want 0", got)
}
if got := m.Calls().Get().PendingStubs(); got != 0 {
	t.Fatalf("after Reset: Get PendingStubs() = %d, want 0", got)
}
`, Options{})
	})

	// --- Async / concurrency ---

	It("--async: overlap panic still fires under concurrent calls", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x" }).Always().Return().Id("first").Err(nil)
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "x" }).Always().Return().Id("second").Err(nil)

panics := make(chan string, 100)
var wg syncPkg.WaitGroup
for i := 0; i < 100; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				msg, _ := r.(string)
				panics <- msg
			}
		}()
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
close(panics)

count := 0
for msg := range panics {
	if !stringsPkg.Contains(msg, "matched 2 When predicates") {
		t.Fatalf("goroutine panicked with unexpected message %q", msg)
	}
	count++
}
if count != 100 {
	t.Fatalf("panicked goroutines = %d, want 100 (overlap is deterministic per call)", count)
}
`, Options{Async: true}, `import syncPkg "sync"`, `import timePkg "time"`, `import stringsPkg "strings"`)
	})

	It("--async: a panicking predicate releases the lock", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { panic("boom") }).Always().Id("x")

func() {
	defer func() { recover() }()
	m.Send("a")
}()

done := make(chan int)
go func() { done <- m.Calls().Send().Count() }()
select {
case n := <-done:
	if n != 1 {
		t.Fatalf("Count() = %d, want 1", n)
	}
case <-timePkg.After(timePkg.Second):
	t.Fatal("mock still locked after a predicate panicked")
}
`, Options{Async: true}, `import timePkg "time"`)
	})

	It("--async: predicate evaluation is race-clean with 100 concurrent calls", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return true }).Always().Return().Id("ok").Err(nil)

var wg syncPkg.WaitGroup
for i := 0; i < 100; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.Send("anything")
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

calls := m.Calls().Send().All()
if len(calls) != 100 {
	t.Fatalf("recorded %d calls, want 100", len(calls))
}
for i, c := range calls {
	if c.Return.Id != "ok" {
		t.Fatalf("call %d Return.Id = %q, want ok", i, c.Return.Id)
	}
}
`, Options{Async: true}, `import syncPkg "sync"`, `import timePkg "time"`)
	})

	// --- Mixing rules: conditional ↔ unconditional ---

	It("Conditional Once + unconditional Once: no panic at registration; each fires correctly", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("special").Err(nil)
m.On().Send().Once().Return().Id("uncond").Err(nil)

idVip, _ := m.Send("vip@x.com")
if idVip != "special" {
	t.Fatalf("matching call: Send() = %q, want special (conditional wins)", idVip)
}

idOther, _ := m.Send("other@x.com")
if idOther != "uncond" {
	t.Fatalf("non-matching call: Send() = %q, want uncond", idOther)
}
`, Options{})
	})

	It("Conditional Always + unconditional Always: no panic; conditional wins on match", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Always().Return().Id("special").Err(nil)
m.On().Send().Always().Return().Id("default").Err(nil)

idVip, _ := m.Send("vip@x.com")
if idVip != "special" {
	t.Fatalf("matching call: Send() = %q, want special", idVip)
}

idOther, _ := m.Send("other@x.com")
if idOther != "default" {
	t.Fatalf("non-matching call: Send() = %q, want default", idOther)
}
`, Options{})
	})

	It("Unconditional ReturnCustom handler wins over conditional static stub on matching call", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Always().Return().Id("static-conditional").Err(nil)
m.On().Send().ReturnCustom(func(email string) (string, error) {
	return "handler-" + email, nil
})

// Even on a matching call, ReturnCustom takes precedence — its handler
// sees the actual param and returns its own value.
id, err := m.Send("vip@x.com")
if id != "handler-vip@x.com" || err != nil {
	t.Fatalf("matching call: Send() = (%q, %v), want (handler-vip@x.com, nil) — handler must win", id, err)
}
`, Options{})
	})

	It("Unconditional ReturnCustom + conditional Once: handler wins; the Once is NOT silently consumed", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("conditional-once").Err(nil)
m.On().Send().ReturnCustom(func(email string) (string, error) {
	return "handler-" + email, nil
})

// Matching call: handler must win (per spec), and the conditional Once
// must NOT be consumed — its value never reached the caller.
id, _ := m.Send("vip@x.com")
if id != "handler-vip@x.com" {
	t.Fatalf("matching call: Send() = %q, want handler-vip@x.com", id)
}
if got := m.Calls().Send().PendingStubs(); got != 1 {
	t.Fatalf("after handler-overriden call: PendingStubs() = %d, want 1 (Once must not be consumed when handler runs)", got)
}
`, Options{})
	})

	It("Conditional Once + unconditional sticky Return: no panic; conditional wins on match, sticky on non-match", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Once().Return().Id("special").Err(nil)
m.On().Send().Return().Id("sticky").Err(nil)

idVip, _ := m.Send("vip@x.com")
if idVip != "special" {
	t.Fatalf("matching call: Send() = %q, want special", idVip)
}

idOther, _ := m.Send("other@x.com")
if idOther != "sticky" {
	t.Fatalf("non-matching call: Send() = %q, want sticky", idOther)
}
`, Options{})
	})

})
