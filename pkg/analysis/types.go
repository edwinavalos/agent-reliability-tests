package analysis

import "time"

type LogEntry struct {
	Loop              int
	Timestamp         time.Time
	Prompt            string
	MainAgentResponse string // "What I told the agent"
	SubAgentResponse  string // "Agent's response"
	RawResponse       string // Original full response
	Errors            string
	ExecutionTime     time.Duration
}

type SimilarityScore struct {
	Index1         int
	Index2         int
	LevenshteinSim float64
	JaccardSim     float64
	OverallSim     float64
}

type AnalysisResult struct {
	TotalResponses    int
	AverageSimilarity float64
	MostCommonPattern string
	MostCommonCount   int
	MostAbnormal      LogEntry
	AbnormalityScore  float64
	SimilarityMatrix  [][]float64
	Clusters          []ResponseCluster
}

type DualAgentAnalysisResult struct {
	TotalEntries       int
	MainAgentAnalysis  *AnalysisResult
	SubAgentAnalysis   *AnalysisResult
	MainAgentResponses []string
	SubAgentResponses  []string
	Entries            []LogEntry
}

type ResponseCluster struct {
	Responses []int // indices of responses in this cluster
	Centroid  string
	Size      int
}
