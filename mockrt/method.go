package mockrt

// Method is the per-method state of a generated mock: the call log and the
// configured returns. The zero value is ready to use.
type Method struct {
	core
}

// Call records a call with Param value p and returns the Return value it
// resolves to, or nil for the zero value. Resolution order: the
// ReturnCustom handler; otherwise the one matching When() conditional (its
// Once queue, then its Always slot); otherwise the Once queue, the Always
// slot, and finally the ReturnAll/Return() values. More than one matching
// When() predicate panics.
//
// Under --async the lock is released before the handler runs, so a handler
// may call back into the mock; the resolved returns are written back to the
// call record afterwards.
func (m *Method) Call(ref Ref, p any) any { return m.call(ref, p) }

// OnVoid is the configuration namespace of a method without return values:
// there is nothing to arm, only Reset.
type OnVoid struct {
	ref Ref
	c   *core
}

// OnVoid returns the configuration namespace for a method without returns.
func (m *Method) OnVoid(ref Ref) OnVoid { return OnVoid{ref: ref, c: &m.core} }

// Reset clears configured returns.
func (o OnVoid) Reset() { o.c.resetConfigLocked(o.ref) }

// On is the untyped configuration namespace of a method with return
// values. The generated per-method type embeds it, shadows Return, Once,
// Always and When with versions typed to the method, and adds ReturnAll and
// ReturnCustom.
//
// When(pred) hands back the same namespace scoped to one conditional: Once
// and Always then arm that conditional, and the operations that only make
// sense for the method as a whole panic.
type On struct {
	OnVoid
	conditional bool
	index       int
}

// On returns the configuration namespace for a method with returns.
func (m *Method) On(ref Ref) On { return On{OnVoid: m.OnVoid(ref)} }

func (o On) notOnWhen(op string) {
	if o.conditional {
		panic(o.ref.Name + "." + o.ref.Method + ": " + op + " is not available on a When() handle; arm it with Once() or Always()")
	}
}

// Reset clears configured returns and the installed handler.
func (o On) Reset() {
	o.notOnWhen("Reset()")
	o.OnVoid.Reset()
}

// SetAll sets every sticky return value at once. Backs ReturnAll.
func (o On) SetAll(r any) {
	o.notOnWhen("ReturnAll()")
	o.c.setAll(o.ref, r)
}

// Custom installs h as the handler that computes the returns of every call.
// Backs ReturnCustom; set is false when ReturnCustom got a nil function,
// which marks the handler as installed without one, so calls fall back to
// the sticky values.
func (o On) Custom(set bool, h func(p any) any) {
	o.notOnWhen("ReturnCustom()")
	if !set {
		h = nil
	}
	o.c.custom(o.ref, h)
}

// Return points at the sticky return values.
func (o On) Return() Stub {
	o.notOnWhen("Return()")
	return o.c.stickySlot(o.ref)
}

// Once arms a single-call return; the stub is consumed by the next call
// and then discarded. On a When() handle it fires only for calls matching
// the predicate.
func (o On) Once() Stub {
	if o.conditional {
		return o.c.whenOnce(o.ref, o.index)
	}
	return o.c.once(o.ref)
}

// Always arms a fallback return used when the Once() queue is exhausted;
// subsequent calls replace the previous fallback. On a When() handle it
// fires on every call matching the predicate once that conditional's
// Once() queue is exhausted.
func (o On) Always() Stub {
	if o.conditional {
		return o.c.whenAlways(o.ref, o.index)
	}
	return o.c.alwaysSlot(o.ref)
}

// When registers a conditional stub set whose entries fire only when the
// call's Param satisfies pred, and returns the handle to arm it with.
func (o On) When(pred func(p any) bool) On {
	o.notOnWhen("When()")
	return On{OnVoid: o.OnVoid, conditional: true, index: o.c.when(o.ref, pred)}
}

// Stub points at one set of return values: the sticky values, a Once()
// queue entry or an Always() slot. Generated per-return setter types embed
// it and write through Set.
type Stub struct {
	ref  Ref
	slot *slot
}

// Set applies f to the Return value s points at and returns s. R is the
// method's Return struct.
func Set[R any](s Stub, f func(r *R)) Stub {
	s.ref.lock()
	defer s.ref.unlock()
	r, _ := s.slot.v.(R)
	f(&r)
	s.slot.v = r
	return s
}
