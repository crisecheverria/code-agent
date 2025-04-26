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