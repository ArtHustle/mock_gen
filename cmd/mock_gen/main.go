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
	// Skip the write if the existing file is byte-identical. Preserves mtime,
	// avoiding spurious git noise and unnecessary build-cache invalidations
	// when `mock_gen` runs as part of `go generate` over an unchanged source.
	if existing, err := os.ReadFile(filePath); err == nil && string(existing) == content {
		fmt.Println("Mock unchanged:", filePath)
		return
	}

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
	var threadSafe bool
	var noHeaderNote bool
	var inspectEntry string
	var configEntry string

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
  mock_gen --thread-safe service.go
  mock_gen --inspect-entry Inspect --config-entry Configure service.go
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
				MakeThreadSafe: threadSafe,
				AddHeaderNote:  !noHeaderNote,
				CustomTemplate: customTemplate,
				InspectEntry:   inspectEntry,
				ConfigEntry:    configEntry,
			}

			// Recursive: path ends in /... — walk the tree, generating
			// per-source mocks in every package directory that contains
			// at least one interface.
			if strings.HasSuffix(inputPath, "/...") || inputPath == "..." {
				root := strings.TrimSuffix(inputPath, "/...")
				if root == "" || root == "..." {
					root = "."
				}
				if outputFile != "" {
					fmt.Println("--output is not supported with recursive input (multiple outputs)")
					os.Exit(1)
				}
				generateRecursive(root, opts)
				return
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
	rootCmd.Flags().BoolVar(&threadSafe, "thread-safe", false,
		"If set, adds a mutex to the mock struct and uses it to make all methods thread-safe")
	rootCmd.Flags().StringVar(&inspectEntry, "inspect-entry", "",
		"Override the inspection namespace entry method name (default: Calls). "+
			"Use this when the source interface has a method named Calls.")
	rootCmd.Flags().StringVar(&configEntry, "config-entry", "",
		"Override the configuration namespace entry method name (default: On). "+
			"Use this when the source interface has a method named On.")
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
	written := generateDirCount(dir, opts)
	if written == 0 {
		fmt.Printf("No interfaces found in %s\n", dir)
		os.Exit(1)
	}
}

// generateDirCount processes one directory and returns the number of mock
// files written. Used as a building block for generateDir (single dir) and
// generateRecursive (walk the tree). It does not exit on empty input.
func generateDirCount(dir string, opts mock_gen.Options) int {
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
	return written
}

// generateRecursive walks rootDir, invoking generateDirCount on every
// subdirectory that contains non-test .go files. Skips "test" directories
// (they hold our own output and should not be re-input).
func generateRecursive(rootDir string, opts mock_gen.Options) {
	total := 0
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}
		// Skip our own output directories — they're generated mocks, not
		// inputs. Re-feeding them would create mocks-of-mocks.
		if d.Name() == "test" && path != rootDir {
			return filepath.SkipDir
		}
		// Skip hidden dirs (.git, .idea, ...) and vendor.
		name := d.Name()
		if path != rootDir && (strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules") {
			return filepath.SkipDir
		}
		total += generateDirCount(path, opts)
		return nil
	})
	if err != nil {
		fmt.Println("Failed to walk:", err)
		os.Exit(1)
	}
	if total == 0 {
		fmt.Printf("No interfaces found under %s\n", rootDir)
		os.Exit(1)
	}
}
