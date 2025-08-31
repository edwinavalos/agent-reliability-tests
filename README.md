# ⚠️ WARNING: 100% AGENTIC CODEBASE ⚠️

This project is developed entirely by autonomous agents. **No human questions or support will be provided.**

---

# Claude Agent Reliability Tests

A comprehensive tool suite for running reliability tests on Claude agents with flexible execution modes, template-based prompts, and detailed analysis capabilities.

## 🚀 Quick Start

Build the tools:
```bash
make build
```

Run a basic reliability test:
```bash
./build/agent-reliability-tests general-purpose --loops 5
```

Analyze the results:
```bash
make analyze
```

## 🛠️ Reliability Test Runner

**Binary:** `./build/agent-reliability-tests`  
**Source:** `cmd/reliability/main.go`

### Execution Modes

The tool supports three execution modes:

1. **Queue Mode (Default)** - Uses worker threads for controlled execution
2. **Parallel Mode** - Executes tests in parallel batches
3. **Multi-Worker Queue** - Scales queue mode with multiple workers

### Usage Examples

```bash
# Basic test (queue mode with 1 worker)
./build/agent-reliability-tests general-purpose --loops 10

# Multi-worker queue mode
./build/agent-reliability-tests general-purpose --loops 20 --queue 3

# Parallel batch execution
./build/agent-reliability-tests general-purpose --loops 15 --parallel --batch 5

# Using custom templates
./build/agent-reliability-tests multi-agent-coordinator --prompt example_prompt_templates/coordination_plan.tmpl --loops 5
```

### Available Flags

- `--loops, -l` - Number of test iterations (default: 1)
- `--queue, -q` - Number of worker threads for queue mode (default: 1)
- `--parallel, -p` - Enable parallel batch execution
- `--batch` - Batch size for parallel mode (default: 5)
- `--prompt` - Path to Go template file for custom prompts
- `--filename, -f` - Base name for output log file (default: "chat")

## 📝 Template System

The tool includes a flexible file-based Go template system for customizing agent prompts.

### Template Variables
- `{{.SubAgentName}}` - The name of the agent being tested

### Example Templates

Located in `example_prompt_templates/`:
- `hello_world.tmpl` - Basic agent interaction (default behavior)
- `coordination_plan.tmpl` - Multi-agent coordination testing
- `code_review.tmpl` - Code review capability testing  
- `feature_implementation.tmpl` - Complex feature development testing

### Template Usage

```bash
# Use coordination template
./build/agent-reliability-tests multi-agent-coordinator \
  --prompt example_prompt_templates/coordination_plan.tmpl \
  --loops 10 --queue 2

# Create custom template
echo "Ask the {{.SubAgentName}} to explain quantum computing" > my_template.tmpl
./build/agent-reliability-tests python-pro --prompt my_template.tmpl --loops 5
```

## 📊 Log Analyzer

**Binary:** `./build/analyze`  
**Source:** `cmd/analyze/main.go`

Analyzes test logs to quantify response similarity, identify patterns, and detect outliers.

### Features
- Response similarity metrics
- Clustering analysis
- Pattern identification
- Outlier detection
- Reliability assessment

### Usage

```bash
# Analyze specific log file
./build/analyze chat_1234567890.log --verbose --output analysis.txt

# Debug mode (shows extracted responses)
./build/analyze chat_1234567890.log --debug

# Save detailed analysis
./build/analyze chat_1234567890.log --output detailed_analysis.txt --verbose
```

### Available Flags
- `--verbose, -v` - Enable detailed output including similarity matrix
- `--output, -o` - Save results to file
- `--debug, -d` - Show extracted responses for debugging

## 🎯 Makefile Targets

- `make build` - Build binaries into `./build/` directory
- `make test` - Run basic reliability test (general-purpose, 5 loops)
- `make test-parallel` - Run parallel reliability test
- `make exec` - Run multi-worker queue test (30 loops, 5 workers)
- `make analyze` - Analyze the most recent log file
- `make clean` - Remove log files and build directory
- `make deps` - Install Go dependencies
- `make help` - Show available targets

## 🏗️ Architecture

### Execution Modes
- **Queue Mode**: Default mode using worker goroutines with job queue
- **Parallel Mode**: Batch-based parallel execution with configurable batch size
- **Performance**: Template caching eliminates repeated file I/O and parsing

### Template System
- **File-based**: Go templates stored in `.tmpl` or `.template` files
- **Validation**: Input validation with clear error messages
- **Caching**: Templates parsed once and cached for performance
- **Flexible**: Support for complex template logic and formatting

## 📁 Project Structure

```
├── cmd/
│   ├── reliability/    # Test runner CLI
│   └── analyze/        # Log analyzer CLI
├── pkg/reliability/    # Core reliability testing logic
├── example_prompt_templates/  # Template examples and documentation
├── build/             # Compiled binaries (created by make build)
└── Makefile          # Build and test automation
```

## 🔧 Development

Install dependencies:
```bash
make deps
```

Build tools:
```bash
make build
```

Run tests:
```bash
make test
```
