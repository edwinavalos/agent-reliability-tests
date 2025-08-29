package reliability

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
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

var logMutex sync.Mutex

// RunReliabilityTest executes the reliability test with the given configuration
func RunReliabilityTest(config TestConfig) (*TestResult, error) {
	// Generate timestamped filename
	timestamp := time.Now().Unix()
	outputFile := fmt.Sprintf("%s_%d.log", config.Filename, timestamp)

	mode := "sequential"
	if config.Parallel {
		mode = "parallel"
	}
	fmt.Printf("Running %d loop(s) with agent: %s (%s mode)\n", config.Loops, config.AgentName, mode)
	fmt.Printf("Output file: %s\n", outputFile)

	startTime := time.Now()

	if config.Parallel {
		return runParallelLoops(config, outputFile, startTime)
	}
	return runSequentialLoops(config, outputFile, startTime)
}

// runSequentialLoops executes loops one after another
func runSequentialLoops(config TestConfig, outputFile string, startTime time.Time) (*TestResult, error) {

	for i := 1; i <= config.Loops; i++ {
		fmt.Printf("\n=== Loop %d/%d ===\n", i, config.Loops)

		if err := executeSingleLoop(i, config, outputFile); err != nil {
			log.Printf("Error in loop %d: %v", i, err)
		}

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

// runParallelLoops executes all loops concurrently
func runParallelLoops(config TestConfig, outputFile string, startTime time.Time) (*TestResult, error) {
	var wg sync.WaitGroup
	errorChan := make(chan error, config.Loops)

	fmt.Printf("\n=== Starting %d parallel loops ===\n", config.Loops)

	for i := 1; i <= config.Loops; i++ {
		wg.Add(1)
		go func(loopNum int) {
			defer wg.Done()
			if err := executeSingleLoop(loopNum, config, outputFile); err != nil {
				errorChan <- fmt.Errorf("loop %d: %v", loopNum, err)
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// Collect any errors
	for err := range errorChan {
		log.Printf("Parallel execution error: %v", err)
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\n=== All %d parallel loops completed ===\n", config.Loops)

	return &TestResult{
		OutputFile: outputFile,
		Duration:   totalDuration,
	}, nil
}

// executeSingleLoop runs a single test loop
func executeSingleLoop(loopNum int, config TestConfig, outputFile string) error {
	// Create the prompt using the specified pattern
	prompt := fmt.Sprintf("use the %s agent and ask it to say 'hello', return what you told the agent, and just its response to you asking it to say 'hello'", config.AgentName)

	fmt.Printf("Loop %d: Executing claude with agent: %s\n", loopNum, config.AgentName)
	fmt.Printf("Loop %d: Prompt: %s\n\n", loopNum, prompt)

	// Execute claude with the specified flags and prompt
	claudeCmd := exec.Command("claude", "-p", "--permission-mode", "acceptEdits", prompt)

	// Capture output in buffers
	var stdout, stderr bytes.Buffer
	claudeCmd.Stdout = &stdout
	claudeCmd.Stderr = &stderr
	// Don't set claudeCmd.Stdin to avoid interactive prompts

	// Record start time
	loopStartTime := time.Now()
	fmt.Printf("Loop %d: Starting execution at: %s\n", loopNum, loopStartTime.Format("2006-01-02 15:04:05"))

	// Run the command
	err := claudeCmd.Run()

	// Record end time
	loopEndTime := time.Now()

	// Display output to console
	if stdout.Len() > 0 {
		fmt.Printf("Loop %d output:\n%s\n", loopNum, stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Fprintf(os.Stderr, "Loop %d stderr:\n%s\n", loopNum, stderr.String())
	}

	// Log the interaction
	logEntry := fmt.Sprintf("=== Loop %d/%d - %s ===\n", loopNum, config.Loops, loopEndTime.UTC().Format("2006-01-02 15:04:05 UTC"))
	logEntry += fmt.Sprintf("Prompt: %s\n", prompt)
	logEntry += fmt.Sprintf("Response:\n%s\n", strings.TrimSpace(stdout.String()))
	if stderr.Len() > 0 {
		logEntry += fmt.Sprintf("Errors:\n%s\n", strings.TrimSpace(stderr.String()))
	}
	logEntry += fmt.Sprintf("Execution time: %v\n", loopEndTime.Sub(loopStartTime))
	logEntry += "---\n\n"

	// Append to log file with thread-safe logging
	if logErr := appendToLogThreadSafe(outputFile, logEntry); logErr != nil {
		log.Printf("Error writing to log file: %v", logErr)
	}

	if err != nil {
		return fmt.Errorf("claude execution failed: %v", err)
	}

	fmt.Printf("Loop %d: Execution completed at: %s\n", loopNum, loopEndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Loop %d: Total execution time: %v\n", loopNum, loopEndTime.Sub(loopStartTime))

	return nil
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

// appendToLogThreadSafe provides thread-safe logging for parallel execution
func appendToLogThreadSafe(filename, content string) error {
	logMutex.Lock()
	defer logMutex.Unlock()
	return appendToLog(filename, content)
}
