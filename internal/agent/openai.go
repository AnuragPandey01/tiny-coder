package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/anuragpandey01/tiny-coder/internal/agent/tools"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// OpenAI runs an agent loop against an OpenAI-compatible Responses API.
type OpenAI struct {
	client openai.Client
	model  string
}

// NewOpenAI creates an agent backed by the local Ollama-compatible endpoint.
func NewOpenAI() *OpenAI {
	return &OpenAI{
		client: openai.NewClient(
			option.WithBaseURL("http://127.0.0.1:11434/v1/"),
			option.WithAPIKey("ollama"),
		),
		model: "gemma4:12b-mlx",
	}
}

type runCommandArgs struct {
	Dir  string   `json:"dir"`
	Name string   `json:"name"`
	Args []string `json:"args"`
}

func (a *OpenAI) Run(prompt string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)

		ctx := context.Background()
		inputItems := []responses.ResponseInputItemUnionParam{
			responses.ResponseInputItemParamOfMessage(prompt, responses.EasyInputMessageRoleUser),
		}
		toolsDef := []responses.ToolUnionParam{{
			OfFunction: &responses.FunctionToolParam{
				Name:        "run_command",
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
			},
		}}

		for {
			params := responses.ResponseNewParams{
				Model: a.model,
				Input: responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems},
				Tools: toolsDef,
			}

			stream := a.client.Responses.NewStreaming(ctx, params, option.WithJSONSet("think", false))

			// a map to collect all tool calls
			toolCalls := make(map[int64]*responses.ResponseFunctionToolCall)

			for stream.Next() {
				e := stream.Current()
				switch e.Type {
				case "response.output_text.delta":
					ch <- e.Delta
				case "response.output_item.done":
					if e.Item.Type == "function_call" {
						fc := e.Item.AsFunctionCall()
						toolCalls[e.OutputIndex] = &fc
					}
				}
			}

			if err := stream.Err(); err != nil {
				ch <- fmt.Sprintf("\n[stream error: %v]\n", err)
				return
			}
			_ = stream.Close()

			// if no tool calls, agent already gave the final answer, so return
			if len(toolCalls) == 0 {
				return
			}

			// sort all tool calls
			indices := make([]int64, 0, len(toolCalls))
			for idx := range toolCalls {
				indices = append(indices, idx)
			}
			slices.Sort(indices)

			for _, idx := range indices {
				tc := toolCalls[idx]
				if tc.Name != "run_command" {
					continue
				}
				var cmdArgs runCommandArgs
				if err := json.Unmarshal([]byte(tc.Arguments), &cmdArgs); err != nil {
					ch <- fmt.Sprintf("[tool parse error: %v]\n", err) //todo: error can be passed to the agent
					continue
				}
				dir := cmdArgs.Dir
				if dir == "" {
					dir = "."
				}
				cmdStr := cmdArgs.Name
				if len(cmdArgs.Args) > 0 {
					cmdStr += " " + strings.Join(cmdArgs.Args, " ")
				}
				ch <- fmt.Sprintf("[running: %s in %s...]\n\n", cmdStr, dir)

				res, err := tools.RunCommand(ctx, dir, cmdArgs.Name, cmdArgs.Args...)
				var outputStr string
				if err != nil {
					outBytes, _ := json.Marshal(map[string]any{"error": err.Error(), "stdout": res.Stdout, "stderr": res.Stderr, "exit_code": res.ExitCode})
					outputStr = string(outBytes)
				} else {
					outBytes, _ := json.Marshal(res)
					outputStr = string(outBytes)
				}
				inputItems = append(inputItems, responses.ResponseInputItemParamOfFunctionCall(tc.Arguments, tc.CallID, tc.Name))
				inputItems = append(inputItems, responses.ResponseInputItemUnionParam{
					OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
						CallID: openai.String(tc.CallID),
						Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{OfString: openai.String(outputStr)},
					},
				})
			}
		}
	}()
	return ch
}
