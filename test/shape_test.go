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

// ServiceMockProcessCall records one call to Service.Process.
// Index is the 0-based position of this call in the per-method log, useful
// for ordering assertions in WaitForCall predicates and All() iteration.
// Index resets to 0 when the log is cleared via Reset.
type ServiceMockProcessCall struct {
	Index  int
	Param  ServiceMockProcessParam
	Return ServiceMockProcessReturn
}

type ServiceMockProcessParam struct {
}

type ServiceMockProcessReturn struct {
}

type serviceMockProcessCfg struct {
}

type ServiceMock struct {
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
	this.processLog = append(this.processLog, ServiceMockProcessCall{
		Index: len(this.processLog),
		Param: ServiceMockProcessParam{},
	})
}

func (this *ServiceMock) Reset() {
	this.processLog = nil
	this.processCfg = serviceMockProcessCfg{}
}

type serviceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock without touching
// configured returns or installed handlers.
func (this serviceMockCalls) Reset() {
	this.m.processLog = nil
}

func (this serviceMockCalls) Process() serviceMockProcessCalls {
	return serviceMockProcessCalls{m: this.m}
}

type serviceMockProcessCalls struct{ m *ServiceMock }

// Count returns the number of recorded calls to Service.Process.
func (this serviceMockProcessCalls) Count() int {
	return len(this.m.processLog)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this serviceMockProcessCalls) At(i int) ServiceMockProcessCall {
	return this.m.processLog[this.m.callIndex(len(this.m.processLog), i)]
}

// Last is sugar for At(-1).
func (this serviceMockProcessCalls) Last() ServiceMockProcessCall {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this serviceMockProcessCalls) All() []ServiceMockProcessCall {
	out := make([]ServiceMockProcessCall, len(this.m.processLog))
	copy(out, this.m.processLog)
	return out
}

// Reset clears the call log for Process.
func (this serviceMockProcessCalls) Reset() {
	this.m.processLog = nil
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

// ServiceMockSendCall records one call to Service.Send.
// Index is the 0-based position of this call in the per-method log, useful
// for ordering assertions in WaitForCall predicates and All() iteration.
// Index resets to 0 when the log is cleared via Reset.
type ServiceMockSendCall struct {
	Index  int
	Param  ServiceMockSendParam
	Return ServiceMockSendReturn
}

type ServiceMockSendParam struct {
	Ctx   context.Context
	Email string
}

type ServiceMockSendReturn struct {
	Id  string
	Err error
}

type serviceMockSendCfg struct {
	id      string
	err     error
	handler func(context.Context, string) (string, error)
}

type ServiceMock struct {
	mu sync.Mutex

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
	this.sendLog = append(this.sendLog, ServiceMockSendCall{
		Index: len(this.sendLog),
		Param: ServiceMockSendParam{Ctx: ctx, Email: email},
	})
	__mockCallIdx := len(this.sendLog) - 1
	__mockHandler := this.sendCfg.handler
	__mockR0 := this.sendCfg.id
	__mockR1 := this.sendCfg.err
	this.mu.Unlock()

	var __mockRet0 string
	var __mockRet1 error
	if __mockHandler != nil {
		__mockRet0, __mockRet1 = __mockHandler(ctx, email)
	} else {
		__mockRet0, __mockRet1 = __mockR0, __mockR1
	}

	this.mu.Lock()
	this.sendLog[__mockCallIdx].Return = ServiceMockSendReturn{Id: __mockRet0, Err: __mockRet1}
	this.mu.Unlock()

	return __mockRet0, __mockRet1
}

func (this *ServiceMock) Reset() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.sendLog = nil
	this.sendCfg = serviceMockSendCfg{}
}

type serviceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock without touching
// configured returns or installed handlers.
func (this serviceMockCalls) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendLog = nil
}

func (this serviceMockCalls) Send() serviceMockSendCalls {
	return serviceMockSendCalls{m: this.m}
}

type serviceMockSendCalls struct{ m *ServiceMock }

// Count returns the number of recorded calls to Service.Send.
func (this serviceMockSendCalls) Count() int {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	return len(this.m.sendLog)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this serviceMockSendCalls) At(i int) ServiceMockSendCall {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	return this.m.sendLog[this.m.callIndex(len(this.m.sendLog), i)]
}

// Last is sugar for At(-1).
func (this serviceMockSendCalls) Last() ServiceMockSendCall {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this serviceMockSendCalls) All() []ServiceMockSendCall {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	out := make([]ServiceMockSendCall, len(this.m.sendLog))
	copy(out, this.m.sendLog)
	return out
}

// Reset clears the call log for Send.
func (this serviceMockSendCalls) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendLog = nil
}

// WaitForCount blocks until at least n calls to Send have been
// recorded, or until timeout elapses. Returns nil on success, an error
// describing the current count on timeout. Polls every 1ms.
//
// Implemented in terms of WaitForCall with an Index-based predicate; the
// timeout error is rebuilt so it surfaces n and the current count.
func (this serviceMockSendCalls) WaitForCount(n int, timeout time.Duration) error {
	if n <= 0 {
		return nil
	}
	if err := this.WaitForCall(func(c ServiceMockSendCall) bool { return c.Index >= n-1 }, timeout); err != nil {
		return errors.New("WaitForCount(" + strconv.Itoa(n) + ", " + timeout.String() + "): timed out, count = " + strconv.Itoa(this.Count()))
	}
	return nil
}

// WaitForCall blocks until at least one recorded call to Send
// satisfies the predicate, or until timeout elapses. Returns nil on
// success, an error on timeout. Polls every 1ms; on each poll, scans every
// recorded call from oldest to newest.
func (this serviceMockSendCalls) WaitForCall(predicate func(ServiceMockSendCall) bool, timeout time.Duration) error {
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
	this.m.sendCfg.id = id
	this.m.sendCfg.err = err
}

// Return returns the per-return setter namespace — one method per return
// value of Send, each overriding just that return without
// touching the others.
func (this serviceMockSendOn) Return() serviceMockSendReturn {
	return serviceMockSendReturn{m: this.m}
}

// ReturnCustom installs a dynamic function called instead of returning the
// configured static values.
func (this serviceMockSendOn) ReturnCustom(value func(context.Context, string) (string, error)) {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg.handler = value
}

// Reset clears configured returns and installed handler for Send.
func (this serviceMockSendOn) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg = serviceMockSendCfg{}
}

// serviceMockSendReturn is the per-return setter namespace for
// Send. Each method overrides one return value without
// touching the others.
type serviceMockSendReturn struct{ m *ServiceMock }

// Id overrides just the id return value.
func (this serviceMockSendReturn) Id(value string) {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg.id = value
}

// Err overrides just the err return value.
func (this serviceMockSendReturn) Err(value error) {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg.err = value
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

// ServiceMockDoSomethingCall records one call to Service.DoSomething.
// Index is the 0-based position of this call in the per-method log, useful
// for ordering assertions in WaitForCall predicates and All() iteration.
// Index resets to 0 when the log is cleared via Reset.
type ServiceMockDoSomethingCall struct {
	Index  int
	Param  ServiceMockDoSomethingParam
	Return ServiceMockDoSomethingReturn
}

type ServiceMockDoSomethingParam struct {
}

type ServiceMockDoSomethingReturn struct {
}

type serviceMockDoSomethingCfg struct {
}

type ServiceMock struct {
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
	this.doSomethingLog = append(this.doSomethingLog, ServiceMockDoSomethingCall{
		Index: len(this.doSomethingLog),
		Param: ServiceMockDoSomethingParam{},
	})
}

func (this *ServiceMock) Reset() {
	this.doSomethingLog = nil
	this.doSomethingCfg = serviceMockDoSomethingCfg{}
}

type serviceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock without touching
// configured returns or installed handlers.
func (this serviceMockCalls) Reset() {
	this.m.doSomethingLog = nil
}

func (this serviceMockCalls) DoSomething() serviceMockDoSomethingCalls {
	return serviceMockDoSomethingCalls{m: this.m}
}

type serviceMockDoSomethingCalls struct{ m *ServiceMock }

// Count returns the number of recorded calls to Service.DoSomething.
func (this serviceMockDoSomethingCalls) Count() int {
	return len(this.m.doSomethingLog)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this serviceMockDoSomethingCalls) At(i int) ServiceMockDoSomethingCall {
	return this.m.doSomethingLog[this.m.callIndex(len(this.m.doSomethingLog), i)]
}

// Last is sugar for At(-1).
func (this serviceMockDoSomethingCalls) Last() ServiceMockDoSomethingCall {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this serviceMockDoSomethingCalls) All() []ServiceMockDoSomethingCall {
	out := make([]ServiceMockDoSomethingCall, len(this.m.doSomethingLog))
	copy(out, this.m.doSomethingLog)
	return out
}

// Reset clears the call log for DoSomething.
func (this serviceMockDoSomethingCalls) Reset() {
	this.m.doSomethingLog = nil
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
