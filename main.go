package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <agent_name>")
		fmt.Println("Example: go run main.go general-purpose")
		os.Exit(1)
	}

	agentName := os.Args[1]
	
	// Create the prompt using the specified pattern
	prompt := fmt.Sprintf("use the %s and tell it to say 'hello', and have it append its response with basic metadata, the timestamp of the metadata should be to the minute to 'edwin.txt'", agentName)
	
	fmt.Printf("Executing claude with agent: %s\n", agentName)
	fmt.Printf("Prompt: %s\n\n", prompt)
	
	// Execute claude with the specified flags and prompt
	cmd := exec.Command("claude", "-p", "--permission-mode", "acceptEdits", prompt)
	
	// Set up output handling - don't connect stdin to avoid hanging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Don't set cmd.Stdin to avoid interactive prompts
	
	// Record start time
	startTime := time.Now()
	fmt.Printf("Starting execution at: %s\n", startTime.Format("2006-01-02 15:04:05"))
	
	// Run the command
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error executing claude: %v", err)
	}
	
	// Record end time
	endTime := time.Now()
	fmt.Printf("\nExecution completed at: %s\n", endTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total execution time: %v\n", endTime.Sub(startTime))
}