package test

import (
	. "github.com/onsi/ginkgo/v2"
)

// Default-flag generation: no Options are set. Covers signature shapes,
// type rendering (slices, channels, pointers, generics, aliases), package
// import handling, multi-interface files, and embedded interfaces.
var _ = Describe("Default generation", func() {
	DescribeTable("emits the expected mock",
		runGenEntry,

		// --- signature shapes ---

		Entry("no params, no return", `
package pkg

type Service interface {
	Add()
}`,
			`package test

type ServiceMock struct {
	Add_calls int
}

func (this *ServiceMock) Add() {
	this.Add_calls++
}
`),
		Entry("1 param, no return", `
package pkg

type Service interface {
	Add(one int)
}`,
			`package test

type ServiceMock struct {
	Add_calls     int
	Add_param_one int
}

func (this *ServiceMock) Add(one int) {
	this.Add_calls++
	this.Add_param_one = one
}
`),
		Entry("2 params of different types, no return", `
package pkg

type Service interface {
	Add(one int, two string)
}`,
			`package test

type ServiceMock struct {
	Add_calls     int
	Add_param_one int
	Add_param_two string
}

func (this *ServiceMock) Add(one int, two string) {
	this.Add_calls++
	this.Add_param_one = one
	this.Add_param_two = two
}
`),
		Entry("2 params of same type, no return", `
package pkg

type Service interface {
	Add(one, two float32)
}`,
			`package test

type ServiceMock struct {
	Add_calls     int
	Add_param_one float32
	Add_param_two float32
}

func (this *ServiceMock) Add(one, two float32) {
	this.Add_calls++
	this.Add_param_one = one
	this.Add_param_two = two
}
`),
		Entry("1 param, 1 return", `
package pkg

type Service interface {
	Add(one int) string
}`,
			`package test

type ServiceMock struct {
	Add_calls         int
	Add_param_one     int
	Add_return_string string
}

func (this *ServiceMock) Add(one int) string {
	this.Add_calls++
	this.Add_param_one = one
	return this.Add_return_string
}
`),
		Entry("1 named return", `
package pkg

type Service interface {
	Add() (name string)
}`,
			`package test

type ServiceMock struct {
	Add_calls       int
	Add_return_name string
}

func (this *ServiceMock) Add() (name string) {
	this.Add_calls++
	return this.Add_return_name
}
`),
		Entry("2 named returns", `
package pkg

type Service interface {
	Add() (name string, e error)
}`,
			`package test

type ServiceMock struct {
	Add_calls       int
	Add_return_name string
	Add_return_e    error
}

func (this *ServiceMock) Add() (name string, e error) {
	this.Add_calls++
	return this.Add_return_name, this.Add_return_e
}
`),
		Entry("anonymous returns disambiguated by type repetition", `
package pkg

type Service interface {
	Add() (int, string, int)
}`,
			`package test

type ServiceMock struct {
	Add_calls         int
	Add_return_int1   int
	Add_return_string string
	Add_return_int2   int
}

func (this *ServiceMock) Add() (int, string, int) {
	this.Add_calls++
	return this.Add_return_int1, this.Add_return_string, this.Add_return_int2
}
`),
		Entry("anonymous param gets a generated name", `
package pkg

type Service interface {
	Add(int)
}`,
			`package test

type ServiceMock struct {
	Add_calls     int
	Add_param_int int
}

func (this *ServiceMock) Add(param int) {
	this.Add_calls++
	this.Add_param_int = param
}
`),
		Entry("2 anonymous params get sequential names", `
package pkg

type Service interface {
	Add(int, string)
}`,
			`package test

type ServiceMock struct {
	Add_calls        int
	Add_param_int    int
	Add_param_string string
}

func (this *ServiceMock) Add(param1 int, param2 string) {
	this.Add_calls++
	this.Add_param_int = param1
	this.Add_param_string = param2
}
`),

		// --- slices, pointers, channels ---

		Entry("anonymous slice param", `
package pkg

type Service interface {
	Add([]int)
}`,
			`package test

type ServiceMock struct {
	Add_calls        int
	Add_param_intArr []int
}

func (this *ServiceMock) Add(param []int) {
	this.Add_calls++
	this.Add_param_intArr = param
}
`),
		Entry("anonymous slice-of-pointers param", `
package pkg

type Service interface {
	Add([]*int)
}`,
			`package test

type ServiceMock struct {
	Add_calls        int
	Add_param_intArr []*int
}

func (this *ServiceMock) Add(param []*int) {
	this.Add_calls++
	this.Add_param_intArr = param
}
`),
		Entry("variadic param (anonymous)", `
package pkg

type Service interface {
	Add(...int)
}`,
			`package test

type ServiceMock struct {
	Add_calls        int
	Add_param_intArr []int
}

func (this *ServiceMock) Add(param ...int) {
	this.Add_calls++
	this.Add_param_intArr = param
}
`),
		Entry("variadic param (named)", `
package pkg

type Service interface {
	Add(values ...int)
}`,
			`package test

type ServiceMock struct {
	Add_calls        int
	Add_param_values []int
}

func (this *ServiceMock) Add(values ...int) {
	this.Add_calls++
	this.Add_param_values = values
}
`),
		Entry("receive-only channel return", `
package pkg

type Service interface {
	Listen() <-chan string
}`,
			`package test

type ServiceMock struct {
	Listen_calls             int
	Listen_return_chanString <-chan string
}

func (this *ServiceMock) Listen() <-chan string {
	this.Listen_calls++
	return this.Listen_return_chanString
}
`),
		Entry("send-only channel param", `
package pkg

type Service interface {
	Send(chan<- int)
}`,
			`package test

type ServiceMock struct {
	Send_calls         int
	Send_param_chanInt chan<- int
}

func (this *ServiceMock) Send(param chan<- int) {
	this.Send_calls++
	this.Send_param_chanInt = param
}
`),
		Entry("bidirectional channel return", `
package pkg

type Service interface {
	Stream() chan string
}`,
			`package test

type ServiceMock struct {
	Stream_calls             int
	Stream_return_chanString chan string
}

func (this *ServiceMock) Stream() chan string {
	this.Stream_calls++
	return this.Stream_return_chanString
}
`),
		Entry("channel of imported type", `
package pkg

import "net/http"

type Service interface {
	Listen() <-chan *http.Request
}`,
			`package test

import (
	"net/http"
)

type ServiceMock struct {
	Listen_calls              int
	Listen_return_chanRequest <-chan *http.Request
}

func (this *ServiceMock) Listen() <-chan *http.Request {
	this.Listen_calls++
	return this.Listen_return_chanRequest
}
`),

		// --- imported types ---

		Entry("slice of imported struct as param", `
package pkg

import "net/http"

type Service interface {
	Add([]*http.Request)
}`,
			`package test

import (
	"net/http"
)

type ServiceMock struct {
	Add_calls            int
	Add_param_requestArr []*http.Request
}

func (this *ServiceMock) Add(param []*http.Request) {
	this.Add_calls++
	this.Add_param_requestArr = param
}
`),
		Entry("slice of imported struct as return", `
package pkg

import "io"

type Service interface {
	Add() []*io.PipeReader
}`,
			`package test

import (
	"io"
)

type ServiceMock struct {
	Add_calls                int
	Add_return_pipeReaderArr []*io.PipeReader
}

func (this *ServiceMock) Add() []*io.PipeReader {
	this.Add_calls++
	return this.Add_return_pipeReaderArr
}
`),
		Entry("imported struct pointer as param", `
package pkg

import "io"

type Service interface {
	Add(*io.PipeReader)
}`,
			`package test

import (
	"io"
)

type ServiceMock struct {
	Add_calls            int
	Add_param_pipeReader *io.PipeReader
}

func (this *ServiceMock) Add(param *io.PipeReader) {
	this.Add_calls++
	this.Add_param_pipeReader = param
}
`),
		Entry("imported struct pointer as return", `
package pkg

import "net/http"

type Service interface {
	Add() *http.Request
}`,
			`package test

import (
	"net/http"
)

type ServiceMock struct {
	Add_calls          int
	Add_return_request *http.Request
}

func (this *ServiceMock) Add() *http.Request {
	this.Add_calls++
	return this.Add_return_request
}
`),
		Entry("slice of imported package type as return", `
package pkg

import (
	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
)

type Service interface {
	Add() []model.A
}`,
			`package test

import (
	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
)

type ServiceMock struct {
	Add_calls       int
	Add_return_aArr []model.A
}

func (this *ServiceMock) Add() []model.A {
	this.Add_calls++
	return this.Add_return_aArr
}
`),

		// --- multi-interface ---

		Entry("multiple interfaces in one file", `
package pkg

type Service interface {
	Add()
}

type Service2 interface {
	Sub()
}`,
			`package test

type ServiceMock struct {
	Add_calls int
}

func (this *ServiceMock) Add() {
	this.Add_calls++
}

type Service2Mock struct {
	Sub_calls int
}

func (this *Service2Mock) Sub() {
	this.Sub_calls++
}
`),

		// --- same-package types ---

		Entry("struct defined in the source package", `
package pkg

type MyParam struct {}
type MyReturn struct {}

type Service interface {
	Add(MyParam) MyReturn
}`,
			`package test

import (
	"mock_gen"
)

type ServiceMock struct {
	Add_calls           int
	Add_param_myParam   mock_gen.MyParam
	Add_return_myReturn mock_gen.MyReturn
}

func (this *ServiceMock) Add(param mock_gen.MyParam) mock_gen.MyReturn {
	this.Add_calls++
	this.Add_param_myParam = param
	return this.Add_return_myReturn
}
`),

		// --- aliases ---

		Entry("alias to type from another package", `
package pkg

import "github.com/ArtHustle/mock_gen/test/dummy"

type MyID = dummy.IntAlias

type Service interface {
	Add(MyID) (MyID)
}
`,
			`package test

import (
	"mock_gen"
)

type ServiceMock struct {
	Add_calls       int
	Add_param_myID  mock_gen.MyID
	Add_return_myID mock_gen.MyID
}

func (this *ServiceMock) Add(param mock_gen.MyID) mock_gen.MyID {
	this.Add_calls++
	this.Add_param_myID = param
	return this.Add_return_myID
}
`),
		Entry("alias from imported package, used as param and return", `
package pkg

import "github.com/ArtHustle/mock_gen/test/dummy"

type Service interface {
	Get(dummy.IntAlias) dummy.IntAlias
}
`,
			`package test

import (
	"github.com/ArtHustle/mock_gen/test/dummy"
)

type ServiceMock struct {
	Get_calls           int
	Get_param_intAlias  dummy.IntAlias
	Get_return_intAlias dummy.IntAlias
}

func (this *ServiceMock) Get(param dummy.IntAlias) dummy.IntAlias {
	this.Get_calls++
	this.Get_param_intAlias = param
	return this.Get_return_intAlias
}
`),
		Entry("multiple params/returns of same alias type", `
package pkg

type MyID = int

type Service interface {
	Add(MyID, MyID) (MyID, error)
}
`,
			`package test

import (
	"mock_gen"
)

type ServiceMock struct {
	Add_calls        int
	Add_param_myID1  mock_gen.MyID
	Add_param_myID2  mock_gen.MyID
	Add_return_myID  mock_gen.MyID
	Add_return_error error
}

func (this *ServiceMock) Add(param1, param2 mock_gen.MyID) (mock_gen.MyID, error) {
	this.Add_calls++
	this.Add_param_myID1 = param1
	this.Add_param_myID2 = param2
	return this.Add_return_myID, this.Add_return_error
}
`),

		// --- generics ---

		Entry("generic with multiple type params", `
package pkg

type gen[T1, T2 any] struct {}

type amount int

type Service interface {
	Add(gen[amount, int])
}`,
			`package test

import (
	"mock_gen"
)

type ServiceMock struct {
	Add_calls              int
	Add_param_genAmountInt mock_gen.gen[mock_gen.amount, int]
}

func (this *ServiceMock) Add(param mock_gen.gen[mock_gen.amount, int]) {
	this.Add_calls++
	this.Add_param_genAmountInt = param
}
`),
		Entry("generic containing a slice of generics", `
package pkg

type gen[T any] struct {}

type Service interface {
	Add(gen[[]gen[int]])
}`,
			`package test

import (
	"mock_gen"
)

type ServiceMock struct {
	Add_calls           int
	Add_param_genGenInt mock_gen.gen[[]mock_gen.gen[int]]
}

func (this *ServiceMock) Add(param mock_gen.gen[[]mock_gen.gen[int]]) {
	this.Add_calls++
	this.Add_param_genGenInt = param
}
`),
		Entry("generic containing a packaged type", `
package pkg

import "io"

type gen[T any] struct {}

type Service interface {
	Add(gen[gen[io.Closer]])
}`,
			`package test

import (
	"io"

	"mock_gen"
)

type ServiceMock struct {
	Add_calls              int
	Add_param_genGenCloser mock_gen.gen[mock_gen.gen[io.Closer]]
}

func (this *ServiceMock) Add(param mock_gen.gen[mock_gen.gen[io.Closer]]) {
	this.Add_calls++
	this.Add_param_genGenCloser = param
}
`),
		Entry("generic as param and return", `
package pkg

type Option[T any] struct{}

type Service interface {
	Get(Option[int]) Option[string]
}
`,
			`package test

import (
	"mock_gen"
)

type ServiceMock struct {
	Get_calls               int
	Get_param_optionInt     mock_gen.Option[int]
	Get_return_optionString mock_gen.Option[string]
}

func (this *ServiceMock) Get(param mock_gen.Option[int]) mock_gen.Option[string] {
	this.Get_calls++
	this.Get_param_optionInt = param
	return this.Get_return_optionString
}
`),
		Entry("imported generic type", `
package pkg

import "github.com/ArtHustle/mock_gen/test/dummy"

type Service interface {
	Add(dummy.One[int])
}`,
			`package test

import (
	"github.com/ArtHustle/mock_gen/test/dummy"
)

type ServiceMock struct {
	Add_calls        int
	Add_param_oneInt dummy.One[int]
}

func (this *ServiceMock) Add(param dummy.One[int]) {
	this.Add_calls++
	this.Add_param_oneInt = param
}
`),

		// --- embedded interfaces ---

		Entry("embedded local interfaces (chain of three)", `
package pkg

type Service interface {
	Base()
}

type GrandParentService interface {
	ParentService
	GrandParent()
}
type ParentService interface {
	Service
	Parent()
}`,
			`package test

type ServiceMock struct {
	Base_calls int
}

func (this *ServiceMock) Base() {
	this.Base_calls++
}

type GrandParentServiceMock struct {
	Base_calls int

	GrandParent_calls int

	Parent_calls int
}

func (this *GrandParentServiceMock) Base() {
	this.Base_calls++
}

func (this *GrandParentServiceMock) GrandParent() {
	this.GrandParent_calls++
}

func (this *GrandParentServiceMock) Parent() {
	this.Parent_calls++
}

type ParentServiceMock struct {
	Base_calls int

	Parent_calls int
}

func (this *ParentServiceMock) Base() {
	this.Base_calls++
}

func (this *ParentServiceMock) Parent() {
	this.Parent_calls++
}
`),
		Entry("embedded interfaces from another package", `
package pkg

import "io"

type Service interface {
	io.Closer
	io.Reader
	Add()
}`,
			`package test

type ServiceMock struct {
	Add_calls int

	Close_calls        int
	Close_return_error error

	Read_calls      int
	Read_param_p    []byte
	Read_return_n   int
	Read_return_err error
}

func (this *ServiceMock) Add() {
	this.Add_calls++
}

func (this *ServiceMock) Close() error {
	this.Close_calls++
	return this.Close_return_error
}

func (this *ServiceMock) Read(p []byte) (n int, err error) {
	this.Read_calls++
	this.Read_param_p = p
	return this.Read_return_n, this.Read_return_err
}
`),

		// --- import aliasing / collision handling ---

		Entry("imports of distinct packages with the same package name", `
package pkg

import (
	a "github.com/ArtHustle/mock_gen/test/dummy/a/model"
	b "github.com/ArtHustle/mock_gen/test/dummy/b/model"
	c "github.com/ArtHustle/mock_gen/test/dummy/c/model"
)

type Service interface {
	Add(a.A, b.B, c.C)
}`,
			`package test

import (
	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
	bModel "github.com/ArtHustle/mock_gen/test/dummy/b/model"
	cModel "github.com/ArtHustle/mock_gen/test/dummy/c/model"
)

type ServiceMock struct {
	Add_calls   int
	Add_param_a model.A
	Add_param_b bModel.B
	Add_param_c cModel.C
}

func (this *ServiceMock) Add(param1 model.A, param2 bModel.B, param3 cModel.C) {
	this.Add_calls++
	this.Add_param_a = param1
	this.Add_param_b = param2
	this.Add_param_c = param3
}
`),
		Entry("same struct name across different packages (params)", `
package pkg

import (
	a "github.com/ArtHustle/mock_gen/test/dummy/a/model"
	b "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)

type Service interface {
	Add(a.A, b.A)
}`,
			`package test

import (
	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
	bModel "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)

type ServiceMock struct {
	Add_calls    int
	Add_param_a1 model.A
	Add_param_a2 bModel.A
}

func (this *ServiceMock) Add(param1 model.A, param2 bModel.A) {
	this.Add_calls++
	this.Add_param_a1 = param1
	this.Add_param_a2 = param2
}
`),
		Entry("same struct name across different packages (returns)", `
package pkg

import (
	a "github.com/ArtHustle/mock_gen/test/dummy/a/model"
	b "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)

type Service interface {
	Add() (a.A, b.A)
}`,
			`package test

import (
	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
	bModel "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)

type ServiceMock struct {
	Add_calls     int
	Add_return_a1 model.A
	Add_return_a2 bModel.A
}

func (this *ServiceMock) Add() (model.A, bModel.A) {
	this.Add_calls++
	return this.Add_return_a1, this.Add_return_a2
}
`),
		Entry("import sorting: std then third-party then local", `
package pkg

import (
	"context"
	"time"

	"github.com/ArtHustle/mock_gen/test/dummy"
)

type MyStruct struct{}

type Service interface {
	Add(context.Context, time.Duration, dummy.Struct, MyStruct)
}
`,
			`package test

import (
	"context"
	"time"

	"github.com/ArtHustle/mock_gen/test/dummy"
	"mock_gen"
)

type ServiceMock struct {
	Add_calls          int
	Add_param_context  context.Context
	Add_param_duration time.Duration
	Add_param_struct   dummy.Struct
	Add_param_myStruct mock_gen.MyStruct
}

func (this *ServiceMock) Add(param1 context.Context, param2 time.Duration, param3 dummy.Struct, param4 mock_gen.MyStruct) {
	this.Add_calls++
	this.Add_param_context = param1
	this.Add_param_duration = param2
	this.Add_param_struct = param3
	this.Add_param_myStruct = param4
}
`),
	)
})
