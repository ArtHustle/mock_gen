package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// --method-api: unexport mock fields and expose them through getter/setter
// methods. Designed so that flipping --thread-safe on later changes only
// internal locking, never the call-site API. This is the future default;
// the legacy public-field shape will eventually be deprecated.
var _ = Describe("Method API", func() {

	// --- shape tests: literal expected output ---

	DescribeTable("emits the expected mock (no thread-safe)",
		runGenEntry,

		Entry("mixed methods (with and without returns)", `
package pkg

type Service interface {
	Add(one int) string
	Process()
}`,
			`package test

type ServiceMock_Add_Call struct {
	One int
}

type ServiceMock_Process_Call struct {
}

type ServiceMock struct {
	add_calls_log     []ServiceMock_Add_Call
	add_param_one     int
	add_return_string string

	process_calls_log []ServiceMock_Process_Call
}

func (this *ServiceMock) Add(one int) string {
	this.add_param_one = one
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{One: one})
	return this.add_return_string
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[i]
}

func (this *ServiceMock) Add_param_one() int {
	return this.add_param_one
}

func (this *ServiceMock) Add_return_string(value string) {
	this.add_return_string = value
}

func (this *ServiceMock) Process() {
	this.process_calls_log = append(this.process_calls_log, ServiceMock_Process_Call{})
}

func (this *ServiceMock) Process_calls() []ServiceMock_Process_Call {
	out := make([]ServiceMock_Process_Call, len(this.process_calls_log))
	copy(out, this.process_calls_log)
	return out
}

func (this *ServiceMock) Process_call(i int) ServiceMock_Process_Call {
	return this.process_calls_log[i]
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
	this.process_calls_log = nil
}
`, Options{MethodAPI: true}),

		Entry("with return handler (handler invoked under no lock)", `
package pkg

type Service interface {
	Add(one int) string
}`,
			`package test

type ServiceMock_Add_Call struct {
	One int
}

type ServiceMock struct {
	add_calls_log     []ServiceMock_Add_Call
	add_param_one     int
	add_return        func(int) string
	add_return_string string
}

func (this *ServiceMock) Add(one int) string {
	this.add_param_one = one
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{One: one})
	if this.add_return != nil {
		return this.add_return(one)
	}
	return this.add_return_string
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[i]
}

func (this *ServiceMock) Add_param_one() int {
	return this.add_param_one
}

func (this *ServiceMock) Add_return(value func(int) string) {
	this.add_return = value
}

func (this *ServiceMock) Add_return_string(value string) {
	this.add_return_string = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
}
`, Options{MethodAPI: true, GenerateReturnHandler: true}),

		Entry("multi-return method (no handler)", `
package pkg

type Service interface {
	Add() (string, error)
}`,
			`package test

type ServiceMock_Add_Call struct {
}

type ServiceMock struct {
	add_calls_log     []ServiceMock_Add_Call
	add_return_string string
	add_return_error  error
}

func (this *ServiceMock) Add() (string, error) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{})
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

func (this *ServiceMock) Add_return_string(value string) {
	this.add_return_string = value
}

func (this *ServiceMock) Add_return_error(value error) {
	this.add_return_error = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
}
`, Options{MethodAPI: true}),

		Entry("imported types in accessor signatures", `
package pkg

import "net/http"

type Service interface {
	Send(r *http.Request) (*http.Response, error)
}`,
			`package test

import (
	"net/http"
)

type ServiceMock_Send_Call struct {
	R *http.Request
}

type ServiceMock struct {
	send_calls_log       []ServiceMock_Send_Call
	send_param_r         *http.Request
	send_return_response *http.Response
	send_return_error    error
}

func (this *ServiceMock) Send(r *http.Request) (*http.Response, error) {
	this.send_param_r = r
	this.send_calls_log = append(this.send_calls_log, ServiceMock_Send_Call{R: r})
	return this.send_return_response, this.send_return_error
}

func (this *ServiceMock) Send_calls() []ServiceMock_Send_Call {
	out := make([]ServiceMock_Send_Call, len(this.send_calls_log))
	copy(out, this.send_calls_log)
	return out
}

func (this *ServiceMock) Send_call(i int) ServiceMock_Send_Call {
	return this.send_calls_log[i]
}

func (this *ServiceMock) Send_param_r() *http.Request {
	return this.send_param_r
}

func (this *ServiceMock) Send_return_response(value *http.Response) {
	this.send_return_response = value
}

func (this *ServiceMock) Send_return_error(value error) {
	this.send_return_error = value
}

func (this *ServiceMock) Reset_history() {
	this.send_calls_log = nil
}
`, Options{MethodAPI: true}),

		Entry("embedded interface from another package", `
package pkg

import "io"

type Service interface {
	io.Closer
	Add()
}`,
			`package test

type ServiceMock_Add_Call struct {
}

type ServiceMock_Close_Call struct {
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call

	close_calls_log    []ServiceMock_Close_Call
	close_return_error error
}

func (this *ServiceMock) Add() {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[i]
}

func (this *ServiceMock) Close() error {
	this.close_calls_log = append(this.close_calls_log, ServiceMock_Close_Call{})
	return this.close_return_error
}

func (this *ServiceMock) Close_calls() []ServiceMock_Close_Call {
	out := make([]ServiceMock_Close_Call, len(this.close_calls_log))
	copy(out, this.close_calls_log)
	return out
}

func (this *ServiceMock) Close_call(i int) ServiceMock_Close_Call {
	return this.close_calls_log[i]
}

func (this *ServiceMock) Close_return_error(value error) {
	this.close_return_error = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
	this.close_calls_log = nil
}
`, Options{MethodAPI: true}),
	)

	// --- property tests ---

	// The painless-transition guarantee: enabling --thread-safe on top of
	// --method-api must not change the API surface of the mock. Test code
	// written against --method-api alone must compile and behave identically
	// when --thread-safe is added — only the internal locking changes.

	DescribeTable("flipping --thread-safe on top of --method-api preserves exported signatures",
		func(source string) {
			methodAPIOnly, err := GenerateMockContentWithOptions(source, Options{
				MethodAPI:             true,
				GenerateReturnHandler: true,
			})
			Expect(err).NotTo(HaveOccurred())

			bothFlags, err := GenerateMockContentWithOptions(source, Options{
				MethodAPI:             true,
				MakeThreadSafe:        true,
				GenerateReturnHandler: true,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(extractServiceMockMethodSignatures(methodAPIOnly)).
				To(Equal(extractServiceMockMethodSignatures(bothFlags)),
					"flipping --thread-safe on must not change exported signatures")
		},
		Entry("single param, single return", `package pkg
type Service interface { Send(ctx int, msg string) error }`),
		Entry("multi return", `package pkg
type Service interface { Get() (string, int, error) }`),
		Entry("no params no returns", `package pkg
type Service interface { Process() }`),
		Entry("variadic", `package pkg
type Service interface { Log(values ...int) error }`),
		Entry("imported types", `package pkg
import "net/http"
type Service interface { Send(r *http.Request) (*http.Response, error) }`),
		Entry("embedded interface", `package pkg
import "io"
type Service interface { io.Closer; Add() }`),
		Entry("generic type", `package pkg
type Option[T any] struct{}
type Service interface { Get(Option[int]) Option[string] }`),
		Entry("named returns", `package pkg
type Service interface { Fetch(id int) (name string, age int, err error) }`),
	)

	DescribeTable("--method-api and --thread-safe produce identical exported signatures",
		func(source string) {
			methodAPIOnly, err := GenerateMockContentWithOptions(source, Options{MethodAPI: true})
			Expect(err).NotTo(HaveOccurred())

			threadSafeOnly, err := GenerateMockContentWithOptions(source, Options{MakeThreadSafe: true})
			Expect(err).NotTo(HaveOccurred())

			Expect(extractServiceMockMethodSignatures(methodAPIOnly)).
				To(Equal(extractServiceMockMethodSignatures(threadSafeOnly)),
					"--method-api and --thread-safe must produce identical exported signatures")
		},
		Entry("single param, single return", `package pkg
type Service interface { Send(ctx int, msg string) error }`),
		Entry("multi return", `package pkg
type Service interface { Get() (string, int, error) }`),
		Entry("no params no returns", `package pkg
type Service interface { Process() }`),
		Entry("variadic", `package pkg
type Service interface { Log(values ...int) error }`),
		Entry("imported types", `package pkg
import "net/http"
type Service interface { Send(r *http.Request) (*http.Response, error) }`),
		Entry("embedded interface", `package pkg
import "io"
type Service interface { io.Closer; Add() }`),
		Entry("generic type", `package pkg
type Option[T any] struct{}
type Service interface { Get(Option[int]) Option[string] }`),
		Entry("named returns", `package pkg
type Service interface { Fetch(id int) (name string, age int, err error) }`),
	)

	It("does not import sync, declare mu, or emit Lock/Unlock without --thread-safe", func() {
		result, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Add(one int) string
}`, Options{MethodAPI: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(ContainSubstring(`"sync"`),
			"sync should not be imported without --thread-safe")
		Expect(result).NotTo(ContainSubstring("sync.Mutex"),
			"mutex field should not be added without --thread-safe")
		Expect(result).NotTo(ContainSubstring("this.mu."),
			"no Lock/Unlock calls without --thread-safe")
	})
})
