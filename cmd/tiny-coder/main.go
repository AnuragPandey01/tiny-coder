package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/anuragpandey01/tiny-coder/internal/agent"
	"github.com/anuragpandey01/tiny-coder/internal/config"
	"github.com/anuragpandey01/tiny-coder/internal/logger"
	"github.com/anuragpandey01/tiny-coder/internal/tui"
	"github.com/google/uuid"
)

func main() {
	appConfig, err := config.GetDefaultAppConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config:", err)
		os.Exit(1)
	}

	sessionID := uuid.New().String()
	if err := logger.InitAppLogger(appConfig, sessionID); err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger:", err)
		os.Exit(1)
	}
	slog.Info("TINY-CODER Started ＼(＾O＾)／ ")

	agentRunner := agent.NewOpenAI()
	app := tui.NewApp(agentRunner)
	if err := app.Run(); err != nil {
		slog.Error("application stopped with error", "error", err)
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	slog.Info("application stopped")
}
