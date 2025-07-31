package accountManagementUI

import (
	"testAutomationSuiteGO/app/shared"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func AccountManagementWindow(deps shared.AppDependencies) fyne.CanvasObject {

	tabs := []AccountManagementTab{
		{
			name:    "Workspaces",
			content: WorkspacesTab(deps),
		},
		{
			name:    "Accounts",
			content: AccountsTab(deps),
		},
	}

	tabItems := make([]*container.TabItem, len(tabs))

	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}

	tabsContainer := container.NewAppTabs(tabItems...)

	return container.NewBorder(nil, nil, nil, nil, tabsContainer)
}
