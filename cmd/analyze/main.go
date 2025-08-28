package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agent-reliability-tests/pkg/analysis"

	"github.com/spf13/cobra"
)

var (
	verbose    bool
	outputFile string
	debug      bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "analyze [log_file]",
		Short: "Analyze Claude agent reliability test logs",
		Long: `Analyze Claude agent reliability test logs to quantify response similarity,
identify common patterns, and detect abnormal responses.

The analysis provides:
- Overall similarity metrics between responses
- Clustering of similar responses  
- Most common response pattern
- Most abnormal/outlier response
- Reliability assessment`,
		Args: cobra.ExactArgs(1),
		Run:  runAnalysis,
	}

	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output including similarity matrix")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Save detailed results to file")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Show extracted responses for debugging")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runAnalysis(cmd *cobra.Command, args []string) {
	logFile := args[0]

	// Check if file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Printf("Error: Log file '%s' does not exist\n", logFile)
		os.Exit(1)
	}

	fmt.Printf("Analyzing log file: %s\n", logFile)
	fmt.Println("Processing...")

	// Parse entries for debug mode (if needed)
	var entries []analysis.LogEntry
	var err error
	if debug {
		entries, err = analysis.ParseLogFile(logFile)
		if err != nil {
			fmt.Printf("Error parsing log file: %v\n", err)
			os.Exit(1)
		}
	}

	// Perform dual agent analysis
	result, err := analysis.AnalyzeLogFile(logFile)
	if err != nil {
		fmt.Printf("Error analyzing log file: %v\n", err)
		os.Exit(1)
	}

	if result.TotalEntries == 0 {
		fmt.Println("No log entries found in log file")
		return
	}

	// Print debug output if requested (before main results)
	if debug {
		if entries == nil {
			entries = result.Entries
		}
		printDualAgentDebugOutput(entries)
	}

	// Print results
	analysis.PrintDualAgentAnalysisResult(result)

	// Print verbose output if requested
	if verbose {
		printVerboseOutput(result)
	}

	// Save to file if requested
	if outputFile != "" {
		if err := saveDualAgentResults(result, outputFile); err != nil {
			fmt.Printf("Warning: Failed to save results to file: %v\n", err)
		} else {
			fmt.Printf("\nDetailed results saved to: %s\n", outputFile)
		}
	}
}

func printVerboseOutput(result *analysis.DualAgentAnalysisResult) {
	// Print verbose output for Main Agent
	if result.MainAgentAnalysis != nil {
		fmt.Println("\n" + strings.Repeat("=", 40))
		fmt.Println("MAIN AGENT VERBOSE OUTPUT")
		fmt.Println(strings.Repeat("=", 40))
		printVerboseAnalysis(result.MainAgentAnalysis, "Main Agent")
	}

	// Print verbose output for Sub Agent
	if result.SubAgentAnalysis != nil {
		fmt.Println("\n" + strings.Repeat("=", 40))
		fmt.Println("SUB AGENT VERBOSE OUTPUT")
		fmt.Println(strings.Repeat("=", 40))
		printVerboseAnalysis(result.SubAgentAnalysis, "Sub Agent")
	}
}

func printVerboseAnalysis(result *analysis.AnalysisResult, agentName string) {
	fmt.Printf("\n--- %s SIMILARITY MATRIX ---\n", strings.ToUpper(agentName))
	matrix := result.SimilarityMatrix
	n := len(matrix)

	if n > 10 {
		fmt.Printf("Matrix too large (%dx%d), showing first 10x10 subset:\n", n, n)
		n = 10
	}

	// Print header
	fmt.Print("     ")
	for j := 0; j < n; j++ {
		fmt.Printf("%6d", j+1)
	}
	fmt.Println()

	// Print matrix
	for i := 0; i < n; i++ {
		fmt.Printf("%3d: ", i+1)
		for j := 0; j < n; j++ {
			fmt.Printf("%6.3f", matrix[i][j])
		}
		fmt.Println()
	}

	fmt.Printf("\n--- %s DETAILED CLUSTERS ---\n", strings.ToUpper(agentName))
	for i, cluster := range result.Clusters {
		fmt.Printf("Cluster %d (%d responses):\n", i+1, cluster.Size)
		fmt.Printf("  Representative: \"%s\"\n", truncateString(cluster.Centroid, 100))
		fmt.Printf("  Response indices: %v\n", cluster.Responses)
		if i >= 4 { // Limit to first 5 clusters
			remaining := len(result.Clusters) - 5
			if remaining > 0 {
				fmt.Printf("... and %d more clusters\n", remaining)
			}
			break
		}
	}
}

func saveDualAgentResults(result *analysis.DualAgentAnalysisResult, filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write detailed dual agent analysis
	fmt.Fprintf(file, "Dual Claude Agent Reliability Analysis Report\n")
	fmt.Fprintf(file, "==========================================\n\n")

	fmt.Fprintf(file, "Total Log Entries: %d\n", result.TotalEntries)
	fmt.Fprintf(file, "Main Agent Responses: %d\n", len(result.MainAgentResponses))
	fmt.Fprintf(file, "Sub Agent Responses: %d\n\n", len(result.SubAgentResponses))

	// Save Main Agent Analysis
	if result.MainAgentAnalysis != nil {
		fmt.Fprintf(file, "=== MAIN AGENT ANALYSIS ===\n")
		saveAnalysisToFile(file, result.MainAgentAnalysis, "Main Agent")
		fmt.Fprintf(file, "\n")
	}

	// Save Sub Agent Analysis
	if result.SubAgentAnalysis != nil {
		fmt.Fprintf(file, "=== SUB AGENT ANALYSIS ===\n")
		saveAnalysisToFile(file, result.SubAgentAnalysis, "Sub Agent")
	}

	return nil
}

func saveAnalysisToFile(file *os.File, result *analysis.AnalysisResult, agentName string) {
	fmt.Fprintf(file, "Total Responses: %d\n", result.TotalResponses)
	fmt.Fprintf(file, "Average Similarity: %.4f\n", result.AverageSimilarity)
	fmt.Fprintf(file, "Most Common Pattern Count: %d\n", result.MostCommonCount)
	fmt.Fprintf(file, "Abnormality Score: %.4f\n\n", result.AbnormalityScore)

	fmt.Fprintf(file, "Most Common Pattern:\n%s\n\n", result.MostCommonPattern)

	fmt.Fprintf(file, "Most Abnormal Response (Loop %d):\n%s\n\n",
		result.MostAbnormal.Loop, result.MostAbnormal.MainAgentResponse+" "+result.MostAbnormal.SubAgentResponse)

	fmt.Fprintf(file, "Similarity Matrix:\n")
	for i, row := range result.SimilarityMatrix {
		fmt.Fprintf(file, "Row %d: ", i+1)
		for _, val := range row {
			fmt.Fprintf(file, "%.4f ", val)
		}
		fmt.Fprintf(file, "\n")
	}

	fmt.Fprintf(file, "\nClusters:\n")
	for i, cluster := range result.Clusters {
		fmt.Fprintf(file, "Cluster %d: %d responses - %v\n",
			i+1, cluster.Size, cluster.Responses)
		fmt.Fprintf(file, "  Centroid: %s\n", cluster.Centroid)
	}
}

func printDualAgentDebugOutput(entries []analysis.LogEntry) {
	fmt.Println("\n=== DEBUG: DUAL AGENT EXTRACTED RESPONSES ===")
	for i, entry := range entries {
		fmt.Printf("Loop %d:\n", entry.Loop)

		if entry.MainAgentResponse != "" {
			fmt.Printf("  Main Agent: \"%s\"\n", truncateString(entry.MainAgentResponse, 100))
		} else {
			fmt.Printf("  Main Agent: [none]\n")
		}

		if entry.SubAgentResponse != "" {
			fmt.Printf("  Sub Agent:  \"%s\"\n", truncateString(entry.SubAgentResponse, 100))
		} else {
			fmt.Printf("  Sub Agent:  [none]\n")
		}

		if i < len(entries)-1 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
