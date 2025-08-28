# ⚠️ WARNING: 100% AGENTIC CODEBASE ⚠️

This project is developed entirely by autonomous agents. **No human questions or support will be provided.**

---

# Claude Agent Reliability Tests

This project provides tools to run reliability tests on Claude agents and analyze their output.

## Reliability Test Runner
- **Location:** `cmd/reliability/main.go`
- **Purpose:** Runs a specified agent multiple times, saving results to a timestamped log file.
- **Usage:**
  ```
  go run cmd/reliability/main.go [agent_name] --loops N --filename base_name
  ```
  Or use the Makefile:
  ```
  make test
  ```
  (Runs with agent "general-purpose" for 5 loops.)

## Log Analyzer
- **Location:** `cmd/analyze/main.go`
- **Purpose:** Analyzes log files for response similarity, clustering, common patterns, and outliers.
- **Usage:**
  ```
  go run cmd/analyze/main.go [log_file] --verbose --output result.txt --debug
  ```
  Or use the Makefile:
  ```
  make analyze
  ```
  (Analyzes the most recent log file.)

## Other Makefile Targets
- `make clean` — Removes all `.log` files.
- `make build` — Builds binaries for both tools.

## Notes on Folders
- `hello-world` and `hello-world-again` contain earlier experiments to determine the best structure for prompt interaction and testing.

## Prompting
- The current implementation uses hard-coded, basic prompts for agent interaction.
- A future update will add support for templated prompts, allowing more flexible and customizable agent instructions.

## Next Steps
The next development step is to implement parallel execution of Claude agent runs for improved efficiency.
