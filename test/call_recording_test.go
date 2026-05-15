package test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// Call recording is the per-method log of recorded arguments. Generated
// artifacts (struct, log field, getter, indexer, Reset_history) are exercised
// by the literal-output entries in method_api_test.go and concurrency_test.go.
//
// This file covers what those literal entries cannot: behaviors observed by
// *negation*, by *ordering*, or that depend on edge-case param shapes that
// don't appear in the canonical entries.
var _ = Describe("Call recording", func() {

	// --- recorded-struct field naming for edge-case param shapes ---
	//
	// Param-name derivation is the only thing that varies meaningfully across
	// methods. These entries cover what `method_api_test.go` and
	// `concurrency_test.go` don't already exercise via their literal outputs.

	DescribeTable("recorded-struct field naming",
		func(source, expectedStruct string) {
			result, err := GenerateMockContentWithOptions(source, Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring(expectedStruct))
		},

		Entry("unnamed single param gets a type-derived field name", `
package pkg

type Service interface {
	Send(string)
}`,
			"type ServiceMock_Send_Call struct {\n\tString string\n}"),

		Entry("multiple unnamed same-typed params get indexed field names", `
package pkg

type Service interface {
	Send(string, string)
}`,
			"type ServiceMock_Send_Call struct {\n\tString1 string\n\tString2 string\n}"),

		Entry("generic instantiation in param renders qualified type, type-derived field name", `
package pkg

type Option[T any] struct{}

type Service interface {
	Get(Option[int])
}`,
			"type ServiceMock_Get_Call struct {\n\tOptionInt mock_gen.Option[int]\n}"),

		Entry("map param", `
package pkg

type Service interface {
	Index(by map[string]int)
}`,
			"type ServiceMock_Index_Call struct {\n\tBy map[string]int\n}"),

		Entry("send-only channel param", `
package pkg

type Service interface {
	Emit(out chan<- int)
}`,
			"type ServiceMock_Emit_Call struct {\n\tOut chan<- int\n}"),
	)

	// --- ordering invariants the literal entries verify implicitly, but
	// which are easy to break silently across template refactors ---

	It("appends to the call log before invoking the handler (method-API, no thread-safe)", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type Service interface {
	Send(x int) error
}`, Options{GenerateReturnHandler: true})
		Expect(err).NotTo(HaveOccurred())

		body := result[strings.Index(result, "func (this *ServiceMock) Send("):]
		body = body[:strings.Index(body, "\n}\n")]

		appendIdx := strings.Index(body, "this.send_calls_log = append(")
		handlerIdx := strings.Index(body, "if this.send_return != nil")
		Expect(appendIdx).To(BeNumerically(">", 0))
		Expect(handlerIdx).To(BeNumerically(">", 0))
		Expect(appendIdx).To(BeNumerically("<", handlerIdx),
			"append must precede handler invocation so a panicking handler still leaves the call recorded")
	})

	It("under --thread-safe, append is inside the lock and the handler runs after Unlock", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type Service interface {
	Send(x int) error
}`, Options{MakeThreadSafe: true, GenerateReturnHandler: true})
		Expect(err).NotTo(HaveOccurred())

		body := result[strings.Index(result, "func (this *ServiceMock) Send("):]
		body = body[:strings.Index(body, "\n}\n")]

		lockIdx := strings.Index(body, "this.mu.Lock()")
		appendIdx := strings.Index(body, "this.send_calls_log = append(")
		unlockIdx := strings.Index(body, "this.mu.Unlock()")
		handlerIdx := strings.Index(body, "if handler != nil")

		Expect(lockIdx).To(BeNumerically(">", 0))
		Expect(appendIdx).To(BeNumerically(">", lockIdx))
		Expect(appendIdx).To(BeNumerically("<", unlockIdx))
		Expect(unlockIdx).To(BeNumerically("<", handlerIdx))
	})

	// --- Reset_history scope: a negative assertion the literal entries can't make ---

	It("Reset_history does not touch installed handlers or static return values", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type Service interface {
	Send(x int) error
}`, Options{GenerateReturnHandler: true})
		Expect(err).NotTo(HaveOccurred())

		reset := result[strings.Index(result, "func (this *ServiceMock) Reset_history()"):]
		reset = reset[:strings.Index(reset, "\n}\n")]

		Expect(reset).NotTo(ContainSubstring("this.send_return ="),
			"Reset_history must not clear the installed handler")
		Expect(reset).NotTo(ContainSubstring("this.send_return_error ="),
			"Reset_history must not clear armed static return values")
	})

	// --- negative-index semantics on Method_call(i) ---
	//
	// The generated mock emits a `mockCallAt[T]` helper at file scope and
	// every `Method_call(i int)` delegates to it. We assert both: (a) the
	// helper is emitted exactly once per generated file and the indexer
	// delegates to it, and (b) the helper's behavior in isolation (negative
	// index = offset from end, out-of-range = panic).

	It("emits the mockCallAt helper once and delegates from Method_call", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type Service interface {
	Send(x int)
	Process()
}`, Options{})
		Expect(err).NotTo(HaveOccurred())

		Expect(strings.Count(result, "func mockCallAt[T any](log []T, i int) T")).To(Equal(1),
			"helper must be emitted exactly once per file even with multiple methods")
		Expect(result).To(ContainSubstring("return mockCallAt(this.send_calls_log, i)"))
		Expect(result).To(ContainSubstring("return mockCallAt(this.process_calls_log, i)"))
	})

	DescribeTable("mockCallAt indexing semantics",
		func(log []int, i int, want int) {
			Expect(testMockCallAt(log, i)).To(Equal(want))
		},
		Entry("i=0 returns first", []int{10, 20, 30}, 0, 10),
		Entry("i=len-1 returns last", []int{10, 20, 30}, 2, 30),
		Entry("i=-1 returns last", []int{10, 20, 30}, -1, 30),
		Entry("i=-2 returns second-to-last", []int{10, 20, 30}, -2, 20),
		Entry("i=-len returns first", []int{10, 20, 30}, -3, 10),
	)

	DescribeTable("mockCallAt panics on out-of-range",
		func(log []int, i int) {
			Expect(func() { testMockCallAt(log, i) }).To(Panic())
		},
		Entry("i past end", []int{10, 20}, 2),
		Entry("i more negative than -len", []int{10, 20}, -3),
		Entry("empty log, i=0", []int{}, 0),
		Entry("empty log, i=-1", []int{}, -1),
	)

	// --- cross-mock uniqueness: two interfaces in one source must produce
	// distinct Call struct names per (mock, method) pair, so they don't collide
	// at package scope. ---

	It("Call struct names are unique across multiple interfaces in the same file", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type A interface {
	Send(x int)
}

type B interface {
	Send(x string)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(ContainSubstring("type AMock_Send_Call struct {\n\tX int\n}"))
		Expect(result).To(ContainSubstring("type BMock_Send_Call struct {\n\tX string\n}"))
	})
})
