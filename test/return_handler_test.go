package test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/ArtHustle/mock_gen"
)

// --generate-return-handler without other flags. Adds a per-method handler
// func field that, when non-nil, is invoked instead of returning the static
// fields. No locking, no API shape change vs default mode (handler is a
// public field).
var _ = Describe("Return handler", func() {
	DescribeTable("emits the expected mock",
		runGenEntry,

		Entry("1 param, 1 return", `
package pkg

type Service interface {
	Add(one int) string
}`,
			`package test

type ServiceMock struct {
	Add_calls         int
	Add_param_one     int
	Add_return        func(int) string
	Add_return_string string
}

func (this *ServiceMock) Add(one int) string {
	this.Add_calls++
	this.Add_param_one = one
	if this.Add_return != nil {
		return this.Add_return(one)
	}
	return this.Add_return_string
}
`, Options{GenerateReturnHandler: true}),

		Entry("no params, 2 returns", `
package pkg

type Service interface {
	Add() (string, error)
}`,
			`package test

type ServiceMock struct {
	Add_calls         int
	Add_return        func() (string, error)
	Add_return_string string
	Add_return_error  error
}

func (this *ServiceMock) Add() (string, error) {
	this.Add_calls++
	if this.Add_return != nil {
		return this.Add_return()
	}
	return this.Add_return_string, this.Add_return_error
}
`, Options{GenerateReturnHandler: true}),

		Entry("multiple params and returns", `
package pkg

type Service interface {
	Add(a int, b string) (int, error)
}`,
			`package test

type ServiceMock struct {
	Add_calls        int
	Add_param_a      int
	Add_param_b      string
	Add_return       func(int, string) (int, error)
	Add_return_int   int
	Add_return_error error
}

func (this *ServiceMock) Add(a int, b string) (int, error) {
	this.Add_calls++
	this.Add_param_a = a
	this.Add_param_b = b
	if this.Add_return != nil {
		return this.Add_return(a, b)
	}
	return this.Add_return_int, this.Add_return_error
}
`, Options{GenerateReturnHandler: true}),
	)
})
