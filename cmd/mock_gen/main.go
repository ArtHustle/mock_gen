package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	mock_gen "github.com/ArtHustle/moc_gen"
)

func determineOutputFile(inputFilePath string) string {
	getFileNameWithoutExt := func(fileName string) string {
		return fileName[:len(fileName)-len(filepath.Ext(fileName))]
	}
	return filepath.Join(filepath.Dir(inputFilePath), "test",
		getFileNameWithoutExt(filepath.Base(inputFilePath))+"_mock.go")
}

func storeToFile(filePath, content string) {
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

func main() {
	var outputFile string
	var generateReturnHandler bool
	var threadSafe bool
	var noHeaderNote bool

	var rootCmd = &cobra.Command{
		Use:   "mock_gen [input file]",
		Short: "Generate Go mocks for interfaces",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			inputFilePath := args[0]

			mockContent, e := mock_gen.GenerateMockFileWithOptions(inputFilePath, mock_gen.Options{
				GenerateReturnHandler: generateReturnHandler,
				MakeThreadSafe:        threadSafe,
				AddHeaderNote:         !noHeaderNote, // Invert the flag - by default add header
			})
			if e != nil {
				fmt.Println("Failed to generate mock content:", e)
				os.Exit(1)
			}

			if outputFile == "" {
				outputFile = determineOutputFile(inputFilePath)
			}
			storeToFile(outputFile, mockContent)
		},
	}

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (optional)")
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
