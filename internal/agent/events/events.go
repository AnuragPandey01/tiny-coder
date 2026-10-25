// Package events defines updates emitted by an agent while it handles a turn.
package events

// Event is a UI-neutral update emitted during an agent turn.
type Event any

type TextDelta struct {
	Text string
}

type ToolStarted struct {
	Name   string
	CallID string
}

type ToolFinished struct {
	Name   string
	CallID string
	Result any
	Error  string
}

type Error struct {
	Error string
}

type TurnFinished struct{}
