package main

import (
	"fmt"
	"os"

	"agent-reliability-tests/pkg/reliability"

	"github.com/spf13/cobra"
)

var (
	loops        int
	filename     string
	parallel     bool
	batchSize    int
	queue        int
	promptTemplate string
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
	rootCmd.Flags().BoolVarP(&parallel, "parallel", "p", false, "Run tests in parallel batches (default: false, uses queue mode)")
	rootCmd.Flags().IntVar(&batchSize, "batch", 5, "Number of parallel executions to run at once (default: 5, only used with --parallel)")
	rootCmd.Flags().IntVarP(&queue, "queue", "q", 0, "Number of worker threads for queue mode (default: 1, mutually exclusive with --parallel)")
	rootCmd.Flags().StringVar(&promptTemplate, "prompt", "", "Path to Go template file for custom prompts (if not provided, uses default prompt)")

	// Make --parallel and --queue mutually exclusive
	rootCmd.MarkFlagsMutuallyExclusive("parallel", "queue")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runTest(cmd *cobra.Command, args []string) {
	config := reliability.TestConfig{
		AgentName:      args[0],
		Loops:          loops,
		Filename:       filename,
		Parallel:       parallel,
		BatchSize:      batchSize,
		Queue:          queue,
		PromptTemplate: promptTemplate,
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
