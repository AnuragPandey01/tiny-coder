package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/anuragpandey01/tiny-coder/internal/agent/events"
	"github.com/anuragpandey01/tiny-coder/internal/agent/tools"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// OpenAI runs an agent loop against an OpenAI-compatible Responses API.
type OpenAI struct {
	client  openai.Client
	model   string
	mu      sync.Mutex
	history []responses.ResponseInputItemUnionParam
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

func (a *OpenAI) Do(prompt string) <-chan events.Event {
	eventStream := make(chan events.Event)
	go func() {
		defer close(eventStream)
		defer func() { eventStream <- events.TurnFinished{} }()
		a.mu.Lock()
		defer a.mu.Unlock()
		slog.Debug("agent is thinking...", "model", a.model, "prompt_length", len(prompt))

		ctx := context.Background()
		a.history = append(a.history,
			responses.ResponseInputItemParamOfMessage(prompt, responses.EasyInputMessageRoleUser),
		)

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
				Input: responses.ResponseNewParamsInputUnion{OfInputItemList: a.history},
				Tools: toolsDef,
			}

			stream := a.client.Responses.NewStreaming(ctx, params, option.WithJSONSet("think", false))

			// a map to collect all tool calls
			toolCalls := make(map[int64]*responses.ResponseFunctionToolCall)
			var responseText strings.Builder

			for stream.Next() {
				e := stream.Current()
				switch e.Type {
				case "response.output_text.delta":
					eventStream <- events.TextDelta{Text: e.Delta}
					responseText.WriteString(e.Delta)
				case "response.output_item.done":
					if e.Item.Type == "function_call" {
						fc := e.Item.AsFunctionCall()
						toolCalls[e.OutputIndex] = &fc
					}
				}
			}

			if err := stream.Err(); err != nil {
				slog.Error("response stream failed", "model", a.model, "error", err)
				eventStream <- events.Error{Error: err.Error()}
				return
			}
			_ = stream.Close()
			if responseText.String() != "" {
				a.history = append(a.history,
					responses.ResponseInputItemParamOfMessage(responseText.String(), responses.EasyInputMessageRoleAssistant),
				)
			}

			// if no tool calls, agent already gave the final answer, so return
			if len(toolCalls) == 0 {
				slog.Debug("agent finished answering", "model", a.model)
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
				slog.Info("tool call received", "tool_name", tc.Name, "call_id", tc.CallID)

				//get the actual Tool interface implementing struct that owns this name
				tool, ok := toolsByName[tc.Name]
				if !ok {
					slog.Warn("unknown tool", "tool_name", tc.Name)
					eventStream <- events.Error{Error: fmt.Sprintf("unknown tool: %s", tc.Name)}
					a.appendToolResult(tc, fmt.Sprintf(`{"error":"unknown tool: %s"}`, tc.Name))
					continue
				}

				eventStream <- events.ToolStarted{Name: tool.Name(), CallID: tc.CallID}
				slog.Debug("Tool Call start", "tool_name", tc.Name, "call_id", tc.CallID, "args", tc.Arguments)

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
				slog.Debug("Tool Call completed", "tool_name", tc.Name, "call_id", tc.CallID, "failed", err != nil)
				eventStream <- events.ToolFinished{Name: tc.Name, CallID: tc.CallID, Result: res, Error: errorString(err)}

				a.appendToolResult(tc, outputStr)
			}
		}
	}()
	return eventStream
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (a *OpenAI) appendToolResult(tc *responses.ResponseFunctionToolCall, output string) {
	a.history = append(a.history, responses.ResponseInputItemParamOfFunctionCall(tc.Arguments, tc.CallID, tc.Name))
	a.history = append(a.history, responses.ResponseInputItemUnionParam{
		OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
			CallID: openai.String(tc.CallID),
			Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{OfString: openai.String(output)},
		},
	})
}
