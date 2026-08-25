package openApiInternal

func TestCaseJSONTemplate(testSectorCamelCase, testSectorAbbr, pathName, operationKeyAllCaps, testCaseTitle, contentType, requestBodyName, responseName, responseBodyName, fileNumber, timeoutProfile, priority string) string {
	return ""
}

func MethodWithBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, fileNumberDigits, method, pathName, testCaseTitle, endpointPrefixGetter, endpointConstantReference, requestBodyLiteral, contentType, bodyKind, responseName, responseCondition, queryParamsLiteral, headerParamsLiteral, responseContractLiteral string) string {
	return ""
}

func MethodWithoutBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, fileNumberDigits, method, pathName, testCaseTitle, endpointPrefixGetter, endpointConstantReference, responseName, responseCondition, queryParamsLiteral, headerParamsLiteral, responseContractLiteral string) string {
	return ""
}

func httpMethodConstant(method string) string {
	return ""
}

func InsertTestCodeIntoTestCodeFile(testSectorCamelCase, testCaseCode string) error {
	return nil
}

func InsertNewCode(path, insert string) error {
	return nil
}

func extractGeneratedFunctionName(insert string) string {
	return ""
}

func TestSectorEndpointGenFile(testSectorCamelCase string, paths []string, endpointValues map[string]string) error {
	return nil
}

func ConvertPathToConstName(path string) string {
	return ""
}

func pascalCaseSegement(segament string) string {
	return ""
}
