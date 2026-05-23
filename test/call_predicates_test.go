package test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/ArtHustle/mock_gen"
)

// Call-predicate behavior: Any(pred) and Filter(pred) per-method helpers.
// Phase 1: all 9 tests are expected to FAIL because Any/Filter do not exist
// on the generated mock yet. They compile fine here (outer binary); the inner
// `go test` spawned by runMockBehavior fails with undefined identifiers.
var _ = Describe("Call predicates (Any/Filter)", func() {

	It("Any returns true when at least one call matches the predicate", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("vip@x.com")
m.Send("c")

if !m.Calls().Send().Any(func(c ServiceMockSendCall) bool { return c.Param.Email == "vip@x.com" }) {
	t.Fatal("Any(vip@x.com) = false, want true")
}
`, Options{})
	})

	It("Any returns false when no call matches the predicate", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("vip@x.com")
m.Send("c")

if m.Calls().Send().Any(func(c ServiceMockSendCall) bool { return c.Param.Email == "notpresent@x.com" }) {
	t.Fatal("Any(notpresent) = true, want false")
}
`, Options{})
	})

	It("Any returns false on an empty call log", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}

if m.Calls().Send().Any(func(c ServiceMockSendCall) bool { return true }) {
	t.Fatal("Any on empty log = true, want false")
}
`, Options{})
	})

	It("Filter returns matching calls in invocation order", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("a")
m.Send("b")
m.Send("a")

out := m.Calls().Send().Filter(func(c ServiceMockSendCall) bool { return c.Param.Email == "a" })
if len(out) != 3 {
	t.Fatalf("Filter(a): len = %d, want 3", len(out))
}
wantIdx := []int{0, 2, 4}
for i, c := range out {
	if c.Index != wantIdx[i] {
		t.Fatalf("Filter(a)[%d].Index = %d, want %d", i, c.Index, wantIdx[i])
	}
}
`, Options{})
	})

	It("Filter returns a defensive copy; mutating the returned slice does not affect subsequent calls", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.Send("a")

out := m.Calls().Send().Filter(func(c ServiceMockSendCall) bool { return true })
out[0].Param.Email = "X"

out2 := m.Calls().Send().Filter(func(c ServiceMockSendCall) bool { return true })
if out2[0].Param.Email != "a" {
	t.Fatalf("Filter defensive copy broken: out2[0].Param.Email = %q, want a", out2[0].Param.Email)
}
`, Options{})
	})

	It("Filter returns a non-nil empty slice when no calls match", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.Send("a")

out := m.Calls().Send().Filter(func(c ServiceMockSendCall) bool { return c.Param.Email == "z" })
if len(out) != 0 {
	t.Fatalf("Filter(no match): len = %d, want 0", len(out))
}
if out == nil {
	t.Fatal("Filter(no match) returned nil slice, want non-nil empty slice")
}
`, Options{})
	})

	It("Predicate can read Return values recorded in the call log", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.On().Send().When(func(p ServiceMockSendParam) bool { return p.Email == "vip@x.com" }).Always().Return().Id("vip").Err(nil)
m.Send("vip@x.com")
m.Send("other@x.com")

out := m.Calls().Send().Filter(func(c ServiceMockSendCall) bool { return c.Return.Id == "vip" })
if len(out) != 1 {
	t.Fatalf("Filter by Return.Id=vip: len = %d, want 1", len(out))
}
if out[0].Param.Email != "vip@x.com" {
	t.Fatalf("Filter by Return.Id: Param.Email = %q, want vip@x.com", out[0].Param.Email)
}
`, Options{})
	})

	It("--async: predicate calling back into the mock does not deadlock", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.Send("a")
m.Send("b")
m.Send("c")

done := make(chan struct{})
go func() {
	// The predicate calls back into the mock. If Filter holds the lock
	// during predicate evaluation, this deadlocks under --async.
	m.Calls().Send().Filter(func(c ServiceMockSendCall) bool {
		return m.Calls().Send().Count() > 0
	})
	close(done)
}()
select {
case <-done:
case <-timePkg.After(2 * timePkg.Second):
	t.Fatal("deadlock: Filter did not return within 2s — predicate called under lock")
}
`, Options{Async: true}, `import timePkg "time"`)
	})

	It("nil predicate passed to Any panics with substring \"predicate is nil\"", func() {
		runMockBehavior(`package svc

type Service interface {
	Send(email string) (id string, err error)
}`, `
m := &ServiceMock{}
m.Send("a")

func() {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Any(nil): expected panic, got none")
		}
		msg, _ := r.(string)
		if !stringsPkg.Contains(msg, "predicate is nil") {
			t.Fatalf("Any(nil) panic message %q missing \"predicate is nil\"", msg)
		}
	}()
	m.Calls().Send().Any(nil)
}()

func() {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Filter(nil): expected panic, got none")
		}
		msg, _ := r.(string)
		if !stringsPkg.Contains(msg, "predicate is nil") {
			t.Fatalf("Filter(nil) panic message %q missing \"predicate is nil\"", msg)
		}
	}()
	m.Calls().Send().Filter(nil)
}()
`, Options{}, `import stringsPkg "strings"`)
	})

})
