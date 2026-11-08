package components

import (
	"github.com/rivo/tview"
)

type Spinner struct {
	*tview.TextView
	frames     []rune
	frameIndex int
}

func NewSpinner() *Spinner {
	return &Spinner{
		TextView:   tview.NewTextView(),
		frames:     []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'},
		frameIndex: 0,
	}
}

// Tick updates the spinner animation.
//
// Usage
//
//	spinner := NewSpinner()
//	app := tview.NewApplication().SetRoot(spinner, true)
//
//	// Run the ticker from the main app
//	go func() {
//		ticker := time.NewTicker(100 * time.Millisecond)
//		defer ticker.Stop()
//		for range ticker.C {
//			app.QueueUpdateDraw(func() {
//				spinner.Tick()
//			})
//		}
//	}()
func (s *Spinner) Tick() {
	s.frameIndex = (s.frameIndex + 1) % len(s.frames)
	s.TextView.SetText(string(s.frames[s.frameIndex]))
}
