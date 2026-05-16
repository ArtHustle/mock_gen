package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// The generator emits three mock-level method names that can collide with
// source-interface method names: the inspection entry (Calls), the config
// entry (On), and Reset. Collisions are rejected with a clear error
// directing the user to override (for Calls/On) or rename (for Reset).
//
// Behavioral verification that overrides actually work end-to-end lives in
// behavior_test.go ("InspectEntry override produces accessors with the
// overridden name"). These tests focus on the *error path* — the override
// behavior tests in behavior_test.go don't overlap.
var _ = Describe("Entry-point collisions", func() {
	DescribeTable("rejects sources that collide with the default entry methods",
		func(source string, opts Options, expectedSubstring string) {
			_, err := GenerateMockContentWithOptions(source, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedSubstring))
		},

		Entry("source method named Calls without override", `
package pkg

type Service interface {
	Calls()
}`, Options{}, "collides with the inspection entry-point"),

		Entry("source method named On without override", `
package pkg

type Service interface {
	On()
}`, Options{}, "collides with the configuration entry-point"),

		Entry("source method named Reset (always errors, no override path)", `
package pkg

type Service interface {
	Reset()
}`, Options{}, "collides with the mock-level Reset method"),

		Entry("override points the collision at the *new* entry name", `
package pkg

type Service interface {
	Calls()
	Inspect()
}`, Options{InspectEntry: "Inspect"}, "collides with the inspection entry-point"),
	)
})
