# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go application that provides a CLI interface for interacting with Claude AI using Anthropic's Go SDK. The application creates a conversational agent that can use tools to interact with the local file system and Git operations.

## Core Architecture

- **Main Entry Point**: `main.go` - Contains the entire application logic including the agent, tool definitions, and tool implementations
- **Agent System**: The `Agent` struct manages conversation flow and tool execution
- **Tool System**: Each tool has a `ToolDefinition` with JSON schema validation and corresponding implementation functions

## Available Tools

The agent provides Claude with these file system and Git tools:
- `read_file` - Read file contents
- `list_files` - List directory contents
- `edit_file` - Create or modify files with find/replace
- `make_dir` - Create directories
- `delete_dir` - Delete directories recursively  
- `git_status` - Show Git working tree status
- `git_add` - Stage files for commit
- `git_commit` - Create commits (uses "AI Agent" as author)
- `git_push` - Push to remote repository
- `git_pull` - Pull from remote repository

## Development Commands

Build and run the application:
```bash
go run main.go
```

Build executable:
```bash
go build -o claude-agent
./claude-agent
```

## Dependencies

- `github.com/anthropics/anthropic-sdk-go` - Anthropic API client
- `github.com/invopop/jsonschema` - JSON schema generation for tool parameters
- `github.com/go-git/go-git/v5` - Git operations

## Configuration

The agent can be configured via `config.yml`:

```yaml
git:
  auto_commit: true  # Set to false to disable git_add and git_commit tools
```

If no config file exists, git auto-commit defaults to `true`.

## Key Implementation Details

- Uses Claude 3.7 Sonnet Latest model
- Tool schemas are auto-generated from Go structs using JSON schema reflection
- File operations use relative paths from the current working directory
- Git operations assume the current directory is a Git repository
- Git tools (git_add, git_commit) respect the `auto_commit` configuration setting
- Error handling returns user-friendly messages through the tool result system