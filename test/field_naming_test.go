package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// Param/Return field naming for the recorded sub-structs. These cases
// would be tedious to assert via full literal-output entries (each one is
// ~120-300 lines), so we use targeted substring assertions instead.
//
// The literal-output shape itself is covered by shape_test.go. The
// compile-and-build assertions for these cases live in compile_check_test.go
// — these tests just verify the field-naming rules.
var _ = Describe("Param and Return field naming", func() {
	DescribeTable("recorded-struct fields",
		func(source string, expectedSubstrings []string) {
			result, err := GenerateMockContentWithOptions(source, Options{})
			Expect(err).NotTo(HaveOccurred())
			for _, s := range expectedSubstrings {
				Expect(result).To(ContainSubstring(s))
			}
		},

		Entry("named param gets title-cased exported field", `
package pkg

type Service interface {
	Send(email string)
}`, []string{
			"type ServiceMockSendParam struct {\n\tEmail string\n}",
		}),

		Entry("unnamed single param gets type-derived exported field", `
package pkg

type Service interface {
	Send(string)
}`, []string{
			"type ServiceMockSendParam struct {\n\tString string\n}",
		}),

		Entry("multiple unnamed same-typed params get numeric suffixes", `
package pkg

type Service interface {
	Send(string, string)
}`, []string{
			"type ServiceMockSendParam struct {\n\tString1 string\n\tString2 string\n}",
		}),

		Entry("leading underscore in source param name gets stripped before title-casing", `
package pkg

type Service interface {
	Send(_handler func(int) error)
}`, []string{
			"type ServiceMockSendParam struct {\n\tHandler func(int) error\n}",
		}),

		Entry("named return gets title-cased exported field", `
package pkg

type Service interface {
	Get() (id string, err error)
}`, []string{
			"type ServiceMockGetReturn struct {\n\tId  string\n\tErr error\n}",
		}),

		Entry("anonymous returns of same type get numeric suffixes", `
package pkg

type Service interface {
	Get() (int, string, int)
}`, []string{
			"type ServiceMockGetReturn struct {\n\tInt1   int\n\tString string\n\tInt2   int\n}",
		}),

		Entry("map param", `
package pkg

type Service interface {
	Index(by map[string]int)
}`, []string{
			"type ServiceMockIndexParam struct {\n\tBy map[string]int\n}",
		}),

		Entry("send-only channel param", `
package pkg

type Service interface {
	Emit(out chan<- int)
}`, []string{
			"type ServiceMockEmitParam struct {\n\tOut chan<- int\n}",
		}),

		Entry("variadic param renders as slice in record", `
package pkg

type Service interface {
	Log(values ...int)
}`, []string{
			"type ServiceMockLogParam struct {\n\tValues []int\n}",
		}),

		Entry("function-typed param", `
package pkg

type Service interface {
	Subscribe(handler func(int) error)
}`, []string{
			"type ServiceMockSubscribeParam struct {\n\tHandler func(int) error\n}",
		}),

		Entry("imported type as param", `
package pkg

import "net/http"

type Service interface {
	Do(r *http.Request)
}`, []string{
			"type ServiceMockDoParam struct {\n\tR *http.Request\n}",
		}),

		Entry("generic type instantiation", `
package pkg

type Option[T any] struct{}

type Service interface {
	Get(o Option[int])
}`, []string{
			"type ServiceMockGetParam struct {\n\tO mock_gen.Option[int]\n}",
		}),

		Entry("inline struct as param", `
package pkg

type Service interface {
	Configure(opts struct{ Host string; Port int })
}`, []string{
			"Opts struct {",
			"Host string",
			"Port int",
		}),

		Entry("empty interface{} renders as any", `
package pkg

type Service interface {
	Process(v interface{})
}`, []string{
			"type ServiceMockProcessParam struct {\n\tV any\n}",
		}),

		Entry("chan struct{}", `
package pkg

type Service interface {
	Done() <-chan struct{}
}`, []string{
			"type ServiceMockDoneReturn struct {\n\tChanStruct <-chan struct{}\n}",
		}),
	)
})
