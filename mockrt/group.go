package mockrt

import "strconv"

// methodState is implemented by every *Method[P, R]; it lets the mock-level
// operations walk methods of different P and R.
type methodState interface {
	pending() int
	resetLog()
	resetConfig()
}

// Named pairs a source method name with its state.
type Named struct {
	Name   string
	Method methodState
}

// Group is a whole mock: its shared state plus every method, in source
// order. It backs the mock-level Reset and the Calls()/On() namespaces.
type Group struct {
	Ref     Ref
	Methods []Named
}

// Reset clears every call log, the timeline and every configured return.
func (g Group) Reset() {
	g.Ref.lock()
	defer g.Ref.unlock()
	g.resetCalls()
	for _, m := range g.Methods {
		m.Method.resetConfig()
	}
}

func (g Group) resetCalls() {
	g.Ref.Mock.global = 0
	g.Ref.Mock.timeline = nil
	for _, m := range g.Methods {
		m.Method.resetLog()
	}
}

// MockCalls returns the mock-level inspection namespace.
func (g Group) MockCalls() MockCalls { return MockCalls{g: g} }

// MockOn returns the mock-level configuration namespace.
func (g Group) MockOn() MockOn { return MockOn{g: g} }

// MockCalls is the mock-level part of a generated Calls() namespace;
// the generated type embeds it next to its per-method accessors.
type MockCalls struct {
	g Group
}

// Reset clears every per-method call log on the mock and the cross-method
// timeline, without touching configured returns or installed handlers.
func (c MockCalls) Reset() {
	c.g.Ref.lock()
	defer c.g.Ref.unlock()
	c.g.resetCalls()
}

// PendingStubs returns the total number of unconsumed Once() stub entries
// across all methods on this mock.
func (c MockCalls) PendingStubs() int {
	c.g.Ref.lock()
	defer c.g.Ref.unlock()
	n := 0
	for _, m := range c.g.Methods {
		n += m.Method.pending()
	}
	return n
}

// Timeline returns a defensive copy of the cross-method call timeline.
// Each entry names the source method, its per-method index, and a
// mock-wide GlobalIndex assigned at recording time.
func (c MockCalls) Timeline() []TimelineEntry {
	c.g.Ref.lock()
	defer c.g.Ref.unlock()
	out := make([]TimelineEntry, len(c.g.Ref.Mock.timeline))
	copy(out, c.g.Ref.Mock.timeline)
	return out
}

// AssertOrder calls t.Errorf when the given entries were not recorded in
// the given order. Adjacent entries with the same GlobalIndex (same call
// passed twice) are treated as in-order. 0 or 1 entries are a no-op.
func (c MockCalls) AssertOrder(t T, entries ...Call) {
	t.Helper()
	if len(entries) < 2 {
		return
	}
	prev := entries[0].Global()
	for i := 1; i < len(entries); i++ {
		cur := entries[i].Global()
		if cur < prev {
			t.Errorf(c.g.Ref.Name + ": AssertOrder violated — entry " + strconv.Itoa(i) + " (global #" + strconv.Itoa(cur) + ") happened before entry " + strconv.Itoa(i-1) + " (global #" + strconv.Itoa(prev) + ")")
			return
		}
		prev = cur
	}
}

// Verify calls t.Errorf if any method has unconsumed Once() stub entries.
// It does not call t.Fatal so other teardown checks still run. The mock
// lock is released before t.Errorf is invoked, so a user-supplied T
// implementation can safely call back into the mock without deadlocking.
func (c MockCalls) Verify(t T) {
	t.Helper()
	c.g.Ref.lock()
	total, breakdown := 0, ""
	for _, m := range c.g.Methods {
		if n := m.Method.pending(); n > 0 {
			total += n
			if breakdown != "" {
				breakdown += ", "
			}
			breakdown += m.Name + ": " + strconv.Itoa(n)
		}
	}
	c.g.Ref.unlock()
	if total > 0 {
		t.Errorf(c.g.Ref.Name + ": " + strconv.Itoa(total) + " unconsumed stubs (" + breakdown + ")")
	}
}

// MockOn is the mock-level part of a generated On() namespace; the
// generated type embeds it next to its per-method accessors.
type MockOn struct {
	g Group
}

// Reset clears configured returns and installed handlers for every method
// on the mock, without touching the call log.
func (o MockOn) Reset() {
	o.g.Ref.lock()
	defer o.g.Ref.unlock()
	for _, m := range o.g.Methods {
		m.Method.resetConfig()
	}
}
