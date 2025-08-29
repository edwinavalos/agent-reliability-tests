# Implementation Plan: Parallel Execution Flag

## Overview
Add a `--parallel` flag to the reliability test runner that enables concurrent execution of Claude agent calls instead of sequential execution.

## Current Implementation Analysis
- **Entry Point**: `cmd/reliability/main.go` uses Cobra CLI framework
- **Core Logic**: `pkg/reliability/runner.go` contains `RunReliabilityTest()` function
- **Sequential Flow**: Currently runs loops 1 to N sequentially with 1-second delays between iterations
- **Output**: Each loop writes to a shared timestamped log file using `appendToLog()`

## Proposed Changes

### 1. CLI Flag Addition
**File**: `cmd/reliability/main.go`
- Add new boolean flag: `--parallel` (default: false)
- Update `TestConfig` struct to include `Parallel bool` field
- Pass parallel flag to reliability package

### 2. Core Logic Modifications  
**File**: `pkg/reliability/runner.go`

#### TestConfig Struct Update
```go
type TestConfig struct {
    AgentName string
    Loops     int
    Filename  string
    Parallel  bool  // NEW
}
```

#### Parallel Execution Implementation
- Create new function: `runParallelLoops()` for concurrent execution
- Keep existing `runSequentialLoops()` for backward compatibility
- Use Go goroutines with `sync.WaitGroup` for synchronization
- Implement thread-safe logging with mutex protection

#### Key Changes:
1. **Goroutine Management**: Launch N goroutines (one per loop)
2. **Synchronization**: Use `sync.WaitGroup` to wait for all goroutines to complete
3. **Thread-Safe Logging**: Add mutex protection to `appendToLog()` function
4. **Error Handling**: Collect errors from all goroutines and report them
5. **Timing**: Remove artificial delays in parallel mode, maintain overall timing metrics

### 3. Thread-Safe Logging
- Add `sync.Mutex` to protect log file writes
- Ensure log entries from concurrent executions don't interleave
- Maintain chronological order where possible

## Implementation Steps

1. **Update CLI Interface**
   - Add `--parallel` flag to cobra command
   - Update TestConfig struct

2. **Refactor Core Runner**
   - Extract current loop logic into `runSequentialLoops()`
   - Implement new `runParallelLoops()` function
   - Add conditional execution based on parallel flag

3. **Thread-Safe Logging**
   - Add mutex protection to log operations
   - Test concurrent write safety

4. **Testing & Validation**
   - Test both sequential and parallel modes
   - Verify log file integrity with concurrent writes
   - Ensure error handling works correctly in both modes

## Benefits
- **Performance**: Significant speedup for multiple agent calls
- **Backward Compatibility**: Default behavior unchanged
- **Optional**: Users opt-in with explicit flag
- **Minimal Intrusion**: Core API and output format remain the same

## Risks & Considerations
- **Resource Usage**: Parallel execution may consume more system resources
- **Rate Limiting**: Claude API may have rate limits that could cause issues
- **Log Ordering**: Log entries may not be in strict execution order in parallel mode
- **Error Debugging**: Parallel errors may be harder to debug

## Success Criteria
- `--parallel` flag successfully launches concurrent agent calls
- Log file maintains integrity with no corrupted entries
- Overall execution time significantly reduced for multiple loops
- Sequential mode remains unchanged and default