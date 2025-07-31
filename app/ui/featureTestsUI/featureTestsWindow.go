package featureTestsUI

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type FeatureTestTab struct {
	name    string
	content fyne.CanvasObject
}

func FeatureTestsWindow() fyne.CanvasObject {

	tabs := GenerateFeatureTestTabs()

	tabItems := make([]*container.TabItem, len(tabs))

	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}

	tabsContainer := container.NewAppTabs(tabItems...)

	return container.NewBorder(nil, nil, nil, nil, tabsContainer)
}
