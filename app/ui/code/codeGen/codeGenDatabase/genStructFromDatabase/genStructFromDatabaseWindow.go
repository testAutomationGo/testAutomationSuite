package genStructFromDatabase

import (
	"log"
	"os"
	"strconv"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/code/codeGen/codeGenDatabase/databaseCache"
	"testAutomationSuiteGO/app/ui/runParametersForUI"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func GenStructFromDatabaseWindow(deps shared.AppDependencies) fyne.CanvasObject {

	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Console output will appear here...")
	appendToConsole := func(text string) {
		fyne.Do(func() {
			consoleOutput.SetText(consoleOutput.Text + text + "\n")
		})
	}

	clearConsole := func() {
		fyne.Do(func() {
			consoleOutput.SetText("")
		})
	}

	databaseCache.Initialize()
	go func() {
		prewarmed := databaseCache.PrewarmKnownTables(false, func(serverName, dbName string) []string {
			return GetDatabaseTables(serverName, dbName, deps)
		})
		if prewarmed > 0 {
			log.Printf("Prewarmed table cache entries: %d", prewarmed)
		}
	}()

	servers := databaseCache.GetServers()
	databases := databaseCache.GetDatabases()
	if len(servers) == 0 {
		servers = runParametersForUI.GetDatabaseServers()
	}
	if len(databases) == 0 {
		databases = runParametersForUI.GetDatabaseNames()
	}
	showDatabasesButton := widget.NewButton("Show Databases", func() {
		appendToConsole("Databases:")
		for i, row := range databases {
			appendToConsole("\nRow " + strconv.Itoa(i) + ":")
			for j, value := range row {
				appendToConsole(" Col " + strconv.Itoa(j) + ": " + value)
			}
		}
	})

	serversSelect := widget.NewSelect(servers, func(string) {})
	if len(servers) > 0 {
		serversSelect.SetSelectedIndex(0)
	}
	var databasesSelect *widget.Select
	if len(databases) > 0 {
		databasesSelect = widget.NewSelect(databases[serversSelect.SelectedIndex()], func(string) {})
	}
	if databasesSelect == nil {
		databasesSelect = widget.NewSelect([]string{}, func(string) {})
	}
	if len(databasesSelect.Options) > 0 {
		databasesSelect.SetSelectedIndex(0)
	}

	loadTablesForSelection := func(forceRefresh bool) []string {
		if serversSelect.Selected == "" || databasesSelect.Selected == "" {
			return []string{}
		}
		return databaseCache.GetOrLoadTables(serversSelect.Selected, databasesSelect.Selected, forceRefresh, func() []string {
			return GetDatabaseTables(serversSelect.Selected, databasesSelect.Selected, deps)
		})
	}
	serversSelect.OnChanged = func(selected string) {
		selectedIndex := -1
		for i, server := range servers {
			if server == selected {
				selectedIndex = i
				break
			}
		}
		if selectedIndex >= 0 && selectedIndex < len(databases) {
			databasesSelect.Options = databases[selectedIndex]
			if len(databasesSelect.Options) > 0 {
				databasesSelect.SetSelectedIndex(0)
			}
			databasesSelect.Refresh()
		} else {
			databasesSelect.Options = []string{}
			databasesSelect.Refresh()
		}
	}

	var tablesSelect *widget.Select
	if len(databasesSelect.Options) > 0 && serversSelect.Selected != "" && databasesSelect.Selected != "" {
		tables := loadTablesForSelection(false)
		tablesSelect = widget.NewSelect(tables, func(string) {})
	}
	if tablesSelect == nil {
		tablesSelect = widget.NewSelect([]string{}, func(string) {})
	}

	databasesSelect.OnChanged = func(selected string) {
		if selected != "" {
			tables := loadTablesForSelection(false)
			tablesSelect.Options = tables
			tablesSelect.Refresh()
		} else {
			tablesSelect.Options = []string{}
			tablesSelect.Refresh()
		}
	}

	refreshTablesButton := widget.NewButton("Refresh Tables", func() {
		if serversSelect.Selected == "" || databasesSelect.Selected == "" {
			return
		}
		tables := loadTablesForSelection(true)
		tablesSelect.Options = tables
		tablesSelect.Refresh()
	})

	genTableStructButton := widget.NewButton("Generate Table Struct", func() {
		appendToConsole("Generating table struct for server: " + serversSelect.Selected + ", database: " + databasesSelect.Selected + ", and table: " + tablesSelect.Selected)
		go func() {
			appendToConsole("Starting database struct generation...")
			theStruct := BuildTableRowStruct(serversSelect.Selected, databasesSelect.Selected, tablesSelect.Selected, deps)
			appendToConsole(theStruct)
		}()
	})

	openCodeGenDatabaseFolder := widget.NewButton("Open Test Data Folder", func() {
		folderPath := testingToolkit.CurrPath() + "/testData/codeGen/codeGenDatabase"
		if !testingToolkit.VerifyFolderIsPresent(folderPath) {
			os.MkdirAll(folderPath, 0755)
		}
		uiFunctions.OpenFolder(folderPath)
	})

	bottomClearConsoleButton := widget.NewButton("Clear Console", func() {
		fyne.Do(func() {
			clearConsole()
		})
	})

	leftContent := container.NewVBox(
		serversSelect,
		databasesSelect,
		tablesSelect,
		genTableStructButton,
		layout.NewSpacer(),
		showDatabasesButton,
		refreshTablesButton,
		layout.NewSpacer(),
		openCodeGenDatabaseFolder,
		bottomClearConsoleButton,
	)
	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, consoleOutput)
	return contentWithConsole
}
