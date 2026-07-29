package apiUI

import (
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/apiUI/requestor"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type ApiWindowTab struct {
	name    string
	content fyne.CanvasObject
}

func ApiWindow(deps shared.AppDependencies) fyne.CanvasObject {
	tabs := []ApiWindowTab{
		{
			name:    "Requestor",
			content: requestor.RequestorWindow(deps),
		}, /*
			{
				name:    "Query ADO",
				content: queryADOUI.QueryADOWindow(deps),
			},*/
	}
	tabItems := make([]*container.TabItem, len(tabs))
	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}
	tabsContainer := container.NewAppTabs(tabItems...)
	return tabsContainer
}
