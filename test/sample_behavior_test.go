package test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"

	rt "github.com/ArtHustle/mock_gen/test/dummy/runtime_fixture"
	rtmock "github.com/ArtHustle/mock_gen/test/dummy/runtime_fixture/test"
)

// Sample-mock behavior suite. Drives a shape-rich generated mock through
// every v0.13/v0.14 method on the specific method shapes that exercise
// distinct codegen branches (variadic, function-typed Param, multi-return,
// embedded interface, etc.).
//
// Per-feature contract tests (timeline_test.go, call_predicates_test.go,
// conditional_test.go, sequenced_stubs_test.go) remain authoritative for
// "what does method X guarantee?". This file answers "does method X work
// across the cross-shape matrix a real mock has to handle?".
//
// Operational rule (decisions.md D6): an assertion that would also pass
// on the Send(email string) error shape used by per-feature tests does
// NOT belong here — it belongs in the per-feature file.

// sampleMockT is a minimal VerifyT for the AssertOrder + Verify tests.
type sampleMockT struct {
	errs []string
}

func (mt *sampleMockT) Helper() {}
func (mt *sampleMockT) Errorf(format string, args ...any) {
	mt.errs = append(mt.errs, fmt.Sprintf(format, args...))
}

var _ = Describe("Sample mock: Timeline + ordering", func() {

	It("Timeline names every method by its exact source-name with monotonic GlobalIndex", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		m.Close()
		m.Activate(ctx, &rt.Trigger{ID: "x"})
		m.Fetch(ctx, "a", "b")
		m.Stats()
		m.History(time.Second)

		tl := m.Calls().Timeline()
		if len(tl) != 5 {
			Fail(fmt.Sprintf("Timeline length = %d, want 5", len(tl)))
		}
		wantMethods := []string{"Close", "Activate", "Fetch", "Stats", "History"}
		for i, want := range wantMethods {
			if tl[i].Method != want || tl[i].GlobalIndex != i {
				Fail(fmt.Sprintf("tl[%d] = {Method=%q, GlobalIndex=%d}, want {%q, %d}",
					i, tl[i].Method, tl[i].GlobalIndex, want, i))
			}
		}
	})

	It("AssertOrder accepts call handles from three different methods", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		m.Activate(ctx, &rt.Trigger{})
		m.Fetch(ctx, "id")
		m.Close()

		mt := &sampleMockT{}
		m.Calls().AssertOrder(mt,
			m.Calls().Activate().Last(),
			m.Calls().Fetch().Last(),
			m.Calls().Close().Last(),
		)
		if len(mt.errs) != 0 {
			Fail(fmt.Sprintf("AssertOrder over three methods called Errorf: %v", mt.errs))
		}
	})

	It("HappenedBefore between a variadic Call and a named-returns Call", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		m.Fetch(ctx, "a", "b")
		m.Stats()

		fetchCall := m.Calls().Fetch().Last()
		statsCall := m.Calls().Stats().Last()
		if !fetchCall.HappenedBefore(statsCall) {
			Fail("Fetch.HappenedBefore(Stats) = false, want true")
		}
		if statsCall.HappenedBefore(fetchCall) {
			Fail("Stats.HappenedBefore(Fetch) = true, want false")
		}
	})

	It("Per-method Reset preserves timeline; mock-level Reset clears it", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		m.Activate(ctx, &rt.Trigger{})
		m.Activate(ctx, &rt.Trigger{})
		m.Close()

		m.Calls().Activate().Reset()
		if got := len(m.Calls().Timeline()); got != 3 {
			Fail(fmt.Sprintf("after per-method Reset: Timeline len = %d, want 3 (preserved)", got))
		}
		if got := m.Calls().Activate().Count(); got != 0 {
			Fail(fmt.Sprintf("after per-method Reset: Activate Count = %d, want 0", got))
		}

		m.Reset()
		if got := len(m.Calls().Timeline()); got != 0 {
			Fail(fmt.Sprintf("after mock Reset: Timeline len = %d, want 0", got))
		}
	})
})

var _ = Describe("Sample mock: Predicates (Any/Filter)", func() {

	It("Filter reads variadic Param.Ids slice", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		m.Fetch(ctx, "a", "b", "c")
		m.Fetch(ctx, "x")

		out := m.Calls().Fetch().Filter(func(c rtmock.RuntimeMockFetchCall) bool {
			return len(c.Param.Ids) == 3
		})
		if len(out) != 1 || len(out[0].Param.Ids) != 3 {
			Fail(fmt.Sprintf("Filter(variadic): got %d matches, want 1 with Ids length 3", len(out)))
		}
	})

	It("Filter reads all three Return slots of Fetch", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		m.On().Fetch().Always().Return().Events([]rt.Event{{Topic: "t"}}).Total(7).Err(nil)
		m.Fetch(ctx, "a")
		m.Fetch(ctx, "b")

		out := m.Calls().Fetch().Filter(func(c rtmock.RuntimeMockFetchCall) bool {
			return c.Return.Total == 7 && c.Return.Err == nil && len(c.Return.Events) == 1
		})
		if len(out) != 2 {
			Fail(fmt.Sprintf("Filter(three-return): got %d matches, want 2", len(out)))
		}
	})

	It("Any inspects function-typed Param.Handler field", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		_, _ = m.Subscribe(ctx, []string{"t"}, func(rt.Event) error { return nil })

		got := m.Calls().Subscribe().Any(func(c rtmock.RuntimeMockSubscribeCall) bool {
			return c.Param.Handler != nil
		})
		if !got {
			Fail("Any(Handler!=nil) = false, want true")
		}
	})

	It("Filter on never-called no-returns method returns non-nil empty slice", func() {
		m := rtmock.NewRuntimeMock()
		// Close is never called.
		out := m.Calls().Close().Filter(func(c rtmock.RuntimeMockCloseCall) bool { return true })
		if len(out) != 0 || out == nil {
			Fail(fmt.Sprintf("Filter on empty no-returns log: len=%d, nil=%v, want len=0 non-nil", len(out), out == nil))
		}
	})
})

var _ = Describe("Sample mock: Conditional dispatch (When)", func() {

	It("When(pred) on a slice-param method (Fetch) dispatches by slice content", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		m.On().Fetch().When(func(p rtmock.RuntimeMockFetchParam) bool {
			return len(p.Ids) == 0
		}).Always().Return().Total(-1).Err(nil)

		_, total1, _ := m.Fetch(ctx) // no ids → matches
		if total1 != -1 {
			Fail(fmt.Sprintf("Fetch() (no ids): total = %d, want -1 (conditional)", total1))
		}
		_, total2, _ := m.Fetch(ctx, "x") // has ids → falls through
		if total2 != 0 {
			Fail(fmt.Sprintf("Fetch(x): total = %d, want 0 (zero fallthrough)", total2))
		}
	})

	It("When(pred).Once() on an embedded-interface method (Close) dispatches and consumes", func() {
		m := rtmock.NewRuntimeMock()
		closed := errors.New("closed")
		m.On().Close().When(func(p rtmock.RuntimeMockCloseParam) bool { return true }).
			Once().Return().Error(closed)

		if err := m.Close(); !errors.Is(err, closed) {
			Fail(fmt.Sprintf("Close() #1: err = %v, want closed", err))
		}
		if err := m.Close(); err != nil {
			Fail(fmt.Sprintf("Close() #2: err = %v, want nil (Once consumed)", err))
		}
	})

	It("Overlap panic fires on a function-typed-param method (Subscribe)", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		m.On().Subscribe().When(func(p rtmock.RuntimeMockSubscribeParam) bool {
			return p.Handler != nil
		}).Always().Return().Err(errors.New("a"))
		m.On().Subscribe().When(func(p rtmock.RuntimeMockSubscribeParam) bool {
			return p.Handler != nil
		}).Always().Return().Err(errors.New("b"))

		defer func() {
			r := recover()
			if r == nil {
				Fail("expected overlap panic, got none")
			}
			msg, _ := r.(string)
			if !strings.Contains(msg, "matched 2 When predicates") {
				Fail(fmt.Sprintf("panic message %q missing \"matched 2 When predicates\"", msg))
			}
		}()
		_, _ = m.Subscribe(ctx, []string{"t"}, func(rt.Event) error { return nil })
	})
})

var _ = Describe("Sample mock: Custom return handlers", func() {

	It("ReturnCustom handler on variadic Fetch receives []string", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		called := false
		gotIds := []string(nil)
		m.On().Fetch().ReturnCustom(func(_ context.Context, ids []string) ([]rt.Event, int, error) {
			called = true
			gotIds = ids
			return nil, 0, nil
		})

		_, _, _ = m.Fetch(ctx, "a", "b", "c")
		if !called {
			Fail("ReturnCustom handler not invoked")
		}
		if len(gotIds) != 3 || gotIds[0] != "a" || gotIds[2] != "c" {
			Fail(fmt.Sprintf("handler received ids=%v, want [a b c]", gotIds))
		}
	})

	It("ReturnCustom handler on Stats emits all three returns", func() {
		m := rtmock.NewRuntimeMock()
		m.On().Stats().ReturnCustom(func() (int, int, error) {
			return 11, 22, nil
		})

		a, q, err := m.Stats()
		if a != 11 || q != 22 || err != nil {
			Fail(fmt.Sprintf("Stats() = (%d, %d, %v), want (11, 22, nil)", a, q, err))
		}
	})
})

var _ = Describe("Sample mock: Sequenced stubs + per-slot setters", func() {

	It("Per-slot Once chain sets Events + Total + Err in one statement", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		events := []rt.Event{{Topic: "x"}, {Topic: "y"}}
		m.On().Fetch().Once().Return().Events(events).Total(2).Err(nil)

		got, total, err := m.Fetch(ctx)
		if total != 2 || err != nil || len(got) != 2 {
			Fail(fmt.Sprintf("Fetch() = (%v, %d, %v), want (2 events, 2, nil)", got, total, err))
		}
	})

	It("Always per-slot setter accepts a slice value", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		events := []rt.Event{{Topic: "steady"}}
		m.On().Fetch().Always().Return().Events(events)

		for i := 0; i < 3; i++ {
			got, _, _ := m.Fetch(ctx)
			if len(got) != 1 || got[0].Topic != "steady" {
				Fail(fmt.Sprintf("Fetch() #%d Events = %v, want one {steady}", i, got))
			}
		}
	})

	It("ReturnAll positional ordering on three-return Fetch", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		events := []rt.Event{{Topic: "p"}}
		m.On().Fetch().ReturnAll(events, 99, nil)

		got, total, err := m.Fetch(ctx)
		if total != 99 || err != nil || len(got) != 1 || got[0].Topic != "p" {
			Fail(fmt.Sprintf("Fetch() = (%v, %d, %v), want ([{p}], 99, nil)", got, total, err))
		}
	})

	It("Per-slot setter writes the type-synthesized field names for unnamed multi-returns (Publish)", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		published := rt.Event{Topic: "out", Payload: []byte("p")}
		m.On().Publish().Always().Return().Event(published).Error(nil)

		ev, err := m.Publish(ctx, "t", []byte("body"))
		if err != nil || ev.Topic != "out" || string(ev.Payload) != "p" {
			Fail(fmt.Sprintf("Publish() = (%+v, %v), want (Event{Topic:out, Payload:p}, nil)", ev, err))
		}
	})

	It("Per-slot setter accepts a func-typed return value (Subscribe.Cancel)", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		cancelled := false
		cancelFn := func() { cancelled = true }
		m.On().Subscribe().Always().Return().Cancel(cancelFn).Err(nil)

		cancel, err := m.Subscribe(ctx, []string{"t"}, func(rt.Event) error { return nil })
		if err != nil {
			Fail(fmt.Sprintf("Subscribe() err = %v, want nil", err))
		}
		if cancel == nil {
			Fail("Subscribe() returned nil cancel func, want non-nil")
		}
		cancel()
		if !cancelled {
			Fail("returned cancel func did not invoke the armed value")
		}
	})
})

var _ = Describe("Sample mock: Pending + Verify across many methods", func() {

	It("PendingStubs aggregates per-method and mock-level across 3+ methods", func() {
		m := rtmock.NewRuntimeMock()
		m.On().Activate().Once().Return().Error(nil)
		m.On().Fetch().Once().Return().Total(1)
		m.On().Fetch().Once().Return().Total(2)
		m.On().Close().Once().Return().Error(nil)
		m.On().Close().Once().Return().Error(nil)
		m.On().Close().Once().Return().Error(nil)

		if got := m.Calls().Activate().PendingStubs(); got != 1 {
			Fail(fmt.Sprintf("Activate PendingStubs = %d, want 1", got))
		}
		if got := m.Calls().Fetch().PendingStubs(); got != 2 {
			Fail(fmt.Sprintf("Fetch PendingStubs = %d, want 2", got))
		}
		if got := m.Calls().Close().PendingStubs(); got != 3 {
			Fail(fmt.Sprintf("Close PendingStubs = %d, want 3", got))
		}
		if got := m.Calls().PendingStubs(); got != 6 {
			Fail(fmt.Sprintf("mock-level PendingStubs = %d, want 6", got))
		}
	})

	It("Verify(t) breakdown names each method with pending stubs", func() {
		m := rtmock.NewRuntimeMock()
		m.On().Activate().Once().Return().Error(nil)
		m.On().Fetch().Once().Return().Total(1)
		m.On().Close().Once().Return().Error(nil)

		mt := &sampleMockT{}
		m.Calls().Verify(mt)
		if len(mt.errs) == 0 {
			Fail("Verify did not call Errorf")
		}
		joined := strings.Join(mt.errs, " | ")
		for _, name := range []string{"Activate", "Fetch", "Close"} {
			if !strings.Contains(joined, name) {
				Fail(fmt.Sprintf("Verify message %q missing method %q", joined, name))
			}
		}
	})
})

var _ = Describe("Sample mock: Legacy inspection on shape-rich CallTypes", func() {

	It("Last round-trips a function-typed Param.Handler field", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()
		_, _ = m.Subscribe(ctx, []string{"t"}, func(rt.Event) error { return nil })

		last := m.Calls().Subscribe().Last()
		if last.Param.Handler == nil {
			Fail("Subscribe.Last().Param.Handler = nil, want non-nil")
		}
	})

})

var _ = Describe("Sample mock: Async waiters under method diversity", func() {

	It("WaitForCount(5, ...) on one method while other methods receive concurrent calls", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()

		var wg sync.WaitGroup
		// 5 Fetch calls + noise on Activate and Close, all concurrent.
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, _ = m.Fetch(ctx, "id")
			}()
		}
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = m.Activate(ctx, &rt.Trigger{})
				_ = m.Close()
			}()
		}

		if err := m.Calls().Fetch().WaitForCount(5, 2*time.Second); err != nil {
			Fail(fmt.Sprintf("WaitForCount(5): %v", err))
		}
		wg.Wait()
	})

	It("WaitForCall predicate matches a variadic Param", func() {
		m := rtmock.NewRuntimeMock()
		ctx := context.Background()

		go func() {
			time.Sleep(20 * time.Millisecond)
			_, _, _ = m.Fetch(ctx, "needle")
		}()

		err := m.Calls().Fetch().WaitForCall(func(c rtmock.RuntimeMockFetchCall) bool {
			return len(c.Param.Ids) == 1 && c.Param.Ids[0] == "needle"
		}, 2*time.Second)
		if err != nil {
			Fail(fmt.Sprintf("WaitForCall(needle): %v", err))
		}
	})
})
