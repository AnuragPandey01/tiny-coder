package widgets

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Spinner struct {
	*tview.Box

	a          *tview.Application
	frames     []rune
	frameIndex int
	cancel     chan struct{}
}

func NewSpinner(app *tview.Application) *Spinner {
	return &Spinner{
		Box:        tview.NewBox(),
		frames:     []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'},
		frameIndex: 0,
		a:          app,
		cancel:     make(chan struct{}),
	}
}

func (s *Spinner) Draw(screen tcell.Screen) {
	s.Box.Draw(screen)

	x, y, width, height := s.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	screen.SetContent(x+width/2, y+height/2, s.frames[s.frameIndex], nil, tcell.StyleDefault)
}

// Start begins the background animation loop. Call this exactly once.
func (s *Spinner) Start() {
	if s.cancel != nil {
		return // Already running
	}

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.cancel:
				return
			case <-ticker.C:
				s.a.QueueUpdateDraw(func() {
					s.frameIndex = (s.frameIndex + 1) % len(s.frames)
				})
			}
		}
	}()
}

// Stop halts the animation and cleans up the goroutine.
func (s *Spinner) Stop() {
	if s.cancel != nil {
		close(s.cancel)
		s.cancel = nil
	}
}
