package components

import (
	"fmt"
	"time"

	"github.com/anuragpandey01/tiny-coder/internal/tui/style"
	"github.com/anuragpandey01/tiny-coder/internal/tui/utils"
	"github.com/rivo/tview"
)

type StatusBar struct {
	*tview.Flex
	spinner   *Spinner
	content   *tview.TextView
	queueDraw func(func())
	stopTicks chan struct{}
	startedAt time.Time
	working   bool
	lastRun   time.Duration
}

// NewStatusBar creates a status bar. queueDraw schedules updates on the tview
// application's UI loop, typically with Application.QueueUpdateDraw.
func NewStatusBar(queueDraw func(func())) *StatusBar {
	sb := &StatusBar{
		Flex:      tview.NewFlex().SetDirection(tview.FlexColumn),
		spinner:   NewSpinner(),
		content:   tview.NewTextView().SetDynamicColors(true),
		queueDraw: queueDraw,
	}

	sb.AddItem(sb.spinner, 2, 0, false).
		AddItem(sb.content, 0, 1, false)

	sb.render()

	return sb
}

func (sb *StatusBar) render() {
	if sb.working {
		sb.content.SetText(utils.StyledString(style.HintColorTag, "the tiny-coder is working…"))
		return
	}
	sb.spinner.SetText("")
	if sb.lastRun > 0 {
		sb.content.SetText(utils.StyledString(style.HintColorTag, fmt.Sprintf("Last run: %s", sb.lastRun.Round(10*time.Millisecond))))
		return
	}
}

func (sb *StatusBar) SetAgentWorking() {
	if sb.working {
		return
	}
	sb.working = true
	sb.startedAt = time.Now()
	sb.stopTicks = make(chan struct{})
	stop := sb.stopTicks
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				sb.queueDraw(func() { sb.Tick() })
			}
		}
	}()
	sb.render()
}

func (sb *StatusBar) SetAgentFinishedWorking() {
	if !sb.working {
		return
	}
	close(sb.stopTicks)
	sb.stopTicks = nil
	sb.lastRun = time.Since(sb.startedAt)
	sb.working = false
	sb.render()
}

func (sb *StatusBar) Tick() {
	if sb.working {
		sb.spinner.Tick()
	}
}
