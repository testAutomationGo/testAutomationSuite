package fileCreatorUI

import (
	"fmt"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testData"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var (
	FileCreatorWindowConsoleOutput *widget.Entry
)

func FileCreatorWindow(deps shared.AppDependencies) fyne.CanvasObject {
	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Console Output:")
	FileCreatorWindowConsoleOutput = consoleOutput

	rootProjectPath := testingToolkit.CurrPath()
	appOutputFullPath := rootProjectPath + "/testData/appOutput"
	if !testingToolkit.DoesPathExist(appOutputFullPath) {
		if !testingToolkit.DoesPathExist(rootProjectPath + "/testData") {
			testingToolkit.CreateFolder(rootProjectPath, "/testData")
		}
		testingToolkit.CreateFolder(rootProjectPath+"/testData", "appOutput")
	}

	appendToConsole := func(text string) {
		currentText := FileCreatorWindowConsoleOutput.Text
		FileCreatorWindowConsoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, text))
	}

	buttonTxt := widget.NewButton("Create Text File", func() {
		textFileName := testData.CreateTextFile(appOutputFullPath, "FileCreator")
		appendToConsole("Created Text File: " + textFileName)
		appendToConsole("File Location: " + appOutputFullPath + "/" + textFileName)
	})

	buttonHtml := widget.NewButton("Create HTML File", func() {
		htmlFileName := testData.CreateHTMLFile(appOutputFullPath, "FileCreator")
		appendToConsole("Created HTML File: " + htmlFileName)
		appendToConsole("File Location: " + appOutputFullPath + "/" + htmlFileName)
	})

	buttonPng := widget.NewButton("Create PNG File", func() {
		pngFileName := testData.CreatePNGFile(400, 400, appOutputFullPath, "FileCreator")
		appendToConsole("Created PNG File: " + pngFileName)
		appendToConsole("File Location: " + appOutputFullPath + "/" + pngFileName)
	})

	buttonJson := widget.NewButton("Create JSON File", func() {
		jsonFileName := testData.CreateJSONFile(appOutputFullPath, "FileCreator")
		appendToConsole("Created JSON File: " + jsonFileName)
		appendToConsole("File Location: " + appOutputFullPath + "/" + jsonFileName)
	})

	buttonJpeg := widget.NewButton("Create JPEG File", func() {
		jpegFileName := testData.CreateJPEGFile(400, 400, appOutputFullPath, "FileCreator")
		appendToConsole("Created JPEG File: " + jpegFileName)
		appendToConsole("File Location: " + appOutputFullPath + "/" + jpegFileName)
	})

	buttonJpg := widget.NewButton("Create JPG File", func() {
		jpgFileName := testData.CreateJPGFile(400, 400, appOutputFullPath, "FileCreator")
		appendToConsole("Created JPG File: " + jpgFileName)
		appendToConsole("File Location: " + appOutputFullPath + "/" + jpgFileName)
	})

	buttonPdf := widget.NewButton("Create PDF File", func() {
		pdfFileName := testData.CreatePDFFile(appOutputFullPath, "FileCreator")
		appendToConsole("Created PDF File: " + pdfFileName)
		appendToConsole("File Location: " + appOutputFullPath + "/" + pdfFileName)
	})
	/*
		buttonMp4 := widget.NewButton("Create MP4 File", func() {
			mp4FileName := testData.CreateMP4File(appOutputFullPath, "FileCreator")
			appendToConsole("Created MP4 File: " + mp4FileName)
			appendToConsole("File Location: " + appOutputFullPath + "/" + mp4FileName)
		})*/

	buttonCreateAFolder := widget.NewButton("Create A Folder", func() {
		folderName := "Folder" + testingToolkit.CurrentTimeForNamingWithMS()
		testingToolkit.CreateFolder(appOutputFullPath, folderName)
		testData.CreateJPGFile(400, 400, appOutputFullPath+"/"+folderName, "FileCreator")
		testData.CreatePNGFile(400, 400, appOutputFullPath+"/"+folderName, "FileCreator")
		testData.CreatePNGFile(400, 400, appOutputFullPath+"/"+folderName, "FileCreator")
		appendToConsole("Created Folder: FolderCreator")
		appendToConsole("Folder Location: " + appOutputFullPath + "/FolderCreator")
	})

	datFileSizeInput := widget.NewEntry()
	datFileSizeInput.SetPlaceHolder("Enter size in MB")

	buttonCreateLargeDatFileINMB := widget.NewButton("Create Large Dat File ()", func() {
		datFileSize := datFileSizeInput.Text
		datFileSizeInt := testingToolkit.ConvertStringToInt(datFileSize)
		fileName := "Dat" + testingToolkit.CurrentTimeForNaming()
		LargeFileCreator(int64(datFileSizeInt), fileName)
		appendToConsole("Created Large Dat File: " + fileName)
	})

	buttonOpenFolderLocation := widget.NewButton("Open Folder Location", func() {
		uiFunctions.OpenFolder(rootProjectPath + "testData/appOutput")
	})

	BottomButtonClearConsole := widget.NewButton("Clear Console", func() {
		FileCreatorWindowConsoleOutput.SetText("Console Output:")
	})

	leftContent := container.NewVBox(
		widget.NewLabel("Create Files:"),
		buttonTxt,
		buttonHtml,
		buttonPdf,
		buttonJson,
		buttonJpeg,
		buttonJpg,
		buttonPng,
		//buttonMp4,
		buttonCreateAFolder,
		widget.NewLabel("Create Large Dat File:"),
		datFileSizeInput,
		buttonCreateLargeDatFileINMB,
		//widget.NewLabel("Test App Updater"),
		layout.NewSpacer(),
		buttonOpenFolderLocation,
		BottomButtonClearConsole,
	)

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, FileCreatorWindowConsoleOutput)

	return contentWithConsole
}
