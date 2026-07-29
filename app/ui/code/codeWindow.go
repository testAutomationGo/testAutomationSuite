package code

import (
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/code/codeGen"
	"testAutomationSuiteGO/app/ui/code/codeRunner"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type CodeWindowTab struct {
	name    string
	content fyne.CanvasObject
}

func CodeWindow(deps shared.AppDependencies) fyne.CanvasObject {
	tabs := []CodeWindowTab{
		{
			name:    "Code Runner",
			content: codeRunner.CodeRunnerWindow(deps),
		},
		{
			name:    "Code Generator",
			content: codeGen.CodeGenWindow(deps),
		},
	}
	tabItems := make([]*container.TabItem, len(tabs))
	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}
	tabsContainer := container.NewAppTabs(tabItems...)
	return tabsContainer
}
