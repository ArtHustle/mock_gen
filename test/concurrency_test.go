package test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// --thread-safe: implies --method-api and additionally adds a sync.Mutex
// guarding the mock state. Methods that return values use the snapshot
// pattern (lock, mutate, capture handler+returns into locals, unlock, then
// invoke handler / return) so the lock is never held while user code runs.
var _ = Describe("Thread-safe generation", func() {

	// --- shape tests: literal expected output ---

	DescribeTable("emits the expected mock",
		runGenEntry,

		Entry("mixed methods, no handler", `
package pkg

type Service interface {
	Add(one int) string
	Process()
}`,
			`package test

import (
	"sync"
)

type ServiceMock struct {
	mu sync.Mutex

	add_calls         int
	add_param_one     int
	add_return_string string

	process_calls int
}

func (this *ServiceMock) Add(one int) string {
	this.mu.Lock()
	this.add_calls++
	this.add_param_one = one
	_r0 := this.add_return_string
	this.mu.Unlock()

	return _r0
}

func (this *ServiceMock) Add_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.add_calls
}

func (this *ServiceMock) Add_param_one() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.add_param_one
}

func (this *ServiceMock) Add_return_string(value string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.add_return_string = value
}

func (this *ServiceMock) Process() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.process_calls++
}

func (this *ServiceMock) Process_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.process_calls
}
`, Options{MakeThreadSafe: true}),

		Entry("with return handler, mixed methods", `
package pkg

type Service interface {
	Send(one int) error
	Process()
}`,
			`package test

import (
	"sync"
)

type ServiceMock struct {
	mu sync.Mutex

	process_calls int

	send_calls        int
	send_param_one    int
	send_return       func(int) error
	send_return_error error
}

func (this *ServiceMock) Process() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.process_calls++
}

func (this *ServiceMock) Process_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.process_calls
}

func (this *ServiceMock) Send(one int) error {
	this.mu.Lock()
	this.send_calls++
	this.send_param_one = one
	handler := this.send_return
	_r0 := this.send_return_error
	this.mu.Unlock()

	if handler != nil {
		return handler(one)
	}
	return _r0
}

func (this *ServiceMock) Send_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.send_calls
}

func (this *ServiceMock) Send_param_one() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.send_param_one
}

func (this *ServiceMock) Send_return(value func(int) error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.send_return = value
}

func (this *ServiceMock) Send_return_error(value error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.send_return_error = value
}
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),

		Entry("with return handler, no params", `
package pkg

type Service interface {
	Get() string
}`,
			`package test

import (
	"sync"
)

type ServiceMock struct {
	mu sync.Mutex

	get_calls         int
	get_return        func() string
	get_return_string string
}

func (this *ServiceMock) Get() string {
	this.mu.Lock()
	this.get_calls++
	handler := this.get_return
	_r0 := this.get_return_string
	this.mu.Unlock()

	if handler != nil {
		return handler()
	}
	return _r0
}

func (this *ServiceMock) Get_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.get_calls
}

func (this *ServiceMock) Get_return(value func() string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.get_return = value
}

func (this *ServiceMock) Get_return_string(value string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.get_return_string = value
}
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),

		Entry("with return handler, imported types in handler signature", `
package pkg

import (
	"context"
	"net/http"
)

type Service interface {
	Do(ctx context.Context, r *http.Request) (*http.Response, error)
}`,
			`package test

import (
	"context"
	"net/http"
	"sync"
)

type ServiceMock struct {
	mu sync.Mutex

	do_calls           int
	do_param_ctx       context.Context
	do_param_r         *http.Request
	do_return          func(context.Context, *http.Request) (*http.Response, error)
	do_return_response *http.Response
	do_return_error    error
}

func (this *ServiceMock) Do(ctx context.Context, r *http.Request) (*http.Response, error) {
	this.mu.Lock()
	this.do_calls++
	this.do_param_ctx = ctx
	this.do_param_r = r
	handler := this.do_return
	_r0 := this.do_return_response
	_r1 := this.do_return_error
	this.mu.Unlock()

	if handler != nil {
		return handler(ctx, r)
	}
	return _r0, _r1
}

func (this *ServiceMock) Do_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.do_calls
}

func (this *ServiceMock) Do_param_ctx() context.Context {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.do_param_ctx
}

func (this *ServiceMock) Do_param_r() *http.Request {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.do_param_r
}

func (this *ServiceMock) Do_return(value func(context.Context, *http.Request) (*http.Response, error)) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.do_return = value
}

func (this *ServiceMock) Do_return_response(value *http.Response) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.do_return_response = value
}

func (this *ServiceMock) Do_return_error(value error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.do_return_error = value
}
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),

		Entry("with return handler, variadic param renders as slice", `
package pkg

type Service interface {
	Log(values ...int) error
}`,
			`package test

import (
	"sync"
)

type ServiceMock struct {
	mu sync.Mutex

	log_calls        int
	log_param_values []int
	log_return       func([]int) error
	log_return_error error
}

func (this *ServiceMock) Log(values ...int) error {
	this.mu.Lock()
	this.log_calls++
	this.log_param_values = values
	handler := this.log_return
	_r0 := this.log_return_error
	this.mu.Unlock()

	if handler != nil {
		return handler(values)
	}
	return _r0
}

func (this *ServiceMock) Log_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.log_calls
}

func (this *ServiceMock) Log_param_values() []int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.log_param_values
}

func (this *ServiceMock) Log_return(value func([]int) error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.log_return = value
}

func (this *ServiceMock) Log_return_error(value error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.log_return_error = value
}
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),

		Entry("with return handler, named multi-returns drop names in handler type", `
package pkg

type Service interface {
	Fetch(id int) (name string, age int, err error)
}`,
			`package test

import (
	"sync"
)

type ServiceMock struct {
	mu sync.Mutex

	fetch_calls       int
	fetch_param_id    int
	fetch_return      func(int) (string, int, error)
	fetch_return_name string
	fetch_return_age  int
	fetch_return_err  error
}

func (this *ServiceMock) Fetch(id int) (name string, age int, err error) {
	this.mu.Lock()
	this.fetch_calls++
	this.fetch_param_id = id
	handler := this.fetch_return
	_r0 := this.fetch_return_name
	_r1 := this.fetch_return_age
	_r2 := this.fetch_return_err
	this.mu.Unlock()

	if handler != nil {
		return handler(id)
	}
	return _r0, _r1, _r2
}

func (this *ServiceMock) Fetch_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.fetch_calls
}

func (this *ServiceMock) Fetch_param_id() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.fetch_param_id
}

func (this *ServiceMock) Fetch_return(value func(int) (string, int, error)) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.fetch_return = value
}

func (this *ServiceMock) Fetch_return_name(value string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.fetch_return_name = value
}

func (this *ServiceMock) Fetch_return_age(value int) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.fetch_return_age = value
}

func (this *ServiceMock) Fetch_return_err(value error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.fetch_return_err = value
}
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),
	)

	// --- property tests ---

	// Regression: a thread-safe mock with a return handler must NOT hold the
	// mutex while invoking the handler. Otherwise a handler that calls back
	// into the mock (e.g. installs a new return value, reads call counts,
	// invokes another method) deadlocks. The fix snapshots the handler and
	// static fallbacks under the lock, releases, and only then calls user code.
	It("releases the mutex before invoking the handler (re-entrancy safety)", func() {
		source := `package pkg

type Service interface {
	Do(x int) error
}`
		result, err := GenerateMockContentWithOptions(source, Options{
			MakeThreadSafe:        true,
			GenerateReturnHandler: true,
		})
		Expect(err).NotTo(HaveOccurred())

		// Inside the body of Do(), the unlock must precede the handler invocation.
		body := result[strings.Index(result, "func (this *ServiceMock) Do("):]
		body = body[:strings.Index(body, "\n}\n")]

		unlockIdx := strings.Index(body, "this.mu.Unlock()")
		handlerIdx := strings.Index(body, "if handler != nil")
		Expect(unlockIdx).To(BeNumerically(">", 0), "Unlock() should appear in body")
		Expect(handlerIdx).To(BeNumerically(">", 0), "handler check should appear in body")
		Expect(unlockIdx).To(BeNumerically("<", handlerIdx),
			"mutex must be released before invoking the handler (re-entrancy safety)")

		// The body must not use `defer this.mu.Unlock()` — that would hold the lock
		// across the handler invocation.
		Expect(body).NotTo(ContainSubstring("defer this.mu.Unlock()"),
			"body must not defer Unlock; that holds the lock across user code")
	})
})
