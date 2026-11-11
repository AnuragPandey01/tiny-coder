package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anuragpandey01/tiny-coder/internal/agent/events"
)

type streamAgent struct {
	stream chan events.Event
	calls  int
}

func (a *streamAgent) Do(string) <-chan events.Event {
	a.calls++
	return a.stream
}

func runStreamCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("stream listener was not scheduled")
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		return msg
	case <-time.After(time.Second):
		t.Fatal("stream listener blocked")
		return nil
	}
}

func TestAgentStreamContinuesThroughToolsAndErrors(t *testing.T) {
	for _, explicitFinish := range []bool{false, true} {
		a := &streamAgent{stream: make(chan events.Event, 8)}
		for _, event := range []events.Event{
			events.TextDeltaMsg{Text: "before tool "},
			events.ToolStartedMsg{Name: "run_command", CallID: "1"},
			events.ToolFinishedMsg{Name: "run_command", CallID: "1"},
			events.TextDeltaMsg{Text: "after tool"},
			events.Error{Error: "stream failed"},
		} {
			a.stream <- event
		}
		if explicitFinish {
			a.stream <- events.TurnFinishedMsg{}
		}
		close(a.stream)
		app := *NewApp(a)
		model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
		app = model.(App)
		app.textInputView.SetValue("hello")
		model, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		app = model.(App)

		// A second prompt must not replace the active stream.
		app.textInputView.SetValue("queued draft")
		model, ignoredCmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		app = model.(App)
		if ignoredCmd != nil || a.calls != 1 || app.textInputView.Value() != "queued draft" {
			t.Fatal("submission during a turn replaced the stream or cleared the draft")
		}

		for i := 0; i < 6; i++ {
			msg := runStreamCmd(t, cmd)
			model, cmd = app.Update(msg)
			app = model.(App)
		}
		if app.activeAgentResponseChannel != nil {
			t.Fatal("turn did not release the active channel")
		}
		view := app.chatView.View()
		if !strings.Contains(view, "before tool after tool") || !strings.Contains(view, "stream failed") || strings.Contains(view, "▍") {
			t.Fatalf("incorrect completed view: %q", view)
		}
	}
}

func TestNilAgentStreamDoesNotScheduleReceive(t *testing.T) {
	if handleAgentEventStream(nil) != nil {
		t.Fatal("nil stream should not schedule a blocking receive")
	}
}
