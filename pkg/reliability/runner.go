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

type ExecutionMode int

const (
	Sequential ExecutionMode = iota
	Parallel
	Queue
)

type TestConfig struct {
	AgentName string
	Loops     int
	Filename  string
	Parallel  bool
	BatchSize int
	Queue     int
}

type TestResult struct {
	OutputFile string
	Duration   time.Duration
}

// GetExecutionMode determines the execution mode based on config flags
func (c TestConfig) GetExecutionMode() ExecutionMode {
	if c.Queue > 0 {
		return Queue
	} else if c.Parallel {
		return Parallel
	}
	return Sequential
}

var logMutex sync.Mutex

// RunReliabilityTest executes the reliability test with the given configuration
func RunReliabilityTest(config TestConfig) (*TestResult, error) {
	// Generate timestamped filename
	timestamp := time.Now().Unix()
	outputFile := fmt.Sprintf("%s_%d.log", config.Filename, timestamp)

	execMode := config.GetExecutionMode()
	var mode string
	switch execMode {
	case Queue:
		mode = fmt.Sprintf("queue (workers: %d)", config.Queue)
	case Parallel:
		mode = "parallel"
	default:
		mode = "sequential"
	}
	fmt.Printf("Running %d loop(s) with agent: %s (%s mode)\n", config.Loops, config.AgentName, mode)
	fmt.Printf("Output file: %s\n", outputFile)

	startTime := time.Now()

	// Use unified execution method for both serial and parallel
	return runLoops(config, outputFile, startTime)
}


// runLoops executes loops based on the configured execution mode
func runLoops(config TestConfig, outputFile string, startTime time.Time) (*TestResult, error) {
	execMode := config.GetExecutionMode()
	
	switch execMode {
	case Queue:
		return runLoopsQueue(config, outputFile, startTime)
	case Parallel:
		return runLoopsParallel(config, outputFile, startTime)
	default:
		return runLoopsSequential(config, outputFile, startTime)
	}
}

// runLoopsQueue executes loops using a worker queue system
func runLoopsQueue(config TestConfig, outputFile string, startTime time.Time) (*TestResult, error) {
	totalLoops := config.Loops
	workerCount := config.Queue
	
	fmt.Printf("\n=== Starting %d loops with %d workers ===\n", totalLoops, workerCount)
	
	// Create channels
	workQueue := make(chan int, totalLoops)
	errorChan := make(chan error, totalLoops)
	
	// Load the queue with all loop numbers
	for i := 1; i <= totalLoops; i++ {
		workQueue <- i
	}
	close(workQueue) // No more work will be added
	
	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			fmt.Printf("Worker %d started\n", workerID)
			
			for loopNum := range workQueue {
				fmt.Printf("Worker %d processing loop %d\n", workerID, loopNum)
				if err := executeLoop(loopNum, config, outputFile); err != nil {
					errorChan <- fmt.Errorf("worker %d, loop %d: %v", workerID, loopNum, err)
				}
				fmt.Printf("Worker %d completed loop %d\n", workerID, loopNum)
			}
			
			fmt.Printf("Worker %d finished\n", workerID)
		}(w)
	}
	
	// Wait for all workers to complete
	wg.Wait()
	close(errorChan)
	
	// Collect any errors
	for err := range errorChan {
		log.Printf("Execution error: %v", err)
	}
	
	totalDuration := time.Since(startTime)
	fmt.Printf("\n=== All %d loops completed with %d workers ===\n", totalLoops, workerCount)
	
	return &TestResult{
		OutputFile: outputFile,
		Duration:   totalDuration,
	}, nil
}

// runLoopsParallel executes loops in parallel batches (existing implementation)
func runLoopsParallel(config TestConfig, outputFile string, startTime time.Time) (*TestResult, error) {
	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 5 // Default batch size for parallel
	}

	totalLoops := config.Loops
	fmt.Printf("\n=== Starting %d loops in batches of %d ===\n", totalLoops, batchSize)

	errorChan := make(chan error, totalLoops)
	loopIndex := 1

	for loopIndex <= totalLoops {
		// Determine the size of the current batch
		currentBatchSize := batchSize
		if loopIndex+batchSize-1 > totalLoops {
			currentBatchSize = totalLoops - loopIndex + 1
		}

		fmt.Printf("\n--- Starting batch: loops %d-%d ---\n", loopIndex, loopIndex+currentBatchSize-1)

		var wg sync.WaitGroup

		// Start the current batch
		for i := 0; i < currentBatchSize; i++ {
			currentLoop := loopIndex + i
			wg.Add(1)
			go func(loopNum int) {
				defer wg.Done()
				if err := executeLoop(loopNum, config, outputFile); err != nil {
					errorChan <- fmt.Errorf("loop %d: %v", loopNum, err)
				}
			}(currentLoop)
		}

		// Wait for current batch to complete
		wg.Wait()
		
		fmt.Printf("--- Batch completed: loops %d-%d ---\n", loopIndex, loopIndex+currentBatchSize-1)

		// Move to next batch
		loopIndex += currentBatchSize
	}

	close(errorChan)

	// Collect any errors
	for err := range errorChan {
		log.Printf("Execution error: %v", err)
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\n=== All %d loops completed in batches ===\n", totalLoops)

	return &TestResult{
		OutputFile: outputFile,
		Duration:   totalDuration,
	}, nil
}

// runLoopsSequential executes loops one after another (existing implementation)
func runLoopsSequential(config TestConfig, outputFile string, startTime time.Time) (*TestResult, error) {
	totalLoops := config.Loops
	fmt.Printf("\n=== Starting %d loops sequentially ===\n", totalLoops)

	errorChan := make(chan error, totalLoops)

	for loopNum := 1; loopNum <= totalLoops; loopNum++ {
		if err := executeLoop(loopNum, config, outputFile); err != nil {
			errorChan <- fmt.Errorf("loop %d: %v", loopNum, err)
		}
		
		if loopNum < totalLoops {
			// Add delay between sequential loops (except after the last one)
			fmt.Printf("Waiting before next iteration...\n")
			time.Sleep(1 * time.Second)
		}
	}

	close(errorChan)

	// Collect any errors
	for err := range errorChan {
		log.Printf("Execution error: %v", err)
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\n=== All %d loops completed ===\n", totalLoops)

	return &TestResult{
		OutputFile: outputFile,
		Duration:   totalDuration,
	}, nil
}

// executeLoop runs a single test loop
func executeLoop(loopNum int, config TestConfig, outputFile string) error {
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
