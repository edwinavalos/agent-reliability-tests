package reliability

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type TestConfig struct {
	AgentName string
	Loops     int
	Filename  string
	Parallel  bool
}

type TestResult struct {
	OutputFile string
	Duration   time.Duration
}

// RunReliabilityTest executes the reliability test with the given configuration
func RunReliabilityTest(config TestConfig) (*TestResult, error) {
	// Generate timestamped filename
	timestamp := time.Now().Unix()
	outputFile := fmt.Sprintf("%s_%d.log", config.Filename, timestamp)

	fmt.Printf("Running %d loop(s) with agent: %s\n", config.Loops, config.AgentName)
	fmt.Printf("Output file: %s\n", outputFile)

	startTime := time.Now()

	for i := 1; i <= config.Loops; i++ {
		fmt.Printf("\n=== Loop %d/%d ===\n", i, config.Loops)

		// Create the prompt using the specified pattern
		prompt := fmt.Sprintf("use the %s agent and ask it to say 'hello', return what you told the agent, and just its response to you asking it to say 'hello'", config.AgentName)

		fmt.Printf("Executing claude with agent: %s\n", config.AgentName)
		fmt.Printf("Prompt: %s\n\n", prompt)

		// Execute claude with the specified flags and prompt
		claudeCmd := exec.Command("claude", "-p", "--permission-mode", "acceptEdits", prompt)

		// Capture output in buffers
		var stdout, stderr bytes.Buffer
		claudeCmd.Stdout = &stdout
		claudeCmd.Stderr = &stderr
		// Don't set claudeCmd.Stdin to avoid interactive prompts

		// Record start time
		loopStartTime := time.Now()
		fmt.Printf("Starting execution at: %s\n", loopStartTime.Format("2006-01-02 15:04:05"))

		// Run the command
		err := claudeCmd.Run()

		// Record end time
		loopEndTime := time.Now()

		// Display output to console
		if stdout.Len() > 0 {
			fmt.Print(stdout.String())
		}
		if stderr.Len() > 0 {
			fmt.Fprint(os.Stderr, stderr.String())
		}

		// Log the interaction
		logEntry := fmt.Sprintf("=== Loop %d/%d - %s ===\n", i, config.Loops, loopEndTime.UTC().Format("2006-01-02 15:04:05 UTC"))
		logEntry += fmt.Sprintf("Prompt: %s\n", prompt)
		logEntry += fmt.Sprintf("Response:\n%s\n", strings.TrimSpace(stdout.String()))
		if stderr.Len() > 0 {
			logEntry += fmt.Sprintf("Errors:\n%s\n", strings.TrimSpace(stderr.String()))
		}
		logEntry += fmt.Sprintf("Execution time: %v\n", loopEndTime.Sub(loopStartTime))
		logEntry += "---\n\n"

		// Append to log file
		if err := appendToLog(outputFile, logEntry); err != nil {
			log.Printf("Error writing to log file: %v", err)
		}

		if err != nil {
			log.Printf("Error executing claude in loop %d: %v", i, err)
			continue
		}

		fmt.Printf("\nExecution completed at: %s\n", loopEndTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Total execution time: %v\n", loopEndTime.Sub(loopStartTime))

		if i < config.Loops {
			fmt.Printf("Waiting before next iteration...\n")
			time.Sleep(1 * time.Second)
		}
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\n=== All %d loops completed ===\n", config.Loops)

	return &TestResult{
		OutputFile: outputFile,
		Duration:   totalDuration,
	}, nil
}

func appendToLog(filename, content string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}
