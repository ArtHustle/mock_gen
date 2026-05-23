package test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/ArtHustle/mock_gen"
)

// Canonical literal-output entries. The new API generates significant code
// per method (Param/Return/Cfg sub-structs, per-method Calls/On types,
// callIndex helper), so we keep this file to a small number of canonical
// shapes that pin down the overall structure. Behavior is verified in
// behavior_test.go; field-naming edge cases in field_naming_test.go.
var _ = Describe("Generated mock shape", func() {
	DescribeTable("emits the expected mock",
		runGenEntry,

		Entry("default options: method with no params/returns", `
package pkg

type Service interface {
	Process()
}`, `package test

import (
	"strconv"
)

type serviceMockCallsHelper[CT any] struct {
	log       *[]CT
	pendingFn func() int
	indexer   func(length, i int) int
}

// Count returns the number of recorded calls.
func (this serviceMockCallsHelper[CT]) Count() int {
	return len(*this.log)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this serviceMockCallsHelper[CT]) At(i int) CT {
	return (*this.log)[this.indexer(len(*this.log), i)]
}

// Last is sugar for At(-1).
func (this serviceMockCallsHelper[CT]) Last() CT {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this serviceMockCallsHelper[CT]) All() []CT {
	out := make([]CT, len(*this.log))
	copy(out, *this.log)
	return out
}

// Reset clears the call log.
func (this serviceMockCallsHelper[CT]) Reset() {
	*this.log = nil
}

// PendingStubs returns the number of unconsumed Once() stub entries.
// Returns 0 for methods with no return values.
func (this serviceMockCallsHelper[CT]) PendingStubs() int {
	return this.pendingFn()
}

// Any returns true iff at least one recorded call satisfies pred.
// Pred runs after All() returns (no lock held); pred may safely call
// back into the mock without deadlocking.
func (this serviceMockCallsHelper[CT]) Any(pred func(CT) bool) bool {
	if pred == nil {
		panic("serviceMockCallsHelper.Any: predicate is nil")
	}
	for _, c := range this.All() {
		if pred(c) {
			return true
		}
	}
	return false
}

// Filter returns every recorded call satisfying pred, in invocation order.
// Returns an empty (non-nil) slice when no call matches. Pred runs after
// All() returns; pred may safely call back into the mock.
func (this serviceMockCallsHelper[CT]) Filter(pred func(CT) bool) []CT {
	if pred == nil {
		panic("serviceMockCallsHelper.Filter: predicate is nil")
	}
	out := []CT{}
	for _, c := range this.All() {
		if pred(c) {
			out = append(out, c)
		}
	}
	return out
}

// ServiceMockCall is satisfied by every per-method <CallType>
// and by ServiceMockTimelineEntry. Enables HappenedBefore comparisons
// across concrete call types.
type ServiceMockCall interface {
	Global() int
}

// ServiceMockTimelineEntry records one call across any method in
// invocation order. Method names the source method; Index is the
// per-method position at the time the call was recorded; GlobalIndex is
// monotonically increasing across all methods.
type ServiceMockTimelineEntry struct {
	Method      string
	Index       int
	GlobalIndex int
}

// Global returns the global ordering position for this entry.
func (this ServiceMockTimelineEntry) Global() int { return this.GlobalIndex }

// HappenedBefore returns true iff this entry was recorded before other.
func (this ServiceMockTimelineEntry) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

// ServiceMockProcessCall records one call to Service.Process.
// Index is the 0-based position of this call in the per-method log, useful
// for ordering assertions in WaitForCall predicates and All() iteration.
// Index resets to 0 when the log is cleared via Reset. GlobalIndex is the
// mock-wide position assigned when the call was recorded; see Timeline().
type ServiceMockProcessCall struct {
	Index       int
	GlobalIndex int
	Param       ServiceMockProcessParam
	Return      ServiceMockProcessReturn
}

// Global returns the mock-wide ordering position for this call.
func (this ServiceMockProcessCall) Global() int { return this.GlobalIndex }

// HappenedBefore returns true iff this call was recorded before other.
func (this ServiceMockProcessCall) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

type ServiceMockProcessParam struct {
}

type ServiceMockProcessReturn struct {
}

type serviceMockProcessCfg struct {
	sticky    struct{}
	stickySet bool
}

type ServiceMock struct {
	globalCallCounter int
	timelineLog       []ServiceMockTimelineEntry

	processLog []ServiceMockProcessCall
	processCfg serviceMockProcessCfg
}

// NewServiceMock returns a zero-valued ServiceMock. Equivalent
// to &ServiceMock{}; provided as a consistent entry point that may
// initialize internal state in future versions.
func NewServiceMock() *ServiceMock {
	return &ServiceMock{}
}

func (this *ServiceMock) Process() {
	__mockGlobal := this.globalCallCounter
	this.globalCallCounter++
	this.processLog = append(this.processLog, ServiceMockProcessCall{
		Index:       len(this.processLog),
		GlobalIndex: __mockGlobal,
		Param:       ServiceMockProcessParam{},
	})
	this.timelineLog = append(this.timelineLog, ServiceMockTimelineEntry{
		Method:      "Process",
		Index:       len(this.processLog) - 1,
		GlobalIndex: __mockGlobal,
	})
}

func (this *ServiceMock) Reset() {
	this.globalCallCounter = 0
	this.timelineLog = nil
	this.processLog = nil
	this.processCfg = serviceMockProcessCfg{}
}

type serviceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock and the cross-method
// timeline, without touching configured returns or installed handlers.
func (this serviceMockCalls) Reset() {
	this.m.globalCallCounter = 0
	this.m.timelineLog = nil
	this.m.processLog = nil
}

func (this serviceMockCalls) Process() serviceMockProcessCalls {
	return serviceMockProcessCalls{serviceMockCallsHelper: serviceMockCallsHelper[ServiceMockProcessCall]{log: &this.m.processLog,
		pendingFn: func() int {
			return 0
		},
		indexer: this.m.callIndex,
	}}
}

// PendingStubs returns the total number of unconsumed Once() stub entries
// across all methods on this mock.
func (this serviceMockCalls) PendingStubs() int {
	n := 0
	return n
}

// Timeline returns a defensive copy of the cross-method call timeline.
// Each entry names the source method, its per-method index, and a
// mock-wide GlobalIndex assigned at recording time.
func (this serviceMockCalls) Timeline() []ServiceMockTimelineEntry {
	out := make([]ServiceMockTimelineEntry, len(this.m.timelineLog))
	copy(out, this.m.timelineLog)
	return out
}

// AssertOrder calls t.Errorf when the given entries were not recorded in
// the given order. Adjacent entries with the same GlobalIndex (same call
// passed twice) are treated as in-order. 0 or 1 entries are a no-op.
func (this serviceMockCalls) AssertOrder(t ServiceMockVerifyT, entries ...ServiceMockCall) {
	t.Helper()
	if len(entries) < 2 {
		return
	}
	prev := entries[0].Global()
	for i := 1; i < len(entries); i++ {
		cur := entries[i].Global()
		if cur < prev {
			t.Errorf("ServiceMock: AssertOrder violated — entry " + strconv.Itoa(i) + " (global #" + strconv.Itoa(cur) + ") happened before entry " + strconv.Itoa(i-1) + " (global #" + strconv.Itoa(prev) + ")")
			return
		}
		prev = cur
	}
}

// ServiceMockVerifyT is the interface Verify accepts. *testing.T satisfies it.
type ServiceMockVerifyT interface {
	Helper()
	Errorf(format string, args ...any)
}

// Verify calls t.Errorf if any method has unconsumed Once() stub entries.
// It does not call t.Fatal so other teardown checks still run. The mock
// mutex is released before t.Errorf is invoked, so a user-supplied T
// implementation can safely call back into the mock without deadlocking.
func (this serviceMockCalls) Verify(t ServiceMockVerifyT) {
	t.Helper()
	total, breakdown := 0, ""
	if total > 0 {
		t.Errorf("ServiceMock: " + strconv.Itoa(total) + " unconsumed stubs (" + breakdown + ")")
	}
}

// serviceMockProcessCalls is the per-method inspection namespace for
// Service.Process. Methods are provided by the embedded serviceMockCallsHelper.
type serviceMockProcessCalls struct {
	serviceMockCallsHelper[ServiceMockProcessCall]
}

type serviceMockOn struct{ m *ServiceMock }

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() serviceMockOn {
	return serviceMockOn{m: this}
}

// Reset clears configured returns and installed handlers for every method
// on the mock, without touching the call log.
func (this serviceMockOn) Reset() {
	this.m.processCfg = serviceMockProcessCfg{}
}

func (this serviceMockOn) Process() serviceMockProcessOn {
	return serviceMockProcessOn{m: this.m}
}

type serviceMockProcessOn struct{ m *ServiceMock }

// Reset clears configured returns and installed handler for Process.
func (this serviceMockProcessOn) Reset() {
	this.m.processCfg = serviceMockProcessCfg{}
}

// callIndex resolves the index argument of Method_call(i)-style indexers:
// negative i counts from the end; out-of-range i panics with a message
// that surfaces both the original input and the log length.
func (this *ServiceMock) callIndex(length, i int) int {
	resolved := i
	if resolved < 0 {
		resolved += length
	}
	if resolved < 0 || resolved >= length {
		if i < 0 {
			panic("ServiceMock call index " + strconv.Itoa(i) + " (resolved to " + strconv.Itoa(resolved) + ") out of range, log has " + strconv.Itoa(length) + " entries")
		}
		panic("ServiceMock call index " + strconv.Itoa(i) + " out of range, log has " + strconv.Itoa(length) + " entries")
	}
	return resolved
}
`, Options{}),

		Entry("default options: method with params and returns (non-async)", `
package pkg

type Service interface {
	Send(email string) (id string, err error)
}`, `package test

import (
	"strconv"
)

type serviceMockCallsHelper[CT any] struct {
	log       *[]CT
	pendingFn func() int
	indexer   func(length, i int) int
}

// Count returns the number of recorded calls.
func (this serviceMockCallsHelper[CT]) Count() int {
	return len(*this.log)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this serviceMockCallsHelper[CT]) At(i int) CT {
	return (*this.log)[this.indexer(len(*this.log), i)]
}

// Last is sugar for At(-1).
func (this serviceMockCallsHelper[CT]) Last() CT {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this serviceMockCallsHelper[CT]) All() []CT {
	out := make([]CT, len(*this.log))
	copy(out, *this.log)
	return out
}

// Reset clears the call log.
func (this serviceMockCallsHelper[CT]) Reset() {
	*this.log = nil
}

// PendingStubs returns the number of unconsumed Once() stub entries.
// Returns 0 for methods with no return values.
func (this serviceMockCallsHelper[CT]) PendingStubs() int {
	return this.pendingFn()
}

// Any returns true iff at least one recorded call satisfies pred.
// Pred runs after All() returns (no lock held); pred may safely call
// back into the mock without deadlocking.
func (this serviceMockCallsHelper[CT]) Any(pred func(CT) bool) bool {
	if pred == nil {
		panic("serviceMockCallsHelper.Any: predicate is nil")
	}
	for _, c := range this.All() {
		if pred(c) {
			return true
		}
	}
	return false
}

// Filter returns every recorded call satisfying pred, in invocation order.
// Returns an empty (non-nil) slice when no call matches. Pred runs after
// All() returns; pred may safely call back into the mock.
func (this serviceMockCallsHelper[CT]) Filter(pred func(CT) bool) []CT {
	if pred == nil {
		panic("serviceMockCallsHelper.Filter: predicate is nil")
	}
	out := []CT{}
	for _, c := range this.All() {
		if pred(c) {
			out = append(out, c)
		}
	}
	return out
}

// ServiceMockCall is satisfied by every per-method <CallType>
// and by ServiceMockTimelineEntry. Enables HappenedBefore comparisons
// across concrete call types.
type ServiceMockCall interface {
	Global() int
}

// ServiceMockTimelineEntry records one call across any method in
// invocation order. Method names the source method; Index is the
// per-method position at the time the call was recorded; GlobalIndex is
// monotonically increasing across all methods.
type ServiceMockTimelineEntry struct {
	Method      string
	Index       int
	GlobalIndex int
}

// Global returns the global ordering position for this entry.
func (this ServiceMockTimelineEntry) Global() int { return this.GlobalIndex }

// HappenedBefore returns true iff this entry was recorded before other.
func (this ServiceMockTimelineEntry) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

// ServiceMockSendCall records one call to Service.Send.
// Index is the 0-based position of this call in the per-method log, useful
// for ordering assertions in WaitForCall predicates and All() iteration.
// Index resets to 0 when the log is cleared via Reset. GlobalIndex is the
// mock-wide position assigned when the call was recorded; see Timeline().
type ServiceMockSendCall struct {
	Index       int
	GlobalIndex int
	Param       ServiceMockSendParam
	Return      ServiceMockSendReturn
}

// Global returns the mock-wide ordering position for this call.
func (this ServiceMockSendCall) Global() int { return this.GlobalIndex }

// HappenedBefore returns true iff this call was recorded before other.
func (this ServiceMockSendCall) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

type ServiceMockSendParam struct {
	Email string
}

type ServiceMockSendReturn struct {
	Id  string
	Err error
}

// serviceMockSendStub holds one set of return values — used for sticky cfg,
// Once() queue entries, and Always() fallback. Queue entries and
// conditional Always slots are referenced through pointers so per-slot
// setters held across subsequent armings remain valid.
type serviceMockSendStub struct {
	id  string
	err error
}

// serviceMockSendConditional is per-method conditional state — one entry
// per When() registration. The outer slice holds *Conditional (rather than
// values) so the conditional's own fields stay at a stable address across
// subsequent When() registrations.
type serviceMockSendConditional struct {
	pred       func(ServiceMockSendParam) bool
	stubQueue  []*serviceMockSendStub
	alwaysStub *serviceMockSendStub
}

type serviceMockSendCfg struct {
	sticky       serviceMockSendStub
	stickySet    bool
	stubQueue    []*serviceMockSendStub
	alwaysStub   *serviceMockSendStub
	conditionals []*serviceMockSendConditional
	handlerSet   bool

	handler func(string) (string, error)
}

type ServiceMock struct {
	globalCallCounter int
	timelineLog       []ServiceMockTimelineEntry

	sendLog []ServiceMockSendCall
	sendCfg serviceMockSendCfg
}

// NewServiceMock returns a zero-valued ServiceMock. Equivalent
// to &ServiceMock{}; provided as a consistent entry point that may
// initialize internal state in future versions.
func NewServiceMock() *ServiceMock {
	return &ServiceMock{}
}

func (this *ServiceMock) Send(email string) (id string, err error) {
	__mockParam := ServiceMockSendParam{Email: email}
	__mockGlobal := this.globalCallCounter
	this.globalCallCounter++
	this.sendLog = append(this.sendLog, ServiceMockSendCall{
		Index:       len(this.sendLog),
		GlobalIndex: __mockGlobal,
		Param:       __mockParam,
	})
	__mockCallIdx := len(this.sendLog) - 1
	this.timelineLog = append(this.timelineLog, ServiceMockTimelineEntry{
		Method:      "Send",
		Index:       __mockCallIdx,
		GlobalIndex: __mockGlobal,
	})
	__mockHandler := this.sendCfg.handler
	var __mockSource serviceMockSendStub
	__mockMatched := false
	if len(this.sendCfg.conditionals) > 0 {
		__mockFirstMatch := -1
		__mockSecondMatch := -1
		__mockMatchCount := 0
		for __mockIdx := range this.sendCfg.conditionals {
			if this.sendCfg.conditionals[__mockIdx].pred(__mockParam) {
				__mockMatchCount++
				if __mockFirstMatch == -1 {
					__mockFirstMatch = __mockIdx
				} else if __mockSecondMatch == -1 {
					__mockSecondMatch = __mockIdx
				}
			}
		}
		if __mockMatchCount >= 2 {
			panic("ServiceMock.Send: call matched " + strconv.Itoa(__mockMatchCount) + " When predicates (indices " + strconv.Itoa(__mockFirstMatch) + " and " + strconv.Itoa(__mockSecondMatch) + "); refine the predicates or remove the overlap")
		}
		if __mockMatchCount == 1 && __mockHandler == nil {
			__mockCond := this.sendCfg.conditionals[__mockFirstMatch]
			if len(__mockCond.stubQueue) > 0 {
				__mockSource = *__mockCond.stubQueue[0]
				__mockCond.stubQueue = __mockCond.stubQueue[1:]
				__mockMatched = true
			} else if __mockCond.alwaysStub != nil {
				__mockSource = *__mockCond.alwaysStub
				__mockMatched = true
			}
		}
	}
	if !__mockMatched && __mockHandler == nil {
		if len(this.sendCfg.stubQueue) > 0 {
			__mockSource = *this.sendCfg.stubQueue[0]
			this.sendCfg.stubQueue = this.sendCfg.stubQueue[1:]
		} else if this.sendCfg.alwaysStub != nil {
			__mockSource = *this.sendCfg.alwaysStub
		} else {
			__mockSource = this.sendCfg.sticky
		}
	}
	__mockR0 := __mockSource.id
	__mockR1 := __mockSource.err
	var __mockRet0 string
	var __mockRet1 error
	if __mockHandler != nil {
		__mockRet0, __mockRet1 = __mockHandler(email)
	} else {
		__mockRet0, __mockRet1 = __mockR0, __mockR1
	}

	if __mockCallIdx < len(this.sendLog) {
		this.sendLog[__mockCallIdx].Return = ServiceMockSendReturn{Id: __mockRet0, Err: __mockRet1}
	}
	return __mockRet0, __mockRet1
}

func (this *ServiceMock) Reset() {
	this.globalCallCounter = 0
	this.timelineLog = nil
	this.sendLog = nil
	this.sendCfg = serviceMockSendCfg{}
}

type serviceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock and the cross-method
// timeline, without touching configured returns or installed handlers.
func (this serviceMockCalls) Reset() {
	this.m.globalCallCounter = 0
	this.m.timelineLog = nil
	this.m.sendLog = nil
}

func (this serviceMockCalls) Send() serviceMockSendCalls {
	return serviceMockSendCalls{serviceMockCallsHelper: serviceMockCallsHelper[ServiceMockSendCall]{log: &this.m.sendLog,
		pendingFn: func() int {
			n := len(this.m.sendCfg.stubQueue)
			for __mockI := range this.m.sendCfg.conditionals {
				n += len(this.m.sendCfg.conditionals[__mockI].stubQueue)
			}
			return n
		},
		indexer: this.m.callIndex,
	}}
}

// PendingStubs returns the total number of unconsumed Once() stub entries
// across all methods on this mock.
func (this serviceMockCalls) PendingStubs() int {
	n := 0
	n += len(this.m.sendCfg.stubQueue)
	for __mockI := range this.m.sendCfg.conditionals {
		n += len(this.m.sendCfg.conditionals[__mockI].stubQueue)
	}
	return n
}

// Timeline returns a defensive copy of the cross-method call timeline.
// Each entry names the source method, its per-method index, and a
// mock-wide GlobalIndex assigned at recording time.
func (this serviceMockCalls) Timeline() []ServiceMockTimelineEntry {
	out := make([]ServiceMockTimelineEntry, len(this.m.timelineLog))
	copy(out, this.m.timelineLog)
	return out
}

// AssertOrder calls t.Errorf when the given entries were not recorded in
// the given order. Adjacent entries with the same GlobalIndex (same call
// passed twice) are treated as in-order. 0 or 1 entries are a no-op.
func (this serviceMockCalls) AssertOrder(t ServiceMockVerifyT, entries ...ServiceMockCall) {
	t.Helper()
	if len(entries) < 2 {
		return
	}
	prev := entries[0].Global()
	for i := 1; i < len(entries); i++ {
		cur := entries[i].Global()
		if cur < prev {
			t.Errorf("ServiceMock: AssertOrder violated — entry " + strconv.Itoa(i) + " (global #" + strconv.Itoa(cur) + ") happened before entry " + strconv.Itoa(i-1) + " (global #" + strconv.Itoa(prev) + ")")
			return
		}
		prev = cur
	}
}

// ServiceMockVerifyT is the interface Verify accepts. *testing.T satisfies it.
type ServiceMockVerifyT interface {
	Helper()
	Errorf(format string, args ...any)
}

// Verify calls t.Errorf if any method has unconsumed Once() stub entries.
// It does not call t.Fatal so other teardown checks still run. The mock
// mutex is released before t.Errorf is invoked, so a user-supplied T
// implementation can safely call back into the mock without deadlocking.
func (this serviceMockCalls) Verify(t ServiceMockVerifyT) {
	t.Helper()
	total, breakdown := 0, ""
	{
		n := len(this.m.sendCfg.stubQueue)
		for __mockI := range this.m.sendCfg.conditionals {
			n += len(this.m.sendCfg.conditionals[__mockI].stubQueue)
		}
		if n > 0 {
			total += n
			if breakdown != "" {
				breakdown += ", "
			}
			breakdown += "Send: " + strconv.Itoa(n)
		}
	}
	if total > 0 {
		t.Errorf("ServiceMock: " + strconv.Itoa(total) + " unconsumed stubs (" + breakdown + ")")
	}
}

// serviceMockSendCalls is the per-method inspection namespace for
// Service.Send. Methods are provided by the embedded serviceMockCallsHelper.
type serviceMockSendCalls struct {
	serviceMockCallsHelper[ServiceMockSendCall]
}

type serviceMockOn struct{ m *ServiceMock }

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() serviceMockOn {
	return serviceMockOn{m: this}
}

// Reset clears configured returns and installed handlers for every method
// on the mock, without touching the call log.
func (this serviceMockOn) Reset() {
	this.m.sendCfg = serviceMockSendCfg{}
}

func (this serviceMockOn) Send() serviceMockSendOn {
	return serviceMockSendOn{m: this.m}
}

type serviceMockSendOn struct{ m *ServiceMock }

// ReturnAll sets all static return values for Send
// positionally. Used when no ReturnCustom handler is installed.
func (this serviceMockSendOn) ReturnAll(id string, err error) {
	if this.m.sendCfg.alwaysStub != nil || len(this.m.sendCfg.stubQueue) > 0 {
		panic("ServiceMock.Send: cannot set sticky return while Once()/Always() entries are armed; call .Reset() first")
	}
	this.m.sendCfg.stickySet = true
	this.m.sendCfg.sticky.id = id
	this.m.sendCfg.sticky.err = err
}

// Return returns the per-return setter namespace — one method per return
// value of Send, each overriding just that return without
// touching the others.
func (this serviceMockSendOn) Return() serviceMockSendReturn {
	if this.m.sendCfg.alwaysStub != nil || len(this.m.sendCfg.stubQueue) > 0 {
		panic("ServiceMock.Send: cannot set sticky return while Once()/Always() entries are armed; call .Reset() first")
	}
	this.m.sendCfg.stickySet = true
	return serviceMockSendReturn{m: this.m, target: &this.m.sendCfg.sticky}
}

// ReturnCustom installs a dynamic function called instead of returning the
// configured static values.
func (this serviceMockSendOn) ReturnCustom(value func(string) (string, error)) {
	if this.m.sendCfg.alwaysStub != nil || len(this.m.sendCfg.stubQueue) > 0 {
		panic("ServiceMock.Send: cannot set sticky return while Once()/Always() entries are armed; call .Reset() first")
	}
	this.m.sendCfg.handlerSet = true
	this.m.sendCfg.handler = value
}

// Once arms a single-call return; the stub is consumed by the next call
// to Send and then discarded.
func (this serviceMockSendOn) Once() serviceMockSendReturn {
	if this.m.sendCfg.stickySet || this.m.sendCfg.handlerSet {
		panic("ServiceMock.Send: cannot arm Once()/Always() while a sticky return is set; call .Reset() first")
	}
	__mockStub := &serviceMockSendStub{}
	this.m.sendCfg.stubQueue = append(this.m.sendCfg.stubQueue, __mockStub)
	return serviceMockSendReturn{m: this.m, target: __mockStub}
}

// Always arms a fallback return used when the Once() queue is exhausted.
// Subsequent calls to Always() replace the previous fallback.
func (this serviceMockSendOn) Always() serviceMockSendReturn {
	if this.m.sendCfg.stickySet || this.m.sendCfg.handlerSet {
		panic("ServiceMock.Send: cannot arm Once()/Always() while a sticky return is set; call .Reset() first")
	}
	this.m.sendCfg.alwaysStub = &serviceMockSendStub{}
	return serviceMockSendReturn{m: this.m, target: this.m.sendCfg.alwaysStub}
}

// When registers a conditional stub set whose entries fire only when
// the call's Param satisfies the predicate. Use Once()/Always() on the
// returned namespace to attach armings.
func (this serviceMockSendOn) When(pred func(ServiceMockSendParam) bool) serviceMockSendOnWhen {
	this.m.sendCfg.conditionals = append(
		this.m.sendCfg.conditionals,
		&serviceMockSendConditional{pred: pred},
	)
	return serviceMockSendOnWhen{m: this.m, condIndex: len(this.m.sendCfg.conditionals) - 1}
}

// Reset clears configured returns and installed handler for Send.
func (this serviceMockSendOn) Reset() {
	this.m.sendCfg = serviceMockSendCfg{}
}

// serviceMockSendReturn is the per-return setter namespace for
// Send. Each method overrides one return value without
// touching the others. Methods return the setter itself for chaining.
type serviceMockSendReturn struct {
	m      *ServiceMock
	target *serviceMockSendStub
}

// Return is a no-op that returns this setter, allowing Once().Return().X()
// and Always().Return().X() chaining syntax.
func (this serviceMockSendReturn) Return() serviceMockSendReturn {
	return this
}

// Id overrides just the id return value.
func (this serviceMockSendReturn) Id(value string) serviceMockSendReturn {
	this.target.id = value
	return this
}

// Err overrides just the err return value.
func (this serviceMockSendReturn) Err(value error) serviceMockSendReturn {
	this.target.err = value
	return this
}

// serviceMockSendOnWhen is the conditional namespace returned by When(pred).
// Exposes Once() (queue a single-shot stub) and Always() (install/replace
// the fallback stub) — both scoped to the matching predicate.
type serviceMockSendOnWhen struct {
	m         *ServiceMock
	condIndex int
}

// Once arms a single-call return that fires only when this conditional's
// predicate matches.
func (this serviceMockSendOnWhen) Once() serviceMockSendReturn {
	if this.condIndex >= len(this.m.sendCfg.conditionals) {
		panic("ServiceMock.Send: stale When() handle; conditionals were cleared (via Reset) after the When() call that produced this handle")
	}
	cond := this.m.sendCfg.conditionals[this.condIndex]
	__mockStub := &serviceMockSendStub{}
	cond.stubQueue = append(cond.stubQueue, __mockStub)
	return serviceMockSendReturn{m: this.m, target: __mockStub}
}

// Always arms a fallback return that fires on every call matching this
// conditional's predicate, used when the conditional's Once() queue is
// exhausted.
func (this serviceMockSendOnWhen) Always() serviceMockSendReturn {
	if this.condIndex >= len(this.m.sendCfg.conditionals) {
		panic("ServiceMock.Send: stale When() handle; conditionals were cleared (via Reset) after the When() call that produced this handle")
	}
	cond := this.m.sendCfg.conditionals[this.condIndex]
	cond.alwaysStub = &serviceMockSendStub{}
	return serviceMockSendReturn{m: this.m, target: cond.alwaysStub}
}

// callIndex resolves the index argument of Method_call(i)-style indexers:
// negative i counts from the end; out-of-range i panics with a message
// that surfaces both the original input and the log length.
func (this *ServiceMock) callIndex(length, i int) int {
	resolved := i
	if resolved < 0 {
		resolved += length
	}
	if resolved < 0 || resolved >= length {
		if i < 0 {
			panic("ServiceMock call index " + strconv.Itoa(i) + " (resolved to " + strconv.Itoa(resolved) + ") out of range, log has " + strconv.Itoa(length) + " entries")
		}
		panic("ServiceMock call index " + strconv.Itoa(i) + " out of range, log has " + strconv.Itoa(length) + " entries")
	}
	return resolved
}
`, Options{}),

		Entry("--thread-safe: method with params and returns", `
package pkg

import "context"

type Service interface {
	Send(ctx context.Context, email string) (id string, err error)
}`, `package test

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"sync"
	"time"
)

type serviceMockCallsHelper[CT any] struct {
	mu          *sync.Mutex
	log         *[]CT
	pendingFn   func() int
	indexer     func(length, i int) int
	indexGetter func(CT) int
}

// Count returns the number of recorded calls.
func (this serviceMockCallsHelper[CT]) Count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(*this.log)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this serviceMockCallsHelper[CT]) At(i int) CT {
	this.mu.Lock()
	defer this.mu.Unlock()
	return (*this.log)[this.indexer(len(*this.log), i)]
}

// Last is sugar for At(-1).
func (this serviceMockCallsHelper[CT]) Last() CT {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this serviceMockCallsHelper[CT]) All() []CT {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]CT, len(*this.log))
	copy(out, *this.log)
	return out
}

// Reset clears the call log.
func (this serviceMockCallsHelper[CT]) Reset() {
	this.mu.Lock()
	defer this.mu.Unlock()
	*this.log = nil
}

// PendingStubs returns the number of unconsumed Once() stub entries.
// Returns 0 for methods with no return values.
func (this serviceMockCallsHelper[CT]) PendingStubs() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.pendingFn()
}

// Any returns true iff at least one recorded call satisfies pred.
// Pred runs after All() returns (no lock held); pred may safely call
// back into the mock without deadlocking.
func (this serviceMockCallsHelper[CT]) Any(pred func(CT) bool) bool {
	if pred == nil {
		panic("serviceMockCallsHelper.Any: predicate is nil")
	}
	for _, c := range this.All() {
		if pred(c) {
			return true
		}
	}
	return false
}

// Filter returns every recorded call satisfying pred, in invocation order.
// Returns an empty (non-nil) slice when no call matches. Pred runs after
// All() returns; pred may safely call back into the mock.
func (this serviceMockCallsHelper[CT]) Filter(pred func(CT) bool) []CT {
	if pred == nil {
		panic("serviceMockCallsHelper.Filter: predicate is nil")
	}
	out := []CT{}
	for _, c := range this.All() {
		if pred(c) {
			out = append(out, c)
		}
	}
	return out
}

// WaitForCount blocks until at least n calls have been recorded, or until
// timeout elapses. Returns nil on success, an error describing the current
// count on timeout. Polls every 1ms.
//
// Implemented in terms of WaitForCall with an Index-based predicate; the
// timeout error is rebuilt so it surfaces n and the current count.
func (this serviceMockCallsHelper[CT]) WaitForCount(n int, timeout time.Duration) error {
	if n <= 0 {
		return nil
	}
	if err := this.WaitForCall(func(c CT) bool { return this.indexGetter(c) >= n-1 }, timeout); err != nil {
		return errors.New("WaitForCount(" + strconv.Itoa(n) + ", " + timeout.String() + "): timed out, count = " + strconv.Itoa(this.Count()))
	}
	return nil
}

// WaitForCall blocks until at least one recorded call satisfies the
// predicate, or until timeout elapses. Returns nil on success, an error on
// timeout. Polls every 1ms; on each poll, scans every recorded call from
// oldest to newest.
func (this serviceMockCallsHelper[CT]) WaitForCall(predicate func(CT) bool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if slices.ContainsFunc(this.All(), predicate) {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("WaitForCall(" + timeout.String() + "): timed out, no matching call in " + strconv.Itoa(this.Count()) + " recorded entries")
		}
		time.Sleep(time.Millisecond)
	}
}

// ServiceMockCall is satisfied by every per-method <CallType>
// and by ServiceMockTimelineEntry. Enables HappenedBefore comparisons
// across concrete call types.
type ServiceMockCall interface {
	Global() int
}

// ServiceMockTimelineEntry records one call across any method in
// invocation order. Method names the source method; Index is the
// per-method position at the time the call was recorded; GlobalIndex is
// monotonically increasing across all methods.
type ServiceMockTimelineEntry struct {
	Method      string
	Index       int
	GlobalIndex int
}

// Global returns the global ordering position for this entry.
func (this ServiceMockTimelineEntry) Global() int { return this.GlobalIndex }

// HappenedBefore returns true iff this entry was recorded before other.
func (this ServiceMockTimelineEntry) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

// ServiceMockSendCall records one call to Service.Send.
// Index is the 0-based position of this call in the per-method log, useful
// for ordering assertions in WaitForCall predicates and All() iteration.
// Index resets to 0 when the log is cleared via Reset. GlobalIndex is the
// mock-wide position assigned when the call was recorded; see Timeline().
type ServiceMockSendCall struct {
	Index       int
	GlobalIndex int
	Param       ServiceMockSendParam
	Return      ServiceMockSendReturn
}

// Global returns the mock-wide ordering position for this call.
func (this ServiceMockSendCall) Global() int { return this.GlobalIndex }

// HappenedBefore returns true iff this call was recorded before other.
func (this ServiceMockSendCall) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

type ServiceMockSendParam struct {
	Ctx   context.Context
	Email string
}

type ServiceMockSendReturn struct {
	Id  string
	Err error
}

// serviceMockSendStub holds one set of return values — used for sticky cfg,
// Once() queue entries, and Always() fallback. Queue entries and
// conditional Always slots are referenced through pointers so per-slot
// setters held across subsequent armings remain valid.
type serviceMockSendStub struct {
	id  string
	err error
}

// serviceMockSendConditional is per-method conditional state — one entry
// per When() registration. The outer slice holds *Conditional (rather than
// values) so the conditional's own fields stay at a stable address across
// subsequent When() registrations.
type serviceMockSendConditional struct {
	pred       func(ServiceMockSendParam) bool
	stubQueue  []*serviceMockSendStub
	alwaysStub *serviceMockSendStub
}

type serviceMockSendCfg struct {
	sticky       serviceMockSendStub
	stickySet    bool
	stubQueue    []*serviceMockSendStub
	alwaysStub   *serviceMockSendStub
	conditionals []*serviceMockSendConditional
	handlerSet   bool

	handler func(context.Context, string) (string, error)
}

type ServiceMock struct {
	mu                sync.Mutex
	globalCallCounter int
	timelineLog       []ServiceMockTimelineEntry

	sendLog []ServiceMockSendCall
	sendCfg serviceMockSendCfg
}

// NewServiceMock returns a zero-valued ServiceMock. Equivalent
// to &ServiceMock{}; provided as a consistent entry point that may
// initialize internal state in future versions.
func NewServiceMock() *ServiceMock {
	return &ServiceMock{}
}

func (this *ServiceMock) Send(ctx context.Context, email string) (id string, err error) {
	this.mu.Lock()
	__mockParam := ServiceMockSendParam{Ctx: ctx, Email: email}
	__mockGlobal := this.globalCallCounter
	this.globalCallCounter++
	this.sendLog = append(this.sendLog, ServiceMockSendCall{
		Index:       len(this.sendLog),
		GlobalIndex: __mockGlobal,
		Param:       __mockParam,
	})
	__mockCallIdx := len(this.sendLog) - 1
	this.timelineLog = append(this.timelineLog, ServiceMockTimelineEntry{
		Method:      "Send",
		Index:       __mockCallIdx,
		GlobalIndex: __mockGlobal,
	})
	__mockHandler := this.sendCfg.handler
	var __mockSource serviceMockSendStub
	__mockMatched := false
	if len(this.sendCfg.conditionals) > 0 {
		__mockFirstMatch := -1
		__mockSecondMatch := -1
		__mockMatchCount := 0
		for __mockIdx := range this.sendCfg.conditionals {
			if this.sendCfg.conditionals[__mockIdx].pred(__mockParam) {
				__mockMatchCount++
				if __mockFirstMatch == -1 {
					__mockFirstMatch = __mockIdx
				} else if __mockSecondMatch == -1 {
					__mockSecondMatch = __mockIdx
				}
			}
		}
		if __mockMatchCount >= 2 {
			this.mu.Unlock()
			panic("ServiceMock.Send: call matched " + strconv.Itoa(__mockMatchCount) + " When predicates (indices " + strconv.Itoa(__mockFirstMatch) + " and " + strconv.Itoa(__mockSecondMatch) + "); refine the predicates or remove the overlap")
		}
		if __mockMatchCount == 1 && __mockHandler == nil {
			__mockCond := this.sendCfg.conditionals[__mockFirstMatch]
			if len(__mockCond.stubQueue) > 0 {
				__mockSource = *__mockCond.stubQueue[0]
				__mockCond.stubQueue = __mockCond.stubQueue[1:]
				__mockMatched = true
			} else if __mockCond.alwaysStub != nil {
				__mockSource = *__mockCond.alwaysStub
				__mockMatched = true
			}
		}
	}
	if !__mockMatched && __mockHandler == nil {
		if len(this.sendCfg.stubQueue) > 0 {
			__mockSource = *this.sendCfg.stubQueue[0]
			this.sendCfg.stubQueue = this.sendCfg.stubQueue[1:]
		} else if this.sendCfg.alwaysStub != nil {
			__mockSource = *this.sendCfg.alwaysStub
		} else {
			__mockSource = this.sendCfg.sticky
		}
	}
	__mockR0 := __mockSource.id
	__mockR1 := __mockSource.err
	this.mu.Unlock()

	var __mockRet0 string
	var __mockRet1 error
	if __mockHandler != nil {
		__mockRet0, __mockRet1 = __mockHandler(ctx, email)
	} else {
		__mockRet0, __mockRet1 = __mockR0, __mockR1
	}

	this.mu.Lock()
	if __mockCallIdx < len(this.sendLog) {
		this.sendLog[__mockCallIdx].Return = ServiceMockSendReturn{Id: __mockRet0, Err: __mockRet1}
	}
	this.mu.Unlock()

	return __mockRet0, __mockRet1
}

func (this *ServiceMock) Reset() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.globalCallCounter = 0
	this.timelineLog = nil
	this.sendLog = nil
	this.sendCfg = serviceMockSendCfg{}
}

type serviceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock and the cross-method
// timeline, without touching configured returns or installed handlers.
func (this serviceMockCalls) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.globalCallCounter = 0
	this.m.timelineLog = nil
	this.m.sendLog = nil
}

func (this serviceMockCalls) Send() serviceMockSendCalls {
	return serviceMockSendCalls{serviceMockCallsHelper: serviceMockCallsHelper[ServiceMockSendCall]{
		mu:          &this.m.mu,
		indexGetter: func(c ServiceMockSendCall) int { return c.Index },
		log:         &this.m.sendLog,
		pendingFn: func() int {
			n := len(this.m.sendCfg.stubQueue)
			for __mockI := range this.m.sendCfg.conditionals {
				n += len(this.m.sendCfg.conditionals[__mockI].stubQueue)
			}
			return n
		},
		indexer: this.m.callIndex,
	}}
}

// PendingStubs returns the total number of unconsumed Once() stub entries
// across all methods on this mock.
func (this serviceMockCalls) PendingStubs() int {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	n := 0
	n += len(this.m.sendCfg.stubQueue)
	for __mockI := range this.m.sendCfg.conditionals {
		n += len(this.m.sendCfg.conditionals[__mockI].stubQueue)
	}
	return n
}

// Timeline returns a defensive copy of the cross-method call timeline.
// Each entry names the source method, its per-method index, and a
// mock-wide GlobalIndex assigned at recording time.
func (this serviceMockCalls) Timeline() []ServiceMockTimelineEntry {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	out := make([]ServiceMockTimelineEntry, len(this.m.timelineLog))
	copy(out, this.m.timelineLog)
	return out
}

// AssertOrder calls t.Errorf when the given entries were not recorded in
// the given order. Adjacent entries with the same GlobalIndex (same call
// passed twice) are treated as in-order. 0 or 1 entries are a no-op.
func (this serviceMockCalls) AssertOrder(t ServiceMockVerifyT, entries ...ServiceMockCall) {
	t.Helper()
	if len(entries) < 2 {
		return
	}
	prev := entries[0].Global()
	for i := 1; i < len(entries); i++ {
		cur := entries[i].Global()
		if cur < prev {
			t.Errorf("ServiceMock: AssertOrder violated — entry " + strconv.Itoa(i) + " (global #" + strconv.Itoa(cur) + ") happened before entry " + strconv.Itoa(i-1) + " (global #" + strconv.Itoa(prev) + ")")
			return
		}
		prev = cur
	}
}

// ServiceMockVerifyT is the interface Verify accepts. *testing.T satisfies it.
type ServiceMockVerifyT interface {
	Helper()
	Errorf(format string, args ...any)
}

// Verify calls t.Errorf if any method has unconsumed Once() stub entries.
// It does not call t.Fatal so other teardown checks still run. The mock
// mutex is released before t.Errorf is invoked, so a user-supplied T
// implementation can safely call back into the mock without deadlocking.
func (this serviceMockCalls) Verify(t ServiceMockVerifyT) {
	t.Helper()
	this.m.mu.Lock()
	total, breakdown := 0, ""
	{
		n := len(this.m.sendCfg.stubQueue)
		for __mockI := range this.m.sendCfg.conditionals {
			n += len(this.m.sendCfg.conditionals[__mockI].stubQueue)
		}
		if n > 0 {
			total += n
			if breakdown != "" {
				breakdown += ", "
			}
			breakdown += "Send: " + strconv.Itoa(n)
		}
	}
	this.m.mu.Unlock()
	if total > 0 {
		t.Errorf("ServiceMock: " + strconv.Itoa(total) + " unconsumed stubs (" + breakdown + ")")
	}
}

// serviceMockSendCalls is the per-method inspection namespace for
// Service.Send. Methods are provided by the embedded serviceMockCallsHelper.
type serviceMockSendCalls struct {
	serviceMockCallsHelper[ServiceMockSendCall]
}

type serviceMockOn struct{ m *ServiceMock }

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() serviceMockOn {
	return serviceMockOn{m: this}
}

// Reset clears configured returns and installed handlers for every method
// on the mock, without touching the call log.
func (this serviceMockOn) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg = serviceMockSendCfg{}
}

func (this serviceMockOn) Send() serviceMockSendOn {
	return serviceMockSendOn{m: this.m}
}

type serviceMockSendOn struct{ m *ServiceMock }

// ReturnAll sets all static return values for Send
// positionally. Used when no ReturnCustom handler is installed.
func (this serviceMockSendOn) ReturnAll(id string, err error) {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	if this.m.sendCfg.alwaysStub != nil || len(this.m.sendCfg.stubQueue) > 0 {
		panic("ServiceMock.Send: cannot set sticky return while Once()/Always() entries are armed; call .Reset() first")
	}
	this.m.sendCfg.stickySet = true
	this.m.sendCfg.sticky.id = id
	this.m.sendCfg.sticky.err = err
}

// Return returns the per-return setter namespace — one method per return
// value of Send, each overriding just that return without
// touching the others.
func (this serviceMockSendOn) Return() serviceMockSendReturn {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	if this.m.sendCfg.alwaysStub != nil || len(this.m.sendCfg.stubQueue) > 0 {
		panic("ServiceMock.Send: cannot set sticky return while Once()/Always() entries are armed; call .Reset() first")
	}
	this.m.sendCfg.stickySet = true
	return serviceMockSendReturn{m: this.m, target: &this.m.sendCfg.sticky}
}

// ReturnCustom installs a dynamic function called instead of returning the
// configured static values.
func (this serviceMockSendOn) ReturnCustom(value func(context.Context, string) (string, error)) {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	if this.m.sendCfg.alwaysStub != nil || len(this.m.sendCfg.stubQueue) > 0 {
		panic("ServiceMock.Send: cannot set sticky return while Once()/Always() entries are armed; call .Reset() first")
	}
	this.m.sendCfg.handlerSet = true
	this.m.sendCfg.handler = value
}

// Once arms a single-call return; the stub is consumed by the next call
// to Send and then discarded.
func (this serviceMockSendOn) Once() serviceMockSendReturn {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	if this.m.sendCfg.stickySet || this.m.sendCfg.handlerSet {
		panic("ServiceMock.Send: cannot arm Once()/Always() while a sticky return is set; call .Reset() first")
	}
	__mockStub := &serviceMockSendStub{}
	this.m.sendCfg.stubQueue = append(this.m.sendCfg.stubQueue, __mockStub)
	return serviceMockSendReturn{m: this.m, target: __mockStub}
}

// Always arms a fallback return used when the Once() queue is exhausted.
// Subsequent calls to Always() replace the previous fallback.
func (this serviceMockSendOn) Always() serviceMockSendReturn {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	if this.m.sendCfg.stickySet || this.m.sendCfg.handlerSet {
		panic("ServiceMock.Send: cannot arm Once()/Always() while a sticky return is set; call .Reset() first")
	}
	this.m.sendCfg.alwaysStub = &serviceMockSendStub{}
	return serviceMockSendReturn{m: this.m, target: this.m.sendCfg.alwaysStub}
}

// When registers a conditional stub set whose entries fire only when
// the call's Param satisfies the predicate. Use Once()/Always() on the
// returned namespace to attach armings.
func (this serviceMockSendOn) When(pred func(ServiceMockSendParam) bool) serviceMockSendOnWhen {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg.conditionals = append(
		this.m.sendCfg.conditionals,
		&serviceMockSendConditional{pred: pred},
	)
	return serviceMockSendOnWhen{m: this.m, condIndex: len(this.m.sendCfg.conditionals) - 1}
}

// Reset clears configured returns and installed handler for Send.
func (this serviceMockSendOn) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg = serviceMockSendCfg{}
}

// serviceMockSendReturn is the per-return setter namespace for
// Send. Each method overrides one return value without
// touching the others. Methods return the setter itself for chaining.
type serviceMockSendReturn struct {
	m      *ServiceMock
	target *serviceMockSendStub
}

// Return is a no-op that returns this setter, allowing Once().Return().X()
// and Always().Return().X() chaining syntax.
func (this serviceMockSendReturn) Return() serviceMockSendReturn {
	return this
}

// Id overrides just the id return value.
func (this serviceMockSendReturn) Id(value string) serviceMockSendReturn {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.target.id = value
	return this
}

// Err overrides just the err return value.
func (this serviceMockSendReturn) Err(value error) serviceMockSendReturn {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.target.err = value
	return this
}

// serviceMockSendOnWhen is the conditional namespace returned by When(pred).
// Exposes Once() (queue a single-shot stub) and Always() (install/replace
// the fallback stub) — both scoped to the matching predicate.
type serviceMockSendOnWhen struct {
	m         *ServiceMock
	condIndex int
}

// Once arms a single-call return that fires only when this conditional's
// predicate matches.
func (this serviceMockSendOnWhen) Once() serviceMockSendReturn {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	if this.condIndex >= len(this.m.sendCfg.conditionals) {
		panic("ServiceMock.Send: stale When() handle; conditionals were cleared (via Reset) after the When() call that produced this handle")
	}
	cond := this.m.sendCfg.conditionals[this.condIndex]
	__mockStub := &serviceMockSendStub{}
	cond.stubQueue = append(cond.stubQueue, __mockStub)
	return serviceMockSendReturn{m: this.m, target: __mockStub}
}

// Always arms a fallback return that fires on every call matching this
// conditional's predicate, used when the conditional's Once() queue is
// exhausted.
func (this serviceMockSendOnWhen) Always() serviceMockSendReturn {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	if this.condIndex >= len(this.m.sendCfg.conditionals) {
		panic("ServiceMock.Send: stale When() handle; conditionals were cleared (via Reset) after the When() call that produced this handle")
	}
	cond := this.m.sendCfg.conditionals[this.condIndex]
	cond.alwaysStub = &serviceMockSendStub{}
	return serviceMockSendReturn{m: this.m, target: cond.alwaysStub}
}

// callIndex resolves the index argument of Method_call(i)-style indexers:
// negative i counts from the end; out-of-range i panics with a message
// that surfaces both the original input and the log length.
func (this *ServiceMock) callIndex(length, i int) int {
	resolved := i
	if resolved < 0 {
		resolved += length
	}
	if resolved < 0 || resolved >= length {
		if i < 0 {
			panic("ServiceMock call index " + strconv.Itoa(i) + " (resolved to " + strconv.Itoa(resolved) + ") out of range, log has " + strconv.Itoa(length) + " entries")
		}
		panic("ServiceMock call index " + strconv.Itoa(i) + " out of range, log has " + strconv.Itoa(length) + " entries")
	}
	return resolved
}
`, Options{Async: true}),

		Entry("with header note", `
package pkg

type Service interface {
	DoSomething()
}`, `// Code generated by MockGen. DO NOT EDIT.

package test

import (
	"strconv"
)

type serviceMockCallsHelper[CT any] struct {
	log       *[]CT
	pendingFn func() int
	indexer   func(length, i int) int
}

// Count returns the number of recorded calls.
func (this serviceMockCallsHelper[CT]) Count() int {
	return len(*this.log)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this serviceMockCallsHelper[CT]) At(i int) CT {
	return (*this.log)[this.indexer(len(*this.log), i)]
}

// Last is sugar for At(-1).
func (this serviceMockCallsHelper[CT]) Last() CT {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this serviceMockCallsHelper[CT]) All() []CT {
	out := make([]CT, len(*this.log))
	copy(out, *this.log)
	return out
}

// Reset clears the call log.
func (this serviceMockCallsHelper[CT]) Reset() {
	*this.log = nil
}

// PendingStubs returns the number of unconsumed Once() stub entries.
// Returns 0 for methods with no return values.
func (this serviceMockCallsHelper[CT]) PendingStubs() int {
	return this.pendingFn()
}

// Any returns true iff at least one recorded call satisfies pred.
// Pred runs after All() returns (no lock held); pred may safely call
// back into the mock without deadlocking.
func (this serviceMockCallsHelper[CT]) Any(pred func(CT) bool) bool {
	if pred == nil {
		panic("serviceMockCallsHelper.Any: predicate is nil")
	}
	for _, c := range this.All() {
		if pred(c) {
			return true
		}
	}
	return false
}

// Filter returns every recorded call satisfying pred, in invocation order.
// Returns an empty (non-nil) slice when no call matches. Pred runs after
// All() returns; pred may safely call back into the mock.
func (this serviceMockCallsHelper[CT]) Filter(pred func(CT) bool) []CT {
	if pred == nil {
		panic("serviceMockCallsHelper.Filter: predicate is nil")
	}
	out := []CT{}
	for _, c := range this.All() {
		if pred(c) {
			out = append(out, c)
		}
	}
	return out
}

// ServiceMockCall is satisfied by every per-method <CallType>
// and by ServiceMockTimelineEntry. Enables HappenedBefore comparisons
// across concrete call types.
type ServiceMockCall interface {
	Global() int
}

// ServiceMockTimelineEntry records one call across any method in
// invocation order. Method names the source method; Index is the
// per-method position at the time the call was recorded; GlobalIndex is
// monotonically increasing across all methods.
type ServiceMockTimelineEntry struct {
	Method      string
	Index       int
	GlobalIndex int
}

// Global returns the global ordering position for this entry.
func (this ServiceMockTimelineEntry) Global() int { return this.GlobalIndex }

// HappenedBefore returns true iff this entry was recorded before other.
func (this ServiceMockTimelineEntry) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

// ServiceMockDoSomethingCall records one call to Service.DoSomething.
// Index is the 0-based position of this call in the per-method log, useful
// for ordering assertions in WaitForCall predicates and All() iteration.
// Index resets to 0 when the log is cleared via Reset. GlobalIndex is the
// mock-wide position assigned when the call was recorded; see Timeline().
type ServiceMockDoSomethingCall struct {
	Index       int
	GlobalIndex int
	Param       ServiceMockDoSomethingParam
	Return      ServiceMockDoSomethingReturn
}

// Global returns the mock-wide ordering position for this call.
func (this ServiceMockDoSomethingCall) Global() int { return this.GlobalIndex }

// HappenedBefore returns true iff this call was recorded before other.
func (this ServiceMockDoSomethingCall) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

type ServiceMockDoSomethingParam struct {
}

type ServiceMockDoSomethingReturn struct {
}

type serviceMockDoSomethingCfg struct {
	sticky    struct{}
	stickySet bool
}

type ServiceMock struct {
	globalCallCounter int
	timelineLog       []ServiceMockTimelineEntry

	doSomethingLog []ServiceMockDoSomethingCall
	doSomethingCfg serviceMockDoSomethingCfg
}

// NewServiceMock returns a zero-valued ServiceMock. Equivalent
// to &ServiceMock{}; provided as a consistent entry point that may
// initialize internal state in future versions.
func NewServiceMock() *ServiceMock {
	return &ServiceMock{}
}

func (this *ServiceMock) DoSomething() {
	__mockGlobal := this.globalCallCounter
	this.globalCallCounter++
	this.doSomethingLog = append(this.doSomethingLog, ServiceMockDoSomethingCall{
		Index:       len(this.doSomethingLog),
		GlobalIndex: __mockGlobal,
		Param:       ServiceMockDoSomethingParam{},
	})
	this.timelineLog = append(this.timelineLog, ServiceMockTimelineEntry{
		Method:      "DoSomething",
		Index:       len(this.doSomethingLog) - 1,
		GlobalIndex: __mockGlobal,
	})
}

func (this *ServiceMock) Reset() {
	this.globalCallCounter = 0
	this.timelineLog = nil
	this.doSomethingLog = nil
	this.doSomethingCfg = serviceMockDoSomethingCfg{}
}

type serviceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock and the cross-method
// timeline, without touching configured returns or installed handlers.
func (this serviceMockCalls) Reset() {
	this.m.globalCallCounter = 0
	this.m.timelineLog = nil
	this.m.doSomethingLog = nil
}

func (this serviceMockCalls) DoSomething() serviceMockDoSomethingCalls {
	return serviceMockDoSomethingCalls{serviceMockCallsHelper: serviceMockCallsHelper[ServiceMockDoSomethingCall]{log: &this.m.doSomethingLog,
		pendingFn: func() int {
			return 0
		},
		indexer: this.m.callIndex,
	}}
}

// PendingStubs returns the total number of unconsumed Once() stub entries
// across all methods on this mock.
func (this serviceMockCalls) PendingStubs() int {
	n := 0
	return n
}

// Timeline returns a defensive copy of the cross-method call timeline.
// Each entry names the source method, its per-method index, and a
// mock-wide GlobalIndex assigned at recording time.
func (this serviceMockCalls) Timeline() []ServiceMockTimelineEntry {
	out := make([]ServiceMockTimelineEntry, len(this.m.timelineLog))
	copy(out, this.m.timelineLog)
	return out
}

// AssertOrder calls t.Errorf when the given entries were not recorded in
// the given order. Adjacent entries with the same GlobalIndex (same call
// passed twice) are treated as in-order. 0 or 1 entries are a no-op.
func (this serviceMockCalls) AssertOrder(t ServiceMockVerifyT, entries ...ServiceMockCall) {
	t.Helper()
	if len(entries) < 2 {
		return
	}
	prev := entries[0].Global()
	for i := 1; i < len(entries); i++ {
		cur := entries[i].Global()
		if cur < prev {
			t.Errorf("ServiceMock: AssertOrder violated — entry " + strconv.Itoa(i) + " (global #" + strconv.Itoa(cur) + ") happened before entry " + strconv.Itoa(i-1) + " (global #" + strconv.Itoa(prev) + ")")
			return
		}
		prev = cur
	}
}

// ServiceMockVerifyT is the interface Verify accepts. *testing.T satisfies it.
type ServiceMockVerifyT interface {
	Helper()
	Errorf(format string, args ...any)
}

// Verify calls t.Errorf if any method has unconsumed Once() stub entries.
// It does not call t.Fatal so other teardown checks still run. The mock
// mutex is released before t.Errorf is invoked, so a user-supplied T
// implementation can safely call back into the mock without deadlocking.
func (this serviceMockCalls) Verify(t ServiceMockVerifyT) {
	t.Helper()
	total, breakdown := 0, ""
	if total > 0 {
		t.Errorf("ServiceMock: " + strconv.Itoa(total) + " unconsumed stubs (" + breakdown + ")")
	}
}

// serviceMockDoSomethingCalls is the per-method inspection namespace for
// Service.DoSomething. Methods are provided by the embedded serviceMockCallsHelper.
type serviceMockDoSomethingCalls struct {
	serviceMockCallsHelper[ServiceMockDoSomethingCall]
}

type serviceMockOn struct{ m *ServiceMock }

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() serviceMockOn {
	return serviceMockOn{m: this}
}

// Reset clears configured returns and installed handlers for every method
// on the mock, without touching the call log.
func (this serviceMockOn) Reset() {
	this.m.doSomethingCfg = serviceMockDoSomethingCfg{}
}

func (this serviceMockOn) DoSomething() serviceMockDoSomethingOn {
	return serviceMockDoSomethingOn{m: this.m}
}

type serviceMockDoSomethingOn struct{ m *ServiceMock }

// Reset clears configured returns and installed handler for DoSomething.
func (this serviceMockDoSomethingOn) Reset() {
	this.m.doSomethingCfg = serviceMockDoSomethingCfg{}
}

// callIndex resolves the index argument of Method_call(i)-style indexers:
// negative i counts from the end; out-of-range i panics with a message
// that surfaces both the original input and the log length.
func (this *ServiceMock) callIndex(length, i int) int {
	resolved := i
	if resolved < 0 {
		resolved += length
	}
	if resolved < 0 || resolved >= length {
		if i < 0 {
			panic("ServiceMock call index " + strconv.Itoa(i) + " (resolved to " + strconv.Itoa(resolved) + ") out of range, log has " + strconv.Itoa(length) + " entries")
		}
		panic("ServiceMock call index " + strconv.Itoa(i) + " out of range, log has " + strconv.Itoa(length) + " entries")
	}
	return resolved
}
`, Options{AddHeaderNote: true}),
	)
})
