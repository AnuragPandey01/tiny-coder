package components

import (
	"strings"

	"github.com/anuragpandey01/tiny-coder/internal/tui/style"
	"github.com/anuragpandey01/tiny-coder/internal/tui/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ChatItemType int
type ToolStatus int

const (
	ItemTypeUserQuery ChatItemType = iota
	ItemTypeAgentResponse
	ItemTypeToolCall
)

const (
	ToolStatusPending ToolStatus = iota
	ToolStatusSuccess
	ToolStatusError
)

type ToolCallData struct {
	ToolCallID string
	Name       string
	Arguments  string
	Result     string
	Status     ToolStatus
}

type ChatItem struct {
	ItemType ChatItemType
	Text     string
	ToolCall *ToolCallData
}

// ChatView renders the conversation as a scrollable transcript. Text can
// contain tview color markup supplied by the caller.
type ChatView struct {
	*tview.Box
	transcript      *tview.TextView
	items           []ChatItem
	currentResponse string
	agentRunning    bool
}

func NewChatView(chat []ChatItem) *ChatView {
	c := &ChatView{
		Box: tview.NewBox().SetBorderPadding(1, 1, 2, 1),
		transcript: tview.NewTextView().
			SetDynamicColors(true).
			SetWrap(true).
			SetWordWrap(true).SetScrollable(true),
		items: append([]ChatItem(nil), chat...),
	}
	c.render()
	return c
}

func (c *ChatView) SetItems(items []ChatItem) {
	c.items = append(c.items[:0], items...)
	c.render()
}

func (c *ChatView) AppendChatItem(item ChatItem) {
	c.items = append(c.items, item)
	c.render()
}

// BeginAgentResponse starts a new live response for the current user turn.
func (c *ChatView) BeginAgentResponse() {
	if c.currentResponse != "" {
		c.items = append(c.items, ChatItem{ItemType: ItemTypeAgentResponse, Text: c.currentResponse})
	}
	c.currentResponse = ""
	c.agentRunning = true
	c.render()
}

// AppendAgentText adds a streamed text delta to the live response.
func (c *ChatView) AppendAgentText(text string) {
	c.currentResponse += text
	c.render()
}

// FinishAgentResponse commits the live response to chat history.
func (c *ChatView) FinishAgentResponse() {
	if c.currentResponse != "" {
		c.items = append(c.items, ChatItem{ItemType: ItemTypeAgentResponse, Text: c.currentResponse})
	}
	c.currentResponse = ""
	c.agentRunning = false
	c.render()
}

func (c *ChatView) IsAgentRunning() bool { return c.agentRunning }

// StartToolCall adds a pending tool call to the conversation.
func (c *ChatView) StartToolCall(id, name string, arg string) {
	c.items = append(c.items, ChatItem{
		ItemType: ItemTypeToolCall,
		ToolCall: &ToolCallData{ToolCallID: id, Name: name, Status: ToolStatusPending, Arguments: arg},
	})
	c.render()
}

// FinishToolCall records a tool's result and final status.
func (c *ChatView) FinishToolCall(id, result, err string) {
	for _, item := range c.items {
		if item.ItemType != ItemTypeToolCall || item.ToolCall == nil || item.ToolCall.ToolCallID != id {
			continue
		}
		item.ToolCall.Result = result
		item.ToolCall.Status = ToolStatusSuccess
		if err != "" {
			item.ToolCall.Result = err
			item.ToolCall.Status = ToolStatusError
		}
		break
	}
	c.render()
}

func (c *ChatView) Draw(screen tcell.Screen) {
	c.Box.DrawForSubclass(screen, c)
	x, y, width, height := c.GetInnerRect()
	c.transcript.SetRect(x, y, width, height)
	c.transcript.Draw(screen)
}

func (c *ChatView) render() {
	var b strings.Builder
	for _, item := range c.items {
		switch item.ItemType {
		case ItemTypeUserQuery:
			b.WriteString(utils.StyledString(style.UserColorTag, "> "+tview.Escape(item.Text)+"\n"))
		case ItemTypeAgentResponse:
			b.WriteString(utils.StyledString(style.AgentColorTag, tview.Escape(item.Text)+"\n\n"))
		case ItemTypeToolCall:
			if item.ToolCall == nil {
				continue
			}
			tool := item.ToolCall
			b.WriteString("[gray]tool: ")
			b.WriteString(tview.Escape(tool.Name) + " " + tool.Arguments)
			//switch tool.Status {
			//case ToolStatusPending:
			//	b.WriteString(" (running)")
			//case ToolStatusSuccess:
			//	b.WriteString(" (completed)")
			//case ToolStatusError:
			//	b.WriteString(" (failed)")
			//}
			//if tool.Arguments != "" {
			//	b.WriteString("\n")
			//	b.WriteString(tview.Escape(tool.Arguments))
			//}
			//if tool.Result != "" {
			//	b.WriteString("\n")
			//	b.WriteString(tview.Escape(tool.Result))
			//}
			b.WriteString("[-]\n\n")
		}
	}
	if c.agentRunning {
		b.WriteString(utils.StyledString(style.AgentColorTag, tview.Escape(c.currentResponse)+style.Cursor))
	}
	c.transcript.SetText(b.String()).ScrollToEnd()
}
