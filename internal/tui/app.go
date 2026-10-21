package tui

import (
	"strings"

	"github.com/anuragpandey01/tiny-coder/internal/agent"
	"github.com/anuragpandey01/tiny-coder/internal/tui/utils"
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

type App struct {
	tv         *tview.Application
	transcript *tview.TextView   // conversation pane
	input      *tview.InputField // what the user is typing
	hint       *tview.TextView   // status line

	history   []turn // full conversation
	current   string // the reply currently being streamed
	streaming bool   // is the agent replying right now?

	agent agent.Agent
}

// New initializes the coder and returns the initial state of the coder
func New(a agent.Agent) *App {

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
		if key != tcell.KeyEnter || app.streaming || prompt == "" {
			return
		}
		app.history = append(app.history, turn{roleUser, prompt})
		app.input.SetText("")
		app.streaming = true
		app.current = ""
		app.render()

		go app.consume(app.agent.Run(prompt))
	})

	app.hint = tview.NewTextView().SetDynamicColors(true)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.transcript, 0, 1, false).
		AddItem(app.input, 1, 0, true).
		AddItem(app.hint, 1, 0, false)

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

	for _, t := range a.history {
		switch t.role {
		case roleUser:
			b.WriteString(utils.StyledString(UserColor, "> "+tview.Escape(t.text)+"\n"))
		case roleAgent:
			b.WriteString(utils.StyledString(AgentColor, tview.Escape(t.text)+"\n\n"))
		}
	}

	if a.streaming {
		b.WriteString(utils.StyledString(AgentColor, tview.Escape(a.current)+cursor))
	}

	a.transcript.SetText(b.String())
	a.transcript.ScrollToEnd()

	hint := "enter to send · ctrl+c to quit"
	if a.streaming {
		hint = "agent is replying… · ctrl+c to quit"
	}

	a.hint.SetText(utils.StyledString(HintColor, hint))
}

func (a *App) consume(ch <-chan string) {
	for tok := range ch {
		a.tv.QueueUpdateDraw(func() {
			a.current += tok
			a.render()
		})
	}
	a.tv.QueueUpdateDraw(func() {
		a.history = append(a.history, turn{roleAgent, a.current})
		a.streaming = false
		a.current = ""
		a.render()
	})
}

// fakeAgent stands in for a real LLM client. Replace its body with a call to
// your model provider's streaming API, sending each text delta into ch.
