package tui

import (
	"fmt"
	"strings"

	"github.com/anuragpandey01/tiny-coder/internal/agent"
	"github.com/anuragpandey01/tiny-coder/internal/agent/events"
	"github.com/anuragpandey01/tiny-coder/internal/tui/components"
	"github.com/anuragpandey01/tiny-coder/internal/tui/style"
	"github.com/anuragpandey01/tiny-coder/internal/tui/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	tv *tview.Application

	chat    *components.ChatView
	input   *tview.InputField
	infoBar *components.InfoBar
	status  *components.StatusBar

	agent agent.Agent
}

// NewApp initializes the TUI around an event emitting agent.
func NewApp(a agent.Agent) *App {

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault

	app := &App{
		tv:    tview.NewApplication(),
		agent: a,
	}

	app.chat = components.NewChatView(nil)

	//user input component
	app.input = tview.NewInputField().
		SetLabel("> ").
		SetPlaceholder("Ask Tinycoder to do anything").
		SetLabelStyle(tcell.StyleDefault.Foreground(style.UserInputBgColor).Background(style.UserInputBgColor).Bold(true)).
		SetPlaceholderStyle(tcell.StyleDefault.Background(style.UserInputBgColor)).
		SetFieldBackgroundColor(style.UserInputBgColor)
	app.input.SetBackgroundColor(style.UserInputBgColor)
	app.input.SetBorderPadding(1, 1, 0, 0)

	app.input.SetDoneFunc(func(key tcell.Key) {
		prompt := strings.TrimSpace(app.input.GetText())
		if key != tcell.KeyEnter || app.chat.IsAgentRunning() || prompt == "" {
			return
		}
		app.chat.AppendChatItem(components.ChatItem{ItemType: components.ItemTypeUserQuery, Text: prompt})
		app.input.SetText("")
		app.chat.BeginAgentResponse()
		app.status.SetAgentWorking()
		app.render()

		go app.onAgentEvent(app.agent.Do(prompt))
	})

	app.infoBar = components.NewInfoBar("gemma4:12b-mlx", "~/test", "N/A")
	app.status = components.NewStatusBar(func(update func()) {
		app.tv.QueueUpdateDraw(update)
	})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.chat, 0, 1, false).
		AddItem(app.status, 2, 0, false).
		AddItem(app.input, 3, 0, true).
		AddItem(app.infoBar, 3, 0, false)

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
	a.infoBar.SetInfoBarHint(utils.StyledString(style.HintColorTag, "enter to send · ctrl+c to quit"))
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
		a.chat.AppendAgentText(event.Text)
	case events.ToolStarted:
		a.chat.StartToolCall(event.CallID, event.Name, event.Arguments)
	case events.ToolFinished:
		a.chat.FinishToolCall(event.CallID, fmt.Sprint(event.Result), event.Error)
	case events.TurnFinished:
		a.chat.FinishAgentResponse()
		a.status.SetAgentFinishedWorking()
	}
}
