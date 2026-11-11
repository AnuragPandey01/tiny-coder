package tui

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anuragpandey01/tiny-coder/internal/agent"
	"github.com/anuragpandey01/tiny-coder/internal/agent/events"
	"github.com/anuragpandey01/tiny-coder/internal/tui/components/chatview"
)

type App struct {
	chatView      chatview.Model
	textInputView textarea.Model

	paddingBottom              int
	paddingLeft                int
	paddingRight               int
	paddingTop                 int
	width                      int
	height                     int
	activeAgentResponseChannel <-chan events.Event

	agent agent.Agent
}

// NewApp initializes the TUI around an event emitting agent.
func NewApp(a agent.Agent) *App {
	app := &App{
		chatView:      chatview.NewChatView(make([]chatview.ChatItem, 0)),
		textInputView: textarea.New(),

		paddingBottom: 2,
		paddingLeft:   1,
		paddingRight:  1,
		paddingTop:    0,
		agent:         a,
	}

	//input styling
	app.textInputView.Placeholder = "ask tinycoder anything, even if it's not a tiny query!"
	app.textInputView.Focus()
	app.textInputView.ShowLineNumbers = false
	s := textarea.DefaultStyles(true)
	s.Focused.CursorLine = lipgloss.NewStyle().UnsetBackground()
	app.textInputView.SetStyles(s)

	return app
}

func (a App) Init() tea.Cmd {
	return textarea.Blink
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd = []tea.Cmd{}
	)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.chatView.SetWidth(msg.Width - a.paddingLeft - a.paddingRight)
		a.chatView.SetHeight(msg.Height - 3 - a.paddingTop - a.paddingBottom)

		a.textInputView.SetWidth(msg.Width - a.paddingLeft - a.paddingRight)
		a.textInputView.SetHeight(3)
		return a, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit

		case "enter":
			userQuery := strings.TrimSpace(a.textInputView.Value())
			if userQuery == "" || a.activeAgentResponseChannel != nil {
				return a, nil
			}

			a.chatView, cmd = a.chatView.Update(chatview.UserQueryMsg{
				Text: userQuery,
			})
			a.textInputView.Reset()

			a.activeAgentResponseChannel = a.agent.Do(userQuery)

			return a, handleAgentEventStream(a.activeAgentResponseChannel)
		}

	case events.TextDeltaMsg:
		a.chatView, cmd = a.chatView.Update(chatview.ResponseDeltaMsg{Text: msg.Text})
		return a, tea.Batch(cmd, handleAgentEventStream(a.activeAgentResponseChannel))

	case events.Error:
		a.chatView, cmd = a.chatView.Update(chatview.ResponseErrorMsg{Text: msg.Error})
		return a, tea.Batch(cmd, handleAgentEventStream(a.activeAgentResponseChannel))

	case events.ToolStartedMsg, events.ToolFinishedMsg:
		return a, handleAgentEventStream(a.activeAgentResponseChannel)

	case events.TurnFinishedMsg:
		a.activeAgentResponseChannel = nil
		a.chatView, cmd = a.chatView.Update(chatview.ResponseFinishedMsg{})
		return a, cmd
	}

	a.chatView, cmd = a.chatView.Update(msg)
	cmds = append(cmds, cmd)

	a.textInputView, cmd = a.textInputView.Update(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a App) View() tea.View {

	if a.width == 0 || a.height == 0 {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		a.chatView.View(),
		a.textInputView.View(),
	)

	view := tea.NewView(
		lipgloss.NewStyle().
			Width(a.width).
			Height(a.height).
			Padding(a.paddingTop, a.paddingRight, a.paddingBottom, a.paddingLeft).
			Render(body),
	)
	view.AltScreen = true
	return view
}

func handleAgentEventStream(stream <-chan events.Event) tea.Cmd {
	if stream == nil {
		return nil
	}
	return func() tea.Msg {
		if msg, ok := <-stream; ok {
			return msg
		}
		return events.TurnFinishedMsg{}
	}
}
