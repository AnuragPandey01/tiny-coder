package style

import "github.com/gdamore/tcell/v2"

type ColorTag string

// Color Tag's
var (
	UserColorTag       ColorTag = "blue::b"
	AgentColorTag      ColorTag = "silver"
	HintColorTag       ColorTag = "gray"
	ModelNameColorTag  ColorTag = "yellow::b"
	WorkingDirColorTag ColorTag = "green::b"

	// Useful ASCII's
	Cursor           = "▍"
	InfoBarSeparator = "∙"

	UserInputBgColor = tcell.GetColor("#303030")
)
