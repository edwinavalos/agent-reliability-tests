package analysis

import (
	"fmt"
	"strings"
)

// AnalyzeLogFile performs comprehensive dual agent analysis on a log file
func AnalyzeLogFile(filename string) (*DualAgentAnalysisResult, error) {
	entries, err := ParseLogFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log file: %w", err)
	}

	if len(entries) == 0 {
		return &DualAgentAnalysisResult{}, nil
	}

	// Extract responses for both agents
	mainResponses := make([]string, 0, len(entries))
	subResponses := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.MainAgentResponse != "" {
			mainResponses = append(mainResponses, entry.MainAgentResponse)
		}
		if entry.SubAgentResponse != "" {
			subResponses = append(subResponses, entry.SubAgentResponse)
		}
	}

	// Analyze Main Agent responses
	var mainAnalysis *AnalysisResult
	if len(mainResponses) > 0 {
		mainAnalysis = analyzeResponses(mainResponses, entries, "main")
	}

	// Analyze Sub Agent responses
	var subAnalysis *AnalysisResult
	if len(subResponses) > 0 {
		subAnalysis = analyzeResponses(subResponses, entries, "sub")
	}

	return &DualAgentAnalysisResult{
		TotalEntries:       len(entries),
		MainAgentAnalysis:  mainAnalysis,
		SubAgentAnalysis:   subAnalysis,
		MainAgentResponses: mainResponses,
		SubAgentResponses:  subResponses,
		Entries:            entries,
	}, nil
}

// analyzeResponses performs analysis on a set of responses
func analyzeResponses(responses []string, allEntries []LogEntry, agentType string) *AnalysisResult {
	if len(responses) == 0 {
		return nil
	}

	// Calculate similarity matrix
	matrix := CalculateSimilarityMatrix(responses)
	avgSimilarity := FindAverageSimilarity(matrix)

	// Find most abnormal response - create entries that properly represent the responses being analyzed
	responseEntries := make([]LogEntry, len(responses))
	for i, response := range responses {
		// Find the entry that contains this response and create a properly structured entry
		for _, entry := range allEntries {
			var matchFound bool
			if agentType == "main" && entry.MainAgentResponse == response {
				matchFound = true
			} else if agentType == "sub" && entry.SubAgentResponse == response {
				matchFound = true
			}

			if matchFound {
				// Create an entry that properly represents which agent we're analyzing
				responseEntries[i] = LogEntry{
					Loop:              entry.Loop,
					Timestamp:         entry.Timestamp,
					Prompt:            entry.Prompt,
					MainAgentResponse: response, // Store the response we're analyzing as MainAgentResponse for consistency
					SubAgentResponse:  "",       // Clear the other to avoid confusion
					RawResponse:       response,
					Errors:            entry.Errors,
					ExecutionTime:     entry.ExecutionTime,
				}
				break
			}
		}
	}

	mostAbnormal, abnormalityScore := FindMostAbnormal(responseEntries, matrix)

	// Cluster responses (threshold of 0.7 for similarity)
	clusters := ClusterResponses(responses, matrix, 0.7)

	// Find most common pattern
	mostCommonPattern, mostCommonCount := findMostCommonPattern(responses, clusters)

	return &AnalysisResult{
		TotalResponses:    len(responses),
		AverageSimilarity: avgSimilarity,
		MostCommonPattern: mostCommonPattern,
		MostCommonCount:   mostCommonCount,
		MostAbnormal:      mostAbnormal,
		AbnormalityScore:  abnormalityScore,
		SimilarityMatrix:  matrix,
		Clusters:          clusters,
	}
}

// findMostCommonPattern identifies the most frequent response pattern
func findMostCommonPattern(responses []string, clusters []ResponseCluster) (string, int) {
	if len(clusters) == 0 {
		return "", 0
	}

	// The largest cluster represents the most common pattern
	largestCluster := clusters[0]

	// Use the first response in the cluster as the representative pattern
	if len(largestCluster.Responses) > 0 {
		patternIndex := largestCluster.Responses[0]
		if patternIndex < len(responses) {
			return responses[patternIndex], largestCluster.Size
		}
	}

	return "", 0
}

// PrintDualAgentAnalysisResult formats and displays the dual agent analysis results
func PrintDualAgentAnalysisResult(result *DualAgentAnalysisResult) {
	fmt.Println("=== DUAL CLAUDE AGENT RELIABILITY ANALYSIS ===")
	fmt.Printf("Total Log Entries: %d\n", result.TotalEntries)
	fmt.Printf("Main Agent Responses: %d\n", len(result.MainAgentResponses))
	fmt.Printf("Sub Agent Responses: %d\n", len(result.SubAgentResponses))

	// Print Main Agent Analysis
	if result.MainAgentAnalysis != nil {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("MAIN AGENT ANALYSIS (\"What I told the agent\")")
		fmt.Println(strings.Repeat("=", 60))
		printSingleAgentAnalysis(result.MainAgentAnalysis, "Main Agent")
	} else {
		fmt.Println("\n--- MAIN AGENT ANALYSIS ---")
		fmt.Println("No main agent responses found")
	}

	// Print Sub Agent Analysis
	if result.SubAgentAnalysis != nil {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("SUB AGENT ANALYSIS (\"Agent's response\")")
		fmt.Println(strings.Repeat("=", 60))
		printSingleAgentAnalysis(result.SubAgentAnalysis, "Sub Agent")
	} else {
		fmt.Println("\n--- SUB AGENT ANALYSIS ---")
		fmt.Println("No sub agent responses found")
	}
}

// printSingleAgentAnalysis prints analysis for a single agent
func printSingleAgentAnalysis(result *AnalysisResult, agentName string) {
	fmt.Printf("Total Responses: %d\n", result.TotalResponses)
	fmt.Printf("Average Similarity: %.3f (%.1f%%)\n", result.AverageSimilarity, result.AverageSimilarity*100)

	fmt.Println("\n--- CLUSTERING ANALYSIS ---")
	fmt.Printf("Found %d distinct response clusters\n", len(result.Clusters))

	for i, cluster := range result.Clusters {
		if i >= 3 { // Only show top 3 clusters
			break
		}
		percentage := float64(cluster.Size) / float64(result.TotalResponses) * 100
		fmt.Printf("Cluster %d: %d responses (%.1f%%) - \"%s\"\n",
			i+1, cluster.Size, percentage, truncateString(cluster.Centroid, 50))
	}

	fmt.Println("\n--- MOST COMMON PATTERN ---")
	if result.MostCommonPattern != "" {
		percentage := float64(result.MostCommonCount) / float64(result.TotalResponses) * 100
		fmt.Printf("Pattern: \"%s\"\n", result.MostCommonPattern)
		fmt.Printf("Frequency: %d/%d (%.1f%%)\n", result.MostCommonCount, result.TotalResponses, percentage)
	} else {
		fmt.Println("No dominant pattern found")
	}

	fmt.Println("\n--- MOST ABNORMAL RESPONSE ---")
	if result.AbnormalityScore > 0 {
		fmt.Printf("Abnormality Score: %.3f (%.1f%%)\n", result.AbnormalityScore, result.AbnormalityScore*100)
		fmt.Printf("Loop: %d\n", result.MostAbnormal.Loop)
		fmt.Printf("Response: \"%s\"\n", truncateString(getResponseFromEntry(result.MostAbnormal), 200))
		fmt.Printf("Timestamp: %s\n", result.MostAbnormal.Timestamp.Format("2006-01-02 15:04:05 UTC"))
	} else {
		fmt.Println("No significantly abnormal responses found")
	}

	fmt.Println("\n--- RELIABILITY ASSESSMENT ---")
	reliability := assessReliability(result)
	fmt.Printf("%s Reliability: %s\n", agentName, reliability)
}

// getResponseFromEntry extracts the appropriate response from a log entry
func getResponseFromEntry(entry LogEntry) string {
	if entry.MainAgentResponse != "" {
		return entry.MainAgentResponse
	}
	if entry.SubAgentResponse != "" {
		return entry.SubAgentResponse
	}
	return entry.RawResponse
}

// assessReliability provides a qualitative assessment based on metrics
func assessReliability(result *AnalysisResult) string {
	avgSim := result.AverageSimilarity
	abnormality := result.AbnormalityScore
	consistency := float64(result.MostCommonCount) / float64(result.TotalResponses)

	if avgSim >= 0.9 && abnormality <= 0.2 && consistency >= 0.8 {
		return "EXCELLENT - Highly consistent responses"
	} else if avgSim >= 0.7 && abnormality <= 0.4 && consistency >= 0.6 {
		return "GOOD - Generally consistent with minor variations"
	} else if avgSim >= 0.5 && abnormality <= 0.6 && consistency >= 0.4 {
		return "MODERATE - Some inconsistency present"
	} else if avgSim >= 0.3 && abnormality <= 0.8 {
		return "POOR - Significant inconsistencies detected"
	} else {
		return "VERY POOR - Highly unreliable responses"
	}
}

// truncateString truncates a string to maxLen with ellipsis
func truncateString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
