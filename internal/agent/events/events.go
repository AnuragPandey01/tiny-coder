// Package events defines updates emitted by an agent while it handles a turn.
package events

// Event is a UI-neutral update emitted during an agent turn.
type Event any

type TextDeltaMsg struct {
	Text string
}

type ToolStartedMsg struct {
	Name      string
	CallID    string
	Arguments string
}

type ToolFinishedMsg struct {
	Name   string
	CallID string
	Result any
	Error  string
}

type Error struct {
	Error string
}

type TurnFinishedMsg struct{}
