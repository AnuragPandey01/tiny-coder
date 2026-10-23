package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// RunCommandTool executes a program with separate arguments.
type RunCommandTool struct{}

func (RunCommandTool) Name() string { return "run_command" }

func (RunCommandTool) Definition() responses.ToolUnionParam {
	return responses.ToolUnionParam{OfFunction: &responses.FunctionToolParam{
		Name:        RunCommandTool{}.Name(),
		Description: openai.String("Execute a command in a specified working directory. Pass the executable and arguments separately; do not use shell syntax."),
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"dir":  map[string]any{"type": "string", "description": "Working directory in which to execute the command."},
				"name": map[string]any{"type": "string", "description": "Executable name, such as go, git, or ls."},
				"args": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Arguments passed to the executable as separate strings."},
			},
			"required": []string{"dir", "name", "args"},
		},
	}}
}

type runCommandArguments struct {
	Dir  string   `json:"dir"`
	Name string   `json:"name"`
	Args []string `json:"args"`
}

func (RunCommandTool) Execute(ctx context.Context, rawArguments string) (any, error) {
	var args runCommandArguments
	if err := json.Unmarshal([]byte(rawArguments), &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if args.Name == "" {
		return nil, fmt.Errorf("invalid arguments: name is required")
	}
	if args.Dir == "" {
		args.Dir = "."
	}
	return RunCommand(ctx, args.Dir, args.Name, args.Args...)
}

type CommandResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// RunCommand executes a command in a given directory and returns the result.
func RunCommand(
	ctx context.Context,
	dir string,
	name string,
	args ...string,
) (CommandResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err == nil {
		return result, nil
	}

	// The command ran but exited with a nonzero status.
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		result.ExitCode = exitErr.ExitCode()

		// Preserve cancellation and timeout information.
		if ctx.Err() != nil {
			return result, fmt.Errorf(
				"command canceled or timed out: %w", ctx.Err(),
			)
		}

		// Let the agent inspect the failure using ExitCode and Stderr.
		return result, nil
	}

	// The command could not start or encountered another execution error.
	result.ExitCode = -1
	return result, fmt.Errorf("failed to run command: %w", err)
}
