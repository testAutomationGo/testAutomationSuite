package apiTesting

import (
	"reflect"
	"strings"
	"sync"
	"testAutomationSuiteGO/internal/awsS3Functions"
	"testAutomationSuiteGO/internal/logger"
	"testAutomationSuiteGO/internal/reporting"
	"testAutomationSuiteGO/internal/testCaseStructuring"
	"testAutomationSuiteGO/internal/testingToolkit"
)

type APITests struct{}

func (a *APITests) TC_API0001() {
	//"title": "Create A Bucket"
	tcNumber, expected, actual, awsKey, awsSecret := testCaseStructuring.TestCaseType1(testingToolkit.GetFunctionName())
	client, err := awsS3Functions.CreateServiceClient(awsKey, awsSecret)
	if err != nil {
		logger.Log("Create ServiceClient failed: "+err.Error(), tcNumber)
		reporting.AssertEquals(tcNumber, expected, actual, "Create ServiceClient failed: "+err.Error(), "")
		return
	}
	bucketName := testingToolkit.GetLowerAlphaNumString(12)
	logger.Log("Bucket Name: "+bucketName, tcNumber)
	createBucketOutput, err := awsS3Functions.CreateBucket(client, bucketName, tcNumber)
	if err != nil {
		logger.Log("Create Bucket failed: "+err.Error(), tcNumber)
		reporting.AssertEquals(tcNumber, expected, actual, "Create Bucket failed: "+err.Error(), "")
		return
	}
	logger.Log("Create Bucket Result: "+createBucketOutput, tcNumber)
	actual = expected
	err = awsS3Functions.DeleteBucket(client, bucketName)
	if err != nil {
		logger.Log("Delete Bucket failed: "+err.Error(), tcNumber)
	}
	reporting.AssertEquals(tcNumber, expected, actual, "Create Bucket succeeded", reporting.Format("Bucket Name", bucketName))
}

func (a *APITests) TC_API0002() {
	//"title": "List Buckets"
	tcNumber, expected, actual, awsKey, awsSecret := testCaseStructuring.TestCaseType1(testingToolkit.GetFunctionName())
	client, err := awsS3Functions.CreateServiceClient(awsKey, awsSecret)
	if err != nil {
		logger.Log("Create ServiceClient failed: "+err.Error(), tcNumber)
		reporting.AssertEquals(tcNumber, expected, actual, "Create ServiceClient failed: "+err.Error(), "")
		return
	}
	bucket1 := "test-bucket-1-" + testingToolkit.GetLowerAlphaNumString(6)
	bucket2 := "test-bucket-2-" + testingToolkit.GetLowerAlphaNumString(6)
	_, err = awsS3Functions.CreateBucket(client, bucket1, tcNumber)
	if err != nil {
		logger.Log("Create Bucket 1 failed: "+err.Error(), tcNumber)
		reporting.AssertEquals(tcNumber, expected, actual, "Create Bucket 1 failed: "+err.Error(), "")
		return
	}
	_, err = awsS3Functions.CreateBucket(client, bucket2, tcNumber)
	if err != nil {
		logger.Log("Create Bucket 2 failed: "+err.Error(), tcNumber)
		reporting.AssertEquals(tcNumber, expected, actual, "Create Bucket 2 failed: "+err.Error(), "")
		return
	}
	buckets, err := awsS3Functions.ListBuckets(client, tcNumber)
	if err != nil {
		logger.Log("List Buckets failed: "+err.Error(), tcNumber)
		reporting.AssertEquals(tcNumber, expected, actual, "List Buckets failed: "+err.Error(), "")
		return
	}
	outputString := "List Buckets Result: "
	for _, bucket := range buckets {
		logger.Log("Bucket Listed, Name: "+bucket, tcNumber)
		outputString += bucket + ", "
	}
	actual = expected
	err = awsS3Functions.DeleteBucket(client, bucket1)
	if err != nil {
		logger.Log("Delete Bucket 1 failed: "+err.Error(), tcNumber)
	}
	err = awsS3Functions.DeleteBucket(client, bucket2)
	if err != nil {
		logger.Log("Delete Bucket 2 failed: "+err.Error(), tcNumber)
	}
	reporting.AssertEquals(tcNumber, expected, actual, "List Buckets succeeded", reporting.Format("Buckets", outputString))
}

func ExecuteAPITests() {
	tester := &APITests{}
	t := reflect.TypeOf(tester)
	v := reflect.ValueOf(tester)
	var wg sync.WaitGroup
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if strings.HasPrefix(method.Name, "TC_") {
			wg.Add(1)
			go func(m reflect.Method) {
				method.Func.Call([]reflect.Value{v})
				wg.Done()
			}(method)
		}
	}
	wg.Wait()
}

func ExecuteSingleAPITest(testToRun string) {
	tester := &APITests{}
	t := reflect.TypeOf(tester)
	v := reflect.ValueOf(tester)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if strings.Contains(method.Name, testToRun) {
			method.Func.Call([]reflect.Value{v})
			return
		}
	}
}

func ExecuteAPITestsSequentially() {
	tester := &APITests{}
	t := reflect.TypeOf(tester)
	v := reflect.ValueOf(tester)
	for i := range t.NumMethod() {
		method := t.Method(i)
		if strings.HasPrefix(method.Name, "TC_") {
			method.Func.Call([]reflect.Value{v})
		}
	}
}
