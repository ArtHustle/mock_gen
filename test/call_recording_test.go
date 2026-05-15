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
		handlerIdx := strings.Index(body, "if _handler != nil")

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
	// Each generated mock has its own unexported `callIndex(length, i int) int`
	// method that resolves negative indices and panics on out-of-range. We
	// assert both: (a) the mock has its own callIndex and every Method_call
	// delegates to it (so multiple mocks in the same package don't collide),
	// and (b) the resolution behavior — covered by `callIndexTester` in
	// helpers_test.go, which mirrors the generated method.

	It("emits a per-mock callIndex method that every Method_call delegates to", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type Service interface {
	Send(x int)
	Process()
}`, Options{})
		Expect(err).NotTo(HaveOccurred())

		Expect(strings.Count(result, "func (this *ServiceMock) callIndex(length, i int) int")).To(Equal(1),
			"callIndex must be emitted once per mock, regardless of method count")
		Expect(result).To(ContainSubstring("return this.send_calls_log[this.callIndex(len(this.send_calls_log), i)]"))
		Expect(result).To(ContainSubstring("return this.process_calls_log[this.callIndex(len(this.process_calls_log), i)]"))
	})

	It("emits a distinct callIndex per mock when multiple mocks share a file", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type A interface { DoA() }
type B interface { DoB() }`, Options{})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(ContainSubstring("func (this *AMock) callIndex(length, i int) int"))
		Expect(result).To(ContainSubstring("func (this *BMock) callIndex(length, i int) int"))
	})

	DescribeTable("callIndex resolves valid indices",
		func(length, i, want int) {
			Expect(callIndexTester{}.callIndex(length, i)).To(Equal(want))
		},
		Entry("i=0 returns 0", 3, 0, 0),
		Entry("i=len-1 returns len-1", 3, 2, 2),
		Entry("i=-1 returns len-1", 3, -1, 2),
		Entry("i=-2 returns len-2", 3, -2, 1),
		Entry("i=-len returns 0", 3, -3, 0),
	)

	DescribeTable("callIndex panics on out-of-range with a descriptive message",
		func(length, i int, msgSubstring string) {
			defer func() {
				r := recover()
				Expect(r).NotTo(BeNil(), "expected panic")
				Expect(r.(string)).To(ContainSubstring(msgSubstring))
			}()
			callIndexTester{}.callIndex(length, i)
		},
		Entry("positive past end mentions input and length", 2, 5, "call index 5 out of range, log has 2 entries"),
		Entry("negative beyond -len shows resolution and length", 2, -5, "call index -5 (resolved to -3) out of range, log has 2 entries"),
		Entry("empty log with i=0 mentions zero entries", 0, 0, "call index 0 out of range, log has 0 entries"),
		Entry("empty log with i=-1 shows resolution", 0, -1, "call index -1 (resolved to -1) out of range, log has 0 entries"),
	)

	// --- Method_call_count() sugar ---

	It("emits Method_call_count() returning len of the log", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type Service interface {
	Send(x int)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(ContainSubstring("func (this *ServiceMock) Send_call_count() int"))
		Expect(result).To(ContainSubstring("return len(this.send_calls_log)"))
	})

	It("Method_call_count() takes the mutex under --thread-safe", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type Service interface {
	Send(x int)
}`, Options{MakeThreadSafe: true})
		Expect(err).NotTo(HaveOccurred())

		body := result[strings.Index(result, "func (this *ServiceMock) Send_call_count()"):]
		body = body[:strings.Index(body, "\n}\n")]
		Expect(body).To(ContainSubstring("this.mu.Lock()"))
		Expect(body).To(ContainSubstring("defer this.mu.Unlock()"))
	})

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
