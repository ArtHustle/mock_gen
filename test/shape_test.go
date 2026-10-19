package test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/ArtHustle/mock_gen"
)

// Canonical literal-output entries. The shared logic lives in mockrt, so a
// mock is its Param/Return structs, the interface methods and thin typed
// wrappers over the runtime; these few canonical shapes pin down that
// structure. Behavior is verified in behavior_test.go; field-naming edge
// cases in field_naming_test.go.
var _ = Describe("Generated mock shape", func() {
	DescribeTable("emits the expected mock",
		runGenEntry,

		Entry("default options: method with no params/returns", `
package pkg

type Service interface {
	Process()
}`, `package test

import (
	"github.com/ArtHustle/mock_gen/mockrt"
)

// ServiceMock implements Service. Inspect recorded calls through
// Calls() and configure returns through On().
type ServiceMock struct {
	rt            mockrt.Mock
	processMethod mockrt.Method
}

// NewServiceMock returns a zero-valued ServiceMock. Equivalent
// to &ServiceMock{}; provided as a consistent entry point that may
// initialize internal state in future versions.
func NewServiceMock() *ServiceMock {
	return &ServiceMock{}
}

type (
	ServiceMockCall          = mockrt.Call
	ServiceMockTimelineEntry = mockrt.TimelineEntry
	ServiceMockVerifyT       = mockrt.T
)

func (this *ServiceMock) ref(method string) mockrt.Ref {
	return mockrt.Ref{Mock: &this.rt, Name: "ServiceMock", Method: method}
}

func (this *ServiceMock) group() mockrt.Group {
	return mockrt.Group{Ref: this.ref(""), Methods: []mockrt.Named{
		{Name: "Process", Method: &this.processMethod},
	}}
}

// Reset clears every call log, the timeline and every configured return.
func (this *ServiceMock) Reset() { this.group().Reset() }

type serviceMockCalls struct {
	mockrt.MockCalls
	m *ServiceMock
}

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{this.group().MockCalls(), this}
}

type serviceMockOn struct {
	mockrt.MockOn
	m *ServiceMock
}

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() serviceMockOn {
	return serviceMockOn{this.group().MockOn(), this}
}

type ServiceMockProcessParam struct{}

type ServiceMockProcessReturn struct{}

// ServiceMockProcessCall records one call to Service.Process.
type ServiceMockProcessCall struct {
	Index       int
	GlobalIndex int
	Param       ServiceMockProcessParam
	Return      ServiceMockProcessReturn
}

func (this ServiceMockProcessCall) Global() int { return this.GlobalIndex }
func (this ServiceMockProcessCall) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

func serviceMockProcessCallOf(r mockrt.Rec) ServiceMockProcessCall {
	ret, _ := r.Return.(ServiceMockProcessReturn)
	return ServiceMockProcessCall{Index: r.Index, GlobalIndex: r.GlobalIndex, Param: r.Param.(ServiceMockProcessParam), Return: ret}
}

func (this *ServiceMock) Process() {
	this.processMethod.Call(this.ref("Process"), ServiceMockProcessParam{})
}

func (this serviceMockCalls) Process() serviceMockProcessCalls {
	return serviceMockProcessCalls{this.m.processMethod.Calls(this.m.ref("Process"))}
}

type serviceMockProcessCalls struct{ mockrt.Calls }

func (this serviceMockProcessCalls) At(i int) ServiceMockProcessCall {
	return serviceMockProcessCallOf(this.Calls.At(i))
}

func (this serviceMockProcessCalls) Last() ServiceMockProcessCall { return this.At(-1) }
func (this serviceMockProcessCalls) All() []ServiceMockProcessCall {
	return mockrt.Map(this.Calls.All(), serviceMockProcessCallOf)
}

func (this serviceMockProcessCalls) Any(pred func(ServiceMockProcessCall) bool) bool {
	return this.Calls.Any(mockrt.Pred(pred, serviceMockProcessCallOf))
}

func (this serviceMockProcessCalls) Filter(pred func(ServiceMockProcessCall) bool) []ServiceMockProcessCall {
	return mockrt.Map(this.Calls.Filter(mockrt.Pred(pred, serviceMockProcessCallOf)), serviceMockProcessCallOf)
}

func (this serviceMockOn) Process() mockrt.OnVoid {
	return this.m.processMethod.OnVoid(this.m.ref("Process"))
}
`, Options{}),

		Entry("default options: method with params and returns (non-async)", `
package pkg

type Service interface {
	Send(email string) (id string, err error)
}`, `package test

import (
	"github.com/ArtHustle/mock_gen/mockrt"
)

// ServiceMock implements Service. Inspect recorded calls through
// Calls() and configure returns through On().
type ServiceMock struct {
	rt         mockrt.Mock
	sendMethod mockrt.Method
}

// NewServiceMock returns a zero-valued ServiceMock. Equivalent
// to &ServiceMock{}; provided as a consistent entry point that may
// initialize internal state in future versions.
func NewServiceMock() *ServiceMock {
	return &ServiceMock{}
}

type (
	ServiceMockCall          = mockrt.Call
	ServiceMockTimelineEntry = mockrt.TimelineEntry
	ServiceMockVerifyT       = mockrt.T
)

func (this *ServiceMock) ref(method string) mockrt.Ref {
	return mockrt.Ref{Mock: &this.rt, Name: "ServiceMock", Method: method}
}

func (this *ServiceMock) group() mockrt.Group {
	return mockrt.Group{Ref: this.ref(""), Methods: []mockrt.Named{
		{Name: "Send", Method: &this.sendMethod},
	}}
}

// Reset clears every call log, the timeline and every configured return.
func (this *ServiceMock) Reset() { this.group().Reset() }

type serviceMockCalls struct {
	mockrt.MockCalls
	m *ServiceMock
}

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{this.group().MockCalls(), this}
}

type serviceMockOn struct {
	mockrt.MockOn
	m *ServiceMock
}

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() serviceMockOn {
	return serviceMockOn{this.group().MockOn(), this}
}

type ServiceMockSendParam struct {
	Email string
}

type ServiceMockSendReturn struct {
	Id  string
	Err error
}

// ServiceMockSendCall records one call to Service.Send.
type ServiceMockSendCall struct {
	Index       int
	GlobalIndex int
	Param       ServiceMockSendParam
	Return      ServiceMockSendReturn
}

func (this ServiceMockSendCall) Global() int { return this.GlobalIndex }
func (this ServiceMockSendCall) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

func serviceMockSendCallOf(r mockrt.Rec) ServiceMockSendCall {
	ret, _ := r.Return.(ServiceMockSendReturn)
	return ServiceMockSendCall{Index: r.Index, GlobalIndex: r.GlobalIndex, Param: r.Param.(ServiceMockSendParam), Return: ret}
}

func (this *ServiceMock) Send(email string) (id string, err error) {
	__mockR, _ := this.sendMethod.Call(this.ref("Send"), ServiceMockSendParam{Email: email}).(ServiceMockSendReturn)
	return __mockR.Id, __mockR.Err
}

func (this serviceMockCalls) Send() serviceMockSendCalls {
	return serviceMockSendCalls{this.m.sendMethod.Calls(this.m.ref("Send"))}
}

type serviceMockSendCalls struct{ mockrt.Calls }

func (this serviceMockSendCalls) At(i int) ServiceMockSendCall {
	return serviceMockSendCallOf(this.Calls.At(i))
}

func (this serviceMockSendCalls) Last() ServiceMockSendCall { return this.At(-1) }
func (this serviceMockSendCalls) All() []ServiceMockSendCall {
	return mockrt.Map(this.Calls.All(), serviceMockSendCallOf)
}

func (this serviceMockSendCalls) Any(pred func(ServiceMockSendCall) bool) bool {
	return this.Calls.Any(mockrt.Pred(pred, serviceMockSendCallOf))
}

func (this serviceMockSendCalls) Filter(pred func(ServiceMockSendCall) bool) []ServiceMockSendCall {
	return mockrt.Map(this.Calls.Filter(mockrt.Pred(pred, serviceMockSendCallOf)), serviceMockSendCallOf)
}

func (this serviceMockOn) Send() serviceMockSendOn {
	return serviceMockSendOn{this.m.sendMethod.On(this.m.ref("Send"))}
}

type serviceMockSendOn struct{ mockrt.On }

func (this serviceMockSendOn) ReturnAll(id string, err error) {
	this.SetAll(ServiceMockSendReturn{Id: id, Err: err})
}

func (this serviceMockSendOn) ReturnCustom(value func(string) (string, error)) {
	this.Custom(value != nil, func(p any) any {
		var r ServiceMockSendReturn
		r.Id, r.Err = value(p.(ServiceMockSendParam).Email)
		return r
	})
}

func (this serviceMockSendOn) Return() serviceMockSendReturn {
	return serviceMockSendReturn{this.On.Return()}
}

func (this serviceMockSendOn) Once() serviceMockSendReturn {
	return serviceMockSendReturn{this.On.Once()}
}

func (this serviceMockSendOn) Always() serviceMockSendReturn {
	return serviceMockSendReturn{this.On.Always()}
}

func (this serviceMockSendOn) When(pred func(ServiceMockSendParam) bool) serviceMockSendOn {
	return serviceMockSendOn{this.On.When(func(p any) bool { return pred(p.(ServiceMockSendParam)) })}
}

type serviceMockSendReturn struct{ mockrt.Stub }

func (this serviceMockSendReturn) Return() serviceMockSendReturn { return this }
func (this serviceMockSendReturn) Id(value string) serviceMockSendReturn {
	return serviceMockSendReturn{mockrt.Set(this.Stub, func(r *ServiceMockSendReturn) { r.Id = value })}
}

func (this serviceMockSendReturn) Err(value error) serviceMockSendReturn {
	return serviceMockSendReturn{mockrt.Set(this.Stub, func(r *ServiceMockSendReturn) { r.Err = value })}
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
	"sync"
	"time"

	"github.com/ArtHustle/mock_gen/mockrt"
)

// ServiceMock implements Service. Inspect recorded calls through
// Calls() and configure returns through On().
type ServiceMock struct {
	mu         sync.Mutex
	rt         mockrt.Mock
	sendMethod mockrt.Method
}

// NewServiceMock returns a zero-valued ServiceMock. Equivalent
// to &ServiceMock{}; provided as a consistent entry point that may
// initialize internal state in future versions.
func NewServiceMock() *ServiceMock {
	return &ServiceMock{}
}

type (
	ServiceMockCall          = mockrt.Call
	ServiceMockTimelineEntry = mockrt.TimelineEntry
	ServiceMockVerifyT       = mockrt.T
)

func (this *ServiceMock) ref(method string) mockrt.Ref {
	return mockrt.Ref{Mock: &this.rt, Mu: &this.mu, Name: "ServiceMock", Method: method}
}

func (this *ServiceMock) group() mockrt.Group {
	return mockrt.Group{Ref: this.ref(""), Methods: []mockrt.Named{
		{Name: "Send", Method: &this.sendMethod},
	}}
}

// Reset clears every call log, the timeline and every configured return.
func (this *ServiceMock) Reset() { this.group().Reset() }

type serviceMockCalls struct {
	mockrt.MockCalls
	m *ServiceMock
}

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{this.group().MockCalls(), this}
}

type serviceMockOn struct {
	mockrt.MockOn
	m *ServiceMock
}

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() serviceMockOn {
	return serviceMockOn{this.group().MockOn(), this}
}

type ServiceMockSendParam struct {
	Ctx   context.Context
	Email string
}

type ServiceMockSendReturn struct {
	Id  string
	Err error
}

// ServiceMockSendCall records one call to Service.Send.
type ServiceMockSendCall struct {
	Index       int
	GlobalIndex int
	Param       ServiceMockSendParam
	Return      ServiceMockSendReturn
}

func (this ServiceMockSendCall) Global() int { return this.GlobalIndex }
func (this ServiceMockSendCall) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

func serviceMockSendCallOf(r mockrt.Rec) ServiceMockSendCall {
	ret, _ := r.Return.(ServiceMockSendReturn)
	return ServiceMockSendCall{Index: r.Index, GlobalIndex: r.GlobalIndex, Param: r.Param.(ServiceMockSendParam), Return: ret}
}

func (this *ServiceMock) Send(ctx context.Context, email string) (id string, err error) {
	__mockR, _ := this.sendMethod.Call(this.ref("Send"), ServiceMockSendParam{Ctx: ctx, Email: email}).(ServiceMockSendReturn)
	return __mockR.Id, __mockR.Err
}

func (this serviceMockCalls) Send() serviceMockSendCalls {
	return serviceMockSendCalls{this.m.sendMethod.AsyncCalls(this.m.ref("Send"))}
}

type serviceMockSendCalls struct{ mockrt.AsyncCalls }

func (this serviceMockSendCalls) At(i int) ServiceMockSendCall {
	return serviceMockSendCallOf(this.AsyncCalls.At(i))
}

func (this serviceMockSendCalls) Last() ServiceMockSendCall { return this.At(-1) }
func (this serviceMockSendCalls) All() []ServiceMockSendCall {
	return mockrt.Map(this.AsyncCalls.All(), serviceMockSendCallOf)
}

func (this serviceMockSendCalls) Any(pred func(ServiceMockSendCall) bool) bool {
	return this.AsyncCalls.Any(mockrt.Pred(pred, serviceMockSendCallOf))
}

func (this serviceMockSendCalls) Filter(pred func(ServiceMockSendCall) bool) []ServiceMockSendCall {
	return mockrt.Map(this.AsyncCalls.Filter(mockrt.Pred(pred, serviceMockSendCallOf)), serviceMockSendCallOf)
}

func (this serviceMockSendCalls) WaitForCall(pred func(ServiceMockSendCall) bool, timeout time.Duration) error {
	return this.AsyncCalls.WaitForCall(mockrt.Pred(pred, serviceMockSendCallOf), timeout)
}

func (this serviceMockOn) Send() serviceMockSendOn {
	return serviceMockSendOn{this.m.sendMethod.On(this.m.ref("Send"))}
}

type serviceMockSendOn struct{ mockrt.On }

func (this serviceMockSendOn) ReturnAll(id string, err error) {
	this.SetAll(ServiceMockSendReturn{Id: id, Err: err})
}

func (this serviceMockSendOn) ReturnCustom(value func(context.Context, string) (string, error)) {
	this.Custom(value != nil, func(p any) any {
		var r ServiceMockSendReturn
		r.Id, r.Err = value(p.(ServiceMockSendParam).Ctx, p.(ServiceMockSendParam).Email)
		return r
	})
}

func (this serviceMockSendOn) Return() serviceMockSendReturn {
	return serviceMockSendReturn{this.On.Return()}
}

func (this serviceMockSendOn) Once() serviceMockSendReturn {
	return serviceMockSendReturn{this.On.Once()}
}

func (this serviceMockSendOn) Always() serviceMockSendReturn {
	return serviceMockSendReturn{this.On.Always()}
}

func (this serviceMockSendOn) When(pred func(ServiceMockSendParam) bool) serviceMockSendOn {
	return serviceMockSendOn{this.On.When(func(p any) bool { return pred(p.(ServiceMockSendParam)) })}
}

type serviceMockSendReturn struct{ mockrt.Stub }

func (this serviceMockSendReturn) Return() serviceMockSendReturn { return this }
func (this serviceMockSendReturn) Id(value string) serviceMockSendReturn {
	return serviceMockSendReturn{mockrt.Set(this.Stub, func(r *ServiceMockSendReturn) { r.Id = value })}
}

func (this serviceMockSendReturn) Err(value error) serviceMockSendReturn {
	return serviceMockSendReturn{mockrt.Set(this.Stub, func(r *ServiceMockSendReturn) { r.Err = value })}
}
`, Options{Async: true}),

		Entry("with header note", `
package pkg

type Service interface {
	DoSomething()
}`, `// Code generated by MockGen. DO NOT EDIT.

package test

import (
	"github.com/ArtHustle/mock_gen/mockrt"
)

// ServiceMock implements Service. Inspect recorded calls through
// Calls() and configure returns through On().
type ServiceMock struct {
	rt                mockrt.Mock
	doSomethingMethod mockrt.Method
}

// NewServiceMock returns a zero-valued ServiceMock. Equivalent
// to &ServiceMock{}; provided as a consistent entry point that may
// initialize internal state in future versions.
func NewServiceMock() *ServiceMock {
	return &ServiceMock{}
}

type (
	ServiceMockCall          = mockrt.Call
	ServiceMockTimelineEntry = mockrt.TimelineEntry
	ServiceMockVerifyT       = mockrt.T
)

func (this *ServiceMock) ref(method string) mockrt.Ref {
	return mockrt.Ref{Mock: &this.rt, Name: "ServiceMock", Method: method}
}

func (this *ServiceMock) group() mockrt.Group {
	return mockrt.Group{Ref: this.ref(""), Methods: []mockrt.Named{
		{Name: "DoSomething", Method: &this.doSomethingMethod},
	}}
}

// Reset clears every call log, the timeline and every configured return.
func (this *ServiceMock) Reset() { this.group().Reset() }

type serviceMockCalls struct {
	mockrt.MockCalls
	m *ServiceMock
}

// Calls returns the inspection namespace — one method per
// interface method, each returning per-call records of Service.
func (this *ServiceMock) Calls() serviceMockCalls {
	return serviceMockCalls{this.group().MockCalls(), this}
}

type serviceMockOn struct {
	mockrt.MockOn
	m *ServiceMock
}

// On returns the configuration namespace — one method per
// interface method, each returning the per-method config object.
func (this *ServiceMock) On() serviceMockOn {
	return serviceMockOn{this.group().MockOn(), this}
}

type ServiceMockDoSomethingParam struct{}

type ServiceMockDoSomethingReturn struct{}

// ServiceMockDoSomethingCall records one call to Service.DoSomething.
type ServiceMockDoSomethingCall struct {
	Index       int
	GlobalIndex int
	Param       ServiceMockDoSomethingParam
	Return      ServiceMockDoSomethingReturn
}

func (this ServiceMockDoSomethingCall) Global() int { return this.GlobalIndex }
func (this ServiceMockDoSomethingCall) HappenedBefore(other ServiceMockCall) bool {
	return this.GlobalIndex < other.Global()
}

func serviceMockDoSomethingCallOf(r mockrt.Rec) ServiceMockDoSomethingCall {
	ret, _ := r.Return.(ServiceMockDoSomethingReturn)
	return ServiceMockDoSomethingCall{Index: r.Index, GlobalIndex: r.GlobalIndex, Param: r.Param.(ServiceMockDoSomethingParam), Return: ret}
}

func (this *ServiceMock) DoSomething() {
	this.doSomethingMethod.Call(this.ref("DoSomething"), ServiceMockDoSomethingParam{})
}

func (this serviceMockCalls) DoSomething() serviceMockDoSomethingCalls {
	return serviceMockDoSomethingCalls{this.m.doSomethingMethod.Calls(this.m.ref("DoSomething"))}
}

type serviceMockDoSomethingCalls struct{ mockrt.Calls }

func (this serviceMockDoSomethingCalls) At(i int) ServiceMockDoSomethingCall {
	return serviceMockDoSomethingCallOf(this.Calls.At(i))
}

func (this serviceMockDoSomethingCalls) Last() ServiceMockDoSomethingCall { return this.At(-1) }
func (this serviceMockDoSomethingCalls) All() []ServiceMockDoSomethingCall {
	return mockrt.Map(this.Calls.All(), serviceMockDoSomethingCallOf)
}

func (this serviceMockDoSomethingCalls) Any(pred func(ServiceMockDoSomethingCall) bool) bool {
	return this.Calls.Any(mockrt.Pred(pred, serviceMockDoSomethingCallOf))
}

func (this serviceMockDoSomethingCalls) Filter(pred func(ServiceMockDoSomethingCall) bool) []ServiceMockDoSomethingCall {
	return mockrt.Map(this.Calls.Filter(mockrt.Pred(pred, serviceMockDoSomethingCallOf)), serviceMockDoSomethingCallOf)
}

func (this serviceMockOn) DoSomething() mockrt.OnVoid {
	return this.m.doSomethingMethod.OnVoid(this.m.ref("DoSomething"))
}
`, Options{AddHeaderNote: true}),
	)
})
