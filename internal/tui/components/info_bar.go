package components

import (
	"strings"

	"github.com/anuragpandey01/tiny-coder/internal/tui/style"
	"github.com/anuragpandey01/tiny-coder/internal/tui/utils"
	"github.com/rivo/tview"
)

type InfoBar struct {
	*tview.TextView

	modelName  string
	workingDir string
	hint       string
}

func NewInfoBar(modelName, workingDir, hint string) *InfoBar {
	ib := &InfoBar{
		TextView:   tview.NewTextView().SetDynamicColors(true),
		modelName:  modelName,
		workingDir: workingDir,
		hint:       hint,
	}
	ib.render()
	return ib
}

func (ib *InfoBar) SetModelName(modelName string) {
	ib.modelName = modelName
	ib.render()
}

func (ib *InfoBar) SetWorkingDir(workingDir string) {
	ib.workingDir = workingDir
	ib.render()
}

func (ib *InfoBar) SetInfoBarHint(hint string) {
	ib.hint = hint
	ib.render()
}

func (ib *InfoBar) render() {

	var infoBarText strings.Builder

	//model
	infoBarText.WriteString(utils.StyledString(style.ModelNameColorTag, ib.modelName))
	infoBarText.WriteString(" " + style.InfoBarSeparator + " ")

	//workingDir
	infoBarText.WriteString(utils.StyledString(style.WorkingDirColorTag, ib.workingDir))
	infoBarText.WriteString(" " + style.InfoBarSeparator + " ")

	infoBarText.WriteString("\n")

	infoBarText.WriteString(utils.StyledString(style.ModelNameColorTag, ib.hint))

	ib.SetText(infoBarText.String())

}
