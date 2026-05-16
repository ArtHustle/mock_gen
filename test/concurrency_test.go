package test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// --thread-safe: adds a sync.Mutex guarding the mock state. Methods that
// return values use the snapshot pattern (lock, mutate, capture
// handler+returns into locals, unlock, then invoke handler / return) so the
// lock is never held while user code runs.
var _ = Describe("Thread-safe generation", func() {

	// --- shape tests: literal expected output ---

	DescribeTable("emits the expected mock",
		runGenEntry,

		Entry("mixed methods, no handler", `
package pkg
type Service interface { Add(one int) string; Process() }`,
			`package test

import (
	"strconv"
	"sync"
)

type ServiceMock_Add_Call struct {
	One int
}

type ServiceMock_Process_Call struct {
}

type ServiceMock struct {
	mu sync.Mutex

	add_calls_log     []ServiceMock_Add_Call
	add_return_string string

	process_calls_log []ServiceMock_Process_Call
}

func (this *ServiceMock) Add(one int) string {
	this.mu.Lock()
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{One: one})
	__mockR0 := this.add_return_string
	this.mu.Unlock()

	return __mockR0
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_string(value string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.add_return_string = value
}

func (this *ServiceMock) Process() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.process_calls_log = append(this.process_calls_log, ServiceMock_Process_Call{})
}

func (this *ServiceMock) Process_calls() []ServiceMock_Process_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]ServiceMock_Process_Call, len(this.process_calls_log))
	copy(out, this.process_calls_log)
	return out
}

func (this *ServiceMock) Process_call(i int) ServiceMock_Process_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.process_calls_log[this.callIndex(len(this.process_calls_log), i)]
}

func (this *ServiceMock) Process_call_count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.process_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.add_calls_log = nil
	this.process_calls_log = nil
}

// callIndex resolves the index argument of a Method_call(i) indexer:
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

		Entry("with return handler, mixed methods", `
package pkg
type Service interface { Send(one int) error; Process() }`,
			`package test

import (
	"strconv"
	"sync"
)

type ServiceMock_Process_Call struct {
}

type ServiceMock_Send_Call struct {
	One int
}

type ServiceMock struct {
	mu sync.Mutex

	process_calls_log []ServiceMock_Process_Call

	send_calls_log    []ServiceMock_Send_Call
	send_return       func(int) error
	send_return_error error
}

func (this *ServiceMock) Process() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.process_calls_log = append(this.process_calls_log, ServiceMock_Process_Call{})
}

func (this *ServiceMock) Process_calls() []ServiceMock_Process_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]ServiceMock_Process_Call, len(this.process_calls_log))
	copy(out, this.process_calls_log)
	return out
}

func (this *ServiceMock) Process_call(i int) ServiceMock_Process_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.process_calls_log[this.callIndex(len(this.process_calls_log), i)]
}

func (this *ServiceMock) Process_call_count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.process_calls_log)
}

func (this *ServiceMock) Send(one int) error {
	this.mu.Lock()
	this.send_calls_log = append(this.send_calls_log, ServiceMock_Send_Call{One: one})
	__mockHandler := this.send_return
	__mockR0 := this.send_return_error
	this.mu.Unlock()

	if __mockHandler != nil {
		return __mockHandler(one)
	}
	return __mockR0
}

func (this *ServiceMock) Send_calls() []ServiceMock_Send_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]ServiceMock_Send_Call, len(this.send_calls_log))
	copy(out, this.send_calls_log)
	return out
}

func (this *ServiceMock) Send_call(i int) ServiceMock_Send_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.send_calls_log[this.callIndex(len(this.send_calls_log), i)]
}

func (this *ServiceMock) Send_call_count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.send_calls_log)
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

func (this *ServiceMock) Reset_history() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.process_calls_log = nil
	this.send_calls_log = nil
}

// callIndex resolves the index argument of a Method_call(i) indexer:
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
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),

		Entry("with return handler, no params", `
package pkg
type Service interface { Get() string }`,
			`package test

import (
	"strconv"
	"sync"
)

type ServiceMock_Get_Call struct {
}

type ServiceMock struct {
	mu sync.Mutex

	get_calls_log     []ServiceMock_Get_Call
	get_return        func() string
	get_return_string string
}

func (this *ServiceMock) Get() string {
	this.mu.Lock()
	this.get_calls_log = append(this.get_calls_log, ServiceMock_Get_Call{})
	__mockHandler := this.get_return
	__mockR0 := this.get_return_string
	this.mu.Unlock()

	if __mockHandler != nil {
		return __mockHandler()
	}
	return __mockR0
}

func (this *ServiceMock) Get_calls() []ServiceMock_Get_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]ServiceMock_Get_Call, len(this.get_calls_log))
	copy(out, this.get_calls_log)
	return out
}

func (this *ServiceMock) Get_call(i int) ServiceMock_Get_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.get_calls_log[this.callIndex(len(this.get_calls_log), i)]
}

func (this *ServiceMock) Get_call_count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.get_calls_log)
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

func (this *ServiceMock) Reset_history() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.get_calls_log = nil
}

// callIndex resolves the index argument of a Method_call(i) indexer:
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
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),

		Entry("with return handler, imported types in handler signature", `
package pkg
import (
	"context"
	"net/http"
)
type Service interface { Do(ctx context.Context, r *http.Request) (*http.Response, error) }`,
			`package test

import (
	"context"
	"net/http"
	"strconv"
	"sync"
)

type ServiceMock_Do_Call struct {
	Ctx context.Context
	R   *http.Request
}

type ServiceMock struct {
	mu sync.Mutex

	do_calls_log       []ServiceMock_Do_Call
	do_return          func(context.Context, *http.Request) (*http.Response, error)
	do_return_response *http.Response
	do_return_error    error
}

func (this *ServiceMock) Do(ctx context.Context, r *http.Request) (*http.Response, error) {
	this.mu.Lock()
	this.do_calls_log = append(this.do_calls_log, ServiceMock_Do_Call{Ctx: ctx, R: r})
	__mockHandler := this.do_return
	__mockR0 := this.do_return_response
	__mockR1 := this.do_return_error
	this.mu.Unlock()

	if __mockHandler != nil {
		return __mockHandler(ctx, r)
	}
	return __mockR0, __mockR1
}

func (this *ServiceMock) Do_calls() []ServiceMock_Do_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]ServiceMock_Do_Call, len(this.do_calls_log))
	copy(out, this.do_calls_log)
	return out
}

func (this *ServiceMock) Do_call(i int) ServiceMock_Do_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.do_calls_log[this.callIndex(len(this.do_calls_log), i)]
}

func (this *ServiceMock) Do_call_count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.do_calls_log)
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

func (this *ServiceMock) Reset_history() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.do_calls_log = nil
}

// callIndex resolves the index argument of a Method_call(i) indexer:
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
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),

		Entry("with return handler, variadic param renders as slice", `
package pkg
type Service interface { Log(values ...int) error }`,
			`package test

import (
	"strconv"
	"sync"
)

type ServiceMock_Log_Call struct {
	Values []int
}

type ServiceMock struct {
	mu sync.Mutex

	log_calls_log    []ServiceMock_Log_Call
	log_return       func([]int) error
	log_return_error error
}

func (this *ServiceMock) Log(values ...int) error {
	this.mu.Lock()
	this.log_calls_log = append(this.log_calls_log, ServiceMock_Log_Call{Values: values})
	__mockHandler := this.log_return
	__mockR0 := this.log_return_error
	this.mu.Unlock()

	if __mockHandler != nil {
		return __mockHandler(values)
	}
	return __mockR0
}

func (this *ServiceMock) Log_calls() []ServiceMock_Log_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]ServiceMock_Log_Call, len(this.log_calls_log))
	copy(out, this.log_calls_log)
	return out
}

func (this *ServiceMock) Log_call(i int) ServiceMock_Log_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.log_calls_log[this.callIndex(len(this.log_calls_log), i)]
}

func (this *ServiceMock) Log_call_count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.log_calls_log)
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

func (this *ServiceMock) Reset_history() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.log_calls_log = nil
}

// callIndex resolves the index argument of a Method_call(i) indexer:
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
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),

		Entry("with return handler, named multi-returns drop names in handler type", `
package pkg
type Service interface { Fetch(id int) (name string, age int, err error) }`,
			`package test

import (
	"strconv"
	"sync"
)

type ServiceMock_Fetch_Call struct {
	Id int
}

type ServiceMock struct {
	mu sync.Mutex

	fetch_calls_log   []ServiceMock_Fetch_Call
	fetch_return      func(int) (string, int, error)
	fetch_return_name string
	fetch_return_age  int
	fetch_return_err  error
}

func (this *ServiceMock) Fetch(id int) (name string, age int, err error) {
	this.mu.Lock()
	this.fetch_calls_log = append(this.fetch_calls_log, ServiceMock_Fetch_Call{Id: id})
	__mockHandler := this.fetch_return
	__mockR0 := this.fetch_return_name
	__mockR1 := this.fetch_return_age
	__mockR2 := this.fetch_return_err
	this.mu.Unlock()

	if __mockHandler != nil {
		return __mockHandler(id)
	}
	return __mockR0, __mockR1, __mockR2
}

func (this *ServiceMock) Fetch_calls() []ServiceMock_Fetch_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]ServiceMock_Fetch_Call, len(this.fetch_calls_log))
	copy(out, this.fetch_calls_log)
	return out
}

func (this *ServiceMock) Fetch_call(i int) ServiceMock_Fetch_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.fetch_calls_log[this.callIndex(len(this.fetch_calls_log), i)]
}

func (this *ServiceMock) Fetch_call_count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.fetch_calls_log)
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

func (this *ServiceMock) Reset_history() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.fetch_calls_log = nil
}

// callIndex resolves the index argument of a Method_call(i) indexer:
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
`, Options{MakeThreadSafe: true, GenerateReturnHandler: true}),
	)

	// --- property tests ---

	// Regression: a thread-safe mock with a return handler must NOT hold the
	// mutex while invoking the handler. Otherwise a handler that calls back
	// into the mock deadlocks. The fix snapshots the handler and static
	// fallbacks under the lock, releases, and only then calls user code.
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

		body := result[strings.Index(result, "func (this *ServiceMock) Do("):]
		body = body[:strings.Index(body, "\n}\n")]

		unlockIdx := strings.Index(body, "this.mu.Unlock()")
		handlerIdx := strings.Index(body, "if __mockHandler != nil")
		Expect(unlockIdx).To(BeNumerically(">", 0), "Unlock() should appear in body")
		Expect(handlerIdx).To(BeNumerically(">", 0), "handler check should appear in body")
		Expect(unlockIdx).To(BeNumerically("<", handlerIdx),
			"mutex must be released before invoking the handler (re-entrancy safety)")

		Expect(body).NotTo(ContainSubstring("defer this.mu.Unlock()"),
			"body must not defer Unlock; that holds the lock across user code")
	})
})
