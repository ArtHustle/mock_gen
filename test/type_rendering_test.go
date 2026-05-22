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
		Expect(out).To(ContainSubstring("handler func([]int) error"))
	})

	It("named multi-returns drop their names in the handler type", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Fetch(id int) (name string, age int, err error)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("handler func(int) (string, int, error)"))
	})

	// --- async-gated emission ---

	It("WaitForCount and WaitForCall methods are NOT emitted without --async", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Send(x int) error
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		// Check for the method declarations specifically, not mere mentions
		// (the Call struct's doc comment references WaitForCall as a use case
		// for the Index field).
		Expect(out).NotTo(ContainSubstring(") WaitForCount("),
			"WaitForCount method must only be emitted under --async")
		Expect(out).NotTo(ContainSubstring(") WaitForCall("),
			"WaitForCall method must only be emitted under --async")
		// No `time` or `errors` imports either when not needed.
		Expect(out).NotTo(ContainSubstring(`"time"`))
		Expect(out).NotTo(ContainSubstring(`"errors"`))
	})

	It("WaitForCount and WaitForCall ARE emitted with --async, on the shared generic helper", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Send(x int) error
	Close() error
}`, Options{Async: true})
		Expect(err).NotTo(HaveOccurred())
		// With the generic helper refactor, WaitForCount/WaitForCall live on
		// serviceMockCallsHelper[CT] (per-mock name) and are promoted to each
		// per-method Calls type.
		Expect(out).To(ContainSubstring("func (this serviceMockCallsHelper[CT]) WaitForCount(n int, timeout time.Duration) error"))
		Expect(out).To(ContainSubstring("func (this serviceMockCallsHelper[CT]) WaitForCall(predicate func(CT) bool, timeout time.Duration) error"))
		// Verify each method's Calls type embeds the helper (and thus has the methods promoted).
		Expect(out).To(ContainSubstring("type serviceMockSendCalls struct {"))
		Expect(out).To(ContainSubstring("type serviceMockCloseCalls struct {"))
	})

	// --- per-method config setter emission ---

	It("a method with no returns gets only Reset on its config object (no ReturnAll/Return/ReturnCustom)", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Tick()
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).NotTo(ContainSubstring(") ReturnAll("),
			"ReturnAll must not be emitted for a method with no returns")
		Expect(out).NotTo(ContainSubstring(") Return()"),
			"Return() per-return namespace must not be emitted for a method with no returns")
		Expect(out).NotTo(ContainSubstring(") ReturnCustom("),
			"ReturnCustom must not be emitted for a method with no returns")
		// The per-return setter type itself must not be declared either.
		Expect(out).NotTo(ContainSubstring("serviceMockTickReturn struct{ m"),
			"the per-return setter type must not be declared for a method with no returns")
		// Reset still exists on the per-method config object.
		Expect(out).To(ContainSubstring("func (this serviceMockTickOn) Reset()"))
		// Once(), Always(), and PendingStubs() must not be emitted for a method with no returns.
		Expect(out).NotTo(ContainSubstring("func (this serviceMockTickOn) Once()"),
			"Once() must not be emitted for a method with no returns")
		Expect(out).NotTo(ContainSubstring("func (this serviceMockTickOn) Always()"),
			"Always() must not be emitted for a method with no returns")
		// PendingStubs is provided by the embedded mockCallsHelper (returns 0 for
		// no-return methods), not as a direct declaration on the per-method type.
		Expect(out).NotTo(ContainSubstring("func (this serviceMockTickCalls) PendingStubs()"),
			"PendingStubs() must not be a direct declaration on the per-method inspect type")
		// The stub type itself must not be declared.
		Expect(out).NotTo(ContainSubstring("serviceMockTickStub"),
			"stub struct must not be declared for a method with no returns")
	})

	It("a single-return method emits ReturnAll, Return().<Name>, and ReturnCustom (no arity-1 special case)", func() {
		out, err := GenerateMockContentWithOptions(`package pkg

type Service interface {
	Greet(name string) string
}`, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) ReturnAll(string string)"))
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) Return() serviceMockGreetReturn"))
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) ReturnCustom(value func(string) string)"))
		// Per-slot setters return the setter type (chainable), not void.
		Expect(out).To(ContainSubstring("func (this serviceMockGreetReturn) String(value string) serviceMockGreetReturn"))
		// Once() and Always() must be emitted, both returning the per-return setter.
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) Once() serviceMockGreetReturn"))
		Expect(out).To(ContainSubstring("func (this serviceMockGreetOn) Always() serviceMockGreetReturn"))
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
