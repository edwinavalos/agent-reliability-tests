package analysis

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ParseLogFile(filename string) ([]LogEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var entries []LogEntry
	var currentEntry LogEntry
	var inResponse bool
	var responseBuilder strings.Builder

	scanner := bufio.NewScanner(file)
	headerRegex := regexp.MustCompile(`^=== Loop (\d+)/\d+ - (.+) ===`)
	promptRegex := regexp.MustCompile(`^Prompt: (.+)`)
	responseStartRegex := regexp.MustCompile(`^Response:`)
	errorStartRegex := regexp.MustCompile(`^Errors:`)
	executionTimeRegex := regexp.MustCompile(`^Execution time: (.+)`)
	separatorRegex := regexp.MustCompile(`^---`)

	for scanner.Scan() {
		line := scanner.Text()

		if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			// Start of new entry
			if currentEntry.Loop != 0 {
				// Save previous entry
				rawResponse := strings.TrimSpace(responseBuilder.String())
				currentEntry.RawResponse = rawResponse

				// Extract both agent responses
				mainResp, subResp := extractBothAgentResponses(rawResponse)
				currentEntry.MainAgentResponse = mainResp
				currentEntry.SubAgentResponse = subResp

				entries = append(entries, currentEntry)
			}

			// Parse loop number
			loop, _ := strconv.Atoi(matches[1])

			// Parse timestamp
			timestamp, err := time.Parse("2006-01-02 15:04:05 UTC", matches[2])
			if err != nil {
				// Try alternative format
				timestamp = time.Now() // fallback
			}

			currentEntry = LogEntry{
				Loop:      loop,
				Timestamp: timestamp,
			}
			responseBuilder.Reset()
			inResponse = false
		} else if matches := promptRegex.FindStringSubmatch(line); matches != nil {
			currentEntry.Prompt = matches[1]
		} else if responseStartRegex.MatchString(line) {
			inResponse = true
			responseBuilder.Reset()
		} else if errorStartRegex.MatchString(line) {
			inResponse = false
		} else if matches := executionTimeRegex.FindStringSubmatch(line); matches != nil {
			duration, _ := time.ParseDuration(matches[1])
			currentEntry.ExecutionTime = duration
			inResponse = false
		} else if separatorRegex.MatchString(line) {
			inResponse = false
		} else if inResponse && strings.TrimSpace(line) != "" {
			if responseBuilder.Len() > 0 {
				responseBuilder.WriteString("\n")
			}
			responseBuilder.WriteString(line)
		}
	}

	// Don't forget the last entry
	if currentEntry.Loop != 0 {
		rawResponse := strings.TrimSpace(responseBuilder.String())
		currentEntry.RawResponse = rawResponse

		// Extract both agent responses
		mainResp, subResp := extractBothAgentResponses(rawResponse)
		currentEntry.MainAgentResponse = mainResp
		currentEntry.SubAgentResponse = subResp

		entries = append(entries, currentEntry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return entries, nil
}

// extractBothAgentResponses extracts both main agent and sub agent responses
func extractBothAgentResponses(rawResponse string) (mainAgent, subAgent string) {
	// Extract main agent response ("What I told the agent")
	mainAgent = extractMainAgentResponse(rawResponse)

	// Extract sub agent response ("Agent's response")
	subAgent = extractSubAgentResponse(rawResponse)

	// If no structured format found, treat entire response as sub agent response
	if mainAgent == "" && subAgent == "" {
		subAgent = strings.TrimSpace(rawResponse)
	}

	return mainAgent, subAgent
}

// extractMainAgentResponse extracts the "What I told the agent" part
func extractMainAgentResponse(rawResponse string) string {
	// Pattern 1: **What I told the agent:** "text"
	mainPattern1 := regexp.MustCompile(`\*\*What I told the agent:\*\*\s*"([^"]+)"`)
	if matches := mainPattern1.FindStringSubmatch(rawResponse); matches != nil {
		return strings.TrimSpace(matches[1])
	}

	// Pattern 2: **What I told the agent:**\n"text"
	mainPattern2 := regexp.MustCompile(`(?s)\*\*What I told the agent:\*\*\s*\n\s*"([^"]+)"`)
	if matches := mainPattern2.FindStringSubmatch(rawResponse); matches != nil {
		return strings.TrimSpace(matches[1])
	}

	// Pattern 3: Look for text after "What I told the agent:" (flexible)
	lines := strings.Split(rawResponse, "\n")
	foundMainAgent := false
	var mainLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "What I told the agent") {
			foundMainAgent = true
			// Check if response starts on same line
			parts := strings.Split(line, "What I told the agent:")
			if len(parts) > 1 {
				remaining := strings.TrimSpace(parts[1])
				if remaining != "" {
					remaining = strings.Trim(remaining, `"`)
					mainLines = append(mainLines, remaining)
				}
			}
			continue
		}

		if foundMainAgent && line != "" {
			// Stop when we hit the agent response section
			if strings.Contains(line, "Agent") && strings.Contains(line, "response") {
				break
			}
			// Remove quotes if this line is wrapped in quotes
			line = strings.Trim(line, `"`)
			mainLines = append(mainLines, line)
		}
	}

	if len(mainLines) > 0 {
		return strings.TrimSpace(strings.Join(mainLines, " "))
	}

	return ""
}

// extractSubAgentResponse extracts the "Agent's response" part
func extractSubAgentResponse(rawResponse string) string {
	// Pattern 1: **Agent's response:** "response text"
	agentPattern1 := regexp.MustCompile(`\*\*Agent[^:]*response[^:]*:\*\*\s*"([^"]+)"`)
	if matches := agentPattern1.FindStringSubmatch(rawResponse); matches != nil {
		return strings.TrimSpace(matches[1])
	}

	// Pattern 2: **Agent's response:** (without quotes, multiline)
	agentPattern2 := regexp.MustCompile(`(?s)\*\*Agent[^:]*response[^:]*:\*\*\s*(.+?)(?:\n\n|$)`)
	if matches := agentPattern2.FindStringSubmatch(rawResponse); matches != nil {
		response := strings.TrimSpace(matches[1])
		// Remove quotes if present
		if strings.HasPrefix(response, `"`) && strings.HasSuffix(response, `"`) {
			response = strings.Trim(response, `"`)
		}
		return strings.TrimSpace(response)
	}

	// Pattern 3: **Agent's response:**\n"response text"
	agentPattern3 := regexp.MustCompile(`(?s)\*\*Agent[^:]*response[^:]*:\*\*\s*\n\s*"([^"]+)"`)
	if matches := agentPattern3.FindStringSubmatch(rawResponse); matches != nil {
		return strings.TrimSpace(matches[1])
	}

	// Pattern 4: Look for text after "Agent response:" or similar (most flexible)
	lines := strings.Split(rawResponse, "\n")
	foundAgentResponse := false
	var responseLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "agent") && strings.Contains(strings.ToLower(line), "response") {
			foundAgentResponse = true
			// Check if response starts on same line
			parts := regexp.MustCompile(`[Aa]gent[^:]*response[^:]*:`).Split(line, 2)
			if len(parts) > 1 {
				remaining := strings.TrimSpace(parts[1])
				if remaining != "" {
					remaining = strings.Trim(remaining, `"`)
					responseLines = append(responseLines, remaining)
				}
			}
			continue
		}

		if foundAgentResponse && line != "" {
			// Skip lines that look like metadata
			if strings.HasPrefix(line, "**What I told") {
				break
			}
			// Remove quotes if this line is wrapped in quotes
			line = strings.Trim(line, `"`)
			responseLines = append(responseLines, line)
		}
	}

	if len(responseLines) > 0 {
		return strings.TrimSpace(strings.Join(responseLines, " "))
	}

	return ""
}
