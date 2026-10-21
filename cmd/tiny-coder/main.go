package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anuragpandey01/tiny-coder/internal/tui"
)

type DummyAgent struct{}

func (a *DummyAgent) Run(prompt string) <-chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(500 * time.Millisecond)
		defer close(ch)
		reply := "You ssaid: \"" + prompt + "\". In the real agent, these words would " +
			"arrive from the LLM's streaming API, one chunk at a time, while the " +
			"UI stays responsive because the event loop never blocks."
		for _, word := range strings.Fields(reply) {
			time.Sleep(50 * time.Millisecond)
			ch <- word + " "
		}
	}()
	return ch
}

func main() {

	agent := &DummyAgent{}

	if err := tui.New(agent).Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
