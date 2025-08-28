package analysis

import (
	"sort"
	"strings"
	"unicode"
)

// LevenshteinDistance calculates the edit distance between two strings
func LevenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // deletion, insertion
				matrix[i-1][j-1]+cost,                   // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// LevenshteinSimilarity converts distance to similarity score (0-1)
func LevenshteinSimilarity(a, b string) float64 {
	distance := LevenshteinDistance(a, b)
	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(distance)/float64(maxLen)
}

// JaccardSimilarity calculates word-based Jaccard similarity
func JaccardSimilarity(a, b string) float64 {
	wordsA := tokenize(strings.ToLower(a))
	wordsB := tokenize(strings.ToLower(b))

	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, word := range wordsA {
		setA[word] = true
	}
	for _, word := range wordsB {
		setB[word] = true
	}

	intersection := 0
	for word := range setA {
		if setB[word] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 1.0
	}

	return float64(intersection) / float64(union)
}

// OverallSimilarity combines multiple similarity measures
func OverallSimilarity(a, b string) float64 {
	lev := LevenshteinSimilarity(a, b)
	jac := JaccardSimilarity(a, b)

	// Weighted average: Levenshtein 40%, Jaccard 60%
	return 0.4*lev + 0.6*jac
}

// CalculateSimilarityMatrix computes pairwise similarities
func CalculateSimilarityMatrix(responses []string) [][]float64 {
	n := len(responses)
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
		matrix[i][i] = 1.0 // self-similarity is 1
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			similarity := OverallSimilarity(responses[i], responses[j])
			matrix[i][j] = similarity
			matrix[j][i] = similarity // symmetric
		}
	}

	return matrix
}

// FindAverageSimilarity calculates the average similarity across all pairs
func FindAverageSimilarity(matrix [][]float64) float64 {
	if len(matrix) <= 1 {
		return 1.0
	}

	total := 0.0
	count := 0
	n := len(matrix)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			total += matrix[i][j]
			count++
		}
	}

	if count == 0 {
		return 1.0
	}
	return total / float64(count)
}

// FindMostAbnormal identifies the response most different from others
func FindMostAbnormal(entries []LogEntry, matrix [][]float64) (LogEntry, float64) {
	if len(entries) == 0 {
		return LogEntry{}, 0.0
	}

	minAvgSimilarity := 1.0
	mostAbnormalIndex := 0

	for i := range entries {
		avgSimilarity := 0.0
		count := 0

		for j := range entries {
			if i != j {
				avgSimilarity += matrix[i][j]
				count++
			}
		}

		if count > 0 {
			avgSimilarity /= float64(count)
			if avgSimilarity < minAvgSimilarity {
				minAvgSimilarity = avgSimilarity
				mostAbnormalIndex = i
			}
		}
	}

	abnormalityScore := 1.0 - minAvgSimilarity
	return entries[mostAbnormalIndex], abnormalityScore
}

// tokenize splits text into words, removing punctuation
func tokenize(text string) []string {
	var words []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// ClusterResponses groups similar responses together
func ClusterResponses(responses []string, matrix [][]float64, threshold float64) []ResponseCluster {
	n := len(responses)
	if n == 0 {
		return nil
	}

	visited := make([]bool, n)
	var clusters []ResponseCluster

	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}

		cluster := ResponseCluster{
			Responses: []int{i},
			Centroid:  responses[i],
			Size:      1,
		}
		visited[i] = true

		// Find all responses similar to this one
		for j := i + 1; j < n; j++ {
			if !visited[j] && matrix[i][j] >= threshold {
				cluster.Responses = append(cluster.Responses, j)
				cluster.Size++
				visited[j] = true
			}
		}

		clusters = append(clusters, cluster)
	}

	// Sort clusters by size (largest first)
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Size > clusters[j].Size
	})

	return clusters
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
