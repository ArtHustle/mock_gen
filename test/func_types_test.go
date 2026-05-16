package test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/ArtHustle/mock_gen"
)

// Function-typed params and returns. Previously crashed in getFieldType
// (regression #14). The generator must render `func(...)` types verbatim in
// the method signature, the recorded-call struct, and the per-return
// fields/setters.
var _ = Describe("Function-typed params and returns", func() {
	DescribeTable("emits the expected mock",
		runGenEntry,

		Entry("named function-typed param", `
package pkg

type Service interface {
	Subscribe(handler func(int) error)
}`,
			`package test

import (
	"strconv"
)

type ServiceMock_Subscribe_Call struct {
	Handler func(int) error
}

type ServiceMock struct {
	subscribe_calls_log []ServiceMock_Subscribe_Call
}

func (this *ServiceMock) Subscribe(handler func(int) error) {
	this.subscribe_calls_log = append(this.subscribe_calls_log, ServiceMock_Subscribe_Call{Handler: handler})
}

func (this *ServiceMock) Subscribe_calls() []ServiceMock_Subscribe_Call {
	out := make([]ServiceMock_Subscribe_Call, len(this.subscribe_calls_log))
	copy(out, this.subscribe_calls_log)
	return out
}

func (this *ServiceMock) Subscribe_call(i int) ServiceMock_Subscribe_Call {
	return this.subscribe_calls_log[this.callIndex(len(this.subscribe_calls_log), i)]
}

func (this *ServiceMock) Subscribe_call_count() int {
	return len(this.subscribe_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.subscribe_calls_log = nil
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
`),

		Entry("named function-typed return", `
package pkg

type Service interface {
	Watch() (cancel func())
}`,
			`package test

import (
	"strconv"
)

type ServiceMock_Watch_Call struct {
}

type ServiceMock struct {
	watch_calls_log     []ServiceMock_Watch_Call
	watch_return_cancel func()
}

func (this *ServiceMock) Watch() (cancel func()) {
	this.watch_calls_log = append(this.watch_calls_log, ServiceMock_Watch_Call{})
	return this.watch_return_cancel
}

func (this *ServiceMock) Watch_calls() []ServiceMock_Watch_Call {
	out := make([]ServiceMock_Watch_Call, len(this.watch_calls_log))
	copy(out, this.watch_calls_log)
	return out
}

func (this *ServiceMock) Watch_call(i int) ServiceMock_Watch_Call {
	return this.watch_calls_log[this.callIndex(len(this.watch_calls_log), i)]
}

func (this *ServiceMock) Watch_call_count() int {
	return len(this.watch_calls_log)
}

func (this *ServiceMock) Watch_return_cancel(value func()) {
	this.watch_return_cancel = value
}

func (this *ServiceMock) Reset_history() {
	this.watch_calls_log = nil
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
`),

		Entry("function param with imported type and named multi-return mixing func", `
package pkg

import "io"

type Service interface {
	Subscribe(topics []string, handler func(io.Reader) error) (cancel func(), err error)
}`,
			`package test

import (
	"io"
	"strconv"
)

type ServiceMock_Subscribe_Call struct {
	Topics  []string
	Handler func(io.Reader) error
}

type ServiceMock struct {
	subscribe_calls_log     []ServiceMock_Subscribe_Call
	subscribe_return_cancel func()
	subscribe_return_err    error
}

func (this *ServiceMock) Subscribe(topics []string, handler func(io.Reader) error) (cancel func(), err error) {
	this.subscribe_calls_log = append(this.subscribe_calls_log, ServiceMock_Subscribe_Call{Topics: topics, Handler: handler})
	return this.subscribe_return_cancel, this.subscribe_return_err
}

func (this *ServiceMock) Subscribe_calls() []ServiceMock_Subscribe_Call {
	out := make([]ServiceMock_Subscribe_Call, len(this.subscribe_calls_log))
	copy(out, this.subscribe_calls_log)
	return out
}

func (this *ServiceMock) Subscribe_call(i int) ServiceMock_Subscribe_Call {
	return this.subscribe_calls_log[this.callIndex(len(this.subscribe_calls_log), i)]
}

func (this *ServiceMock) Subscribe_call_count() int {
	return len(this.subscribe_calls_log)
}

func (this *ServiceMock) Subscribe_return_cancel(value func()) {
	this.subscribe_return_cancel = value
}

func (this *ServiceMock) Subscribe_return_err(value error) {
	this.subscribe_return_err = value
}

func (this *ServiceMock) Reset_history() {
	this.subscribe_calls_log = nil
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
`),

		// Regression for #17: user params named `_handler` or `_r0` (the
		// snapshot-pattern locals we used in v0.7.x/v0.8.0) shadowed the
		// captured handler/returns. Snapshot locals are now `__mockHandler`
		// and `__mockR0`, `__mockR1`, ... to make collision impossible.
		Entry("user params named _handler and _r0 do not collide with snapshot locals", `
package pkg

type Service interface {
	Send(_handler func(int) error, _r0 string) (int, error)
}`,
			`package test

import (
	"strconv"
	"sync"
)

type ServiceMock_Send_Call struct {
	Handler func(int) error
	R0      string
}

type ServiceMock struct {
	mu sync.Mutex

	send_calls_log    []ServiceMock_Send_Call
	send_return       func(func(int) error, string) (int, error)
	send_return_int   int
	send_return_error error
}

func (this *ServiceMock) Send(_handler func(int) error, _r0 string) (int, error) {
	this.mu.Lock()
	this.send_calls_log = append(this.send_calls_log, ServiceMock_Send_Call{Handler: _handler, R0: _r0})
	__mockHandler := this.send_return
	__mockR0 := this.send_return_int
	__mockR1 := this.send_return_error
	this.mu.Unlock()

	if __mockHandler != nil {
		return __mockHandler(_handler, _r0)
	}
	return __mockR0, __mockR1
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

func (this *ServiceMock) Send_return(value func(func(int) error, string) (int, error)) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.send_return = value
}

func (this *ServiceMock) Send_return_int(value int) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.send_return_int = value
}

func (this *ServiceMock) Send_return_error(value error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.send_return_error = value
}

func (this *ServiceMock) Reset_history() {
	this.mu.Lock()
	defer this.mu.Unlock()
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

		// Regression: when a method has a function-typed param named "handler",
		// the snapshot pattern's `handler := this.X_return` collided with the
		// param. Snapshot locals are now `_handler`, `_r0`, ... to avoid this.
		Entry("param named handler does not collide with the snapshot local", `
package pkg

type Service interface {
	Subscribe(handler func(int) error) error
}`,
			`package test

import (
	"strconv"
	"sync"
)

type ServiceMock_Subscribe_Call struct {
	Handler func(int) error
}

type ServiceMock struct {
	mu sync.Mutex

	subscribe_calls_log    []ServiceMock_Subscribe_Call
	subscribe_return       func(func(int) error) error
	subscribe_return_error error
}

func (this *ServiceMock) Subscribe(handler func(int) error) error {
	this.mu.Lock()
	this.subscribe_calls_log = append(this.subscribe_calls_log, ServiceMock_Subscribe_Call{Handler: handler})
	__mockHandler := this.subscribe_return
	__mockR0 := this.subscribe_return_error
	this.mu.Unlock()

	if __mockHandler != nil {
		return __mockHandler(handler)
	}
	return __mockR0
}

func (this *ServiceMock) Subscribe_calls() []ServiceMock_Subscribe_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	out := make([]ServiceMock_Subscribe_Call, len(this.subscribe_calls_log))
	copy(out, this.subscribe_calls_log)
	return out
}

func (this *ServiceMock) Subscribe_call(i int) ServiceMock_Subscribe_Call {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.subscribe_calls_log[this.callIndex(len(this.subscribe_calls_log), i)]
}

func (this *ServiceMock) Subscribe_call_count() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.subscribe_calls_log)
}

func (this *ServiceMock) Subscribe_return(value func(func(int) error) error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.subscribe_return = value
}

func (this *ServiceMock) Subscribe_return_error(value error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.subscribe_return_error = value
}

func (this *ServiceMock) Reset_history() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.subscribe_calls_log = nil
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

		// Regression for #18: inline struct/interface types previously
		// rendered as the empty string, producing broken Go that crashed
		// gofmt. These entries cover the canonical inline-type cases.
		Entry("empty struct{} as channel element renders correctly", `
package pkg

type Service interface {
	Done() <-chan struct{}
}`,
			`package test

import (
	"strconv"
)

type ServiceMock_Done_Call struct {
}

type ServiceMock struct {
	done_calls_log         []ServiceMock_Done_Call
	done_return_chanStruct <-chan struct{}
}

func (this *ServiceMock) Done() <-chan struct{} {
	this.done_calls_log = append(this.done_calls_log, ServiceMock_Done_Call{})
	return this.done_return_chanStruct
}

func (this *ServiceMock) Done_calls() []ServiceMock_Done_Call {
	out := make([]ServiceMock_Done_Call, len(this.done_calls_log))
	copy(out, this.done_calls_log)
	return out
}

func (this *ServiceMock) Done_call(i int) ServiceMock_Done_Call {
	return this.done_calls_log[this.callIndex(len(this.done_calls_log), i)]
}

func (this *ServiceMock) Done_call_count() int {
	return len(this.done_calls_log)
}

func (this *ServiceMock) Done_return_chanStruct(value <-chan struct{}) {
	this.done_return_chanStruct = value
}

func (this *ServiceMock) Reset_history() {
	this.done_calls_log = nil
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
`, Options{}),

		Entry("empty interface{} as param renders as any", `
package pkg

type Service interface {
	Process(v interface{})
}`,
			`package test

import (
	"strconv"
)

type ServiceMock_Process_Call struct {
	V any
}

type ServiceMock struct {
	process_calls_log []ServiceMock_Process_Call
}

func (this *ServiceMock) Process(v any) {
	this.process_calls_log = append(this.process_calls_log, ServiceMock_Process_Call{V: v})
}

func (this *ServiceMock) Process_calls() []ServiceMock_Process_Call {
	out := make([]ServiceMock_Process_Call, len(this.process_calls_log))
	copy(out, this.process_calls_log)
	return out
}

func (this *ServiceMock) Process_call(i int) ServiceMock_Process_Call {
	return this.process_calls_log[this.callIndex(len(this.process_calls_log), i)]
}

func (this *ServiceMock) Process_call_count() int {
	return len(this.process_calls_log)
}

func (this *ServiceMock) Reset_history() {
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
`, Options{}),

		Entry("inline struct with fields as param", `
package pkg

type Service interface {
	Configure(opts struct{ Host string; Port int })
}`,
			`package test

import (
	"strconv"
)

type ServiceMock_Configure_Call struct {
	Opts struct {
		Host string
		Port int
	}
}

type ServiceMock struct {
	configure_calls_log []ServiceMock_Configure_Call
}

func (this *ServiceMock) Configure(opts struct {
	Host string
	Port int
}) {
	this.configure_calls_log = append(this.configure_calls_log, ServiceMock_Configure_Call{Opts: opts})
}

func (this *ServiceMock) Configure_calls() []ServiceMock_Configure_Call {
	out := make([]ServiceMock_Configure_Call, len(this.configure_calls_log))
	copy(out, this.configure_calls_log)
	return out
}

func (this *ServiceMock) Configure_call(i int) ServiceMock_Configure_Call {
	return this.configure_calls_log[this.callIndex(len(this.configure_calls_log), i)]
}

func (this *ServiceMock) Configure_call_count() int {
	return len(this.configure_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.configure_calls_log = nil
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
`, Options{}),

		Entry("anonymous function-typed params get type-derived field names from the rendered type", `
package pkg

type Service interface {
	Register(func(int) error, func() bool)
}`,
			`package test

import (
	"strconv"
)

type ServiceMock_Register_Call struct {
	FuncIntError func(int) error
	FuncBool     func() bool
}

type ServiceMock struct {
	register_calls_log []ServiceMock_Register_Call
}

func (this *ServiceMock) Register(param1 func(int) error, param2 func() bool) {
	this.register_calls_log = append(this.register_calls_log, ServiceMock_Register_Call{FuncIntError: param1, FuncBool: param2})
}

func (this *ServiceMock) Register_calls() []ServiceMock_Register_Call {
	out := make([]ServiceMock_Register_Call, len(this.register_calls_log))
	copy(out, this.register_calls_log)
	return out
}

func (this *ServiceMock) Register_call(i int) ServiceMock_Register_Call {
	return this.register_calls_log[this.callIndex(len(this.register_calls_log), i)]
}

func (this *ServiceMock) Register_call_count() int {
	return len(this.register_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.register_calls_log = nil
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
`),
	)
})
