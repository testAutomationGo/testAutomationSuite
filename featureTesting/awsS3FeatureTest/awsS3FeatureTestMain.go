package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"testAutomationSuiteGO/internal/awsS3Functions"
	"testAutomationSuiteGO/internal/testData"
	"testAutomationSuiteGO/internal/testingToolkit"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

/*
TO-DO:
XXX- [AbortMultipartUpload](https://docs.aws.amazon.com/AmazonS3/latest/API/API_AbortMultipartUpload.html)
XXX- [CompleteMultipartUpload](https://docs.aws.amazon.com/AmazonS3/latest/API/API_CompleteMultipartUpload.html)
XXX- [CreateMultipartUpload](https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateMultipartUpload.html)
XXX- [DeleteObject](https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteObject.html)
XXX- [DeleteObjects](https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteObjects.html)
XXX- [GetObject](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html)
XXX- [HeadObject](https://docs.aws.amazon.com/AmazonS3/latest/API/API_HeadObject.html)
XXX- [ListMultipartUploads](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListMultipartUploads.html)
XXX- [ListObjectsV2](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjectsV2.html)
XXX- [ListParts](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListParts.html)
XXX- [PutObject](https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html)
XXX- [UploadPart](https://docs.aws.amazon.com/AmazonS3/latest/API/API_UploadPart.html)

- [~~ListObjects~~](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html)
*/
func main() {
	env := "DEV"
	bucketName := testingToolkit.CurrentTimeForNamingWithMS()
	//RunFullAWSTestWithoutMultiPart(env, bucketName)
	RunAWSMultiPart(env, bucketName)
}

func RunAWSMultiPart(env, bucketName string) {
	variablePath := testingToolkit.CurrPath() + "/testData/featureTestingVariables/awsS3Functions.txt"
	filesFolder := testingToolkit.CurrPath() + "/featureTesting/awsS3Functions/files"
	var APIKEY1 string
	var APISECRET1 string
	//var JWT1 string
	if env == "DEV" {
		//s3URLPrefix = "https://s3.amazonaws.com"
		//s3URLPrefix = "https://s3-uploads.devpinata.cloud/"
		//APIKEY1 = testingToolkit.GetENVVariable(variablePath, "DEVAPIKEY1")
		//APISECRET1 = testingToolkit.GetENVVariable(variablePath, "DEVAPISECRET1")
		APIKEY1 = testingToolkit.GetENVVariable(variablePath, "AWSKEY")
		APISECRET1 = testingToolkit.GetENVVariable(variablePath, "AWSSECRET")
	}
	if env == "PRD" {
		//s3URLPrefix = "https://s3-uploads.pinata.cloud"
		APIKEY1 = testingToolkit.GetENVVariable(variablePath, "PRDAPIKEY1")
		APISECRET1 = testingToolkit.GetENVVariable(variablePath, "PRDAPISECRET1")
	}
	testFileName := testingToolkit.CurrentTimeForNamingWithMS() + "-awsMultipartTest.txt"
	fileSize := 10 * 1024 * 1024
	partSize := 5 * 1024 * 1024

	fmt.Println("=== S3 Multipart Upload Test Suite ===")

	svc, err := awsS3Functions.CreateServiceClient(APIKEY1, APISECRET1)
	if err != nil {
		fmt.Printf("Failed to create service client: %v\n", err)
		return
	}
	fmt.Println("Service client created successfully")

	fmt.Println("\n--- Creating large test file (10MB) ---")
	err = createLargeTestFile(testFileName, fileSize)
	if err != nil {
		fmt.Printf("Failed to create large test file: %v\n", err)
		return
	}
	fmt.Printf("Large test file created (%d bytes)\n", fileSize)

	fmt.Println("\n--- Creating bucket ---")
	_, err = awsS3Functions.CreateBucket(svc, bucketName, "")
	if err != nil {
		fmt.Printf("Create bucket failed: %v\n", err)
		RemoveFile(testFileName, filesFolder)
		return
	}
	fmt.Println("Bucket created successfully")

	fmt.Println("\n--- Testing CreateMultipartUpload ---")
	createResult, err := awsS3Functions.CreateMultipartUpload(svc, bucketName, testFileName)
	if err != nil {
		fmt.Printf("Create multipart upload failed: %v\n", err)
		RemoveFile(testFileName, filesFolder)
		return
	}
	uploadID := *createResult.UploadId
	fmt.Printf("Multipart upload created - Upload ID: %s\n", uploadID)

	fmt.Println("\n--- Opening file for multipart upload ---")
	file, err := os.Open(testingToolkit.CurrPath() + "/featureTesting/awsS3FilesTest/files/" + testFileName)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		awsS3Functions.AbortMultipartUpload(svc, bucketName, testFileName, uploadID)
		RemoveFile(testFileName, filesFolder)
		return
	}
	defer file.Close()

	fmt.Println("\n--- Testing UploadPart ---")
	var completedParts []types.CompletedPart
	partNumber := int32(1)
	totalParts := (fileSize + partSize - 1) / partSize

	for {
		buffer := make([]byte, partSize)
		bytesRead, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Failed to read file chunk: %v\n", err)
			awsS3Functions.AbortMultipartUpload(svc, bucketName, testFileName, uploadID)
			RemoveFile(testFileName, filesFolder)
			return
		}

		partData := bytes.NewReader(buffer[:bytesRead])

		fmt.Printf("  Uploading part %d/%d (%d bytes)...\n", partNumber, totalParts, bytesRead)
		uploadResult, err := awsS3Functions.UploadPart(svc, bucketName, testFileName, uploadID, partNumber, partData)
		if err != nil {
			fmt.Printf("Upload part %d failed: %v\n", partNumber, err)
			awsS3Functions.AbortMultipartUpload(svc, bucketName, testFileName, uploadID)
			RemoveFile(testFileName, filesFolder)
			return
		}

		completedParts = append(completedParts, types.CompletedPart{
			ETag:       uploadResult.ETag,
			PartNumber: aws.Int32(partNumber),
		})

		partNumber++
	}

	fmt.Printf("All %d parts uploaded successfully\n", len(completedParts))

	fmt.Println("\n--- Testing CompleteMultipartUpload ---")
	err = awsS3Functions.CompleteMultipartUpload(svc, bucketName, testFileName, uploadID, completedParts)
	if err != nil {
		fmt.Printf("Complete multipart upload failed: %v\n", err)
		awsS3Functions.AbortMultipartUpload(svc, bucketName, testFileName, uploadID)
		RemoveFile(testFileName, filesFolder)
		return
	}
	fmt.Println("Multipart upload completed successfully")

	fmt.Println("\n--- Verifying uploaded file ---")
	headResult, err := awsS3Functions.HeadObject(svc, bucketName, testFileName)
	if err != nil {
		fmt.Printf("Head object failed: %v\n", err)
	} else {
		fmt.Printf("✓ File verified - Size: %d bytes\n", headResult.ContentLength)
	}

	fmt.Println("\n--- Testing multipart upload abortion (with new upload) ---")
	abortTestFile := testingToolkit.CurrentTimeForNamingWithMS() + "abort-test-file.txt"
	err = createLargeTestFile(abortTestFile, 7*1024*1024)
	if err != nil {
		fmt.Printf("Failed to create abort test file: %v\n", err)
	} else {
		createResult2, err := awsS3Functions.CreateMultipartUpload(svc, bucketName, abortTestFile)
		if err != nil {
			fmt.Printf("Create second multipart upload failed: %v\n", err)
		} else {
			uploadID2 := *createResult2.UploadId
			fmt.Printf("Second multipart upload created for abortion test - Upload ID: %s\n", uploadID2)

			fmt.Println("\n--- Testing AbortMultipartUpload ---")
			err = awsS3Functions.AbortMultipartUpload(svc, bucketName, abortTestFile, uploadID2)
			if err != nil {
				fmt.Printf("Abort multipart upload failed: %v\n", err)
			} else {
				fmt.Println("✓ Multipart upload aborted successfully")
			}
		}
		os.Remove(abortTestFile)
	}

	fmt.Println("\n--- Listing final objects ---")
	listResult, err := awsS3Functions.ListObjects(svc, bucketName)
	if err != nil {
		fmt.Printf("List objects failed: %v\n", err)
	} else {
		fmt.Printf("Objects in bucket: %d\n", len(listResult.Contents))
		for _, obj := range listResult.Contents {
			fmt.Printf("  - %s (Size: %d bytes)\n", *obj.Key, obj.Size)
		}
	}

	//fmt.Println("\n--- Cleanup: Deleting uploaded file ---")
	//err = awsS3Functions.DeleteObject(svc, bucketName, testFileName)
	//if err != nil {
	//	fmt.Printf("Delete object failed: %v\n", err)
	//} else {
	//	fmt.Println("Uploaded file deleted successfully")
	//}

	//fmt.Println("\n--- Cleanup: Removing local test files ---")
	//RemoveFile(testFileName, filesFolder)
	//fmt.Println("Local test files cleaned up")

	fmt.Println("\n=== All multipart upload tests completed successfully! ===")
}

func createLargeTestFile(fileName string, size int) error {
	file, err := os.Create(testingToolkit.CurrPath() + "/featureTesting/awsS3Functions/files/" + fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	chunk := make([]byte, 1024)
	for i := 0; i < len(chunk); i++ {
		chunk[i] = byte('A' + (i % 26))
	}

	written := 0
	for written < size {
		toWrite := size - written
		if toWrite > len(chunk) {
			toWrite = len(chunk)
		}

		n, err := file.Write(chunk[:toWrite])
		if err != nil {
			return err
		}
		written += n
	}

	return nil
}

func RunFullAWSTestWithoutMultiPart(env, bucketName string) {
	variablePath := testingToolkit.CurrPath() + "/testData/featureTestingVariables/awsS3Functions.txt"
	filesFolder := testingToolkit.CurrPath() + "/featureTesting/awsS3Functions/files"
	var APIKEY1 string
	var APISECRET1 string
	//var JWT1 string
	if env == "DEV" {
		//s3URLPrefix = "https://s3-uploads.devpinata.cloud/"
		APIKEY1 = testingToolkit.GetENVVariable(variablePath, "DEVAPIKEY1")
		APISECRET1 = testingToolkit.GetENVVariable(variablePath, "DEVAPISECRET1")
	}
	if env == "PRD" {
		//s3URLPrefix = "https://s3-uploads.pinata.cloud"
		APIKEY1 = testingToolkit.GetENVVariable(variablePath, "PRDAPIKEY1")
		APISECRET1 = testingToolkit.GetENVVariable(variablePath, "PRDAPISECRET1")
	}
	fmt.Println("=== S3 Wrapper Test Suite ===")

	svc, err := awsS3Functions.CreateServiceClient(APIKEY1, APISECRET1)
	if err != nil {
		fmt.Printf("Failed to create service client: %v\n", err)
		return
	}
	fmt.Println("Service client created successfully")

	fmt.Println("\n--- Creating test file ---")
	testFileName := testData.CreateTextFile(filesFolder, "awsTestFile")

	fmt.Println("Test file created")

	fmt.Println("\n--- Testing CreateBucket ---")
	_, err = awsS3Functions.CreateBucket(svc, bucketName, "")
	if err != nil {
		fmt.Printf("Create bucket failed: %v\n", err)
		RemoveFile(testFileName, filesFolder)
		return
	}
	fmt.Println("Bucket created successfully")

	fmt.Println("\n--- Testing HeadBucket ---")
	err = awsS3Functions.HeadBucket(svc, bucketName)
	if err != nil {
		fmt.Printf("Head bucket failed: %v\n", err)
		RemoveFile(testFileName, filesFolder)
		return
	}
	fmt.Println("Bucket exists confirmed")

	fmt.Println("\n--- Testing UploadFile ---")
	err = awsS3Functions.PutObject(svc, testFileName, filesFolder, bucketName)
	if err != nil {
		fmt.Printf("Upload file failed: %v\n", err)
		RemoveFile(testFileName, filesFolder)
		return
	}
	fmt.Println("File uploaded successfully")

	fmt.Println("\n--- Testing ListObjects ---")
	listResult, err := awsS3Functions.ListObjects(svc, bucketName)
	if err != nil {
		fmt.Printf("List objects failed: %v\n", err)
		RemoveFile(testFileName, filesFolder)
		return
	}
	fmt.Printf("Objects listed - Found %d objects\n", len(listResult.Contents))
	for _, obj := range listResult.Contents {
		fmt.Printf("  - %s (Size: %d bytes)\n", *obj.Key, obj.Size)
	}

	fmt.Println("\n--- Testing HeadObject ---")
	headResult, err := awsS3Functions.HeadObject(svc, bucketName, testFileName)
	if err != nil {
		fmt.Printf("Head object failed: %v\n", err)
		RemoveFile(testFileName, filesFolder)
		return
	}
	fmt.Printf("Object metadata retrieved - Content Length: %d\n", headResult.ContentLength)

	fmt.Println("\n--- Testing GetObject ---")
	getResult, err := awsS3Functions.GetObject(svc, bucketName, testFileName)
	if err != nil {
		fmt.Printf("Get object failed: %v\n", err)
		RemoveFile(testFileName, filesFolder)
		return
	}
	defer getResult.Body.Close()

	content, err := io.ReadAll(getResult.Body)
	if err != nil {
		fmt.Printf("Failed to read object content: %v\n", err)
		RemoveFile(testFileName, filesFolder)
		return
	}
	fmt.Printf("Object retrieved - Content: %s\n", string(content))

	fmt.Println("\n--- Testing upload of additional files ---")
	var fileNames []string
	for i := 0; i < 3; i++ {
		fileName := testData.CreateTextFile(filesFolder, fmt.Sprintf("additionalFile%d", i+1))
		fileNames = append(fileNames, fileName)
	}
	for _, fileName := range fileNames {
		err = awsS3Functions.PutObject(svc, fileName, filesFolder, bucketName)
		if err != nil {
			fmt.Printf("Failed to upload %s: %v\n", fileName, err)
		} else {
			fmt.Printf("✓ %s uploaded successfully\n", fileName)
		}
	}

	fmt.Println("\n--- Testing ListObjects after multiple uploads ---")
	listResult, err = awsS3Functions.ListObjects(svc, bucketName)
	if err != nil {
		fmt.Printf("List objects failed: %v\n", err)
	} else {
		fmt.Printf("Objects listed - Found %d objects\n", len(listResult.Contents))
		var objectKeys []string
		for _, obj := range listResult.Contents {
			objectKeys = append(objectKeys, *obj.Key)
			fmt.Printf("  - %s\n", *obj.Key)
		}

		fmt.Println("\n--- Testing DeleteObjects (batch delete) ---")
		err = awsS3Functions.DeleteObjects(svc, bucketName, objectKeys)
		if err != nil {
			fmt.Printf("Delete objects failed: %v\n", err)
		} else {
			fmt.Printf("All %d objects deleted successfully\n", len(objectKeys))
		}
	}

	fmt.Println("\n--- Verifying all objects deleted ---")
	listResult, err = awsS3Functions.ListObjects(svc, bucketName)
	if err != nil {
		fmt.Printf("List objects failed: %v\n", err)
	} else {
		fmt.Printf("Objects remaining after deletion: %d\n", len(listResult.Contents))
	}

	fmt.Println("\n--- Testing single DeleteObject ---")
	finalTestFileName := testData.CreateTextFile(filesFolder, "final-test")
	if err == nil {
		err = awsS3Functions.PutObject(svc, finalTestFileName, filesFolder, bucketName)
		if err == nil {
			err = awsS3Functions.DeleteObject(svc, bucketName, "final-test.txt")
			if err != nil {
				fmt.Printf("Delete single object failed: %v\n", err)
			} else {
				fmt.Println("Single object deleted successfully")
			}
		}
		RemoveFile(finalTestFileName, filesFolder)
	}

	fmt.Println("\n--- Cleanup: Removing local test files ---")
	RemoveFile(testFileName, filesFolder)

	for _, fileName := range fileNames {
		RemoveFile(fileName, filesFolder)
		if err != nil {
			fmt.Printf("Failed to remove local file %s: %v\n", fileName, err)
		}
	}

	fmt.Println("Local test files cleaned up")

	fmt.Println("\n=== All tests completed successfully! ===")
}

func RemoveFile(fileName, folderPath string) {
	filePath := folderPath + "/" + fileName
	filePath = testingToolkit.RemoveDuplicateFolders(filePath)
	err := os.Remove(filePath)
	if err != nil {
		fmt.Printf("Failed to remove file %s: %v\n", fileName, err)
	} else {
		fmt.Printf("File %s removed successfully.\n", fileName)
	}
}
