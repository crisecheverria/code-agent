package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/invopop/jsonschema"
)

func main() {
	client := anthropic.NewClient()

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	tools := []ToolDefinition{
		ReadFileDefinition,
		ListFilesDefinition,
		EditFileDefinition,
		MakeDirDefinition,
		DeleteDirDefinition,
		GitStatusDefinition,
		GitAddDefinition,
		GitCommitDefinition,
		GitPushDefinition,
		GitPullDefinition,
	}
	agent := NewAgent(&client, getUserMessage, tools)
	err := agent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), tools []ToolDefinition) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	tools          []ToolDefinition
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}
		message, err := a.runInterface(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			case "tool_use":
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
		conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
	}

	return nil
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDef ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	if !found {
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}
	return anthropic.NewToolResultBlock(id, response, false)
}

func (a *Agent) runInterface(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}
	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(1024),
		Messages:  conversation,
		Tools:     anthropicTools,
	})

	return message, err
}

type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path o a file in the working directory."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		panic(err)
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

var ListFilesDefinition = ToolDefinition{
	Name:        "list_files",
	Description: "List files and directories at a given path. If no path is provided, lists files in the current directory.",
	InputSchema: ListFilesInputSchema,
	Function:    ListFiles,
}

type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

var ListFilesInputSchema = GenerateSchema[ListFilesInput]()

func ListFiles(input json.RawMessage) (string, error) {
	listFilesInput := ListFilesInput{}
	err := json.Unmarshal(input, &listFilesInput)
	if err != nil {
		panic(err)
	}

	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath != "." {
			if info.IsDir() {
				files = append(files, relPath+"/")
			} else {
				files = append(files, relPath)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

var EditFileDefinition = ToolDefinition{
	Name: "edit_file",
	Description: `Make edits to a text file.
  Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other.
  If the file specified with path doesn't exist, it will be created.`,
	InputSchema: EditFileInputSchema,
	Function:    EditFile,
}

type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file"`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for - must match exactly and must only have one match exactly"`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with"`
}

var EditFileInputSchema = GenerateSchema[EditFileInput]()

func EditFile(input json.RawMessage) (string, error) {
	editFileInput := EditFileInput{}
	err := json.Unmarshal(input, &editFileInput)
	if err != nil {
		return "", err
	}

	if editFileInput.Path == "" || editFileInput.OldStr == editFileInput.NewStr {
		return "", fmt.Errorf("ivalid input parameters")
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsNotExist(err) && editFileInput.OldStr == "" {
			return createNewFile(editFileInput.Path, editFileInput.NewStr)
		}
		return "", err
	}

	oldContent := string(content)
	newContent := strings.Replace(oldContent, editFileInput.OldStr, editFileInput.NewStr, -1)

	if oldContent == newContent && editFileInput.OldStr != "" {
		return "", fmt.Errorf("old_str not found in file")
	}

	err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func createNewFile(filePath, content string) (string, error) {
	dir := path.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	return fmt.Sprintf("Successfully created file %s", filePath), nil
}

var MakeDirDefinition = ToolDefinition{
	Name:        "make_dir",
	Description: "Creates a new directory at the specified path. If parent directories don't exist, they will be created as well.",
	InputSchema: MakeDirInputSchema,
	Function:    MakeDir,
}

type MakeDirInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of the directory to create"`
}

var MakeDirInputSchema = GenerateSchema[MakeDirInput]()

func MakeDir(input json.RawMessage) (string, error) {
	makeDirInput := MakeDirInput{}
	err := json.Unmarshal(input, &makeDirInput)
	if err != nil {
		return "", err
	}

	if makeDirInput.Path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	err = os.MkdirAll(makeDirInput.Path, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return fmt.Sprintf("Successfully created directory %s", makeDirInput.Path), nil
}

var DeleteDirDefinition = ToolDefinition{
	Name:        "delete_dir",
	Description: "Deletes a directory and all its contents recursively. Use with caution as this operation cannot be undone.",
	InputSchema: DeleteDirInputSchema,
	Function:    DeleteDir,
}

type DeleteDirInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of the directory to delete"`
}

var DeleteDirInputSchema = GenerateSchema[DeleteDirInput]()

var GitStatusDefinition = ToolDefinition{
	Name:        "git_status",
	Description: "Show the working tree status, including staged, unstaged and untracked files",
	InputSchema: GenerateSchema[struct{}](),
	Function:    GitStatus,
}

type GitAddInput struct {
	Path string `json:"path" jsonschema_description:"Path of file(s) to stage. Use '.' for all files"`
}

var GitAddDefinition = ToolDefinition{
	Name:        "git_add",
	Description: "Stage changes for commit",
	InputSchema: GenerateSchema[GitAddInput](),
	Function:    GitAdd,
}

type GitCommitInput struct {
	Message string `json:"message" jsonschema_description:"Commit message"`
}

var GitCommitDefinition = ToolDefinition{
	Name:        "git_commit",
	Description: "Commit staged changes",
	InputSchema: GenerateSchema[GitCommitInput](),
	Function:    GitCommit,
}

var GitPushDefinition = ToolDefinition{
	Name:        "git_push",
	Description: "Push commits to remote repository",
	InputSchema: GenerateSchema[struct{}](),
	Function:    GitPush,
}

var GitPullDefinition = ToolDefinition{
	Name:        "git_pull",
	Description: "Pull changes from remote repository",
	InputSchema: GenerateSchema[struct{}](),
	Function:    GitPull,
}

func GitStatus(input json.RawMessage) (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	return status.String(), nil
}

func GitAdd(input json.RawMessage) (string, error) {
	var addInput GitAddInput
	if err := json.Unmarshal(input, &addInput); err != nil {
		return "", err
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.AddWithOptions(&git.AddOptions{
		Path: addInput.Path,
	})
	if err != nil {
		return "", fmt.Errorf("failed to add files: %w", err)
	}

	return fmt.Sprintf("Successfully staged changes for: %s", addInput.Path), nil
}

func GitCommit(input json.RawMessage) (string, error) {
	var commitInput GitCommitInput
	if err := json.Unmarshal(input, &commitInput); err != nil {
		return "", err
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	commit, err := worktree.Commit(commitInput.Message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "AI Agent",
			Email: "ai@agent.local",
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return fmt.Sprintf("Successfully created commit: %s", commit.String()), nil
}

func GitPush(input json.RawMessage) (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	err = repo.Push(&git.PushOptions{})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return "Everything up-to-date", nil
		}
		return "", fmt.Errorf("failed to push: %w", err)
	}

	return "Successfully pushed changes to remote", nil
}

func GitPull(input json.RawMessage) (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Pull(&git.PullOptions{})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return "Already up-to-date", nil
		}
		return "", fmt.Errorf("failed to pull: %w", err)
	}

	return "Successfully pulled changes from remote", nil
}

func DeleteDir(input json.RawMessage) (string, error) {
	deleteDirInput := DeleteDirInput{}
	err := json.Unmarshal(input, &deleteDirInput)
	if err != nil {
		return "", err
	}

	if deleteDirInput.Path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Check if the path exists
	_, err = os.Stat(deleteDirInput.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory does not exist: %s", deleteDirInput.Path)
		}
		return "", fmt.Errorf("error checking directory: %w", err)
	}

	// Remove the directory and all its contents
	err = os.RemoveAll(deleteDirInput.Path)
	if err != nil {
		return "", fmt.Errorf("failed to delete directory: %w", err)
	}

	return fmt.Sprintf("Successfully deleted directory %s and all its contents", deleteDirInput.Path), nil
}
