package fileCreatorUI

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testAutomationSuiteGO/internal/testingToolkit"
	"time"
)

type FileGenerator struct {
	Size int64  // size in MB
	Path string // output path
}

func NewFileGenerator() *FileGenerator {
	return &FileGenerator{
		Size: 1024,
		Path: filepath.Join("files", "testfile.dat"),
	}
}

// Generate creates the file according to the configuration
func (fg *FileGenerator) Generate() error {
	// Convert MB to bytes
	sizeInBytes := fg.Size * 1024 * 1024

	startTime := time.Now()

	// Create the file
	file, err := os.OpenFile(fg.Path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Create a 1MB buffer for random data
	bufferSize := 1024 * 1024 // 1MB
	buffer := make([]byte, bufferSize)

	// Write random data in chunks
	bytesWritten := int64(0)
	for bytesWritten < sizeInBytes {
		// Fill buffer with random data
		_, err := rand.Read(buffer)
		if err != nil {
			return fmt.Errorf("error generating random data: %v", err)
		}

		// Calculate remaining bytes to write
		remainingBytes := sizeInBytes - bytesWritten
		writeSize := int64(bufferSize)
		if remainingBytes < int64(bufferSize) {
			fmt.Print(writeSize)
			writeSize = remainingBytes
			buffer = buffer[:writeSize]
		}

		// Write the buffer
		n, err := file.Write(buffer)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
		bytesWritten += int64(n)

		// Print progress every 10%
		progress := float64(bytesWritten) / float64(sizeInBytes) * 100
		if int(progress)%10 == 0 {
			fmt.Printf("Progress: %.0f%%\n", progress)
		}
	}

	duration := time.Since(startTime)
	speedMBps := float64(sizeInBytes) / duration.Seconds() / 1024 / 1024

	fmt.Printf("\nFile created successfully!\n")
	fmt.Printf("Path: %s\n", fg.Path)
	fmt.Printf("Size: %d MB\n", fg.Size)
	fmt.Printf("Time taken: %.2f seconds\n", duration.Seconds())
	fmt.Printf("Speed: %.2f MB/s\n", speedMBps)

	return nil
}

func LargeFileCreator(fileSizeInMB int64, fileName string) {
	// Create and configure generator
	generator := NewFileGenerator()
	rootProjectPath := testingToolkit.CurrPath()
	appOutputFullPath := rootProjectPath + "/testData/appOutput/" + fileName + ".dat"
	generator.Size = fileSizeInMB
	generator.Path = appOutputFullPath

	// Generate the file
	if err := generator.Generate(); err != nil {
		fmt.Printf("Error generating file: %v\n", err)
		os.Exit(1)
	}

	testingToolkit.DelaySeconds(1)

	copyFile(appOutputFullPath, rootProjectPath+"featureTesting/resumableUpload/files/"+fileName+".dat")
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		log.Println("Unable to open data file", err)
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		log.Println("Unable to create data file", err)
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
