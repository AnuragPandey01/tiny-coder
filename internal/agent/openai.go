package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

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

func (a *OpenAI) Run(prompt string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)

		ctx := context.Background()
		inputItems := []responses.ResponseInputItemUnionParam{
			responses.ResponseInputItemParamOfMessage(prompt, responses.EasyInputMessageRoleUser),
		}

		//all tools available to the agent
		availableTools := []tools.Tool{tools.RunCommandTool{}}

		toolsByName := make(map[string]tools.Tool, len(availableTools))
		toolsDef := make([]responses.ToolUnionParam, 0, len(availableTools))
		for _, tool := range availableTools {
			toolsByName[tool.Name()] = tool
			toolsDef = append(toolsDef, tool.Definition())
		}

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

				//get the actual Tool interface implementing struct that owns this name
				tool, ok := toolsByName[tc.Name]
				if !ok {
					// todo: consider logging this or/and sending this mistake to the agent
					slog.Warn("unknown tool", "tool_name", tc.Name)
					ch <- fmt.Sprintf("[unknown tool: %s]\n", tc.Name)
					continue
				}

				ch <- fmt.Sprintf("[running tool: %s...]\n\n", tool.Name())

				res, err := tool.Execute(ctx, tc.Arguments)
				var outputStr string
				if err != nil {
					slog.Error("tool execution failed", "tool_name", tc.Name, "error", err)
					outBytes, _ := json.Marshal(map[string]any{"error": err.Error(), "result": res})
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
