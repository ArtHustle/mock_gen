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
type ServiceMockProcessCall struct {
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

func (this *ServiceMock) Process() {
	this.processLog = append(this.processLog, ServiceMockProcessCall{
		Param: ServiceMockProcessParam{},
	})
}

func (this *ServiceMock) Reset() {
	this.processLog = nil
	this.processCfg = serviceMockProcessCfg{}
}

type ServiceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() ServiceMockCalls {
	return ServiceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock without touching
// configured returns or installed handlers.
func (this ServiceMockCalls) Reset() {
	this.m.processLog = nil
}

func (this ServiceMockCalls) Process() ServiceMockProcessCalls {
	return ServiceMockProcessCalls{m: this.m}
}

type ServiceMockProcessCalls struct{ m *ServiceMock }

// Count returns the number of recorded calls to Service.Process.
func (this ServiceMockProcessCalls) Count() int {
	return len(this.m.processLog)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this ServiceMockProcessCalls) At(i int) ServiceMockProcessCall {
	return this.m.processLog[this.m.callIndex(len(this.m.processLog), i)]
}

// Last is sugar for At(-1).
func (this ServiceMockProcessCalls) Last() ServiceMockProcessCall {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this ServiceMockProcessCalls) All() []ServiceMockProcessCall {
	out := make([]ServiceMockProcessCall, len(this.m.processLog))
	copy(out, this.m.processLog)
	return out
}

// Reset clears the call log for Process.
func (this ServiceMockProcessCalls) Reset() {
	this.m.processLog = nil
}

type ServiceMockOn struct{ m *ServiceMock }

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() ServiceMockOn {
	return ServiceMockOn{m: this}
}

// Reset clears configured returns and installed handlers for every method
// on the mock, without touching the call log.
func (this ServiceMockOn) Reset() {
	this.m.processCfg = serviceMockProcessCfg{}
}

func (this ServiceMockOn) Process() ServiceMockProcessOn {
	return ServiceMockProcessOn{m: this.m}
}

type ServiceMockProcessOn struct{ m *ServiceMock }

// Reset clears configured returns and installed handler for Process.
func (this ServiceMockProcessOn) Reset() {
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
	"strconv"
	"sync"
)

// ServiceMockSendCall records one call to Service.Send.
type ServiceMockSendCall struct {
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

func (this *ServiceMock) Send(ctx context.Context, email string) (id string, err error) {
	this.mu.Lock()
	this.sendLog = append(this.sendLog, ServiceMockSendCall{
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

type ServiceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() ServiceMockCalls {
	return ServiceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock without touching
// configured returns or installed handlers.
func (this ServiceMockCalls) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendLog = nil
}

func (this ServiceMockCalls) Send() ServiceMockSendCalls {
	return ServiceMockSendCalls{m: this.m}
}

type ServiceMockSendCalls struct{ m *ServiceMock }

// Count returns the number of recorded calls to Service.Send.
func (this ServiceMockSendCalls) Count() int {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	return len(this.m.sendLog)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this ServiceMockSendCalls) At(i int) ServiceMockSendCall {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	return this.m.sendLog[this.m.callIndex(len(this.m.sendLog), i)]
}

// Last is sugar for At(-1).
func (this ServiceMockSendCalls) Last() ServiceMockSendCall {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this ServiceMockSendCalls) All() []ServiceMockSendCall {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	out := make([]ServiceMockSendCall, len(this.m.sendLog))
	copy(out, this.m.sendLog)
	return out
}

// Reset clears the call log for Send.
func (this ServiceMockSendCalls) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendLog = nil
}

type ServiceMockOn struct{ m *ServiceMock }

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() ServiceMockOn {
	return ServiceMockOn{m: this}
}

// Reset clears configured returns and installed handlers for every method
// on the mock, without touching the call log.
func (this ServiceMockOn) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg = serviceMockSendCfg{}
}

func (this ServiceMockOn) Send() ServiceMockSendOn {
	return ServiceMockSendOn{m: this.m}
}

type ServiceMockSendOn struct{ m *ServiceMock }

// Returns sets all static return values for Send.
// Used when no Handler is installed.
func (this ServiceMockSendOn) Returns(id string, err error) {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg.id = id
	this.m.sendCfg.err = err
}

// ReturnsId overrides just the id return value.
func (this ServiceMockSendOn) ReturnsId(value string) {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg.id = value
}

// ReturnsErr overrides just the err return value.
func (this ServiceMockSendOn) ReturnsErr(value error) {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg.err = value
}

// Handler installs a dynamic function called instead of returning the
// configured static values.
func (this ServiceMockSendOn) Handler(value func(context.Context, string) (string, error)) {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg.handler = value
}

// Reset clears configured returns and installed handler for Send.
func (this ServiceMockSendOn) Reset() {
	this.m.mu.Lock()
	defer this.m.mu.Unlock()
	this.m.sendCfg = serviceMockSendCfg{}
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
`, Options{MakeThreadSafe: true}),

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
type ServiceMockDoSomethingCall struct {
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

func (this *ServiceMock) DoSomething() {
	this.doSomethingLog = append(this.doSomethingLog, ServiceMockDoSomethingCall{
		Param: ServiceMockDoSomethingParam{},
	})
}

func (this *ServiceMock) Reset() {
	this.doSomethingLog = nil
	this.doSomethingCfg = serviceMockDoSomethingCfg{}
}

type ServiceMockCalls struct{ m *ServiceMock }

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() ServiceMockCalls {
	return ServiceMockCalls{m: this}
}

// Reset clears every per-method call log on the mock without touching
// configured returns or installed handlers.
func (this ServiceMockCalls) Reset() {
	this.m.doSomethingLog = nil
}

func (this ServiceMockCalls) DoSomething() ServiceMockDoSomethingCalls {
	return ServiceMockDoSomethingCalls{m: this.m}
}

type ServiceMockDoSomethingCalls struct{ m *ServiceMock }

// Count returns the number of recorded calls to Service.DoSomething.
func (this ServiceMockDoSomethingCalls) Count() int {
	return len(this.m.doSomethingLog)
}

// At returns the i-th recorded call. Negative i counts from the end.
// Panics on out-of-range with a message that surfaces both the input
// and the log length.
func (this ServiceMockDoSomethingCalls) At(i int) ServiceMockDoSomethingCall {
	return this.m.doSomethingLog[this.m.callIndex(len(this.m.doSomethingLog), i)]
}

// Last is sugar for At(-1).
func (this ServiceMockDoSomethingCalls) Last() ServiceMockDoSomethingCall {
	return this.At(-1)
}

// All returns a defensive copy of the full call log.
func (this ServiceMockDoSomethingCalls) All() []ServiceMockDoSomethingCall {
	out := make([]ServiceMockDoSomethingCall, len(this.m.doSomethingLog))
	copy(out, this.m.doSomethingLog)
	return out
}

// Reset clears the call log for DoSomething.
func (this ServiceMockDoSomethingCalls) Reset() {
	this.m.doSomethingLog = nil
}

type ServiceMockOn struct{ m *ServiceMock }

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() ServiceMockOn {
	return ServiceMockOn{m: this}
}

// Reset clears configured returns and installed handlers for every method
// on the mock, without touching the call log.
func (this ServiceMockOn) Reset() {
	this.m.doSomethingCfg = serviceMockDoSomethingCfg{}
}

func (this ServiceMockOn) DoSomething() ServiceMockDoSomethingOn {
	return ServiceMockDoSomethingOn{m: this.m}
}

type ServiceMockDoSomethingOn struct{ m *ServiceMock }

// Reset clears configured returns and installed handler for DoSomething.
func (this ServiceMockDoSomethingOn) Reset() {
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
