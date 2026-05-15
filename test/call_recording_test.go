package test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// Call recording is the per-method log of recorded arguments. It is generated
// under --method-api (and therefore also under --thread-safe). Generated
// artifacts (struct, log field, getter, indexer, Reset_history) are exercised
// by the literal-output entries in method_api_test.go and concurrency_test.go.
//
// This file covers what those literal entries cannot: behaviors that are
// observed by *negation*, by *ordering*, or that depend on edge-case param
// shapes that don't appear in the canonical entries.
var _ = Describe("Call recording", func() {

	// --- recorded-struct field naming for edge-case param shapes ---
	//
	// Param-name derivation is the only thing that varies meaningfully across
	// methods. These entries cover what `method_api_test.go` and
	// `concurrency_test.go` don't already exercise via their literal outputs.

	DescribeTable("recorded-struct field naming",
		func(source, expectedStruct string) {
			result, err := GenerateMockContentWithOptions(source, Options{MethodAPI: true})
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
}`, Options{MethodAPI: true, GenerateReturnHandler: true})
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
}`, Options{MethodAPI: true, GenerateReturnHandler: true})
		Expect(err).NotTo(HaveOccurred())

		reset := result[strings.Index(result, "func (this *ServiceMock) Reset_history()"):]
		reset = reset[:strings.Index(reset, "\n}\n")]

		Expect(reset).NotTo(ContainSubstring("this.send_return ="),
			"Reset_history must not clear the installed handler")
		Expect(reset).NotTo(ContainSubstring("this.send_return_error ="),
			"Reset_history must not clear armed static return values")
	})

	// --- legacy field-API: call recording must stay off, as a hard guarantee ---

	It("legacy field-API does not emit any call-log machinery", func() {
		result, err := GenerateMockContentWithOptions(`
package pkg

type Service interface {
	Send(x int)
}`, Options{})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).NotTo(ContainSubstring("_Call"))
		Expect(result).NotTo(ContainSubstring("_calls_log"))
		Expect(result).NotTo(ContainSubstring("Reset_history"))
		Expect(result).NotTo(ContainSubstring("Send_call("))
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
}`, Options{MethodAPI: true})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(ContainSubstring("type AMock_Send_Call struct {\n\tX int\n}"))
		Expect(result).To(ContainSubstring("type BMock_Send_Call struct {\n\tX string\n}"))
	})
})
