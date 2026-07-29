package codeGenDatabase

import (
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/code/codeGen/codeGenDatabase/genStructFromDatabase"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type CodeGenDatabaseTab struct {
	name    string
	content fyne.CanvasObject
}

func CodeGenDatabaseWindow(deps shared.AppDependencies) fyne.CanvasObject {
	tabs := []CodeGenDatabaseTab{
		{
			name:    "Gen Struct From Database",
			content: genStructFromDatabase.GenStructFromDatabaseWindow(deps),
		},
	}
	tabItems := make([]*container.TabItem, len(tabs))
	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}
	tabsContainer := container.NewAppTabs(tabItems...)
	return tabsContainer
}
