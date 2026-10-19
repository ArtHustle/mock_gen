package mockrt

import (
	"strconv"
	"time"
)

// The runtime is deliberately non-generic: Params and Returns travel as
// `any` and generated code converts at the boundary. Go compiles every
// method of an instantiated generic type in each package that uses the
// type, so generic namespace types would be recompiled, per method, in
// every test package that touches a mock.

// core is the per-method state: the call log and the configured returns.
type core struct {
	log []Rec
	cfg config
}

// Rec is one recorded call in untyped form; generated code converts it to
// the method's Call struct. Param holds the method's Param value; Return
// holds its Return value, or nil for the zero value (also while the call is
// still resolving its returns).
type Rec struct {
	Index, GlobalIndex int
	Param, Return      any
}

// slot holds one set of return values: the sticky values, a Once() queue
// entry or an Always() slot. v holds a Return value or is nil for the zero
// value; setters replace v rather than mutating it, so a copy of v is a
// snapshot. Queue entries and Always slots are pointers so per-return
// setters held across later armings keep writing to the slot they were
// handed.
type slot struct {
	v any
}

type config struct {
	sticky       slot
	stickySet    bool
	queue        []*slot
	always       *slot
	conditionals []*conditional
	handlerSet   bool
	handler      func(any) any
}

// conditional is the state of one When() registration. The config holds
// pointers so a conditional stays at a stable address across later When()
// registrations.
type conditional struct {
	pred   func(any) bool
	queue  []*slot
	always *slot
}

func (c *core) call(ref Ref, p any) any {
	ref.lock()
	global := ref.Mock.global
	ref.Mock.global++
	idx := len(c.log)
	c.log = append(c.log, Rec{Index: idx, GlobalIndex: global, Param: p})
	ref.Mock.timeline = append(ref.Mock.timeline, TimelineEntry{Method: ref.Method, Index: idx, GlobalIndex: global})
	handler := c.cfg.handler
	var ret any
	matched := false
	if len(c.cfg.conditionals) > 0 {
		first, second, count := -1, -1, 0
		for i, cond := range c.cfg.conditionals {
			if cond.pred(p) {
				count++
				if first == -1 {
					first = i
				} else if second == -1 {
					second = i
				}
			}
		}
		if count >= 2 {
			ref.unlock()
			panic(ref.Name + "." + ref.Method + ": call matched " + strconv.Itoa(count) + " When predicates (indices " + strconv.Itoa(first) + " and " + strconv.Itoa(second) + "); refine the predicates or remove the overlap")
		}
		if count == 1 && handler == nil {
			ret, matched = c.cfg.conditionals[first].take()
		}
	}
	if !matched && handler == nil {
		switch {
		case len(c.cfg.queue) > 0:
			ret = c.cfg.queue[0].v
			c.cfg.queue = c.cfg.queue[1:]
		case c.cfg.always != nil:
			ret = c.cfg.always.v
		default:
			ret = c.cfg.sticky.v
		}
	}
	ref.unlock()

	if handler != nil {
		ret = handler(p)
	}

	ref.lock()
	if idx < len(c.log) {
		c.log[idx].Return = ret
	}
	ref.unlock()
	return ret
}

func (c *conditional) take() (any, bool) {
	if len(c.queue) > 0 {
		v := c.queue[0].v
		c.queue = c.queue[1:]
		return v, true
	}
	if c.always != nil {
		return c.always.v, true
	}
	return nil, false
}

// pending is the number of unconsumed Once() entries, conditional ones
// included. Callers hold the lock.
func (c *core) pending() int {
	n := len(c.cfg.queue)
	for _, cond := range c.cfg.conditionals {
		n += len(cond.queue)
	}
	return n
}

func (c *core) resetLog()    { c.log = nil }
func (c *core) resetConfig() { c.cfg = config{} }

// --- inspection ---

func (c *core) count(ref Ref) int {
	ref.lock()
	defer ref.unlock()
	return len(c.log)
}

func (c *core) at(ref Ref, i int) Rec {
	ref.lock()
	defer ref.unlock()
	return c.log[ref.index(len(c.log), i)]
}

func (c *core) all(ref Ref) []Rec {
	ref.lock()
	defer ref.unlock()
	out := make([]Rec, len(c.log))
	copy(out, c.log)
	return out
}

func (c *core) reset(ref Ref) {
	ref.lock()
	defer ref.unlock()
	c.resetLog()
}

func (c *core) pendingStubs(ref Ref) int {
	ref.lock()
	defer ref.unlock()
	return c.pending()
}

func (c *core) waitForCall(ref Ref, pred func(Rec) bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		for _, r := range c.all(ref) {
			if pred(r) {
				return true
			}
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(time.Millisecond)
	}
}

// --- configuration ---

func (c *core) mustNotBeArmed(ref Ref) {
	if c.cfg.always != nil || len(c.cfg.queue) > 0 {
		panic(ref.Name + "." + ref.Method + ": cannot set sticky return while Once()/Always() entries are armed; call .Reset() first")
	}
}

func (c *core) mustNotBeSticky(ref Ref) {
	if c.cfg.stickySet || c.cfg.handlerSet {
		panic(ref.Name + "." + ref.Method + ": cannot arm Once()/Always() while a sticky return is set; call .Reset() first")
	}
}

func (c *core) setAll(ref Ref, v any) {
	ref.lock()
	defer ref.unlock()
	c.mustNotBeArmed(ref)
	c.cfg.stickySet = true
	c.cfg.sticky.v = v
}

func (c *core) custom(ref Ref, handler func(any) any) {
	ref.lock()
	defer ref.unlock()
	c.mustNotBeArmed(ref)
	c.cfg.handlerSet = true
	c.cfg.handler = handler
}

func (c *core) stickySlot(ref Ref) Stub {
	ref.lock()
	defer ref.unlock()
	c.mustNotBeArmed(ref)
	c.cfg.stickySet = true
	return Stub{ref: ref, slot: &c.cfg.sticky}
}

func (c *core) once(ref Ref) Stub {
	ref.lock()
	defer ref.unlock()
	c.mustNotBeSticky(ref)
	s := &slot{}
	c.cfg.queue = append(c.cfg.queue, s)
	return Stub{ref: ref, slot: s}
}

func (c *core) alwaysSlot(ref Ref) Stub {
	ref.lock()
	defer ref.unlock()
	c.mustNotBeSticky(ref)
	c.cfg.always = &slot{}
	return Stub{ref: ref, slot: c.cfg.always}
}

func (c *core) when(ref Ref, pred func(any) bool) int {
	ref.lock()
	defer ref.unlock()
	c.cfg.conditionals = append(c.cfg.conditionals, &conditional{pred: pred})
	return len(c.cfg.conditionals) - 1
}

func (c *core) resetConfigLocked(ref Ref) {
	ref.lock()
	defer ref.unlock()
	c.resetConfig()
}

func (c *core) conditional(ref Ref, index int) *conditional {
	if index >= len(c.cfg.conditionals) {
		panic(ref.Name + "." + ref.Method + ": stale When() handle; conditionals were cleared (via Reset) after the When() call that produced this handle")
	}
	return c.cfg.conditionals[index]
}

func (c *core) whenOnce(ref Ref, index int) Stub {
	ref.lock()
	defer ref.unlock()
	cond := c.conditional(ref, index)
	s := &slot{}
	cond.queue = append(cond.queue, s)
	return Stub{ref: ref, slot: s}
}

func (c *core) whenAlways(ref Ref, index int) Stub {
	ref.lock()
	defer ref.unlock()
	cond := c.conditional(ref, index)
	cond.always = &slot{}
	return Stub{ref: ref, slot: cond.always}
}
