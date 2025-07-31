package featureTestingFunctionRegistry

import (
	"reflect"
	"testAutomationSuiteGO/featureTesting/awsS3FeatureTest/awsS3FeatureTest"
	"testAutomationSuiteGO/featureTesting/zTests/zTests"
)

var FunctionRegistry = map[string]reflect.Value{
	"awsS3FeatureTest.GenerateWindowContent": reflect.ValueOf(awsS3FeatureTest.GenerateWindowContent),
	"zTests.GenerateWindowContent":           reflect.ValueOf(zTests.GenerateWindowContent),
}
