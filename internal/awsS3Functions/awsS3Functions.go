package awsS3Functions

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testAutomationSuiteGO/internal/logger"
	"testAutomationSuiteGO/internal/testingToolkit"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func CreateServiceClient(apiKey, apiSecret string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			apiKey,
			apiSecret,
			"",
		)),
	)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		return nil, err
	}
	svc := s3.NewFromConfig(cfg)
	return svc, nil
}

func CreateBucket(svc *s3.Client, bucketName, tcNumber string) (string, error) {
	createBucketOutput, err := svc.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		logger.Log("Create Bucket failed: "+err.Error(), tcNumber)
		return "", err
	}
	location := ""
	if createBucketOutput.Location != nil {
		location = *createBucketOutput.Location
	}

	result := fmt.Sprintf("Bucket created successfully. Location: %s", location)
	return result, nil
}

func ListBuckets(svc *s3.Client, tcNumber string) ([]string, error) {
	listBucketsOutput, err := svc.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		logger.Log("List Buckets failed: "+err.Error(), tcNumber)
		return nil, err
	}

	var bucketNames []string
	for _, bucket := range listBucketsOutput.Buckets {
		if bucket.Name != nil {
			bucketNames = append(bucketNames, *bucket.Name)
		}
	}

	return bucketNames, nil
}

func DeleteBucket(svc *s3.Client, bucketName string) error {
	_, err := svc.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		fmt.Printf("Failed to delete bucket: %v\n", err)
		return err
	}
	fmt.Println("Bucket deleted.")
	return nil
}

func DeleteBucketWithContents(svc *s3.Client, bucketName string) error {
	paginator := s3.NewListObjectsV2Paginator(svc, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("Failed to list objects: %v\n", err)
			return err
		}

		if len(page.Contents) == 0 {
			break
		}

		var objectsToDelete []types.ObjectIdentifier
		for _, obj := range page.Contents {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		_, err = svc.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: objectsToDelete,
			},
		})
		if err != nil {
			fmt.Printf("Failed to delete objects: %v\n", err)
			return err
		}
	}

	_, err := svc.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		fmt.Printf("Failed to delete bucket: %v\n", err)
		return err
	}
	fmt.Println("Bucket deleted.")
	return nil
}

func HeadBucket(svc *s3.Client, bucketName string) error {
	_, err := svc.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		fmt.Printf("Failed to check bucket existence: %v\n", err)
		return err
	}
	fmt.Println("Bucket exists.")
	return nil
}

func ListObjects(svc *s3.Client, bucketName string) (*s3.ListObjectsV2Output, error) {
	result, err := svc.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		fmt.Printf("Failed to list objects: %v\n", err)
		return nil, err
	}
	fmt.Println("Objects listed.")
	return result, nil
}

func PutObject(svc *s3.Client, fileName, fileFolder, bucketName string) error {
	path := fileFolder + "/" + fileName
	path = testingToolkit.RemoveDuplicateFolders(path)

	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return err
	}
	defer file.Close()

	objectKey := fileName
	_, err = svc.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		return err
	}
	fmt.Println("File Uploaded.")
	return nil
}

func GetObject(svc *s3.Client, bucketName, objectKey string) (*s3.GetObjectOutput, error) {
	result, err := svc.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		fmt.Printf("Failed to get object: %v\n", err)
		return nil, err
	}
	fmt.Println("Object retrieved.")
	return result, nil
}

func GetObjectFullDownload(svc *s3.Client, bucketName, objectKey, downloadFolderPath string) error {
	result, err := svc.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		fmt.Printf("Failed to download file: %v\n", err)
		return err
	}
	defer result.Body.Close()
	downloadFolderPath = testingToolkit.RemoveDuplicateFolders(downloadFolderPath)
	if err := os.MkdirAll(downloadFolderPath, os.ModePerm); err != nil {
		fmt.Printf("Failed to create download directory: %v\n", err)
		return err
	}

	downloadFilePath := filepath.Join(downloadFolderPath, objectKey)
	localFile, err := os.Create(downloadFilePath)
	if err != nil {
		fmt.Printf("Failed to create local file: %v\n", err)
		return err
	}
	defer localFile.Close()

	written, err := io.Copy(localFile, result.Body)
	if err != nil {
		fmt.Printf("Failed to write to local file: %v\n", err)
		return err
	}
	fmt.Printf("Successfully downloaded %d bytes to %s\n", written, downloadFilePath)

	return nil
}

func HeadObject(svc *s3.Client, bucketName, objectKey string) (*s3.HeadObjectOutput, error) {
	result, err := svc.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		fmt.Printf("Failed to get object metadata: %v\n", err)
		return nil, err
	}
	fmt.Println("Object metadata retrieved.")
	return result, nil
}

func DeleteObject(svc *s3.Client, bucketName, objectKey string) error {
	_, err := svc.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		fmt.Printf("Failed to delete object: %v\n", err)
		return err
	}
	fmt.Println("Object deleted.")
	return nil
}

func DeleteObjects(svc *s3.Client, bucketName string, objectKeys []string) error {
	var objects []types.ObjectIdentifier
	for _, key := range objectKeys {
		objects = append(objects, types.ObjectIdentifier{
			Key: aws.String(key),
		})
	}

	_, err := svc.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		fmt.Printf("Failed to delete objects: %v\n", err)
		return err
	}
	fmt.Println("Objects deleted.")
	return nil
}

func CreateMultipartUpload(svc *s3.Client, bucketName, objectKey string) (*s3.CreateMultipartUploadOutput, error) {
	result, err := svc.CreateMultipartUpload(context.TODO(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		fmt.Printf("Failed to create multipart upload: %v\n", err)
		return nil, err
	}
	fmt.Println("Multipart upload created.")
	return result, nil
}

func UploadPart(svc *s3.Client, bucketName, objectKey, uploadID string, partNumber int32, body io.Reader) (*s3.UploadPartOutput, error) {
	result, err := svc.UploadPart(context.TODO(), &s3.UploadPartInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(objectKey),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(partNumber),
		Body:       body,
	})
	if err != nil {
		fmt.Printf("Failed to upload part: %v\n", err)
		return nil, err
	}
	fmt.Printf("Part %d uploaded.\n", partNumber)
	return result, nil
}

func CompleteMultipartUpload(svc *s3.Client, bucketName, objectKey, uploadID string, parts []types.CompletedPart) error {
	_, err := svc.CompleteMultipartUpload(context.TODO(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucketName),
		Key:      aws.String(objectKey),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	if err != nil {
		fmt.Printf("Failed to complete multipart upload: %v\n", err)
		return err
	}
	fmt.Println("Multipart upload completed.")
	return nil
}

func AbortMultipartUpload(svc *s3.Client, bucketName, objectKey, uploadID string) error {
	_, err := svc.AbortMultipartUpload(context.TODO(), &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucketName),
		Key:      aws.String(objectKey),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		fmt.Printf("Failed to abort multipart upload: %v\n", err)
		return err
	}
	fmt.Println("Multipart upload aborted.")
	return nil
}
