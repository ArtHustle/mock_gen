package test

import (
	. "github.com/onsi/ginkgo/v2"
)

// Default-flag generation: no Options are set. Method-based shape with full
// call recording is always on. This file covers cases not already exercised
// by method_api_test.go: edge cases in param/return naming, type rendering
// (slices, channels, generics, aliases), package import handling,
// multi-interface files, and embedded interfaces.
var _ = Describe("Default generation — edge cases", func() {
	DescribeTable("emits the expected mock",
		runGenEntry,

		Entry("2 params of different types, no return", `
package pkg
type Service interface { Add(one int, two string) }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
	One int
	Two string
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(one int, two string) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{One: one, Two: two})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("2 params of same type compact in signature", `
package pkg
type Service interface { Add(one, two float32) }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
	One float32
	Two float32
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(one, two float32) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{One: one, Two: two})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("two anonymous params of different types get sequential param names", `
package pkg
type Service interface { Add(int, string) }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
	Int    int
	String string
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param1 int, param2 string) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{Int: param1, String: param2})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("single named return", `
package pkg
type Service interface { Add() (name string) }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
}

type ServiceMock struct {
	add_calls_log   []ServiceMock_Add_Call
	add_return_name string
}

func (this *ServiceMock) Add() (name string) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{})
	return this.add_return_name
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_name(value string) {
	this.add_return_name = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("two named returns", `
package pkg
type Service interface { Add() (name string, e error) }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
}

type ServiceMock struct {
	add_calls_log   []ServiceMock_Add_Call
	add_return_name string
	add_return_e    error
}

func (this *ServiceMock) Add() (name string, e error) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{})
	return this.add_return_name, this.add_return_e
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_name(value string) {
	this.add_return_name = value
}

func (this *ServiceMock) Add_return_e(value error) {
	this.add_return_e = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("anonymous returns of same type get numeric suffixes", `
package pkg
type Service interface { Add() (int, string, int) }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
}

type ServiceMock struct {
	add_calls_log     []ServiceMock_Add_Call
	add_return_int1   int
	add_return_string string
	add_return_int2   int
}

func (this *ServiceMock) Add() (int, string, int) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{})
	return this.add_return_int1, this.add_return_string, this.add_return_int2
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_int1(value int) {
	this.add_return_int1 = value
}

func (this *ServiceMock) Add_return_string(value string) {
	this.add_return_string = value
}

func (this *ServiceMock) Add_return_int2(value int) {
	this.add_return_int2 = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("anonymous slice param", `
package pkg
type Service interface { Add([]int) }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
	IntArr []int
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param []int) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{IntArr: param})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("slice of pointers as param", `
package pkg
type Service interface { Add([]*int) }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
	IntArr []*int
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param []*int) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{IntArr: param})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("anonymous variadic param", `
package pkg
type Service interface { Add(...int) }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
	IntArr []int
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param ...int) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{IntArr: param})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("receive-only channel return", `
package pkg
type Service interface { Listen() <-chan string }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Listen_Call struct {
}

type ServiceMock struct {
	listen_calls_log         []ServiceMock_Listen_Call
	listen_return_chanString <-chan string
}

func (this *ServiceMock) Listen() <-chan string {
	this.listen_calls_log = append(this.listen_calls_log, ServiceMock_Listen_Call{})
	return this.listen_return_chanString
}

func (this *ServiceMock) Listen_calls() []ServiceMock_Listen_Call {
	out := make([]ServiceMock_Listen_Call, len(this.listen_calls_log))
	copy(out, this.listen_calls_log)
	return out
}

func (this *ServiceMock) Listen_call(i int) ServiceMock_Listen_Call {
	return this.listen_calls_log[this.callIndex(len(this.listen_calls_log), i)]
}

func (this *ServiceMock) Listen_call_count() int {
	return len(this.listen_calls_log)
}

func (this *ServiceMock) Listen_return_chanString(value <-chan string) {
	this.listen_return_chanString = value
}

func (this *ServiceMock) Reset_history() {
	this.listen_calls_log = nil
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

		Entry("bidirectional channel return", `
package pkg
type Service interface { Stream() chan string }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Stream_Call struct {
}

type ServiceMock struct {
	stream_calls_log         []ServiceMock_Stream_Call
	stream_return_chanString chan string
}

func (this *ServiceMock) Stream() chan string {
	this.stream_calls_log = append(this.stream_calls_log, ServiceMock_Stream_Call{})
	return this.stream_return_chanString
}

func (this *ServiceMock) Stream_calls() []ServiceMock_Stream_Call {
	out := make([]ServiceMock_Stream_Call, len(this.stream_calls_log))
	copy(out, this.stream_calls_log)
	return out
}

func (this *ServiceMock) Stream_call(i int) ServiceMock_Stream_Call {
	return this.stream_calls_log[this.callIndex(len(this.stream_calls_log), i)]
}

func (this *ServiceMock) Stream_call_count() int {
	return len(this.stream_calls_log)
}

func (this *ServiceMock) Stream_return_chanString(value chan string) {
	this.stream_return_chanString = value
}

func (this *ServiceMock) Reset_history() {
	this.stream_calls_log = nil
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

		Entry("channel of imported pointer type", `
package pkg
import "net/http"
type Service interface { Listen() <-chan *http.Request }`,
			`package test

import (
	"net/http"
	"strconv"
)

type ServiceMock_Listen_Call struct {
}

type ServiceMock struct {
	listen_calls_log          []ServiceMock_Listen_Call
	listen_return_chanRequest <-chan *http.Request
}

func (this *ServiceMock) Listen() <-chan *http.Request {
	this.listen_calls_log = append(this.listen_calls_log, ServiceMock_Listen_Call{})
	return this.listen_return_chanRequest
}

func (this *ServiceMock) Listen_calls() []ServiceMock_Listen_Call {
	out := make([]ServiceMock_Listen_Call, len(this.listen_calls_log))
	copy(out, this.listen_calls_log)
	return out
}

func (this *ServiceMock) Listen_call(i int) ServiceMock_Listen_Call {
	return this.listen_calls_log[this.callIndex(len(this.listen_calls_log), i)]
}

func (this *ServiceMock) Listen_call_count() int {
	return len(this.listen_calls_log)
}

func (this *ServiceMock) Listen_return_chanRequest(value <-chan *http.Request) {
	this.listen_return_chanRequest = value
}

func (this *ServiceMock) Reset_history() {
	this.listen_calls_log = nil
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

		Entry("slice of imported struct pointer as param", `
package pkg
import "net/http"
type Service interface { Add([]*http.Request) }`,
			`package test

import (
	"net/http"
	"strconv"
)

type ServiceMock_Add_Call struct {
	RequestArr []*http.Request
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param []*http.Request) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{RequestArr: param})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("slice of imported struct pointer as return", `
package pkg
import "io"
type Service interface { Add() []*io.PipeReader }`,
			`package test

import (
	"io"
	"strconv"
)

type ServiceMock_Add_Call struct {
}

type ServiceMock struct {
	add_calls_log            []ServiceMock_Add_Call
	add_return_pipeReaderArr []*io.PipeReader
}

func (this *ServiceMock) Add() []*io.PipeReader {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{})
	return this.add_return_pipeReaderArr
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_pipeReaderArr(value []*io.PipeReader) {
	this.add_return_pipeReaderArr = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("slice of imported package type as return", `
package pkg
import "github.com/ArtHustle/mock_gen/test/dummy/a/model"
type Service interface { Add() []model.A }`,
			`package test

import (
	"strconv"

	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
)

type ServiceMock_Add_Call struct {
}

type ServiceMock struct {
	add_calls_log   []ServiceMock_Add_Call
	add_return_aArr []model.A
}

func (this *ServiceMock) Add() []model.A {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{})
	return this.add_return_aArr
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_aArr(value []model.A) {
	this.add_return_aArr = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("multiple interfaces in one file produce separate mocks", `
package pkg
type Service interface { Add() }
type Service2 interface { Sub() }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
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
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

type Service2Mock_Sub_Call struct {
}

type Service2Mock struct {
	sub_calls_log []Service2Mock_Sub_Call
}

func (this *Service2Mock) Sub() {
	this.sub_calls_log = append(this.sub_calls_log, Service2Mock_Sub_Call{})
}

func (this *Service2Mock) Sub_calls() []Service2Mock_Sub_Call {
	out := make([]Service2Mock_Sub_Call, len(this.sub_calls_log))
	copy(out, this.sub_calls_log)
	return out
}

func (this *Service2Mock) Sub_call(i int) Service2Mock_Sub_Call {
	return this.sub_calls_log[this.callIndex(len(this.sub_calls_log), i)]
}

func (this *Service2Mock) Sub_call_count() int {
	return len(this.sub_calls_log)
}

func (this *Service2Mock) Reset_history() {
	this.sub_calls_log = nil
}

// callIndex resolves the index argument of a Method_call(i) indexer:
// negative i counts from the end; out-of-range i panics with a message
// that surfaces both the original input and the log length.
func (this *Service2Mock) callIndex(length, i int) int {
	resolved := i
	if resolved < 0 {
		resolved += length
	}
	if resolved < 0 || resolved >= length {
		if i < 0 {
			panic("Service2Mock call index " + strconv.Itoa(i) + " (resolved to " + strconv.Itoa(resolved) + ") out of range, log has " + strconv.Itoa(length) + " entries")
		}
		panic("Service2Mock call index " + strconv.Itoa(i) + " out of range, log has " + strconv.Itoa(length) + " entries")
	}
	return resolved
}
`),

		Entry("struct defined in the source package", `
package pkg
type MyParam struct {}
type MyReturn struct {}
type Service interface { Add(MyParam) MyReturn }`,
			`package test

import (
	"strconv"

	"mock_gen"
)

type ServiceMock_Add_Call struct {
	MyParam mock_gen.MyParam
}

type ServiceMock struct {
	add_calls_log       []ServiceMock_Add_Call
	add_return_myReturn mock_gen.MyReturn
}

func (this *ServiceMock) Add(param mock_gen.MyParam) mock_gen.MyReturn {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{MyParam: param})
	return this.add_return_myReturn
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_myReturn(value mock_gen.MyReturn) {
	this.add_return_myReturn = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("local alias to imported type", `
package pkg
import "github.com/ArtHustle/mock_gen/test/dummy"
type MyID = dummy.IntAlias
type Service interface { Add(MyID) (MyID) }`,
			`package test

import (
	"strconv"

	"mock_gen"
)

type ServiceMock_Add_Call struct {
	MyID mock_gen.MyID
}

type ServiceMock struct {
	add_calls_log   []ServiceMock_Add_Call
	add_return_myID mock_gen.MyID
}

func (this *ServiceMock) Add(param mock_gen.MyID) mock_gen.MyID {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{MyID: param})
	return this.add_return_myID
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_myID(value mock_gen.MyID) {
	this.add_return_myID = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("alias used directly from imported package", `
package pkg
import "github.com/ArtHustle/mock_gen/test/dummy"
type Service interface { Get(dummy.IntAlias) dummy.IntAlias }`,
			`package test

import (
	"strconv"

	"github.com/ArtHustle/mock_gen/test/dummy"
)

type ServiceMock_Get_Call struct {
	IntAlias dummy.IntAlias
}

type ServiceMock struct {
	get_calls_log       []ServiceMock_Get_Call
	get_return_intAlias dummy.IntAlias
}

func (this *ServiceMock) Get(param dummy.IntAlias) dummy.IntAlias {
	this.get_calls_log = append(this.get_calls_log, ServiceMock_Get_Call{IntAlias: param})
	return this.get_return_intAlias
}

func (this *ServiceMock) Get_calls() []ServiceMock_Get_Call {
	out := make([]ServiceMock_Get_Call, len(this.get_calls_log))
	copy(out, this.get_calls_log)
	return out
}

func (this *ServiceMock) Get_call(i int) ServiceMock_Get_Call {
	return this.get_calls_log[this.callIndex(len(this.get_calls_log), i)]
}

func (this *ServiceMock) Get_call_count() int {
	return len(this.get_calls_log)
}

func (this *ServiceMock) Get_return_intAlias(value dummy.IntAlias) {
	this.get_return_intAlias = value
}

func (this *ServiceMock) Reset_history() {
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
`),

		Entry("repeated alias gets numeric suffix in field name", `
package pkg
type MyID = int
type Service interface { Add(MyID, MyID) (MyID, error) }`,
			`package test

import (
	"strconv"

	"mock_gen"
)

type ServiceMock_Add_Call struct {
	MyID1 mock_gen.MyID
	MyID2 mock_gen.MyID
}

type ServiceMock struct {
	add_calls_log    []ServiceMock_Add_Call
	add_return_myID  mock_gen.MyID
	add_return_error error
}

func (this *ServiceMock) Add(param1, param2 mock_gen.MyID) (mock_gen.MyID, error) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{MyID1: param1, MyID2: param2})
	return this.add_return_myID, this.add_return_error
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_myID(value mock_gen.MyID) {
	this.add_return_myID = value
}

func (this *ServiceMock) Add_return_error(value error) {
	this.add_return_error = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("generic type with two type params", `
package pkg
type gen[T1, T2 any] struct {}
type amount int
type Service interface { Add(gen[amount, int]) }`,
			`package test

import (
	"strconv"

	"mock_gen"
)

type ServiceMock_Add_Call struct {
	GenAmountInt mock_gen.gen[mock_gen.amount, int]
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param mock_gen.gen[mock_gen.amount, int]) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{GenAmountInt: param})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("generic instantiated with a slice of generics", `
package pkg
type gen[T any] struct {}
type Service interface { Add(gen[[]gen[int]]) }`,
			`package test

import (
	"strconv"

	"mock_gen"
)

type ServiceMock_Add_Call struct {
	GenGenInt mock_gen.gen[[]mock_gen.gen[int]]
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param mock_gen.gen[[]mock_gen.gen[int]]) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{GenGenInt: param})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("nested generic instantiated with an imported type", `
package pkg
import "io"
type gen[T any] struct {}
type Service interface { Add(gen[gen[io.Closer]]) }`,
			`package test

import (
	"io"
	"strconv"

	"mock_gen"
)

type ServiceMock_Add_Call struct {
	GenGenCloser mock_gen.gen[mock_gen.gen[io.Closer]]
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param mock_gen.gen[mock_gen.gen[io.Closer]]) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{GenGenCloser: param})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("imported generic type", `
package pkg
import "github.com/ArtHustle/mock_gen/test/dummy"
type Service interface { Add(dummy.One[int]) }`,
			`package test

import (
	"strconv"

	"github.com/ArtHustle/mock_gen/test/dummy"
)

type ServiceMock_Add_Call struct {
	OneInt dummy.One[int]
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param dummy.One[int]) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{OneInt: param})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("transitively embedded local interfaces", `
package pkg
type Service interface { Base() }
type GrandParentService interface { ParentService; GrandParent() }
type ParentService interface { Service; Parent() }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Base_Call struct {
}

type ServiceMock struct {
	base_calls_log []ServiceMock_Base_Call
}

func (this *ServiceMock) Base() {
	this.base_calls_log = append(this.base_calls_log, ServiceMock_Base_Call{})
}

func (this *ServiceMock) Base_calls() []ServiceMock_Base_Call {
	out := make([]ServiceMock_Base_Call, len(this.base_calls_log))
	copy(out, this.base_calls_log)
	return out
}

func (this *ServiceMock) Base_call(i int) ServiceMock_Base_Call {
	return this.base_calls_log[this.callIndex(len(this.base_calls_log), i)]
}

func (this *ServiceMock) Base_call_count() int {
	return len(this.base_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.base_calls_log = nil
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

type GrandParentServiceMock_Base_Call struct {
}

type GrandParentServiceMock_GrandParent_Call struct {
}

type GrandParentServiceMock_Parent_Call struct {
}

type GrandParentServiceMock struct {
	base_calls_log []GrandParentServiceMock_Base_Call

	grandParent_calls_log []GrandParentServiceMock_GrandParent_Call

	parent_calls_log []GrandParentServiceMock_Parent_Call
}

func (this *GrandParentServiceMock) Base() {
	this.base_calls_log = append(this.base_calls_log, GrandParentServiceMock_Base_Call{})
}

func (this *GrandParentServiceMock) Base_calls() []GrandParentServiceMock_Base_Call {
	out := make([]GrandParentServiceMock_Base_Call, len(this.base_calls_log))
	copy(out, this.base_calls_log)
	return out
}

func (this *GrandParentServiceMock) Base_call(i int) GrandParentServiceMock_Base_Call {
	return this.base_calls_log[this.callIndex(len(this.base_calls_log), i)]
}

func (this *GrandParentServiceMock) Base_call_count() int {
	return len(this.base_calls_log)
}

func (this *GrandParentServiceMock) GrandParent() {
	this.grandParent_calls_log = append(this.grandParent_calls_log, GrandParentServiceMock_GrandParent_Call{})
}

func (this *GrandParentServiceMock) GrandParent_calls() []GrandParentServiceMock_GrandParent_Call {
	out := make([]GrandParentServiceMock_GrandParent_Call, len(this.grandParent_calls_log))
	copy(out, this.grandParent_calls_log)
	return out
}

func (this *GrandParentServiceMock) GrandParent_call(i int) GrandParentServiceMock_GrandParent_Call {
	return this.grandParent_calls_log[this.callIndex(len(this.grandParent_calls_log), i)]
}

func (this *GrandParentServiceMock) GrandParent_call_count() int {
	return len(this.grandParent_calls_log)
}

func (this *GrandParentServiceMock) Parent() {
	this.parent_calls_log = append(this.parent_calls_log, GrandParentServiceMock_Parent_Call{})
}

func (this *GrandParentServiceMock) Parent_calls() []GrandParentServiceMock_Parent_Call {
	out := make([]GrandParentServiceMock_Parent_Call, len(this.parent_calls_log))
	copy(out, this.parent_calls_log)
	return out
}

func (this *GrandParentServiceMock) Parent_call(i int) GrandParentServiceMock_Parent_Call {
	return this.parent_calls_log[this.callIndex(len(this.parent_calls_log), i)]
}

func (this *GrandParentServiceMock) Parent_call_count() int {
	return len(this.parent_calls_log)
}

func (this *GrandParentServiceMock) Reset_history() {
	this.base_calls_log = nil
	this.grandParent_calls_log = nil
	this.parent_calls_log = nil
}

// callIndex resolves the index argument of a Method_call(i) indexer:
// negative i counts from the end; out-of-range i panics with a message
// that surfaces both the original input and the log length.
func (this *GrandParentServiceMock) callIndex(length, i int) int {
	resolved := i
	if resolved < 0 {
		resolved += length
	}
	if resolved < 0 || resolved >= length {
		if i < 0 {
			panic("GrandParentServiceMock call index " + strconv.Itoa(i) + " (resolved to " + strconv.Itoa(resolved) + ") out of range, log has " + strconv.Itoa(length) + " entries")
		}
		panic("GrandParentServiceMock call index " + strconv.Itoa(i) + " out of range, log has " + strconv.Itoa(length) + " entries")
	}
	return resolved
}

type ParentServiceMock_Base_Call struct {
}

type ParentServiceMock_Parent_Call struct {
}

type ParentServiceMock struct {
	base_calls_log []ParentServiceMock_Base_Call

	parent_calls_log []ParentServiceMock_Parent_Call
}

func (this *ParentServiceMock) Base() {
	this.base_calls_log = append(this.base_calls_log, ParentServiceMock_Base_Call{})
}

func (this *ParentServiceMock) Base_calls() []ParentServiceMock_Base_Call {
	out := make([]ParentServiceMock_Base_Call, len(this.base_calls_log))
	copy(out, this.base_calls_log)
	return out
}

func (this *ParentServiceMock) Base_call(i int) ParentServiceMock_Base_Call {
	return this.base_calls_log[this.callIndex(len(this.base_calls_log), i)]
}

func (this *ParentServiceMock) Base_call_count() int {
	return len(this.base_calls_log)
}

func (this *ParentServiceMock) Parent() {
	this.parent_calls_log = append(this.parent_calls_log, ParentServiceMock_Parent_Call{})
}

func (this *ParentServiceMock) Parent_calls() []ParentServiceMock_Parent_Call {
	out := make([]ParentServiceMock_Parent_Call, len(this.parent_calls_log))
	copy(out, this.parent_calls_log)
	return out
}

func (this *ParentServiceMock) Parent_call(i int) ParentServiceMock_Parent_Call {
	return this.parent_calls_log[this.callIndex(len(this.parent_calls_log), i)]
}

func (this *ParentServiceMock) Parent_call_count() int {
	return len(this.parent_calls_log)
}

func (this *ParentServiceMock) Reset_history() {
	this.base_calls_log = nil
	this.parent_calls_log = nil
}

// callIndex resolves the index argument of a Method_call(i) indexer:
// negative i counts from the end; out-of-range i panics with a message
// that surfaces both the original input and the log length.
func (this *ParentServiceMock) callIndex(length, i int) int {
	resolved := i
	if resolved < 0 {
		resolved += length
	}
	if resolved < 0 || resolved >= length {
		if i < 0 {
			panic("ParentServiceMock call index " + strconv.Itoa(i) + " (resolved to " + strconv.Itoa(resolved) + ") out of range, log has " + strconv.Itoa(length) + " entries")
		}
		panic("ParentServiceMock call index " + strconv.Itoa(i) + " out of range, log has " + strconv.Itoa(length) + " entries")
	}
	return resolved
}
`),

		Entry("two embedded cross-package interfaces (io.Reader has params + named returns)", `
package pkg
import "io"
type Service interface { io.Closer; io.Reader; Add() }`,
			`package test

import (
	"strconv"
)

type ServiceMock_Add_Call struct {
}

type ServiceMock_Close_Call struct {
}

type ServiceMock_Read_Call struct {
	P []byte
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call

	close_calls_log    []ServiceMock_Close_Call
	close_return_error error

	read_calls_log  []ServiceMock_Read_Call
	read_return_n   int
	read_return_err error
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
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
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
	return this.close_calls_log[this.callIndex(len(this.close_calls_log), i)]
}

func (this *ServiceMock) Close_call_count() int {
	return len(this.close_calls_log)
}

func (this *ServiceMock) Close_return_error(value error) {
	this.close_return_error = value
}

func (this *ServiceMock) Read(p []byte) (n int, err error) {
	this.read_calls_log = append(this.read_calls_log, ServiceMock_Read_Call{P: p})
	return this.read_return_n, this.read_return_err
}

func (this *ServiceMock) Read_calls() []ServiceMock_Read_Call {
	out := make([]ServiceMock_Read_Call, len(this.read_calls_log))
	copy(out, this.read_calls_log)
	return out
}

func (this *ServiceMock) Read_call(i int) ServiceMock_Read_Call {
	return this.read_calls_log[this.callIndex(len(this.read_calls_log), i)]
}

func (this *ServiceMock) Read_call_count() int {
	return len(this.read_calls_log)
}

func (this *ServiceMock) Read_return_n(value int) {
	this.read_return_n = value
}

func (this *ServiceMock) Read_return_err(value error) {
	this.read_return_err = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
	this.close_calls_log = nil
	this.read_calls_log = nil
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

		Entry("imports of distinct packages with same package name get aliased", `
package pkg
import (
	a "github.com/ArtHustle/mock_gen/test/dummy/a/model"
	b "github.com/ArtHustle/mock_gen/test/dummy/b/model"
	c "github.com/ArtHustle/mock_gen/test/dummy/c/model"
)
type Service interface { Add(a.A, b.B, c.C) }`,
			`package test

import (
	"strconv"

	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
	bModel "github.com/ArtHustle/mock_gen/test/dummy/b/model"
	cModel "github.com/ArtHustle/mock_gen/test/dummy/c/model"
)

type ServiceMock_Add_Call struct {
	A model.A
	B bModel.B
	C cModel.C
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param1 model.A, param2 bModel.B, param3 cModel.C) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{A: param1, B: param2, C: param3})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("same struct name across different packages disambiguated in field names (params)", `
package pkg
import (
	a "github.com/ArtHustle/mock_gen/test/dummy/a/model"
	b "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)
type Service interface { Add(a.A, b.A) }`,
			`package test

import (
	"strconv"

	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
	bModel "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)

type ServiceMock_Add_Call struct {
	A1 model.A
	A2 bModel.A
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param1 model.A, param2 bModel.A) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{A1: param1, A2: param2})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("same struct name across different packages disambiguated in field names (returns)", `
package pkg
import (
	a "github.com/ArtHustle/mock_gen/test/dummy/a/model"
	b "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)
type Service interface { Add() (a.A, b.A) }`,
			`package test

import (
	"strconv"

	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
	bModel "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)

type ServiceMock_Add_Call struct {
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
	add_return_a1 model.A
	add_return_a2 bModel.A
}

func (this *ServiceMock) Add() (model.A, bModel.A) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{})
	return this.add_return_a1, this.add_return_a2
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Add_return_a1(value model.A) {
	this.add_return_a1 = value
}

func (this *ServiceMock) Add_return_a2(value bModel.A) {
	this.add_return_a2 = value
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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

		Entry("import grouping: stdlib then third-party then local module", `
package pkg
import (
	"context"
	"time"
	"github.com/ArtHustle/mock_gen/test/dummy"
)
type MyStruct struct{}
type Service interface { Add(context.Context, time.Duration, dummy.Struct, MyStruct) }`,
			`package test

import (
	"context"
	"strconv"
	"time"

	"github.com/ArtHustle/mock_gen/test/dummy"
	"mock_gen"
)

type ServiceMock_Add_Call struct {
	Context  context.Context
	Duration time.Duration
	Struct   dummy.Struct
	MyStruct mock_gen.MyStruct
}

type ServiceMock struct {
	add_calls_log []ServiceMock_Add_Call
}

func (this *ServiceMock) Add(param1 context.Context, param2 time.Duration, param3 dummy.Struct, param4 mock_gen.MyStruct) {
	this.add_calls_log = append(this.add_calls_log, ServiceMock_Add_Call{Context: param1, Duration: param2, Struct: param3, MyStruct: param4})
}

func (this *ServiceMock) Add_calls() []ServiceMock_Add_Call {
	out := make([]ServiceMock_Add_Call, len(this.add_calls_log))
	copy(out, this.add_calls_log)
	return out
}

func (this *ServiceMock) Add_call(i int) ServiceMock_Add_Call {
	return this.add_calls_log[this.callIndex(len(this.add_calls_log), i)]
}

func (this *ServiceMock) Add_call_count() int {
	return len(this.add_calls_log)
}

func (this *ServiceMock) Reset_history() {
	this.add_calls_log = nil
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
