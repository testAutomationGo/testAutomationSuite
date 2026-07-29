package codeGen

import (
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/code/codeGen/codeGenDatabase"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type CodeGenWindowTab struct {
	name    string
	content fyne.CanvasObject
}

func CodeGenWindow(deps shared.AppDependencies) fyne.CanvasObject {
	tabs := []CodeGenWindowTab{
		{
			name:    "Database Code Gen",
			content: codeGenDatabase.CodeGenDatabaseWindow(deps),
		},
	}
	tabItems := make([]*container.TabItem, len(tabs))
	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}
	tabsContainer := container.NewAppTabs(tabItems...)
	return tabsContainer
}
