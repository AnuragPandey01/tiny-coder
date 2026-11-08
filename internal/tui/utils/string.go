package utils

import "github.com/anuragpandey01/tiny-coder/internal/tui/style"

func StyledString(colorTag style.ColorTag, actualString string) string {
	// tview resets with a single minus [-] inside the closing bracket
	return "[" + string(colorTag) + "]" + actualString + "[-]"
}
