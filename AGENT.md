# Agent Guide

## Build/Test/Lint Commands
- **Run application**: `go run main.go`
- **Build executable**: `go build -o claude-agent` then `./claude-agent`
- **Get dependencies**: `go mod tidy`
- **Format code**: `go fmt ./...`
- **Test**: `go test ./...`

## Code Style Guidelines
- **Go conventions**: Follow standard Go formatting and naming conventions
- **Error handling**: Always handle errors appropriately, return meaningful error messages
- **JSON tags**: Use proper JSON schema descriptions for tool input structs
- **Tool design**: Keep tool functions focused and single-purpose
- **Configuration**: Use YAML for configuration files with clear documentation

## Codebase Structure
- `main.go`: Main application entry point with Agent orchestration and all tool definitions
- `config.yml`: Configuration file controlling agent behavior (git auto-commit, context files)
- `go.mod`: Go module dependencies including Anthropic SDK and git libraries

## Development Notes
- The application uses Anthropic's Claude 3.7 Sonnet model
- Tools are defined with JSON schemas for input validation
- Git operations respect the `auto_commit` configuration setting
- Context files can be configured to provide additional project knowledge to Claude
- All file operations use relative paths from the working directory

## Tool Architecture
Each tool implements:
1. Name and description for Claude
2. JSON schema for input validation using reflection
3. Function that executes the tool logic with proper error handling

The tool system is extensible - new tools can be added by implementing the ToolDefinition interface.