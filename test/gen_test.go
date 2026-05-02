package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	. "github.com/ArtHustle/mock_gen"
	_ "github.com/ArtHustle/mock_gen/test/dummy"
)

const (
	simpleContent = `package service

type Service interface {
	Add() 
}`
)

var _ = Describe("Mock Generator:", func() {
	Context("if no arguments", func() {
		It("should exit with code 1", func() {
			command := exec.Command(mockGeneratorPath)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("if input file doesn't exist", func() {
		It("should exit with code 1", func() {
			command := exec.Command(mockGeneratorPath, "~/tmp/invalidInputFile.go")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	// Regression test: when a package has a struct field with the same name as an interface,
	// findTypeDef must return the TypeName (interface declaration), not the struct field.
	// Previously this was non-deterministic due to map iteration order, causing mock_gen to
	// silently produce an empty mock on subsequent runs.
	Context("when interface name collides with a struct field name in the same package", func() {
		It("generates correct mock on repeated runs", func() {
			tempDir, err := os.MkdirTemp(os.TempDir(), "mock_gen_collision_test_")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { os.RemoveAll(tempDir) })

			Expect(os.WriteFile(
				filepath.Join(tempDir, "go.mod"),
				[]byte("module collision_test\n\ngo 1.23\n"),
				0644,
			)).To(Succeed())

			// domain.go defines a struct with a field named "Provider"
			Expect(os.WriteFile(filepath.Join(tempDir, "domain.go"), []byte(`package service

type Config struct {
	Provider string
}
`), 0644)).To(Succeed())

			// provider.go defines the Provider interface
			providerFile := filepath.Join(tempDir, "provider.go")
			Expect(os.WriteFile(providerFile, []byte(`package service

import "context"

type Provider interface {
	Delete(ctx context.Context, id string) error
}
`), 0644)).To(Succeed())

			testDir := filepath.Join(tempDir, "test")
			Expect(os.MkdirAll(testDir, 0755)).To(Succeed())
			outputFile := filepath.Join(testDir, "provider_mock.go")

			// Run mock_gen 10 times — every run must produce a non-empty mock with the Delete method.
			// Without the TypeName filter in findTypeDef, map iteration non-determinism causes
			// findTypeDef to return the struct field "Provider string" instead of the interface
			// declaration, resulting in an empty mock. Running 10 times makes this highly likely
			// to be caught even though the failure is non-deterministic.
			for i := 1; i <= 10; i++ {
				result, err := GenerateMockFileWithOptions(providerFile, Options{})
				Expect(err).NotTo(HaveOccurred(), "run %d failed", i)
				Expect(result).To(ContainSubstring("ProviderMock"), "run %d: mock struct missing", i)
				Expect(result).To(ContainSubstring("Delete_calls"), "run %d: Delete method missing", i)
				Expect(os.WriteFile(outputFile, []byte(result), 0644)).To(Succeed())
			}
		})
	})

	Context("if input file exists", func() {
		var inputFilePath string
		var outputFilePath string

		BeforeEach(func() {
			tempDir, err := os.MkdirTemp(os.TempDir(), "mock_gen_test_")
			Expect(err).NotTo(HaveOccurred())

			goModFile := filepath.Join(tempDir, "go.mod")
			Expect(os.WriteFile(goModFile, []byte("module mock_gen_test\n\ngo 1.23\n"), 0644)).To(Succeed())

			inputFile, err := os.CreateTemp(tempDir, "service.go")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { os.RemoveAll(tempDir) })

			testDir := filepath.Join(tempDir, "test")
			Expect(os.MkdirAll(testDir, os.ModePerm)).To(Succeed())
			outputFilePath = filepath.Join(testDir, "service_mock.go")

			inputFilePath = inputFile.Name()

			_, err = inputFile.WriteString(simpleContent)
			Expect(err).NotTo(HaveOccurred())

			Expect(inputFile.Close()).To(Succeed())
		})

		Context("if valid parameters provided", func() {
			runWithParams := func(params ...string) {
				command := exec.Command(mockGeneratorPath, params...)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(outputFilePath).To(BeAnExistingFile())
			}

			It("it should output a file", func() {
				By("file name")
				runWithParams(inputFilePath)

				By("file name after --")
				runWithParams("--", inputFilePath)
			})
		})
	})

	Context("by mock generation", func() {
		DescribeTable("generate mock content",
			func(input, expectedOutputTmpl string, extra ...Options) {
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
			},
			Entry("no params and no return", `
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
			Entry("2 param with different types, no return", `
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
			Entry("2 param with the same types, no return", `
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
			Entry("no params, 1 return with name", `
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
			Entry("no params, 2 returns with names", `
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
			Entry("no params, returns without names but same types", `
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
			Entry("param without name", `
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
			Entry("2 params without name", `
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
			Entry("slice param without name", `
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
			Entry("slice param of pointers without name", `
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
			Entry("slice param of structs from package", `
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
			Entry("return slice of structs from package", `
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
			Entry("param of struct pointer from package", `
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
			Entry("return struct pointer from package", `
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
			Entry("2 interfaces", `
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
			Entry("struct from same package", `
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
			Entry("alias for struct from another package", `
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
			Entry("multiple generic params", `
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
			Entry("generic of slice of generics", `
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
			Entry("packaged inside generic", `
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
			Entry("embedded local interface", `
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
			Entry("ellipsis param without name", `
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
			Entry("ellipsis param with name", `
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
			Entry("embedded interface from another package", `
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
			Entry("import generic type", `
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
			Entry("import packages with same package name", `
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
			Entry("import different packages with same struct name (params)", `
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
			Entry("import different packages with same struct name (returns)", `
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
			Entry("return slice of imported package", `
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
			Entry("type alias param and return", `
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
			Entry("generic param and return", `
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
			Entry("packages sorting", `
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
			Entry("1 param, 1 return with handler", `
package pkg

type Service interface {
	Add(one int) string
}`,
				`package test

type ServiceMock struct {
	Add_calls         int
	Add_param_one     int
	Add_return        func(int) string
	Add_return_string string
}

func (this *ServiceMock) Add(one int) string {
	this.Add_calls++
	this.Add_param_one = one
	if this.Add_return != nil {
		return this.Add_return(one)
	}
	return this.Add_return_string
}
`, Options{GenerateReturnHandler: true}),
			Entry("no params, 2 returns with handler", `
package pkg

type Service interface {
	Add() (string, error)
}`,
				`package test

type ServiceMock struct {
	Add_calls         int
	Add_return        func() (string, error)
	Add_return_string string
	Add_return_error  error
}

func (this *ServiceMock) Add() (string, error) {
	this.Add_calls++
	if this.Add_return != nil {
		return this.Add_return()
	}
	return this.Add_return_string, this.Add_return_error
}
`, Options{GenerateReturnHandler: true}),
			Entry("multiple params/returns with handler", `
package pkg

type Service interface {
	Add(a int, b string) (int, error)
}`,
				`package test

type ServiceMock struct {
	Add_calls        int
	Add_param_a      int
	Add_param_b      string
	Add_return       func(int, string) (int, error)
	Add_return_int   int
	Add_return_error error
}

func (this *ServiceMock) Add(a int, b string) (int, error) {
	this.Add_calls++
	this.Add_param_a = a
	this.Add_param_b = b
	if this.Add_return != nil {
		return this.Add_return(a, b)
	}
	return this.Add_return_int, this.Add_return_error
}
`, Options{GenerateReturnHandler: true}),
			Entry("thread-safe mock with returns", `
package pkg

type Service interface {
	Add(one int) string
	Process()
}`,
				`package test

import (
	"sync"
)

type ServiceMock struct {
	mu sync.Mutex

	add_calls         int
	add_param_one     int
	add_return_string string

	process_calls int
}

func (this *ServiceMock) Add(one int) string {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.add_calls++
	this.add_param_one = one
	return this.add_return_string
}

func (this *ServiceMock) Add_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.add_calls
}

func (this *ServiceMock) Add_param_one() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.add_param_one
}

func (this *ServiceMock) Add_return_string(value string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.add_return_string = value
}

func (this *ServiceMock) Process() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.process_calls++
}

func (this *ServiceMock) Process_calls() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.process_calls
}
`, Options{MakeThreadSafe: true}),
			Entry("simple interface with header note", `
package pkg

type Service interface {
	DoSomething()
}`,
				`// Code generated by MockGen. DO NOT EDIT.

package test

type ServiceMock struct {
	DoSomething_calls int
}

func (this *ServiceMock) DoSomething() {
	this.DoSomething_calls++
}
`, Options{AddHeaderNote: true}),
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
		)
	})
})
