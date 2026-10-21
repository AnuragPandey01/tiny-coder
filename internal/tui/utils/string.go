package utils

func StyledString(colorTagBody string, actualString string) string {
	// tview resets with a single minus [-] inside the closing bracket
	return "[" + colorTagBody + "]" + actualString + "[-]"
}
