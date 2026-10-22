package main

import (
	"fmt"
	"os"

	"github.com/anuragpandey01/tiny-coder/internal/agent"
	"github.com/anuragpandey01/tiny-coder/internal/tui"
)

func main() {
	coder := agent.NewOpenAI()
	if err := tui.New(coder).Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
