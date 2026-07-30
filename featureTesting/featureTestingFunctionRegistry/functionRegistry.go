package featureTestingFunctionRegistry

import (
	"reflect"
	"testAutomationSuiteGO/featureTesting/zTests/zTests"
)

var FunctionRegistry = map[string]reflect.Value{
	"zTests.GenerateWindowContent": reflect.ValueOf(zTests.GenerateWindowContent),
}
