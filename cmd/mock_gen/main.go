package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	mock_gen "github.com/ArtHustle/mock_gen"
)

func determineOutputFile(inputFilePath string) string {
	getFileNameWithoutExt := func(fileName string) string {
		return fileName[:len(fileName)-len(filepath.Ext(fileName))]
	}
	return filepath.Join(filepath.Dir(inputFilePath), "test",
		getFileNameWithoutExt(filepath.Base(inputFilePath))+"_mock.go")
}

func storeToFile(filePath, content string) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		fmt.Println("Failed to create output dir:", err)
		os.Exit(1)
	}
	outputFile, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Failed to create output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	fmt.Println("Generating mock file:", filePath)

	_, err = outputFile.WriteString(content)
	if err != nil {
		fmt.Println("Failed to write output file:", err)
		os.Exit(1)
	}
}

// mockStructRe matches `type <Name>Mock struct {`. Used to decide whether a
// generated output file contains at least one mock (i.e. the source file
// declared at least one interface) — if not, we skip writing it in
// directory-mode invocations.
var mockStructRe = regexp.MustCompile(`(?m)^type \w+Mock struct`)

func hasMock(content string) bool {
	return mockStructRe.MatchString(content)
}

// listSourceFiles returns the non-test .go files in dir, sorted by name.
// _test.go files are excluded.
func listSourceFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		out = append(out, filepath.Join(dir, name))
	}
	sort.Strings(out)
	return out, nil
}

func main() {
	var outputFile string
	var templateFile string
	var generateReturnHandler bool
	var threadSafe bool
	var noHeaderNote bool

	var rootCmd = &cobra.Command{
		Use:   "mock_gen [flags] <source_file.go | dir>",
		Short: "Generate Go mocks for interfaces",
		Long: `mock_gen generates mock implementations for all interfaces in a Go source file.

The input is either a single .go file or a directory. For a directory, every
non-test .go file in it is scanned and one mock is generated per source file
that declares at least one interface. The input (file or directory) must be
inside a Go module (a directory tree containing a go.mod file). Generated
mocks are placed in a "test" subdirectory by default.

Examples:
  mock_gen service.go
  mock_gen --output mocks/service_mock.go service.go
  mock_gen --generate-return-handler --thread-safe service.go
  mock_gen ./internal/service`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			inputPath := args[0]

			var customTemplate string
			if templateFile != "" {
				tmplBytes, err := os.ReadFile(templateFile)
				if err != nil {
					fmt.Println("Failed to read template file:", err)
					os.Exit(1)
				}
				customTemplate = string(tmplBytes)
			}

			opts := mock_gen.Options{
				GenerateReturnHandler: generateReturnHandler,
				MakeThreadSafe:        threadSafe,
				AddHeaderNote:         !noHeaderNote,
				CustomTemplate:        customTemplate,
			}

			info, err := os.Stat(inputPath)
			if err != nil {
				fmt.Println("Failed to stat input:", err)
				os.Exit(1)
			}

			if info.IsDir() {
				if outputFile != "" {
					fmt.Println("--output is not supported with a directory input (multiple outputs)")
					os.Exit(1)
				}
				generateDir(inputPath, opts)
				return
			}

			generateSingleFile(inputPath, outputFile, opts)
		},
	}

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (optional, single-file mode only)")
	rootCmd.Flags().StringVarP(&templateFile, "template", "t", "",
		"Path to a custom Go text/template file (.tmpl) for code generation. "+
			"The template receives a MockData struct with parsed interface information "+
			"(imports, interface names, methods, params, returns, signatures). "+
			"Output is run through gofmt, so it must produce valid Go. "+
			"Use mock_gen.DefaultTemplate() programmatically to get the built-in template as a starting point.")
	rootCmd.Flags().BoolVar(&generateReturnHandler, "generate-return-handler", false,
		"If set, generates a single return handler field (e.g., Method_return func(...)), and ignores other return value fields. If the handler is provided, other return values (like Method_return_error) will be ignored.")
	rootCmd.Flags().BoolVar(&threadSafe, "thread-safe", false,
		"If set, adds a mutex to the mock struct and uses it to make all methods thread-safe")
	rootCmd.Flags().BoolVar(&noHeaderNote, "no-header-note", false,
		"If set, skips adding the 'Code generated' header comment to the generated file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateSingleFile(inputFilePath, outputFile string, opts mock_gen.Options) {
	mockContent, e := mock_gen.GenerateMockFileWithOptions(inputFilePath, opts)
	if e != nil {
		fmt.Println("Failed to generate mock content:", e)
		os.Exit(1)
	}

	if outputFile == "" {
		outputFile = determineOutputFile(inputFilePath)
	}
	storeToFile(outputFile, mockContent)
}

func generateDir(dir string, opts mock_gen.Options) {
	files, err := listSourceFiles(dir)
	if err != nil {
		fmt.Println("Failed to list source files:", err)
		os.Exit(1)
	}

	written := 0
	for _, f := range files {
		content, err := mock_gen.GenerateMockFileWithOptions(f, opts)
		if err != nil {
			fmt.Printf("Failed to generate mock for %s: %v\n", f, err)
			os.Exit(1)
		}
		if !hasMock(content) {
			continue
		}
		storeToFile(determineOutputFile(f), content)
		written++
	}

	if written == 0 {
		fmt.Printf("No interfaces found in %s\n", dir)
		os.Exit(1)
	}
}
