# Claude Agent Tool Interface

This Go application provides a CLI interface for interacting with Claude AI using Anthropic's Go SDK. It allows users to have a conversation with Claude while also enabling Claude to use various tools to interact with the local file system.

## Overview

The application creates a chat interface between the user and Claude's 3.7 Sonnet model. Claude has access to a set of tools that allow it to interact with the file system, making it possible to work with files and directories during the conversation.

## Tools Available

The application provides Claude with the following tools:

### 1. read_file

Reads the contents of a file at a specified path.

**Parameters:**

- `path`: The relative path to a file in the working directory.

**Example:**

```
tool: read_file({"path":"main.go"})
```

### 2. list_files

Lists all files and directories at a given path. If no path is provided, it lists files in the current directory.

**Parameters:**

- `path` (optional): Relative path to list files from. Defaults to current directory if not provided.

**Example:**

```
tool: list_files({})
tool: list_files({"path":"src/"})
```

### 3. edit_file

Makes edits to a text file by replacing specified text with new text. If the file doesn't exist, it can create a new file.

**Parameters:**

- `path`: The path to the file
- `old_str`: Text to search for - must match exactly
- `new_str`: Text to replace old_str with

**Example:**

```
tool: edit_file({"path":"hello.txt","old_str":"Hello","new_str":"Hello, World!"})
```

To create a new file:

```
tool: edit_file({"path":"new_file.txt","old_str":"","new_str":"This is a new file content"})
```

### 4. make_dir

Creates a new directory at the specified path. If parent directories don't exist, they will be created as well.

**Parameters:**

- `path`: The relative path of the directory to create

### 5. delete_dir

Deletes a directory and all its contents recursively. Use with caution as this operation cannot be undone.

**Parameters:**

- `path`: The relative path of the directory to delete

### 6. git_status

Shows the working tree status, including staged, unstaged and untracked files.

### 7. git_add

Stage changes for commit. **Note**: This tool respects the `auto_commit` configuration setting.

**Parameters:**

- `path`: Path of file(s) to stage. Use '.' for all files

### 8. git_commit

Commit staged changes. **Note**: This tool respects the `auto_commit` configuration setting.

**Parameters:**

- `message`: Commit message

### 9. git_push

Push commits to remote repository.

### 10. git_pull

Pull changes from remote repository.

## Installation and Setup

### Prerequisites

1. Go programming language (Go 1.18 or later recommended)
2. Anthropic API key for accessing Claude AI

### Setup

1. Clone this repository to your local machine
2. Navigate to the project directory
3. Set up your Anthropic API key as an environment variable:

   ```
   export ANTHROPIC_API_KEY=your_api_key_here
   ```

## Configuration

The agent can be configured using a `config.yml` file in the project directory. If no configuration file exists, the agent will use default settings.

### Git Auto-Commit Configuration

You can control whether Claude can use git staging and commit tools:

```yaml
# config.yml
git:
  auto_commit: true # Set to false to disable git_add and git_commit tools
```

**When `auto_commit: true` (default)**:

- Claude can use `git_add` to stage files
- Claude can use `git_commit` to create commits
- Claude has full control over git operations

**When `auto_commit: false`**:

- `git_add` and `git_commit` tools will return error messages
- Claude cannot stage files or create commits
- `git_status`, `git_push`, and `git_pull` remain available
- Useful for environments where you want manual control over commits

## Running the AI Agent

To run the AI agent, use one of the following methods:

### Method 1: Using Go Run

```bash
go run main.go
```

### Method 2: Build and Execute

```bash
# Build the executable
go build -o claude-agent

# Run the executable
./claude-agent
```

## AGENT.md file

For Claude better understanding of the project create an AGENT.md file with information about your project for example:

```bash
# Agent Guide

## Build/Test/Lint Commands
- **Start development**: `npm run dev` (frontend: http://localhost:5173, backend: http://localhost:3001)
- **Build**: `npm run build` (all), `npm run build:frontend`, `npm run build:backend`
- **Test**: `npm run test` (all), `npm test -w packages/frontend`, `npm test -w packages/backend`
- **Test single file**: Frontend: `npm run test -w packages/frontend -- src/path/to/file.test.tsx`
- **Lint**: `npm run lint` (all), `npm run lint -w packages/frontend`

## Code Style Guidelines
- **TypeScript**: Strong typing required, avoid `any` when possible
- **Imports**: Use absolute imports with `@` alias in frontend (e.g., `@/components`)
- **Error handling**: Use try/catch blocks for async operations
- **Components**: React functional components with hooks, avoid class components
- **Backend**: Express routes in separate files, Mongoose for MongoDB
- **Testing**: Frontend uses Vitest with React Testing Library, backend uses Jest
- **Formatting**: ESLint and Prettier for consistent code style
- **Environment variables**: Use `.env` files (copy from `.env.example`)

## Codebase Structure
- `packages/frontend`: React/TypeScript (Vite) with Tailwind CSS
- `packages/backend`: Node.js/Express/TypeScript API with MongoDB

```

## Usage

Once the application is running, you can chat with Claude in your terminal. Claude can use the above tools to help with file operations.

1. Start a conversation
2. Ask Claude to perform file operations using the tools
3. Claude will use the tools to process your requests
4. Use 'ctrl-c' to quit the application

The interface highlights different parts of the conversation:

- **You**: Your messages
- **Claude**: Claude's responses
- **tool**: When Claude is using a tool

## Dependencies

- github.com/anthropics/anthropic-sdk-go
- github.com/invopop/jsonschema
- github.com/go-git/go-git/v5
- gopkg.in/yaml.v3

## Recent Updates

### Configuration System

- Added `config.yml` support for controlling agent behavior
- Implemented git auto-commit configuration (`git.auto_commit`)
- Users can now disable git staging and commit operations when needed
- Configuration defaults to allowing git operations if no config file exists

### File System Management Enhancements

- Added support for directory creation with the `make_dir` tool
- Added support for directory deletion with the `delete_dir` tool
- Improved file handling capabilities in the AI agent
- Enhanced error messaging for file operations

### Git Integration

- Added comprehensive git tools: `git_status`, `git_add`, `git_commit`, `git_push`, `git_pull`
- Git operations respect configuration settings for better control
- Commits are created with "AI Agent" as the author

These updates allow the AI agent to perform comprehensive file system and git operations while giving users control over potentially destructive actions like automatic commits.

