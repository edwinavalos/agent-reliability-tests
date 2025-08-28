package main

import (
	"fmt"
	"os"

	"agent-reliability-tests/pkg/reliability"

	"github.com/spf13/cobra"
)

var (
	loops    int
	filename string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "agent-reliability-tests [agent_name]",
		Short: "Run Claude agent reliability tests",
		Long:  "A tool to run Claude agent reliability tests with configurable loop counts.",
		Args:  cobra.ExactArgs(1),
		Run:   runTest,
	}

	rootCmd.Flags().IntVarP(&loops, "loops", "l", 1, "Number of times to run the test (default: 1)")
	rootCmd.Flags().StringVarP(&filename, "filename", "f", "chat", "Base name for output file (will be formatted as <name>_<unix_timestamp>.log)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runTest(cmd *cobra.Command, args []string) {
	config := reliability.TestConfig{
		AgentName: args[0],
		Loops:     loops,
		Filename:  filename,
	}

	result, err := reliability.RunReliabilityTest(config)
	if err != nil {
		fmt.Printf("Error running reliability test: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test completed successfully!\n")
	fmt.Printf("Results saved to: %s\n", result.OutputFile)
	fmt.Printf("Total duration: %v\n", result.Duration)
}
