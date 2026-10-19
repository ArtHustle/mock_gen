package test

import (
	"os"
	"path/filepath"
	"strconv"
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

// callIndexTester mirrors the generated `<MockName>.callIndex` method.
// Tested in call_recording_test.go to verify negative-index resolution
// and out-of-range panics without compiling a generated mock.
type callIndexTester struct{}

func (callIndexTester) callIndex(length, i int) int {
	resolved := i
	if resolved < 0 {
		resolved += length
	}
	if resolved < 0 || resolved >= length {
		if i < 0 {
			panic("Mock call index " + strconv.Itoa(i) + " (resolved to " + strconv.Itoa(resolved) + ") out of range, log has " + strconv.Itoa(length) + " entries")
		}
		panic("Mock call index " + strconv.Itoa(i) + " out of range, log has " + strconv.Itoa(length) + " entries")
	}
	return resolved
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

// writeBuildableGoMod writes a go.mod (and go.sum) to dir for a throwaway
// module whose generated mocks are compiled. Generated mocks import mockrt,
// so the module requires mock_gen and replaces it with this checkout; the go
// directive and go.sum are taken from this checkout so the replaced module
// resolves offline.
func writeBuildableGoMod(dir, module string) {
	GinkgoHelper()
	root, err := filepath.Abs("..")
	Expect(err).NotTo(HaveOccurred())
	rootMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	Expect(err).NotTo(HaveOccurred())
	goLine := ""
	for _, line := range strings.Split(string(rootMod), "\n") {
		if strings.HasPrefix(line, "go ") {
			goLine = line
		}
	}
	mod := "module " + module + "\n\n" + goLine + "\n\n" +
		"require github.com/ArtHustle/mock_gen v0.0.0\n\n" +
		"replace github.com/ArtHustle/mock_gen => " + root + "\n"
	Expect(os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0644)).To(Succeed())
	sum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	Expect(err).NotTo(HaveOccurred())
	Expect(os.WriteFile(filepath.Join(dir, "go.sum"), sum, 0644)).To(Succeed())
}
