package chatview

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ChatItemType int
type ToolStatus int

// UserQueryMsg adds a user query and starts a new streaming response.
type UserQueryMsg struct {
	Text string
}

// ResponseDeltaMsg appends text to the current streaming response.
type ResponseDeltaMsg struct {
	Text string
}

// ResponseFinishedMsg commits the streaming response to the transcript.
type ResponseFinishedMsg struct{}

// ResponseErrorMsg preserves any partial response and displays an error.
type ResponseErrorMsg struct {
	Text string
}

const (
	ItemTypeUserQuery ChatItemType = iota
	ItemTypeAgentResponse
	ItemTypeToolCall
	ItemTypeError
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

// ChatView renders the conversation as a scrollable transcript
type Model struct {
	items                    []ChatItem
	currentStreamingResponse string
	agentRunning             bool

	viewport         viewport.Model
	userSenderStyle  lipgloss.Style
	agentSenderStyle lipgloss.Style
	errorStyle       lipgloss.Style
}

func NewChatView(chat []ChatItem) Model {
	m := Model{
		items:            append([]ChatItem(nil), chat...),
		viewport:         viewport.New(),
		userSenderStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true),
		agentSenderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true),
		errorStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),
	}
	m.viewport.FillHeight = true
	m.refreshContent()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	follow := m.viewport.AtBottom()
	switch msg := msg.(type) {
	case UserQueryMsg:
		m.finishResponse()
		m.items = append(m.items, ChatItem{ItemType: ItemTypeUserQuery, Text: msg.Text})
		m.agentRunning = true
		follow = true
	case ResponseDeltaMsg:
		m.agentRunning = true
		m.currentStreamingResponse += msg.Text
	case ResponseFinishedMsg:
		m.finishResponse()
	case ResponseErrorMsg:
		m.finishResponse()
		m.items = append(m.items, ChatItem{ItemType: ItemTypeError, Text: msg.Text})
	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	m.refreshContent()
	if follow {
		m.viewport.GotoBottom()
	}
	return m, nil
}

func (m Model) View() string {
	return m.viewport.View()
}

func (m *Model) SetHeight(h int) {
	m.viewport.SetHeight(h)
}

func (m *Model) SetWidth(h int) {
	m.viewport.SetWidth(h)
}

// finishResponse commits the current text once, including a partial response
// when a turn fails. Empty turns do not add an empty transcript entry.
func (m *Model) finishResponse() {
	if m.currentStreamingResponse != "" {
		m.items = append(m.items, ChatItem{
			ItemType: ItemTypeAgentResponse,
			Text:     m.currentStreamingResponse,
		})
	}
	m.currentStreamingResponse = ""
	m.agentRunning = false
}

func (m *Model) refreshContent() {
	var content strings.Builder
	for _, item := range m.items {
		switch item.ItemType {
		case ItemTypeUserQuery:
			content.WriteString(m.userSenderStyle.Render("You: "))
		case ItemTypeAgentResponse:
			content.WriteString(m.agentSenderStyle.Render("Agent: "))
		case ItemTypeError:
			content.WriteString(m.errorStyle.Render("Error: "))
		}
		content.WriteString(item.Text)
		content.WriteString("\n")
		if item.ItemType == ItemTypeAgentResponse {
			content.WriteString("\n")
		}
	}
	if m.agentRunning {
		content.WriteString(m.agentSenderStyle.Render("Agent: "))
		content.WriteString(m.currentStreamingResponse)
		content.WriteString("▍\n")
	}
	m.viewport.SetContent(content.String())
}
