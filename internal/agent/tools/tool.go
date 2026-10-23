package tools

import (
	"context"

	"github.com/openai/openai-go/v3/responses"
)

// Tool describes a callable agent tool. Implementations own their argument
// schema, validation, and execution so the agent loop only needs to dispatch.
type Tool interface {
	Name() string
	Definition() responses.ToolUnionParam
	Execute(ctx context.Context, rawArguments string) (any, error)
}
