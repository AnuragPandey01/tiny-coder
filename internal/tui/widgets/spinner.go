package widgets

import (
	"github.com/rivo/tview"
)

type Spinner struct {
	*tview.Box

	frames     []rune
	frameIndex int
}

// ['⣾','⣽','⣻','⢿','⡿','⣟','⣯','⣷']
func NewSpinner() *Spinner {
	return &Spinner{
		Box:        tview.NewBox(),
		frames:     []rune{'-', '\\', '|', '/'},
		frameIndex: 0,
	}
}

//func (s *Spinner) Draw(screen tcell.Screen) {
//	s.Box.Draw(screen)
//
//	x, y, width, height := t.GetInnerRect()
//	if width <= 0 || height <= 0 {
//		return
//	}
//
//	screen.SetContent(x+width/2, y+height/2, s.frames[s.frameIndex], nil, tcell.StyleDefault)
//	s.frameIndex = (s.frameIndex + 1) % len(s.frames)
//}
//
//func (s *Spinner) Start() {
//	go func() {
//		for {
//			s.Draw(t)
//			time.Sleep(0.2 * time.Second)
//		}
//	}()
//}
