package test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/ArtHustle/mock_gen"
)

// --generate-return-handler without --thread-safe. Adds a per-method handler
// func field that, when non-nil, is invoked instead of returning the static
// fields. No locking. Covers cases not exercised by method_api_test.go:
// multi-return-no-params, and multi-param-multi-return handler signatures.
var _ = Describe("Return handler", func() {
	DescribeTable("emits the expected mock",
		runGenEntry,

		Entry("no params, two returns (handler captures both)", `
package pkg

type Service interface {
	Add() (string, error)
}`,
			`package test

type ServiceMock_Add_Call struct {
}

type ServiceMock struct {
	add_calls_log     []ServiceMock_Add_Call
	add_return        func() (string, error)
	add_return_string string
	add_return_error  error
}

func (this *ServiceMock) Add() (string, error) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{})
	if this.add_return != nil {
		return this.add_return()
	}
	return this.add_return_string, this.add_return_error
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[i]
}

func (this *ServiceMock) Add_return(value func() (string, error)) {
	this.add_return = value
}

func (this *ServiceMock) Add_return_string(value string) {
	this.add_return_string = value
}

func (this *ServiceMock) Add_return_error(value error) {
	this.add_return_error = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
}
`, Options{GenerateReturnHandler: true}),

		Entry("multiple params and returns (handler signature mirrors method)", `
package pkg

type Service interface {
	Add(a int, b string) (int, error)
}`,
			`package test

type ServiceMock_Add_Call struct {
	A int
	B string
}

type ServiceMock struct {
	add_calls_log    []ServiceMock_Add_Call
	add_param_a      int
	add_param_b      string
	add_return       func(int, string) (int, error)
	add_return_int   int
	add_return_error error
}

func (this *ServiceMock) Add(a int, b string) (int, error) {
	this.add_param_a = a
	this.add_param_b = b
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{A: a, B: b})
	if this.add_return != nil {
		return this.add_return(a, b)
	}
	return this.add_return_int, this.add_return_error
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[i]
}

func (this *ServiceMock) Add_param_a() int {
	return this.add_param_a
}

func (this *ServiceMock) Add_param_b() string {
	return this.add_param_b
}

func (this *ServiceMock) Add_return(value func(int, string) (int, error)) {
	this.add_return = value
}

func (this *ServiceMock) Add_return_int(value int) {
	this.add_return_int = value
}

func (this *ServiceMock) Add_return_error(value error) {
	this.add_return_error = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
}
`, Options{GenerateReturnHandler: true}),
	)
})
