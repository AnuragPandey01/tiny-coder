package agent

import "github.com/anuragpandey01/tiny-coder/internal/agent/events"

// Agent is the interface for an agent that takes a prompt and returns a channel of events
type Agent interface {
	Do(prompt string) <-chan events.Event
}
