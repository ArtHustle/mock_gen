package test

import (
	"strings"
	"text/template"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/ArtHustle/mock_gen"
)

// runGenEntry is the shared body for each Entry in a DescribeTable that
// expresses a "given this interface source, the generator emits this exact
// output" test. It parses the expected output as a text/template (so entries
// can interpolate `{{.Service}}`-style placeholders, though no current entry
// does) and compares against the generator's actual output via MatchString.
//
// Entries pass Options as the optional third argument; when omitted, the
// zero-value Options is used.
func runGenEntry(input, expectedOutputTmpl string, extra ...Options) {
	data := map[string]string{"Service": "Service"}
	tmpl, e := template.New("example").Parse(expectedOutputTmpl)
	if e != nil {
		Fail("Error parsing template: " + e.Error())
	}
	var output strings.Builder
	e = tmpl.Execute(&output, data)
	if e != nil {
		Fail("Error executing template: " + e.Error())
	}

	opts := Options{}
	if len(extra) > 0 {
		opts = extra[0]
	}
	actual, err := GenerateMockContentWithOptions(input, opts)
	Expect(err).NotTo(HaveOccurred())
	Expect(actual).To(MatchString(output.String()))
}

// testMockCallAt mirrors the `mockCallAt` helper emitted into every
// generated mock. Tested in call_recording_test.go to verify negative-index
// behavior independently of compiling a generated mock.
func testMockCallAt[T any](log []T, i int) T {
	if i < 0 {
		i += len(log)
	}
	return log[i]
}

// extractServiceMockMethodSignatures returns the exported method signatures
// of `ServiceMock` from a generated source, one per line, leading
// `func (this *ServiceMock) ` stripped of its trailing `{`. Used by property
// tests that compare exported API surface across flag combinations.
func extractServiceMockMethodSignatures(src string) []string {
	var sigs []string
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, "func (this *ServiceMock) ") {
			idx := strings.Index(line, "{")
			if idx > 0 {
				sigs = append(sigs, strings.TrimSpace(line[:idx]))
			}
		}
	}
	return sigs
}
