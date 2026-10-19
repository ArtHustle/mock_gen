package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// Type-rendering coverage: every distinct Go type that can appear as a param
// or return on an interface method must render correctly in the generated
// mock. Coverage is by substring assertion against representative locations
// in the output (signature, recorded-struct field, mock-struct cfg field).
//
// These tests are kept separate from field_naming_test.go which focuses on
// *param/return naming derivation* — this file focuses on the rendered
// *type strings* for less-common Go types and import behaviors. No overlap:
// field_naming asserts the field's exported Name; type_rendering asserts
// the field's rendered Type (and import lines where relevant).
var _ = Describe("Type rendering", func() {

	// --- signature compaction for same-typed adjacent params ---

	It("two params of the same type render compactly in the signature", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Add(one, two float32)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("func (this *ServiceMock) Add(one, two float32)"),
			"adjacent same-typed params must render with compact `name1, name2 T` syntax")
	})

	It("two params of different types render with explicit types", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Add(one int, two string)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("func (this *ServiceMock) Add(one int, two string)"))
	})

	It("two anonymous params of different types get sequential generated names", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Add(int, string)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("func (this *ServiceMock) Add(param1 int, param2 string)"))
	})

	// --- slices, pointers, channels ---

	It("anonymous slice param", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Add([]int)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("IntArr []int"))
	})

	It("slice of pointers as param", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Add([]*int)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("IntArr []*int"))
	})

	It("anonymous variadic param", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Add(...int)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("func (this *ServiceMock) Add(param ...int)"))
		Expect(out).To(ContainSubstring("IntArr []int"))
	})

	It("receive-only channel return", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Listen() <-chan string
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("ChanString <-chan string"))
	})

	It("bidirectional channel return", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Stream() chan string
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("ChanString chan string"))
	})

	It("channel of imported pointer type", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import "net/http"

type Service interface {
	Listen() <-chan *http.Request
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("ChanRequest <-chan *http.Request"))
		Expect(out).To(ContainSubstring(`"net/http"`))
	})

	// --- imported types in slices ---

	It("slice of imported struct pointer as param brings in the import", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import "net/http"

type Service interface {
	Add([]*http.Request)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("RequestArr []*http.Request"))
		Expect(out).To(ContainSubstring(`"net/http"`))
	})

	It("slice of imported struct pointer as return brings in the import", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import "io"

type Service interface {
	Add() []*io.PipeReader
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("PipeReaderArr []*io.PipeReader"))
		Expect(out).To(ContainSubstring(`"io"`))
	})

	It("slice of imported package type as return", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import (
	"github.com/ArtHustle/mock_gen/test/dummy/a/model"
)

type Service interface {
	Add() []model.A
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("AArr []model.A"))
		Expect(out).To(ContainSubstring(`"github.com/ArtHustle/mock_gen/test/dummy/a/model"`))
	})

	// --- same-package types ---

	It("struct defined in the source package renders with the package alias", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type MyParam struct {}
type MyReturn struct {}

type Service interface {
	Add(MyParam) MyReturn
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("mock_gen.MyParam"))
		Expect(out).To(ContainSubstring("mock_gen.MyReturn"))
	})

	// --- type aliases ---

	It("local alias to imported type renders the local alias name", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import "github.com/ArtHustle/mock_gen/test/dummy"

type MyID = dummy.IntAlias

type Service interface {
	Add(MyID) MyID
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("mock_gen.MyID"))
	})

	It("alias used directly from imported package", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import "github.com/ArtHustle/mock_gen/test/dummy"

type Service interface {
	Get(dummy.IntAlias) dummy.IntAlias
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("dummy.IntAlias"))
	})

	It("repeated alias gets numeric suffix in field name", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type MyID = int

type Service interface {
	Add(MyID, MyID) (MyID, error)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("MyID1 mock_gen.MyID"))
		Expect(out).To(ContainSubstring("MyID2 mock_gen.MyID"))
	})

	// --- generics ---

	It("generic type with two type params", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Pair[T1, T2 any] struct{}
type amount int

type Service interface {
	Add(Pair[amount, int])
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("mock_gen.Pair[mock_gen.amount, int]"))
	})

	It("generic instantiated with a slice of generics", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type gen[T any] struct{}

type Service interface {
	Add(gen[[]gen[int]])
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("mock_gen.gen[[]mock_gen.gen[int]]"))
	})

	It("nested generic with imported packaged type", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import "io"

type gen[T any] struct{}

type Service interface {
	Add(gen[gen[io.Closer]])
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("mock_gen.gen[mock_gen.gen[io.Closer]]"))
		Expect(out).To(ContainSubstring(`"io"`))
	})

	It("imported generic type", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import "github.com/ArtHustle/mock_gen/test/dummy"

type Service interface {
	Add(dummy.One[int])
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("dummy.One[int]"))
	})

	// --- embedded interfaces ---

	It("transitively embedded local interfaces include all reachable methods", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

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
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("func (this *GrandParentServiceMock) Base()"))
		Expect(out).To(ContainSubstring("func (this *GrandParentServiceMock) Parent()"))
		Expect(out).To(ContainSubstring("func (this *GrandParentServiceMock) GrandParent()"))
	})

	It("two embedded cross-package interfaces (io.Closer + io.Reader)", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import "io"

type Service interface {
	io.Closer
	io.Reader
	Add()
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("func (this *ServiceMock) Close() error"))
		Expect(out).To(ContainSubstring("func (this *ServiceMock) Read(p []byte)"))
		Expect(out).To(ContainSubstring("func (this *ServiceMock) Add()"))
	})

	// --- import aliasing / collision handling ---

	It("imports of distinct packages with same package name get aliased", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import (
	a "github.com/ArtHustle/mock_gen/test/dummy/a/model"
	b "github.com/ArtHustle/mock_gen/test/dummy/b/model"
	c "github.com/ArtHustle/mock_gen/test/dummy/c/model"
)

type Service interface {
	Add(a.A, b.B, c.C)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("model.A"))
		Expect(out).To(ContainSubstring("bModel.B"))
		Expect(out).To(ContainSubstring("cModel.C"))
	})

	It("same struct name across different packages disambiguated in field names (params)", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import (
	a "github.com/ArtHustle/mock_gen/test/dummy/a/model"
	b "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)

type Service interface {
	Add(a.A, b.A)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("A1 model.A"))
		Expect(out).To(ContainSubstring("A2 bModel.A"))
	})

	It("same struct name across different packages (returns)", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import (
	a "github.com/ArtHustle/mock_gen/test/dummy/a/model"
	b "github.com/ArtHustle/mock_gen/test/dummy/b/model"
)

type Service interface {
	Add() (a.A, b.A)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("A1 model.A"))
		Expect(out).To(ContainSubstring("A2 bModel.A"))
	})

	It("import grouping: stdlib then third-party (incl. local module)", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import (
	"context"
	"time"

	"github.com/ArtHustle/mock_gen/test/dummy"
)

type MyStruct struct{}

type Service interface {
	Add(context.Context, time.Duration, dummy.Struct, MyStruct)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		ctxIdx := indexOf(out, `"context"`)
		timeIdx := indexOf(out, `"time"`)
		dummyIdx := indexOf(out, `"github.com/ArtHustle/mock_gen/test/dummy"`)
		mockGenIdx := indexOf(out, `"mock_gen"`)
		Expect(ctxIdx).To(BeNumerically(">", 0))
		Expect(timeIdx).To(BeNumerically(">", ctxIdx))
		Expect(dummyIdx).To(BeNumerically(">", timeIdx),
			"third-party imports must come after stdlib")
		Expect(mockGenIdx).To(BeNumerically(">", dummyIdx),
			"local-module imports group with third-party")
	})

	// --- function-typed and inline ---

	It("named function-typed return renders verbatim", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Watch() (cancel func())
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("Cancel func()"))
		Expect(out).To(ContainSubstring("Watch() (cancel func())"))
	})

	It("function param with imported type + named multi-return mixing func", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import "io"

type Service interface {
	Subscribe(topics []string, handler func(io.Reader) error) (cancel func(), err error)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("Handler func(io.Reader) error"))
		Expect(out).To(ContainSubstring("Cancel func()"))
		Expect(out).To(ContainSubstring("Err    error"))
	})

	It("anonymous function-typed params get type-derived field names", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Register(func(int) error, func() bool)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("FuncIntError func(int) error"))
		Expect(out).To(ContainSubstring("FuncBool     func() bool"))
	})

	// --- handler type rendering ---

	It("handler type includes imported types from the method signature", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

import (
	"context"
	"net/http"
)

type Service interface {
	Do(ctx context.Context, r *http.Request) (*http.Response, error)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("func(context.Context, *http.Request) (*http.Response, error)"))
	})

	It("variadic source param renders as []T in the handler type", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Log(values ...int) error
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("ReturnCustom(value func([]int) error)"))
	})

	It("named multi-returns drop their names in the handler type", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Fetch(id int) (name string, age int, err error)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("ReturnCustom(value func(int) (string, int, error))"))
	})

	// --- async-gated emission ---

	It("WaitForCount and WaitForCall methods are NOT emitted without --async", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Send(x int) error
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("type serviceMockSendCalls struct{ mockrt.Calls }"))
		Expect(out).NotTo(ContainSubstring("mockrt.AsyncCalls"),
			"WaitForCount must only be reachable under --async")
		Expect(out).NotTo(ContainSubstring(") WaitForCall("),
			"WaitForCall must only be emitted under --async")
		Expect(out).NotTo(ContainSubstring(`"time"`))
		Expect(out).NotTo(ContainSubstring(`"errors"`))
	})

	It("WaitForCount and WaitForCall ARE reachable with --async", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Send(x int) error
	Close() error
}`, Options{Async: true})
		Expect(err).NotTo(HaveOccurred())
		// WaitForCount is promoted from mockrt.AsyncCalls; WaitForCall is
		// typed to the method's Call struct.
		Expect(out).To(ContainSubstring("type serviceMockSendCalls struct{ mockrt.AsyncCalls }"))
		Expect(out).To(ContainSubstring("func (this serviceMockSendCalls) WaitForCall(pred func(ServiceMockSendCall) bool, timeout time.Duration) error"))
		Expect(out).To(ContainSubstring("func (this serviceMockCloseCalls) WaitForCall(pred func(ServiceMockCloseCall) bool, timeout time.Duration) error"))
	})

	// --- per-method config setter emission ---

	It("a method with no returns gets only Reset on its config object (no ReturnAll/Return/ReturnCustom)", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Tick()
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		// mockrt.OnVoid has only Reset: no ReturnAll/Return/ReturnCustom,
		// Once/Always or When — there's nothing to arm or conditionalize.
		Expect(out).To(ContainSubstring("func (this serviceMockOn) Tick() mockrt.OnVoid"))
		Expect(out).NotTo(ContainSubstring("serviceMockTickOn"),
			"no per-method config type is declared for a method with no returns")
		Expect(out).NotTo(ContainSubstring("type serviceMockTickReturn struct"),
			"the per-return setter type must not be declared for a method with no returns")
	})

	It("a single-return method emits ReturnAll, Return().<Name>, and ReturnCustom (no arity-1 special case)", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Greet(name string) string
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) ReturnAll(string string)"))
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) ReturnCustom(value func(string) string)"))
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) Return() serviceMockGreetReturn"))
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) Once() serviceMockGreetReturn"))
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) Always() serviceMockGreetReturn"))
		// When(pred) returns the same config type scoped to the conditional;
		// its Once()/Always() arm the conditional.
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) When(pred func(ServiceMockGreetParam) bool) serviceMockGreetOn"))
		// Per-slot setters return the setter type (chainable), not void.
		Expect(out).To(ContainSubstring("func (this serviceMockGreetReturn) String(value string) serviceMockGreetReturn"))
		Expect(out).To(ContainSubstring("func (this serviceMockGreetReturn) Return() serviceMockGreetReturn"))
	})

	It("emits the v0.14 timeline surface: per-mock interface, entry type, Timeline + AssertOrder on inspect, GlobalIndex on CallType", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Send(email string) (id string, err error)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		// The per-mock interface, entry and T names are aliases of the
		// runtime types; the per-method Call struct is concrete.
		Expect(out).To(ContainSubstring("ServiceMockCall          = mockrt.Call"))
		Expect(out).To(ContainSubstring("ServiceMockTimelineEntry = mockrt.TimelineEntry"))
		Expect(out).To(ContainSubstring("ServiceMockVerifyT       = mockrt.T"))
		Expect(out).To(ContainSubstring("type ServiceMockSendCall struct {\n\tIndex       int\n\tGlobalIndex int"))
		Expect(out).To(ContainSubstring("func (this ServiceMockSendCall) Global() int"))
		Expect(out).To(ContainSubstring("func (this ServiceMockSendCall) HappenedBefore(other ServiceMockCall) bool"))
		// Timeline and AssertOrder are promoted from the embedded mockrt.MockCalls.
		Expect(out).To(ContainSubstring("type serviceMockCalls struct {\n\tmockrt.MockCalls"))
	})

	It("rejects an interface method that would collide with an embedded runtime namespace", func() {
		_, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	MockCalls() int
}`, Options{})
		Expect(err).To(MatchError(ContainSubstring(`declares method "MockCalls" which collides with the runtime namespace`)))
	})

	It("rejects a return value named Stub, which would collide with the setter's embedded mockrt.Stub", func() {
		_, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Find() (stub string, err error)
}`, Options{})
		Expect(err).To(MatchError(ContainSubstring(`has a return value named "Stub"`)))
	})

	It("non-empty inline interface renders methods inline", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Visit(v interface{ Read(p []byte) (n int, err error); Close() error })
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("V interface {\n\t\tClose() error\n\t\tRead([]byte) (int, error)\n\t}"))
	})
})

// indexOf is a small helper for "find this substring or return -1".
func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
