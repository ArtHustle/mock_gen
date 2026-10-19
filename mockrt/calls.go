package mockrt

import (
	"errors"
	"strconv"
	"time"
)

// Calls is the untyped inspection namespace of one method. The generated
// per-method type embeds it, keeps Count, Reset and PendingStubs, and
// shadows At, Last, All, Any and Filter with versions typed to the method's
// Call struct.
type Calls struct {
	ref Ref
	c   *core
}

// Calls returns the inspection namespace for m.
func (m *Method) Calls(ref Ref) Calls { return Calls{ref: ref, c: &m.core} }

// Count returns the number of recorded calls.
func (c Calls) Count() int { return c.c.count(c.ref) }

// Reset clears the call log of this method.
func (c Calls) Reset() { c.c.reset(c.ref) }

// PendingStubs returns the number of unconsumed Once() stub entries.
// Returns 0 for methods with no return values.
func (c Calls) PendingStubs() int { return c.c.pendingStubs(c.ref) }

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (c Calls) At(i int) Rec { return c.c.at(c.ref, i) }

// All returns a defensive copy of the full call log.
func (c Calls) All() []Rec { return c.c.all(c.ref) }

// Any returns true iff at least one recorded call satisfies pred.
// Pred runs after the log is copied (no lock held); pred may safely call
// back into the mock without deadlocking. A nil pred panics.
func (c Calls) Any(pred func(Rec) bool) bool {
	if pred == nil {
		c.nilPredicate("Any")
	}
	for _, r := range c.All() {
		if pred(r) {
			return true
		}
	}
	return false
}

// Filter returns every recorded call satisfying pred, in invocation order.
// Returns an empty (non-nil) slice when no call matches. Pred runs after
// the log is copied; pred may safely call back into the mock. A nil pred
// panics.
func (c Calls) Filter(pred func(Rec) bool) []Rec {
	if pred == nil {
		c.nilPredicate("Filter")
	}
	out := []Rec{}
	for _, r := range c.All() {
		if pred(r) {
			out = append(out, r)
		}
	}
	return out
}

func (c Calls) nilPredicate(op string) {
	panic(c.ref.helperName() + "." + op + ": predicate is nil")
}

// AsyncCalls is the inspection namespace of one method of an --async mock:
// Calls plus WaitForCount and WaitForCall for calls made by background
// goroutines.
type AsyncCalls struct {
	Calls
}

// AsyncCalls returns the inspection namespace for m on an --async mock.
func (m *Method) AsyncCalls(ref Ref) AsyncCalls { return AsyncCalls{m.Calls(ref)} }

// WaitForCount blocks until at least n calls have been recorded, or until
// timeout elapses. Returns nil on success, an error describing the current
// count on timeout. Polls every 1ms.
func (c AsyncCalls) WaitForCount(n int, timeout time.Duration) error {
	if n <= 0 {
		return nil
	}
	if !c.c.waitForCall(c.ref, func(r Rec) bool { return r.Index >= n-1 }, timeout) {
		return errors.New("WaitForCount(" + strconv.Itoa(n) + ", " + timeout.String() + "): timed out, count = " + strconv.Itoa(c.Count()))
	}
	return nil
}

// WaitForCall blocks until at least one recorded call satisfies the
// predicate, or until timeout elapses. Returns nil on success, an error on
// timeout. Polls every 1ms; on each poll, scans every recorded call from
// oldest to newest.
func (c AsyncCalls) WaitForCall(predicate func(Rec) bool, timeout time.Duration) error {
	if !c.c.waitForCall(c.ref, predicate, timeout) {
		return errors.New("WaitForCall(" + timeout.String() + "): timed out, no matching call in " + strconv.Itoa(c.Count()) + " recorded entries")
	}
	return nil
}

// Map converts recorded calls with conv, the generated per-method
// converter. The result is non-nil even when recs is empty.
func Map[C any](recs []Rec, conv func(Rec) C) []C {
	out := make([]C, len(recs))
	for i, r := range recs {
		out[i] = conv(r)
	}
	return out
}

// Pred adapts a typed call predicate to the untyped log, keeping nil as nil
// so Any and Filter can report it.
func Pred[C any](pred func(C) bool, conv func(Rec) C) func(Rec) bool {
	if pred == nil {
		return nil
	}
	return func(r Rec) bool { return pred(conv(r)) }
}
