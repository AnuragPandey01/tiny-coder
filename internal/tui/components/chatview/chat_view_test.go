package chatview

import (
	"strings"
	"testing"
)

func newTestView() Model {
	m := NewChatView(nil)
	m.SetWidth(80)
	m.SetHeight(10)
	return m
}

func TestStreamingResponseLifecycle(t *testing.T) {
	m := newTestView()
	if strings.Contains(m.View(), "Agent:") {
		t.Fatal("idle view shows an agent response")
	}
	m, _ = m.Update(UserQueryMsg{Text: "hello"})
	if !strings.Contains(m.View(), "▍") {
		t.Fatal("waiting response should show the cursor")
	}
	for _, chunk := range []string{"Hello", " 世界", "\nnext line"} {
		m, _ = m.Update(ResponseDeltaMsg{Text: chunk})
	}
	if !strings.Contains(m.View(), "Hello 世界") || !strings.Contains(m.View(), "next line▍") {
		t.Fatalf("streamed text missing from view: %q", m.View())
	}
	m, _ = m.Update(ResponseFinishedMsg{})
	m, _ = m.Update(ResponseFinishedMsg{})
	if strings.Contains(m.View(), "▍") || len(m.items) != 2 {
		t.Fatalf("completion should commit exactly one response: %+v", m.items)
	}
	if m.items[1].Text != "Hello 世界\nnext line" {
		t.Fatalf("incorrect committed text: %q", m.items[1].Text)
	}
	m, _ = m.Update(UserQueryMsg{Text: "again"})
	m, _ = m.Update(ResponseDeltaMsg{Text: "second"})
	m, _ = m.Update(ResponseFinishedMsg{})
	if len(m.items) != 4 || m.items[3].Text != "second" {
		t.Fatalf("responses from separate turns mixed: %+v", m.items)
	}
}

func TestFailedAndEmptyResponses(t *testing.T) {
	for _, partial := range []string{"", "partial answer"} {
		t.Run(partial, func(t *testing.T) {
			m := newTestView()
			m, _ = m.Update(UserQueryMsg{Text: "hello"})
			if partial != "" {
				m, _ = m.Update(ResponseDeltaMsg{Text: partial})
			}
			m, _ = m.Update(ResponseErrorMsg{Text: "connection lost"})
			m, _ = m.Update(ResponseFinishedMsg{})
			if strings.Contains(m.View(), "▍") || !strings.Contains(m.View(), "connection lost") {
				t.Fatalf("failure not rendered correctly: %q", m.View())
			}
			wantItems := 2
			if partial != "" {
				wantItems++
				if m.items[1].Text != partial {
					t.Fatal("partial response was lost")
				}
			}
			if len(m.items) != wantItems {
				t.Fatalf("got %d transcript entries, want %d", len(m.items), wantItems)
			}
		})
	}
	m := newTestView()
	m, _ = m.Update(UserQueryMsg{Text: "hello"})
	m, _ = m.Update(ResponseFinishedMsg{})
	if len(m.items) != 1 || strings.Contains(m.View(), "Agent:") {
		t.Fatal("empty turn added an agent entry")
	}
}

func TestStreamingRespectsScrollPosition(t *testing.T) {
	m := newTestView()
	m.SetHeight(3)
	m, _ = m.Update(UserQueryMsg{Text: "hello"})
	m, _ = m.Update(ResponseDeltaMsg{Text: strings.Repeat("line\n", 10)})
	if !m.viewport.AtBottom() {
		t.Fatal("stream should follow the bottom initially")
	}
	m.viewport.GotoTop()
	m, _ = m.Update(ResponseDeltaMsg{Text: "more\n"})
	if m.viewport.YOffset() != 0 {
		t.Fatal("stream moved the viewport while reading earlier text")
	}
	m.viewport.GotoBottom()
	m, _ = m.Update(ResponseDeltaMsg{Text: "last\n"})
	if !m.viewport.AtBottom() {
		t.Fatal("stream should resume following at the bottom")
	}
}

func TestInitialHistoryIsRendered(t *testing.T) {
	m := NewChatView([]ChatItem{{ItemType: ItemTypeAgentResponse, Text: "previous answer"}})
	m.SetWidth(80)
	m.SetHeight(10)
	if !strings.Contains(m.View(), "previous answer") || strings.Contains(m.View(), "▍") {
		t.Fatalf("incorrect initial transcript: %q", m.View())
	}
}
