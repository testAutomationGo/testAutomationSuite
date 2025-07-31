package testData

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	testingToolkit "testAutomationSuiteGO/internal/testingToolkit"

	"github.com/jung-kurt/gofpdf"
)

func CreateTextFileSizeInMB(folderPath, tcNumber string, sizeInMB int) string {
	fileName := "Text_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".txt"
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	baseContent := "TC Number: " + tcNumber + ". This is a text file. "
	baseContent += testingToolkit.GetAlNumString(10) + testingToolkit.GetNumericString(5)
	baseContent += testingToolkit.CurrentTimeForLogging()

	baseContentSize := len(baseContent)

	chunkSize := 1024 * 1024
	remainingChunkSize := chunkSize - baseContentSize
	if remainingChunkSize < 0 {
		remainingChunkSize = 0
	}

	for i := 0; i < sizeInMB; i++ {
		content := baseContent
		if remainingChunkSize > 0 {
			content += testingToolkit.GetAlNumString(remainingChunkSize)
		}
		_, err := file.WriteString(content)
		if err != nil {
			log.Fatalf("Failed to write to file: %v\n", err)
		}
	}

	return fileName
}

func CreateTextFileSizeInKB(folderPath, tcNumber string, sizeInKB int) string {
	fileName := "Text_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".txt"
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	baseContent := "TC Number: " + tcNumber + ". This is a text file. "
	baseContent += testingToolkit.GetAlNumString(10) + testingToolkit.GetNumericString(5)
	baseContent += testingToolkit.CurrentTimeForLogging()

	baseContentSize := len(baseContent)

	chunkSize := 1024
	remainingChunkSize := chunkSize - baseContentSize
	if remainingChunkSize < 0 {
		remainingChunkSize = 0
	}

	for i := 0; i < sizeInKB; i++ {
		content := baseContent
		if remainingChunkSize > 0 {
			content += testingToolkit.GetAlNumString(remainingChunkSize)
		}
		_, err := file.WriteString(content)
		if err != nil {
			log.Fatalf("Failed to write to file: %v\n", err)
		}
	}

	return fileName
}

func CreateTextFile(folderPath, tcNumber string) string {
	fileName := "Text_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".txt"
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString("TC Number: " + tcNumber + ". This is a text file. ")
	file.WriteString(testingToolkit.GetAlNumString(10) + testingToolkit.GetNumericString(5))
	file.WriteString(testingToolkit.CurrentTimeForLogging())
	return fileName
}

func CreateTextFileWithSpecificText(folderPath, tcNumber, text string) string {
	fileName := "Text_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".txt"
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString(text)
	return fileName
}

func CreateTextFileWithSpecificFileNameSize(folderPath, tcNumber string, size int) string {
	//no more than 253
	if size > 239 {
		size = 239
	}
	fileName := "Text_" + tcNumber + "_" + testingToolkit.GetLowerAlphaString(size) + ".txt"
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString("TC Number: " + tcNumber + ". This is a text file. ")
	file.WriteString(testingToolkit.GetAlNumString(10) + testingToolkit.GetNumericString(5))
	file.WriteString(testingToolkit.CurrentTimeForLogging())
	return fileName
}

func CreateTextFileWithSpecificFileName(folderPath, fileName string) {
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating file:", err)
	}
	defer file.Close()
}

func CreateJSONFile(folderPath, tcNumber string) string {
	fileName := "JSON_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".json"
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString("{\n")
	file.WriteString("  \"type\": \"JSON File\",\n")
	file.WriteString("  \"createdAt\": \"" + testingToolkit.CurrentTimeForNamingWithMS() + "\",\n")
	file.WriteString("  \"message\": \"" + tcNumber + " Hello, this is a json file.\",\n")
	file.WriteString("  \"data\": \"" + testingToolkit.GetAlNumString(10) + testingToolkit.GetNumericString(5) + "\"\n")
	file.WriteString("}")
	return fileName
}

func CreateJSONFileReturnNameAndJson(folderPath, tcNumber string) (string, string) {
	fileName := "JSON_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".json"
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var s strings.Builder
	s.WriteString("{\n")
	s.WriteString("  \"type\": \"JSON File\",\n")
	s.WriteString("  \"createdAt\": \"" + testingToolkit.CurrentTimeForNamingWithMS() + "\",\n")
	s.WriteString("  \"message\": \"" + tcNumber + " Hello, this is a json file.\",\n")
	s.WriteString("  \"data\": \"" + testingToolkit.GetAlNumString(10) + testingToolkit.GetNumericString(5) + "\"\n")
	s.WriteString("}")
	file.WriteString(s.String())
	return fileName, s.String()
}

func CreateJSONFileFromString(folderPath, tcNumber, jsonString string) string {
	fileName := "JSON_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".json"
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString(jsonString)
	return fileName
}

func CreateJPEGFile(width, height int, folderPath, tcNumber string) string {
	fileName := "JPEG_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".jpeg"
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(rand.Intn(255))
			g := uint8(rand.Intn(255))
			b := uint8(rand.Intn(255))
			a := uint8(rand.Intn(255))
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	outFile, err := os.Create(filepath.Join(folderPath, "/", fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	jpeg.Encode(outFile, img, nil)
	return fileName
}

func CreateJPGFile(width, height int, folderPath, tcNumber string) string {
	fileName := "JPG_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".jpg"
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(rand.Intn(255))
			g := uint8(rand.Intn(255))
			b := uint8(rand.Intn(255))
			a := uint8(rand.Intn(255))
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	outFile, err := os.Create(filepath.Join(folderPath, "/", fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	jpeg.Encode(outFile, img, nil)
	return fileName
}

func CreatePNGFile(width, height int, folderPath, tcNumber string) string {
	fileName := "PNG_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".png"
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(rand.Intn(255))
			g := uint8(rand.Intn(255))
			b := uint8(rand.Intn(255))
			a := uint8(rand.Intn(255))
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	outFile, err := os.Create(filepath.Join(folderPath, "/", fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	jpeg.Encode(outFile, img, nil)
	return fileName
}

func CreatePDFFile(folderPath, tcNumber string) string {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	content := tcNumber + " This is a PDF file. " + testingToolkit.CurrentTimeForNamingWithMS() + " " + testingToolkit.GetAlNumString(10) + testingToolkit.GetNumericString(5)
	pdf.Cell(40, 10, content)
	fileName := "PDF_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".pdf"
	fullPath := folderPath + "/" + fileName
	err := pdf.OutputFileAndClose(fullPath)
	if err != nil {
		log.Fatal(err)
	}
	return fileName
}

func CreateHTMLFile(folderPath, tcNumber string) string {
	fileName := "HTML_" + tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS() + ".html"
	filePath := filepath.Join(folderPath, "/", fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	currentTime := testingToolkit.CurrentTimeForNamingWithMS()
	file.WriteString("<html>\n")
	file.WriteString("<div id=\"text\">" + tcNumber + " This is some html version_" + currentTime + " " + testingToolkit.GetAlNumString(10) + testingToolkit.GetNumericString(5) + "</div>\n")
	file.WriteString("<button id=\"button\">push</button>\n")
	file.WriteString("<script>\n")
	file.WriteString("// Get a reference to the button element\n")
	file.WriteString("var button = document.getElementById(\"button\");\n")
	file.WriteString("\n")
	file.WriteString("// Get a reference to the text element\n")
	file.WriteString("var text = document.getElementById(\"text\");\n")
	file.WriteString("\n")
	file.WriteString("// Add a click event listener to the button\n")
	file.WriteString("button.addEventListener(\"click\", function() {\n")
	file.WriteString("  // Check the current color of the text\n")
	file.WriteString("  if (text.style.color == \"red\") {\n")
	file.WriteString("    // If it is red, change it to green\n")
	file.WriteString("    text.style.color = \"green\";\n")
	file.WriteString("  } else {\n")
	file.WriteString("    // Otherwise, change it to red\n")
	file.WriteString("    text.style.color = \"red\";\n")
	file.WriteString("  }\n")
	file.WriteString("});\n")
	file.WriteString("</script>\n")
	file.WriteString("</html>\n")
	return fileName
}

/*
func CreateMP4File(folderPath, tcNumber string) string {
	mp4FileName := tcNumber + testingToolkit.CurrentTimeForNamingWithMS() + ".mp4"
	var fileNames []string
	for i := 0; i < 10; i++ {
		fileNames = append(fileNames, CreatePNGFile(30, 30, folderPath, tcNumber))
	}

	listFile, err := os.Create("filelist.txt")
	if err != nil {
		fmt.Println("Error creating file list:", err)
	}
	defer listFile.Close()

	for _, fileName := range fileNames {
		_, err := listFile.WriteString(fmt.Sprintf("file '%s'\n", fileName))
		if err != nil {
			fmt.Println("Error writing to file list:", err)
		}
	}

	// Check if FFmpeg is available
	_, err = exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println("FFmpeg not found in PATH")
	}

	// Create the FFmpeg command using ffmpeg-go
	err = ffmpeg.Input("filelist.txt", ffmpeg.KwArgs{"f": "concat", "safe": 0}).
		Output("output.mp4", ffmpeg.KwArgs{"c:v": "libx264", "pix_fmt": "yuv420p"}).
		OverWriteOutput(). // Overwrite output file if it exists
		Run()
	if err != nil {
		fmt.Printf("Error executing FFmpeg: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Video created successfully: output.mp4")

	for _, fileName := range fileNames {
		err := os.Remove(filepath.Join(folderPath, fileName))
		if err != nil {
			log.Fatal(err)
		}
	}

	return mp4FileName
}*/

func CreateFileByType(fileType, folderPath, tcNumber string) string {
	fileType = strings.ToLower(fileType)
	switch fileType {
	case "txt":
		return CreateTextFile(folderPath, tcNumber)
	case "json":
		return CreateJSONFile(folderPath, tcNumber)
	case "jpeg":
		return CreateJPEGFile(100, 100, folderPath, tcNumber)
	case "jpg":
		return CreateJPGFile(100, 100, folderPath, tcNumber)
	case "png":
		return CreatePNGFile(100, 100, folderPath, tcNumber)
	case "pdf":
		return CreatePDFFile(folderPath, tcNumber)
	case "html":
		return CreateHTMLFile(folderPath, tcNumber)
	default:
		return ""
	}
}

func CreateAFolderWith3TextFiles(folderPath, tcNumber string) string {
	folderName := tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS()
	err := os.Mkdir(folderPath+"/"+folderName, 0755)
	if err != nil {
		return "UnableToCreateDir"
	}
	for i := range 3 {
		CreateTextFile(folderPath+"/"+folderName, tcNumber+testingToolkit.ConvertIntToString(i))
	}
	return folderName
}

func CreateAFolderWithXNumberOfTextFiles(folderPath string, x int, tcNumber string) string {
	folderName := tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS()
	err := os.Mkdir(folderPath+"/"+folderName, 0755)
	if err != nil {
		return "UnableToCreateDir"
	}
	for i := range x {
		CreateTextFile(folderPath+"/"+folderName, tcNumber+testingToolkit.ConvertIntToString(i))
	}
	return folderName
}

func CreateAFolderWithXNumberOfPNGFiles(folderPath string, x int, tcNumber string) string {
	folderName := tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS()
	err := os.Mkdir(folderPath+"/"+folderName, 0755)
	if err != nil {
		return "UnableToCreateDir"
	}
	for i := range x {
		CreatePNGFile(5, 5, folderPath+"/"+folderName, tcNumber+testingToolkit.ConvertIntToString(i))
	}
	return folderName
}

func CreateFolderWithXSubLevelFolders(folderPath string, numberOfFilesInMainFolder, numberOfSubFolders, numberOfFilesInSubFolders int, tcNumber string) (string, []string, []string, [][]string) {
	folderName := tcNumber + "_" + testingToolkit.CurrentTimeForNamingWithMS()
	var mainFilesNames []string
	var subFoldersNames []string
	subFoldersAndFiles := make([][]string, numberOfSubFolders)
	for i := range numberOfFilesInMainFolder {
		subFoldersAndFiles[i] = make([]string, numberOfFilesInSubFolders)
	}
	err := os.Mkdir(folderPath+"/"+folderName, 0755)
	if err != nil {
		return "UnableToCreateDir", nil, nil, nil
	}
	for i := range numberOfFilesInMainFolder {
		fileName := CreateTextFile(folderPath+"/"+folderName, tcNumber+testingToolkit.ConvertIntToString(i))
		mainFilesNames = append(mainFilesNames, fileName)
	}
	for i := range numberOfSubFolders {
		subFolderName := folderName + "_" + testingToolkit.ConvertIntToString(i)
		subFoldersNames = append(subFoldersNames, subFolderName)
		err := os.Mkdir(folderPath+"/"+folderName+"/"+subFolderName, 0755)
		if err != nil {
			return "UnableToCreateDir", nil, nil, nil
		}
		for j := range numberOfFilesInSubFolders {
			fileName := CreateTextFile(folderPath+"/"+folderName+"/"+subFolderName, tcNumber+testingToolkit.ConvertIntToString(j))
			subFoldersAndFiles[i][j] = fileName
		}
	}
	return folderName, mainFilesNames, subFoldersNames, subFoldersAndFiles
}
