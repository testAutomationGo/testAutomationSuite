package awsS3FeatureTest

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/awsS3Functions"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func GenerateWindowContent() fyne.CanvasObject {

	consoleOutput := uiFunctions.NewCustomEntry(shared.GetThisWindow())
	appendToConsole := func(text string) {
		currentText := consoleOutput.Text
		consoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, text))
	}

	bucketNameEntry := widget.NewEntry()
	bucketNameEntry.SetPlaceHolder("Enter Bucket Name")
	createBucketButton := widget.NewButton("Create Bucket", func() {
		go func() {
			appendToConsole("Creating Bucket in " + shared.GetCurrentEnvironment() + "...")
			key, secret := "", ""
			if bucketNameEntry.Text == "" {
				appendToConsole("Please enter a bucket name.")
				return
			}
			svc, err := awsS3Functions.CreateServiceClient(key, secret)
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to create S3 client: %v", err))
				return
			}
			_, err = awsS3Functions.CreateBucket(svc, bucketNameEntry.Text, "")
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to create bucket: %v", err))
				return
			}
			appendToConsole(fmt.Sprintf("Created Bucket: %s", bucketNameEntry.Text))
		}()
	})

	pinAnyFileEntry := widget.NewEntry()
	pinAnyFileEntry.SetPlaceHolder("Paste File Path Here")

	var filePath string = ""
	var fileName string = ""

	fileEntryBrowseButton := widget.NewButton("Browse", func() {
		fileDialog := dialog.NewFileOpen(
			func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					log.Println("Error opening file dialog:", err)
					dialog.ShowError(err, shared.GetThisWindow())
					return
				}
				if reader == nil {
					return
				}
				filePath = reader.URI().Path()
				fileName = filepath.Base(filePath)
				filePath = filepath.Dir(filePath)
				pinAnyFileEntry.SetText(filePath)
			}, shared.GetThisWindow())
		initialDir := storage.NewFileURI(testingToolkit.CurrPath())
		lister, err := storage.ListerForURI(initialDir)
		if err != nil {
			log.Println("Error getting lister for URI:", err)
			dialog.ShowError(err, shared.GetThisWindow())
			return
		} else {
			fileDialog.SetLocation(lister)
		}
		fileDialog.Resize(fyne.NewSize(800, 500))
		fileDialog.Show()
	})

	browseBox := container.NewHBox(
		layout.NewSpacer(),
		fileEntryBrowseButton,
	)

	browseContainer := container.NewBorder(nil, nil, nil, browseBox, pinAnyFileEntry)

	uploadFileButton := widget.NewButton("Upload File", func() {
		go func() {
			apiKey, apiSecret := "", ""
			if filePath == "" || fileName == "" {
				appendToConsole("Please select a file to upload.")
				return
			}
			if bucketNameEntry.Text == "" {
				appendToConsole("Please enter a bucket name.")
				return
			}
			svc, err := awsS3Functions.CreateServiceClient(apiKey, apiSecret)
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to create S3 client: %v", err))
				return
			}
			appendToConsole(fmt.Sprintf("Uploading file %s to bucket %s in %s environment...", fileName, bucketNameEntry.Text, shared.GetCurrentEnvironment()))
			appendToConsole(fmt.Sprintf("File Path: %s", filePath))
			err = awsS3Functions.PutObject(svc, fileName, filePath, bucketNameEntry.Text)
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to upload file: %v", err))
				return
			}
			appendToConsole(fmt.Sprintf("File %s uploaded successfully to bucket %s.", fileName, bucketNameEntry.Text))

		}()
	})

	headObjectMetadataButton := widget.NewButton("Head Object Metadata", func() {
		go func() {
			apiKey, apiSecret := "", ""
			if filePath == "" || fileName == "" {
				appendToConsole("Please select a file to query.")
				return
			}
			if bucketNameEntry.Text == "" {
				appendToConsole("Please enter a bucket name.")
				return
			}
			svc, err := awsS3Functions.CreateServiceClient(apiKey, apiSecret)
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to create S3 client: %v", err))
				return
			}
			appendToConsole(fmt.Sprintf("Getting metadata for object %s in bucket %s in %s environment...", fileName, bucketNameEntry.Text, shared.GetCurrentEnvironment()))
			headResult, err := awsS3Functions.HeadObject(svc, bucketNameEntry.Text, fileName)
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to get object metadata: %v", err))
				return
			}
			appendToConsole(fmt.Sprintf("HEAD Object metadata - Size: %d, LastModified: %v, ETag: %s",
				*headResult.ContentLength, *headResult.LastModified, *headResult.ETag))
			appendToConsole(fmt.Sprintf("Metadata for object %s in bucket %s retrieved successfully.", fileName, bucketNameEntry.Text))
		}()
	})

	deleteFileButton := widget.NewButton("Delete Object", func() {
		go func() {
			apiKey, apiSecret := "", ""
			if filePath == "" || fileName == "" {
				appendToConsole("Please select a file to delete.")
				return
			}
			if bucketNameEntry.Text == "" {
				appendToConsole("Please enter a bucket name.")
				return
			}
			svc, err := awsS3Functions.CreateServiceClient(apiKey, apiSecret)
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to create S3 client: %v", err))
				return
			}
			err = awsS3Functions.DeleteObject(svc, bucketNameEntry.Text, fileName)
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to delete object: %v", err))
				return
			}
			appendToConsole(fmt.Sprintf("Object %s deleted successfully from bucket %s.", fileName, bucketNameEntry.Text))
		}()
	})

	getObjectButton := widget.NewButton("Get Object", func() {
		go func() {
			apiKey, apiSecret := "", ""
			if filePath == "" || fileName == "" {
				appendToConsole("Please select a file to download.")
				return
			}
			if bucketNameEntry.Text == "" {
				appendToConsole("Please enter a bucket name.")
				return
			}
			svc, err := awsS3Functions.CreateServiceClient(apiKey, apiSecret)
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to create S3 client: %v", err))
				return
			}
			downloadFolderPath := testingToolkit.CurrPath() + "/featureTesting/awsS3FilesTest/downloads"
			downloadFolderPath = strings.ReplaceAll(downloadFolderPath, "testAutomationSuiteGO/testAutomationSuiteGO", "testAutomationSuiteGO")
			appendToConsole(fmt.Sprintf("Downloading object %s from bucket %s in %s environment...", fileName, bucketNameEntry.Text, shared.GetCurrentEnvironment()))
			err = awsS3Functions.GetObjectFullDownload(svc, bucketNameEntry.Text, fileName, downloadFolderPath)
			if err != nil {
				appendToConsole(fmt.Sprintf("Failed to download object: %v", err))
				return
			}
			appendToConsole(fmt.Sprintf("Object %s downloaded successfully to %s.", fileName, downloadFolderPath))
		}()
	})

	bottomClearConsoleButton := widget.NewButton("Clear Console", func() {
		consoleOutput.SetText("")
	})

	leftContent := container.NewVBox(
		widget.NewLabel("Enter Bucket Name:"),
		bucketNameEntry,
		createBucketButton,
		browseContainer,
		uploadFileButton,
		headObjectMetadataButton,
		getObjectButton,
		deleteFileButton,
		layout.NewSpacer(),
		bottomClearConsoleButton,
	)

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, consoleOutput)

	return contentWithConsole
}

const (
	FeatureTestName = "awsS3FilesTest"
	VariablePath    = "/testData/featureTestingVariables/awsS3FilesTest.txt"
	DEVBUCKETSURL   = "https://uploads.devpinata.cloud/v3/buckets"
	PRDBUCKETSURL   = "https://uploads.pinata.cloud/v3/buckets"
	DEVS3URL        = "https://s3-uploads.devpinata.cloud/"
	PRDS3URL        = "https://s3-uploads.pinata.cloud/"
)

var DEVAPIKEY1 string = testingToolkit.GetENVVariable(testingToolkit.CurrPath()+VariablePath, "DEVAPIKEY1")
var DEVAPISECRET1 string = testingToolkit.GetENVVariable(testingToolkit.CurrPath()+VariablePath, "DEVAPISECRET1")
var DEVJWT1 string = testingToolkit.GetENVVariable(testingToolkit.CurrPath()+VariablePath, "DEVJWT1")
var PRDAPIKEY1 string = testingToolkit.GetENVVariable(testingToolkit.CurrPath()+VariablePath, "PRDAPIKEY1")
var PRDAPISECRET1 string = testingToolkit.GetENVVariable(testingToolkit.CurrPath()+VariablePath, "PRDAPISECRET1")
var PRDJWT1 string = testingToolkit.GetENVVariable(testingToolkit.CurrPath()+VariablePath, "PRDJWT1")
