package pageToolsUI

import (
	"os"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"
	"testAutomationSuiteGO/internal/webAppTesting/playwrightInternal"
	"testAutomationSuiteGO/internal/webAppTesting/webPageTracing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/playwright-community/playwright-go"
)

func PageToolsWindow(deps shared.AppDependencies) fyne.CanvasObject {
	mapperOutputFolder := testingToolkit.CurrPath() + "/testData/appOutput/webPageMappers"
	if !testingToolkit.VerifyFolderIsPresent(mapperOutputFolder) {
		os.MkdirAll(mapperOutputFolder, 0755)
	}
	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Write Your Notes Here:")
	appendToConsole := func(text string) {
		fyne.Do(func() {
			currentText := consoleOutput.Text
			consoleOutput.SetText(currentText + "\n" + text)
		})
	}
	titleLabel := widget.NewLabel("Page Tools")
	var playwrightPage playwright.Page
	var playwrightContext playwright.BrowserContext

	urlTextBox := widget.NewEntry()
	urlTextBox.SetPlaceHolder("Navigate to:")

	navToButton := widget.NewButton("Nav To:", func() {
		_, _, playwrightContext, playwrightPage = playwrightInternal.InitPlaywrightWithContextUtil(false, "normal", playwright.BrowserNewContextOptions{})
		if ValidateURL(urlTextBox.Text) {
			playwrightInternal.GoToUtil(playwrightPage, urlTextBox.Text)
		} else {
			appendToConsole("Invalid URL, edit and try again.")
		}
	})
	navToButton.Importance = widget.HighImportance

	mapWebPageButton := widget.NewButton("Map Web Page", func() {
		go func() {
			if playwrightPage == nil {
				appendToConsole("Please navigate to a page first.")
				return
			}
			fileName := testingToolkit.CurrentTimeForNamingWithMS() + "_webPageMap.json"
			err := webPageTracing.PrintAllPageElementsToFile(playwrightPage, mapperOutputFolder+"/"+fileName)
			if err != nil {
				appendToConsole("Error mapping web page: " + err.Error())
			}
			playwrightContext.Close()
			appendToConsole("Web page mapped successfully. Output saved to: " + mapperOutputFolder + "/" + fileName)
			uiFunctions.OpenFile(mapperOutputFolder + "/" + fileName)
		}()
	})
	openMapperFolderButton := widget.NewButton("Open Mapper Folder ", func() {
		go func() {
			uiFunctions.OpenFolder(mapperOutputFolder)
		}()
	})
	bottomClearConsoleButton := widget.NewButton("Clear Console", func() {
		go consoleOutput.SetText("")
	})

	leftContent := container.NewVBox(
		titleLabel,
		urlTextBox,
		navToButton,
		layout.NewSpacer(),
		openMapperFolderButton,
		bottomClearConsoleButton,
	)
	topBox := container.NewHBox(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), mapWebPageButton)
	rightContent := container.New(
		layout.NewBorderLayout(topBox, nil, nil, nil),
		topBox,
		container.NewStack(consoleOutput),
	)
	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, rightContent)
	return contentWithConsole
}
