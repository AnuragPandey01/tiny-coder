package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

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
