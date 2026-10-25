package tui

import (
	"fmt"
	"strings"

	"github.com/anuragpandey01/tiny-coder/internal/agent"
	"github.com/anuragpandey01/tiny-coder/internal/agent/events"
	"github.com/anuragpandey01/tiny-coder/internal/tui/utils"
	"github.com/anuragpandey01/tiny-coder/internal/tui/widgets"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type role int

const (
	roleUser role = iota
	roleAgent
)

type turn struct {
	role role
	text string
}

// appState contains the conversation state that the TUI presents.
type appState struct {
	history        []turn
	current        string
	streaming      bool
	activeTool     string
	toolStatus     string
	lastToolResult any
	error          string
}

type App struct {
	tv         *tview.Application
	transcript *tview.TextView   // conversation pane
	input      *tview.InputField // what the user is typing
	hint       *tview.TextView   // status line
	spinner    *widgets.Spinner

	state appState
	agent agent.Agent
}

// NewApp initializes the TUI around an event emitting agent.
func NewApp(a agent.Agent) *App {

	app := &App{
		tv:    tview.NewApplication(),
		agent: a,
	}

	app.transcript = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetWordWrap(true).
		SetChangedFunc(func() { app.tv.Draw() })

	app.input = tview.NewInputField().
		SetLabel("> ").
		SetLabelStyle(tcell.StyleDefault.Foreground(tcell.ColorLime).Bold(true)).
		SetFieldBackgroundColor(tcell.ColorDefault)

	app.input.SetDoneFunc(func(key tcell.Key) {
		prompt := strings.TrimSpace(app.input.GetText())
		if key != tcell.KeyEnter || app.state.streaming || prompt == "" {
			return
		}
		app.state.history = append(app.state.history, turn{roleUser, prompt})
		app.input.SetText("")
		app.state.streaming = true
		app.state.current = ""
		app.state.activeTool = ""
		app.state.toolStatus = ""
		app.state.lastToolResult = nil
		app.state.error = ""
		app.render()

		go app.onAgentEvent(app.agent.Do(prompt))
	})

	app.hint = tview.NewTextView().SetDynamicColors(true)
	app.spinner = widgets.NewSpinner(app.tv)
	app.spinner.Start()

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.transcript, 0, 1, false).
		AddItem(app.input, 1, 0, true).
		AddItem(app.spinner, 1, 0, false)

	app.tv.SetRoot(layout, true).
		SetFocus(app.input).
		EnablePaste(true).
		EnableMouse(true)

	app.render()

	return app
}

func (a *App) Run() error {
	return a.tv.Run()
}

func (a *App) render() {
	var b strings.Builder

	for _, t := range a.state.history {
		switch t.role {
		case roleUser:
			b.WriteString(utils.StyledString(UserColor, "> "+tview.Escape(t.text)+"\n"))
		case roleAgent:
			b.WriteString(utils.StyledString(AgentColor, tview.Escape(t.text)+"\n\n"))
		}
	}

	if a.state.streaming {
		b.WriteString(utils.StyledString(AgentColor, tview.Escape(a.state.current)+cursor))
	}

	a.transcript.SetText(b.String())
	a.transcript.ScrollToEnd()

	hint := "enter to send · ctrl+c to quit"
	if a.state.activeTool != "" {
		hint = "running tool: " + a.state.activeTool + " · ctrl+c to quit"
	} else if a.state.toolStatus != "" {
		hint = a.state.toolStatus + " · ctrl+c to quit"
	} else if a.state.error != "" {
		hint = "agent error: " + a.state.error + " · ctrl+c to quit"
	} else if a.state.streaming {
		hint = "agent is replying… · ctrl+c to quit"
	}

	a.hint.SetText(utils.StyledString(HintColor, hint))
}

func (a *App) onAgentEvent(eventStream <-chan events.Event) {
	for event := range eventStream {
		a.tv.QueueUpdateDraw(func() {
			a.onEvent(event)
			a.render()
		})
	}
}

func (a *App) onEvent(event events.Event) {
	switch event := event.(type) {
	case events.TextDelta:
		a.state.current += event.Text
	case events.ToolStarted:
		a.state.activeTool = event.Name
		a.state.toolStatus = ""
		a.state.lastToolResult = nil
	case events.ToolFinished:
		a.state.activeTool = ""
		a.state.toolStatus = "tool completed: " + event.Name
		a.state.lastToolResult = event.Result
		if event.Error != "" {
			a.state.error = fmt.Sprintf("%s: %s", event.Name, event.Error)
			a.state.toolStatus = ""
		}
	case events.Error:
		a.state.error = event.Error
	case events.TurnFinished:
		if a.state.current != "" {
			a.state.history = append(a.state.history, turn{roleAgent, a.state.current})
		}
		a.state.streaming = false
		a.state.current = ""
		a.state.activeTool = ""
	}
}
