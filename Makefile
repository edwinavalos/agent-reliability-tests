# Makefile for Claude Agent Reliability Tests

.PHONY: test analyze clean help build

# Default target
help:
	@echo "Available targets:"
	@echo "  test      - Run reliability test with general-purpose agent (5 loops)"
	@echo "  analyze   - Analyze the most recent log file"  
	@echo "  build     - Build binaries into ./build directory"
	@echo "  clean     - Delete all .log files and build directory"
	@echo "  help      - Show this help message"

# Run reliability test
test:
	@echo "Running reliability test..."
	go run cmd/reliability/main.go general-purpose --loops 5

# Run parallel reliability test
test-parallel:
	@echo "Running reliability test..."
	go run cmd/reliability/main.go general-purpose --loops 5 --parallel

# Analyze most recent log file  
analyze:
	@echo "Finding most recent log file..."
	@LATEST_LOG=$$(ls -t *_*.log 2>/dev/null | head -n1); \
	if [ -z "$$LATEST_LOG" ]; then \
		echo "No log files found in current directory"; \
		exit 1; \
	else \
		echo "Analyzing: $$LATEST_LOG"; \
		go run cmd/analyze/main.go "$$LATEST_LOG"; \
	fi

# Clean up log files and build directory
clean:
	@echo "Cleaning up..."
	@LOG_COUNT=$$(ls *.log 2>/dev/null | wc -l); \
	if [ "$$LOG_COUNT" -eq 0 ]; then \
		echo "No log files to clean"; \
	else \
		echo "Removing $$LOG_COUNT log file(s)..."; \
		rm -f *.log; \
		echo "Log files removed"; \
	fi
	@if [ -d build ]; then \
		echo "Removing build directory..."; \
		rm -rf build; \
		echo "Build directory removed"; \
	else \
		echo "No build directory to clean"; \
	fi

# Build binaries into ./build directory
build:
	@echo "Building binaries..."
	@mkdir -p build
	go build -o build/agent-reliability-tests cmd/reliability/main.go
	go build -o build/analyze cmd/analyze/main.go
	@echo "Built: build/agent-reliability-tests, build/analyze"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download