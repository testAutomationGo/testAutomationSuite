package openApiInternal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testAutomationSuiteGO/internal/apiFunctions"
	"testAutomationSuiteGO/internal/logger"
	"testAutomationSuiteGO/internal/testingToolkit"
	"time"
)

var activeSchemaResolver *schemaResolver
var generatedScenarioNumbers map[string]string
var activeGenerationReport *generationReport
var generatedBodyFactoryCache map[string]map[string]generatedBodyFactory
var generatedResponseContractHelpers map[string]bool

type generationReport struct {
	OpenAPIURL          string               `json:"openAPIURL"`
	TestSector          string               `json:"testSector"`
	StartedAt           string               `json:"startedAt"`
	FinishedAt          string               `json:"finishedAt"`
	ReportPath          string               `json:"reportPath,omitempty"`
	PathsProcessed      int                  `json:"pathsProcessed"`
	OperationsProcessed int                  `json:"operationsProcessed"`
	PositiveScenarios   int                  `json:"positiveScenarios"`
	NegativeScenarios   int                  `json:"negativeScenarios"`
	JSONFilesCreated    int                  `json:"jsonFilesCreated"`
	JSONFilesOverwrote  int                  `json:"jsonFilesOverwrote"`
	CodeInsertions      int                  `json:"codeInsertions"`
	CodeSkips           int                  `json:"codeSkips"`
	Warnings            []string             `json:"warnings,omitempty"`
	Errors              []string             `json:"errors,omitempty"`
	WarningSummary      []messageSummaryItem `json:"warningSummary,omitempty"`
	ErrorSummary        []messageSummaryItem `json:"errorSummary,omitempty"`
}

type messageSummaryItem struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Sample   string `json:"sample"`
}

type schemaResolver struct {
	baseLocation    string
	rootDocument    any
	externalDocs    map[string]any
	failedDocuments map[string]bool
}

type generatedBodyParam struct {
	name       string
	typeName   string
	isOptional bool
	baseType   string
}

type generatedBodyFactory struct {
	name   string
	params []generatedBodyParam
}

func GenerateAllOpenAPITestCases(openAPIURL, testSectorCamelCase, testSectorAbbr, testSectorEndpointPrefix string) (int, error) {
	if err := ArchivePreviousFiles(testSectorCamelCase); err != nil {
		return 0, err
	}
	if err := ensureGenerationTargets(testSectorCamelCase); err != nil {
		return 0, err
	}
	if err := normalizeGeneratedTestFileImports(testSectorCamelCase); err != nil {
		return 0, err
	}

	activeGenerationReport = &generationReport{
		OpenAPIURL: openAPIURL,
		TestSector: testSectorCamelCase,
		StartedAt:  time.Now().UTC().Format(time.RFC3339),
		Warnings:   make([]string, 0),
		Errors:     make([]string, 0),
	}
	defer finalizeGenerationReport(testSectorCamelCase)

	data, responseBody, err := GetOpenAPIJSON(openAPIURL, testSectorCamelCase)
	if err != nil {
		reportError("failed to download or parse OpenAPI JSON: " + err.Error())
		return 0, err
	}

	activeSchemaResolver = newSchemaResolver(openAPIURL, data)
	generatedScenarioNumbers = make(map[string]string)
	generatedResponseContractHelpers = make(map[string]bool)

	defer func() {
		activeSchemaResolver = nil
		generatedScenarioNumbers = nil
		generatedBodyFactoryCache = nil
		generatedResponseContractHelpers = nil
		activeGenerationReport = nil
	}()

	GeneratedOpenAPIObjects(responseBody, openAPIURL, testSectorCamelCase)
	paths := GetOpenAPIPathNames(data)
	if activeGenerationReport != nil {
		activeGenerationReport.PathsProcessed = len(paths)
	}
	endpointValues := buildEndpointConstValues(data, paths)
	for _, pathName := range paths {
		GenerateOpenAPITestCases(testSectorCamelCase, testSectorAbbr, data, pathName, testSectorEndpointPrefix)
	}
	if err := TestSectorEndpointGenFile(testSectorCamelCase, paths, endpointValues); err != nil {
		reportError("failed to generate endpoints file: " + err.Error())
		return 0, err
	}
	if activeGenerationReport != nil {
		totalScenarios := activeGenerationReport.PositiveScenarios + activeGenerationReport.NegativeScenarios
		if totalScenarios > 0 {
			return totalScenarios, nil
		}
	}
	return len(paths), nil
}

func GetOpenAPIJSON(url, testSectorCamelCase string) (OpenAPISpec, string, error) {
	code, response, _, err := apiFunctions.DoRequest(apiFunctions.RequestOptions{
		Method: http.MethodGet,
		URL:    url,
		Headers: map[string]string{
			"accept": "*/*",
		},
	})
	if err != nil {
		return OpenAPISpec{}, "", err
	}

	if code != 200 {
		logger.Log("\"Error! Received non-200 response code: "+testingToolkit.ConvertIntToString(code), "GetOpenAPIJSON")
		return OpenAPISpec{}, "", fmt.Errorf("received non-200 response code: %d", code)
	}

	var data OpenAPISpec
	err = json.Unmarshal(response, &data)
	if err != nil {
		return OpenAPISpec{}, "", err
	}

	responseText := string(response)
	testingToolkit.PrintStringToFile(responseText, testingToolkit.CurrPath()+"/tests/"+testSectorCamelCase+"Tests/openAPIJson"+testingToolkit.CurrentTimeForNaming()+".json")
	return data, responseText, nil
}

func ensureGenerationTargets(testSectorCamelCase string) error {
	testFolder := testingToolkit.CurrPath() + "/tests/" + testSectorCamelCase + "Tests"
	testCasesFolder := testFolder + "/testCases"
	pageDomFolder := testingToolkit.CurrPath() + "/comInternal/pageDoms/" + testSectorCamelCase
	if err := os.MkdirAll(testCasesFolder, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(pageDomFolder, 0755); err != nil {
		return err
	}
	testCodeFile := testFolder + "/" + testSectorCamelCase + "Tests.go"
	if testingToolkit.VerifyFileIsPresent(testCodeFile) {
		return nil
	}
	return os.WriteFile(testCodeFile, []byte(buildBaseTestsFileContent(testSectorCamelCase)), 0644)
}

func buildBaseTestsFileContent(testSectorCamelCase string) string {
	packageName := testSectorCamelCase + "Tests"
	typeName := packageName
	executeName := "Execute" + upperFirst(testSectorCamelCase) + "Tests"

	return fmt.Sprintf(`package %s

import (
	"goSuite/comInternal/pageDoms/%s"
	"goSuite/comInternal/testCaseStructs"
	"goSuite/comInternal/testRunnerParameters"
	"goSuite/internal/apiFunctions"
	"goSuite/internal/logger"
	"goSuite/internal/reporting"
	"goSuite/internal/testingToolkit"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

type %s struct{}

func (s *%s) requestWithRecovery(method reflect.Method, receiver reflect.Value) {
	defer testCaseStructs.RecoverTestFailure(method.Name())
	method.Func.Call([]reflect.Value{receiver})
}

func %s() {
	tester := %s{}
	t := reflect.TypeOf(tester)
	v := reflect.ValueOf(tester)

	var wg sync.WaitGroup
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if strings.HasPrefix(method.Name, "TC_") {
			wg.Add(1)
			go func(m reflect.Method) {
				defer wg.Done()
				logger.Log("Running test case: "+m.Name, m.Name)
				testingToolkit.PrintFastStatusMessage("Running test case: " + m.Name)
				tester.runTestWithRecovery(m, v)
			}(method)
		}		
	}
	wg.Wait()
	%s.ResetCachedContractValidators()
}
`, packageName, testSectorCamelCase, typeName, typeName, executeName, typeName, testSectorCamelCase)
}

func upperFirst(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func normalizeGeneratedTestFileImports(testSectorCamelCase string) error {
	filePath := testingToolkit.CurrPath() + "/tests/" + testSectorCamelCase + "Tests/" + testSectorCamelCase + "Tests.go"
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(contentBytes)
	importBlock := canonicalGeneratedTestsImportBlock(testSectorCamelCase)
	re := regexp.MustCompile(`(?s)import\s*\(.*?\)`)
	if re.MatchString(content) {
		updated := re.ReplaceAllString(content, importBlock)
		if updated != content {
			return os.WriteFile(filePath, []byte(updated), 0644)
		}
		return nil
	}

	packageLine := "package " + testSectorCamelCase + "Tests"
	needle := packageLine + "\n\n"
	if strings.Contains(content, needle) {
		updated := strings.Replace(
			content,
			needle,
			packageLine+"\n\n"+importBlock+"\n\n",
			1,
		)
		return os.WriteFile(filePath, []byte(updated), 0644)
	}

	return nil
}

func canonicalGeneratedTestsImportBlock(testSectorCamelCase string) string {
	return fmt.Sprintf(`import (
	"goSuite/comInternal/pageDoms/%s"
	"goSuite/comInternal/testCaseStructuring"
	"goSuite/comInternal/testRunParameters"
	"goSuite/internal/apiFunctions"
	"goSuite/internal/logger"
	"goSuite/internal/reporting"
	"goSuite/internal/testingToolkit"
	"goSuite"
	"maps"
	"net/http"
	"reflect"
	"strings"
	"sync"
)`, testSectorCamelCase)
}

func GenerateOpenAPITestCases(testSectorCamelCase, testSectorAbbr string, data OpenAPISpec, pathName, endpointPrefixGetter string) {
	var methods []string
	for method := range data.Paths[pathName] {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	for _, method := range methods {
		if activeGenerationReport != nil {
			activeGenerationReport.OperationsProcessed++
		}
		operation := data.Paths[pathName][method]
		endpointConstReference := endpointConstantReference(testSectorCamelCase, pathName)
		methodNameOperationInAllCaps := strings.ToUpper(method)
		pathLevelParameters := pathItemParameters(data, pathName)
		resolvedPathName, queryParams, headerParams, requiredQueryNames, requiredHeaderNames := buildOperationParameterDetails(data, pathName, pathLevelParameters, operation)
		queryParamsLiteral := mapStringLiteral(queryParams)
		headerParamsLiteral := mapStringLiteral(headerParams)
		requestBodyName, requestBodyLiteral, requestContentType, requestBodyKind := buildRequestBodyDetails(data, operation, testSectorCamelCase, pathName, method)
		var responses []string
		for responseCode := range operation.Responses {
			responses = append(responses, responseCode)
		}
		sort.Strings(responses)
		if len(responses) > 1 {
			log.Println("More than 1 response, must create more test cases for path: " + pathName + " method: " + method)
			reportWarning("multiple responses for path " + pathName + " method " + method + "; only one positive scenario currently emitted")
		}
		if len(responses) == 0 {
			log.Println("no responses found for path: " + pathName + " method: " + method)
			reportWarning("no responses found for path " + pathName + " method " + method)
			continue
		}
		selectedResponseCode := responses[0]
		if slices.Contains(responses, "200") {
			selectedResponseCode = "200"
		}
		if selectedResponseCode != "200" {
			log.Println("Response code is not 200 for path: " + pathName + " method: " + method + ", using " + selectedResponseCode)
			reportWarning("non-200 positive response selected for path " + pathName + " method " + method + ": " + selectedResponseCode)
		}
		var responseBodyReference string
		if operation.Responses[selectedResponseCode].Content != nil {
			if _, exists := operation.Responses[selectedResponseCode].
				Content["application/json"]; exists {
				responseBodyReference = operation.Responses[selectedResponseCode].Content["application/json"].Schema.Reference
			}
		}
		responseBodyName := path.Base(responseBodyReference)
		if responseBodyName == "." {
			responseBodyName = "No Response Body"
		}
		responseContractLiteral := buildResponseContractReference(data, operation, selectedResponseCode, testSectorCamelCase, testSectorCamelCase, responseBodyName)
		contentTypeForTestCase := requestContentType
		if contentTypeForTestCase == "" {
			contentTypeForTestCase = "application/json"
		}
		testCaseTitle := buildTestCaseTitle(method, resolvedPathName)
		scenarioKey := scenarioReservationKey("positive", resolvedPathName, method, selectedResponseCode, "")
		currentNumber, _ := reserveTestCaseNumberForScenario(testSectorCamelCase, testSectorAbbr, scenarioKey)
		positiveResponseCondition := "code == " + selectedResponseCode
		testCaseJSON := TestCaseJSONTemplate(testSectorCamelCase, testSectorAbbr, resolvedPathName, methodNameOperationInAllCaps, testCaseTitle, contentTypeForTestCase, requestBodyName, selectedResponseCode, responseBodyName, currentNumber, "normal", "medium")
		filePath := testCaseJSONPath(testSectorCamelCase, testSectorAbbr, currentNumber, GenTestCaseTitleFromPathAndMethod(resolvedPathName, method))
		writeGeneratedTestCaseJSON(filePath, testCaseJSON)
		if activeGenerationReport != nil {
			activeGenerationReport.PositiveScenarios++
		}
		var testCaseCode string
		switch method {
		case "post", "put", "patch":
			contentTypeForCode := requestContentType
			if contentTypeForCode == "" {
				contentTypeForCode = "application/json"
			}
			testCaseCode = MethodWithBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, currentNumber, method, resolvedPathName, testCaseTitle, endpointPrefixGetter, endpointConstReference, requestBodyLiteral, contentTypeForCode, testingToolkit.ConvertIntToString(requestBodyKind), selectedResponseCode, positiveResponseCondition, queryParamsLiteral, headerParamsLiteral, responseContractLiteral)
		case "get", "delete", "options", "head", "trace":
			testCaseCode = MethodWithoutBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, currentNumber, method, resolvedPathName, testCaseTitle, endpointPrefixGetter, endpointConstReference, selectedResponseCode, positiveResponseCondition, queryParamsLiteral, headerParamsLiteral, responseContractLiteral)
		default:
			testCaseCode = ""
		}
		if testCaseCode != "" {
			insertGeneratedTestCode(testSectorCamelCase, testCaseCode, "positive")
		}
		generateNegativeTestCase(testSectorCamelCase, testSectorAbbr, endpointPrefixGetter, endpointConstReference, data, method, operation, resolvedPathName, requestContentType, requestBodyKind, requestBodyLiteral, requestBodyName, requiredQueryNames, requiredHeaderNames, queryParams, headerParams)
	}
}

func buildRequestBodyDetails(spec OpenAPISpec, operation Operation, testSectorCamelCase, pathName, method string) (string, string, string, int) {
	if operation.RequestBody == nil || len(operation.RequestBody.Content) == 0 {
		return "No Request Body", strconv.Quote(""), "", int(apiFunctions.BodyNone)
	}
	contentType, mediaType := selectRequestMediaType(operation.RequestBody.Content)
	requestBodyName := schemaDisplayName(mediaType.Schema)
	schema := resolveSchemaReference(spec, mediaType.Schema, 0)
	value := sampleValueFromSchema(spec, schema, 0)
	if generatedExpression, ok := buildGeneratedRequestBodyExpression(testSectorCamelCase, pathName, method, requestBodyName, value); ok {
		if contentType == "" {
			contentType = "application/json"
		}
		return requestBodyName, generatedExpression, contentType, ReturnBodyKindIntFromContentType(contentType)
	}
	jsonPayload := ""
	if value != nil {
		if bytes, err := json.Marshal(value); err == nil {
			jsonPayload = string(bytes)
		}
	}
	if jsonPayload == "" && requestBodyName != "No Request Body" {
		jsonPayload = "{}"
	}
	if contentType == "" {
		contentType = "application/json"
	}

	return requestBodyName, strconv.Quote(jsonPayload), contentType, ReturnBodyKindIntFromContentType(contentType)
}

func buildOperationParameterLiterals(spec OpenAPISpec, pathName string, pathParameters []Parameter, operation Operation) (string, string, string) {
	resolvedPath, queryParams, headerParams, _, _ := buildOperationParameterDetails(spec, pathName, pathParameters, operation)
	return resolvedPath, mapStringLiteral(queryParams), mapStringLiteral(headerParams)
}

func buildOperationParameterDetails(spec OpenAPISpec, pathName string, pathParameters []Parameter, operation Operation) (string, map[string]string, map[string]string, []string, []string) {
	resolvedPath := pathName
	queryParams := make(map[string]string)
	headerParams := make(map[string]string)
	requiredQueryNames := make([]string, 0)
	requiredHeaderNames := make([]string, 0)
	parameters := mergeParameters(pathParameters, operation.Parameters)
	for _, parameter := range parameters {
		resolvedSchema := resolveSchemaReference(spec, parameter.Schema, 0)
		sampleValue := sampleValueFromSchema(spec, resolvedSchema, 0)
		stringValue := parameterValueAsAString(sampleValue)
		switch strings.ToLower(parameter.In) {
		case "path":
			resolvedPath = strings.ReplaceAll(resolvedPath, "{"+parameter.Name+"}", stringValue)
		case "query":
			queryParams[parameter.Name] = stringValue
			if parameter.Required {
				requiredQueryNames = append(requiredQueryNames, parameter.Name)
			}
		case "header":
			headerParams[parameter.Name] = stringValue
			if parameter.Required {
				requiredHeaderNames = append(requiredHeaderNames, parameter.Name)
			}
		}
	}
	sort.Strings(requiredQueryNames)
	sort.Strings(requiredHeaderNames)
	return resolvedPath, queryParams, headerParams, requiredQueryNames, requiredHeaderNames
}

func generateNegativeTestCase(testSectorCamelCase, testSectorAbbr, endpointPrefixGetter, endpointConstReference string, spec OpenAPISpec, method string, operation Operation, resolvedPathName, requestContentType string, requestBodyKind int, requestBodyLiteral, requestBodyName string, requiredQueryNames, requiredHeaderNames []string, queryParams, headerParams map[string]string) {
	negativeResponseLabel, negativeResponseCondition := selectNegativeStatusExpectation(operation)
	negativeContract := testSectorCamelCase + `.ResponseContract{Type: "any"}`
	contentType := requestContentType
	if contentType == "" {
		contentType = "application/json"
	}
	if generateAuthNegativeTestCase(testSectorCamelCase, testSectorAbbr, endpointPrefixGetter, endpointConstReference, spec, method, operation, resolvedPathName, contentType, requestBodyKind, requestBodyLiteral, requestBodyName, queryParams, headerParams, negativeContract) {
		return
	}
	if len(requiredQueryNames) > 0 {
		missingName := requiredQueryNames[0]
		negativeQuery := cloneStringMap(queryParams)
		delete(negativeQuery, missingName)
		scenarioKey := scenarioReservationKey("negative-missing-query", resolvedPathName, method, negativeResponseLabel, missingName)
		testCaseNumber, _ := reserveTestCaseNumberForScenario(testSectorCamelCase, testSectorAbbr, scenarioKey)
		qualifier := "Negative Missing Query Parameter: " + missingName
		testCaseTitle := buildNegativeTestCaseTitle(method, resolvedPathName, qualifier)
		jsonCase := TestCaseJSONTemplate(testSectorCamelCase, testSectorAbbr, resolvedPathName, strings.ToUpper(method), testCaseTitle, contentType, requestBodyName, negativeResponseLabel, "Error Response", testCaseNumber, "normal", "high")
		filePath := testCaseJSONPath(testSectorCamelCase, testSectorAbbr, testCaseNumber, GenTestCaseTitleFromPathAndMethod(resolvedPathName, method)+"_"+negativeQualifierFileSegment(qualifier)+"_"+sanitizeCaseName(missingName))
		writeGeneratedTestCaseJSON(filePath, jsonCase)
		if activeGenerationReport != nil {
			activeGenerationReport.NegativeScenarios++
		}
		var code string
		switch method {
		case "post", "put", "patch":
			code = MethodWithBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, testCaseNumber, method, resolvedPathName, testCaseTitle, endpointPrefixGetter, endpointConstReference, requestBodyLiteral, contentType, testingToolkit.ConvertIntToString(requestBodyKind), negativeResponseLabel, negativeResponseCondition, mapStringLiteral(negativeQuery), mapStringLiteral(headerParams), negativeContract)
		default:
			code = MethodWithoutBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, testCaseNumber, method, resolvedPathName, testCaseTitle, endpointPrefixGetter, endpointConstReference, negativeResponseLabel, negativeResponseCondition, mapStringLiteral(negativeQuery), mapStringLiteral(headerParams), negativeContract)
		}
		if code != "" {
			insertGeneratedTestCode(testSectorCamelCase, code, "negative-missing-query")
		}
		return
	}

	if len(requiredHeaderNames) > 0 {
		missingName := requiredHeaderNames[0]
		negativeHeader := cloneStringMap(headerParams)
		delete(negativeHeader, missingName)
		scenarioKey := scenarioReservationKey("negative-missing-header", resolvedPathName, method, negativeResponseLabel, missingName)
		testCaseNumber, _ := reserveTestCaseNumberForScenario(testSectorCamelCase, testSectorAbbr, scenarioKey)
		qualifier := "Negative Missing Header: " + missingName
		testCaseTitle := buildNegativeTestCaseTitle(method, resolvedPathName, qualifier)
		jsonCase := TestCaseJSONTemplate(testSectorCamelCase, testSectorAbbr, resolvedPathName, strings.ToUpper(method), testCaseTitle, contentType, requestBodyName, negativeResponseLabel, "Error Response", testCaseNumber, "normal", "high")
		filePath := testCaseJSONPath(testSectorCamelCase, testSectorAbbr, testCaseNumber, GenTestCaseTitleFromPathAndMethod(resolvedPathName, method)+"_"+negativeQualifierFileSegment(qualifier)+"_"+sanitizeCaseName(missingName))
		writeGeneratedTestCaseJSON(filePath, jsonCase)
		if activeGenerationReport != nil {
			activeGenerationReport.NegativeScenarios++
		}
		var code string
		switch method {
		case "post", "put", "patch":
			code = MethodWithBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, testCaseNumber, method, resolvedPathName, testCaseTitle, endpointPrefixGetter, endpointConstReference, requestBodyLiteral, contentType, testingToolkit.ConvertIntToString(requestBodyKind), negativeResponseLabel, negativeResponseCondition, mapStringLiteral(queryParams), mapStringLiteral(negativeHeader), negativeContract)
		default:
			code = MethodWithoutBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, testCaseNumber, method, resolvedPathName, testCaseTitle, endpointPrefixGetter, endpointConstReference, negativeResponseLabel, negativeResponseCondition, mapStringLiteral(queryParams), mapStringLiteral(negativeHeader), negativeContract)
		}
		if code != "" {
			insertGeneratedTestCode(testSectorCamelCase, code, "negative-missing-header")
		}
		return
	}

	if method == "post" || method == "put" || method == "patch" {
		negativeBodyLiteral, missingField, ok := buildMissingRequiredBodyLiteral(spec, operation, requestBodyLiteral)
		if ok {
			scenarioKey := scenarioReservationKey("negative-missing-body", resolvedPathName, method, negativeResponseLabel, missingField)
			testCaseNumber, _ := reserveTestCaseNumberForScenario(testSectorCamelCase, testSectorAbbr, scenarioKey)
			qualifier := "Negative Missing Body: " + missingField
			testCaseTitle := buildNegativeTestCaseTitle(method, resolvedPathName, qualifier)
			jsonCase := TestCaseJSONTemplate(testSectorCamelCase, testSectorAbbr, resolvedPathName, strings.ToUpper(method), testCaseTitle, contentType, requestBodyName, negativeResponseLabel, "Error Response", testCaseNumber, "normal", "high")
			filePath := testCaseJSONPath(testSectorCamelCase, testSectorAbbr, testCaseNumber, GenTestCaseTitleFromPathAndMethod(resolvedPathName, method)+"_"+negativeQualifierFileSegment(qualifier)+"_"+sanitizeCaseName(missingField))
			writeGeneratedTestCaseJSON(filePath, jsonCase)
			if activeGenerationReport != nil {
				activeGenerationReport.NegativeScenarios++
			}
			code := MethodWithBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, testCaseNumber, method, resolvedPathName, testCaseTitle, endpointPrefixGetter, endpointConstReference, negativeBodyLiteral, contentType, testingToolkit.ConvertIntToString(requestBodyKind), negativeResponseLabel, negativeResponseCondition, mapStringLiteral(queryParams), mapStringLiteral(headerParams), negativeContract)
			insertGeneratedTestCode(testSectorCamelCase, code, "negative-missing-body")
		}
	}
}

func generateAuthNegativeTestCase(testSectorCamelCase, testSectorAbbr, endpointPrefixGetter, endpointConstReference string, spec OpenAPISpec, method string, operation Operation, resolvedPathName, contentType string, requestBodyKind int, requestBodyLiteral, requestBodyName string, queryParams, headerParams map[string]string, negativeContract string) bool {
	requirements := effectiveSecurityRequirements(spec, operation)
	if len(requirements) == 0 {
		return false
	}
	schemeName, scheme, locationKey, locationName := selectSecurityTestTarget(spec, requirements)
	if schemeName == "" {
		return false
	}

	negativeResponseLabel, negativeResponseCondition := selectAuthNegativeStatusExpectation(operation)
	scenarioKey := scenarioReservationKey("negative-auth", resolvedPathName, method, negativeResponseLabel, schemeName)
	testCaseNumber, _ := reserveTestCaseNumberForScenario(testSectorCamelCase, testSectorAbbr, scenarioKey)
	qualifier := "Negative Auth: " + schemeName
	testCaseTitle := buildNegativeTestCaseTitle(method, resolvedPathName, qualifier)
	jsonCase := TestCaseJSONTemplate(testSectorCamelCase, testSectorAbbr, resolvedPathName, strings.ToUpper(method), testCaseTitle, contentType, requestBodyName, negativeResponseLabel, "Error Response", testCaseNumber, "normal", "high")
	filePath := testCaseJSONPath(testSectorCamelCase, testSectorAbbr, testCaseNumber, GenTestCaseTitleFromPathAndMethod(resolvedPathName, method)+"_"+negativeQualifierFileSegment(qualifier)+"_"+sanitizeCaseName(schemeName))
	writeGeneratedTestCaseJSON(filePath, jsonCase)
	if activeGenerationReport != nil {
		activeGenerationReport.NegativeScenarios++
	}
	negativeHeaders := cloneStringMap(headerParams)
	negativeQuery := cloneStringMap(queryParams)
	switch strings.ToLower(locationKey) {
	case "header":
		if locationName == "" {
			locationName = "Authorization"
		}
		negativeHeaders[locationName] = invalidSecurityValue(scheme)
	case "query":
		if locationName == "" {
			locationName = "api_key"
		}
		negativeQuery[locationName] = "invalid"
	case "cookie":
		if locationName == "" {
			locationName = "cookie"
		}
		negativeHeaders["Cookie"] = locationName + "=invalid"
	default:
		negativeHeaders["Authorization"] = "Bearer invalid"
	}
	var code string
	switch method {
	case "post", "put", "patch":
		code = MethodWithBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, testCaseNumber, method, resolvedPathName, testCaseTitle, endpointPrefixGetter, endpointConstReference, requestBodyLiteral, contentType, testingToolkit.ConvertIntToString(requestBodyKind), negativeResponseLabel, negativeResponseCondition, mapStringLiteral(negativeQuery), mapStringLiteral(negativeHeaders), negativeContract)
	default:
		code = MethodWithoutBodyTestCaseCodeTemplate(testSectorCamelCase, testSectorAbbr, testCaseNumber, method, resolvedPathName, testCaseTitle, endpointPrefixGetter, endpointConstReference, negativeResponseLabel, negativeResponseCondition, mapStringLiteral(negativeQuery), mapStringLiteral(negativeHeaders), negativeContract)
	}
	if code != "" {
		return false
	}
	insertGeneratedTestCode(testSectorCamelCase, code, "negative-auth")
	return true
}

func effectiveSecurityRequirements(spec OpenAPISpec, operation Operation) []map[string][]string {
	if operation.Security != nil {
		return operation.Security
	}
	return spec.Security
}

func selectSecurityTestTarget(spec OpenAPISpec, requirements []map[string][]string) (string, SecurityScheme, string, string) {
	if len(requirements) == 0 {
		return "", SecurityScheme{}, "", ""
	}
	for _, requirement := range requirements {
		if len(requirement) == 0 {
			continue
		}
		names := make([]string, 0, len(requirement))
		for name := range requirement {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			scheme, ok := spec.Components.SecuritySchemes[name]
			if !ok {
				continue
			}
			locationKey, locationName := authInjectionPoint(scheme)
			return name, scheme, locationKey, locationName
		}
	}
	return "", SecurityScheme{}, "", ""
}

func authInjectionPoint(scheme SecurityScheme) (string, string) {
	switch strings.ToLower(scheme.Type) {
	case "apiKey":
		in := strings.ToLower(strings.TrimSpace(scheme.In))
		if in == "header" || in == "query" || in == "cookie" {
			return in, strings.TrimSpace(scheme.Name)
		}
		return "header", strings.TrimSpace(scheme.Name)
	case "http":
		return "header", "Authorization"
	case "oauth2", "openIdConnect":
		return "header", "Authorization"
	default:
		return "header", "Authorization"
	}
}

func invalidSecurityValue(scheme SecurityScheme) string {
	if strings.EqualFold(scheme.Type, "http") {
		switch strings.ToLower(strings.TrimSpace(scheme.Scheme)) {
		case "basic":
			return "Basic invalid"
		case "bearer":
			return "Bearer invalid"
		}
	}
	if strings.EqualFold(scheme.Type, "oauth2") || strings.EqualFold(scheme.Type, "openIdConnect") || strings.EqualFold(scheme.Type, "openidconnect") {
		return "Bearer invalid"
	}
	if strings.EqualFold(scheme.Type, "apiKey") {
		return "invalid"
	}
	return "Bearer invalid"
}

func selectAuthNegativeStatusExpectation(operation Operation) (string, string) {
	responseCodes := sortedResponseCodes(operation)
	if len(responseCodes) == 0 {
		return "401", "code == 401"
	}
	if slicesContains(responseCodes, "401") && slicesContains(responseCodes, "403") {
		return "403", "code == 403"
	}
	if slicesContains(responseCodes, "401") {
		return "401", "code == 401"
	}
	if slicesContains(responseCodes, "403") {
		return "403", "code == 403"
	}
	for _, code := range responseCodes {
		if strings.HasPrefix(code, "4") {
			return code, "code == " + code
		}
	}
	return "401", "code == 401"
}

func selectNegativeStatusExpectation(operation Operation) (string, string) {
	responseCodes := sortedResponseCodes(operation)
	if len(responseCodes) == 0 {
		return "400", "code == 400"
	}
	fourXX := make([]string, 0)
	for _, code := range responseCodes {
		if strings.HasPrefix(code, "4") {
			fourXX = append(fourXX, code)
		}
	}
	if len(fourXX) == 0 {
		return "400", "code == 400"
	}
	if len(fourXX) == 1 {
		return fourXX[0], "code == " + fourXX[0]
	}
	return "4xx", "code >= 400 && code < 500"
}

func sortedResponseCodes(operation Operation) []string {
	if len(operation.Responses) == 0 {
		return nil
	}
	responseCodes := make([]string, 0, len(operation.Responses))
	for code := range operation.Responses {
		responseCodes = append(responseCodes, code)
	}
	sort.Strings(responseCodes)
	return responseCodes
}

func selectAuthNegativeResponseCode(operation Operation) int {
	if len(operation.Responses) == 0 {
		responsesCodes := make([]string, 0, len(operation.Responses))
		for code := range operation.Responses {
			responsesCodes = append(responsesCodes, code)
		}
		sort.Strings(responsesCodes)
		for _, preferred := range []string{"401", "403"} {
			if slicesContains(responsesCodes, preferred) {
				if parsed, err := strconv.Atoi(preferred); err == nil {
					return parsed
				}
			}
		}
		for _, code := range responsesCodes {
			if strings.HasPrefix(code, "4") {
				if parsed, err := strconv.Atoi(code); err == nil {
					return parsed
				}
			}
		}
	}
	return 401
}

func selectNegativeResponseCode(operation Operation) int {
	if len(operation.Responses) > 0 {
		responseCodes := make([]string, 0, len(operation.Responses))
		for code := range operation.Responses {
			responseCodes = append(responseCodes, code)
		}
		sort.Strings(responseCodes)
		for _, code := range responseCodes {
			if strings.HasPrefix(code, "4") {
				if parsed, err := strconv.Atoi(code); err == nil {
					return parsed
				}
			}
		}
	}
	return 400
}

func endpointConstantReference(testSectorCamelCase, pathName string) string {
	return testSectorCamelCase + "." + ConvertPathToConstName(pathName) + "Endpoint"
}

func buildEndpointConstValues(spec OpenAPISpec, paths []string) map[string]string {
	values := make(map[string]string)
	for _, pathName := range paths {
		values[pathName] = resolveEndpointPathForConstant(spec, pathName)
	}
	return values
}

func resolveEndpointPathForConstant(spec OpenAPISpec, pathName string) string {
	operations := spec.Paths[pathName]
	if len(operations) == 0 {
		return pathName
	}
	methods := make([]string, 0, len(operations))
	for method := range operations {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	pathLevelParameters := pathItemParameters(spec, pathName)
	for _, method := range methods {
		operation := operations[method]
		resolvedPath, _, _, _, _ := buildOperationParameterDetails(spec, pathName, pathLevelParameters, operation)
		if strings.TrimSpace(resolvedPath) != "" {
			return resolvedPath
		}
	}
	return pathName
}

func buildMissingRequiredBodyLiteral(spec OpenAPISpec, operation Operation, baseRequestBodyLiteral string) (string, string, bool) {
	if operation.RequestBody == nil || len(operation.RequestBody.Content) == 0 {
		return "", "", false
	}
	_, mediaType := selectRequestMediaType(operation.RequestBody.Content)
	schema := resolveSchemaReference(spec, mediaType.Schema, 0)
	if len(schema.Required) == 0 {
		return "", "", false
	}
	if strings.Contains(baseRequestBodyLiteral, ".") && strings.Contains(baseRequestBodyLiteral, "(") && strings.Contains(baseRequestBodyLiteral, ")") {
		requiredFields := append([]string{}, schema.Required...)
		sort.Strings(requiredFields)
		missingField := requiredFields[0]
		bodyLiteral := indentMultilineLiteral(baseRequestBodyLiteral, "\t")
		expression := "func() map[string]any {\n\t\tbody := " + bodyLiteral + "\n\t\tdelete(body, " + strconv.Quote(missingField) + ")\n\t\treturn body\n\t}()"
		return expression, missingField, true
	}
	bodySample := sampleValueFromSchema(spec, schema, 0)
	objectBody, ok := bodySample.(map[string]any)
	if !ok {
		return "", "", false
	}
	requiredFields := append([]string{}, schema.Required...)
	sort.Strings(requiredFields)
	missingField := requiredFields[0]
	if _, exists := objectBody[missingField]; !exists {
		return "", "", false
	}
	delete(objectBody, missingField)
	bytes, err := json.Marshal(objectBody)
	if err != nil {
		return "", "", false
	}
	return strconv.Quote(string(bytes)), missingField, true
}

func buildGeneratedRequestBodyExpression(testSectorCamelCase, pathName, method, requestBodyName string, sample any) (string, bool) {
	objectSample, ok := sample.(map[string]any)
	if !ok || len(objectSample) == 0 {
		return "", false
	}
	factories := loadGeneratedBodyFactories(testSectorCamelCase)
	if len(factories) == 0 {
		return "", false
	}
	candidates := []string{
		ConvertPathToConstName(pathName) + upperFirst(strings.ToLower(method)) + "RequestBody",
		sanitizeGeneratedFunctionName(requestBodyName),
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		factory, exists := factories[candidate]
		if !exists {
			continue
		}
		args, ok := buildGeneratedBodyFactoryArgs(testSectorCamelCase, factory, objectSample)
		if !ok {
			continue
		}
		return formatGeneratedBodyFactoryCall(testSectorCamelCase+"."+factory.name, args), true
	}
	return "", false
}

func formatGeneratedBodyFactoryCall(qualifiedFactoryName string, args []string) string {
	if len(args) == 0 {
		return qualifiedFactoryName + "()"
	}
	var builder strings.Builder
	builder.WriteString(qualifiedFactoryName)
	builder.WriteString("(\n")
	for _, arg := range args {
		builder.WriteString("\t\t")
		builder.WriteString(arg)
		builder.WriteString(",\n")
	}
	builder.WriteString("\t)")
	return builder.String()
}

func indentMultilineLiteral(value string, indent string) string {
	if !strings.Contains(value, "\n") {
		return value
	}
	lines := strings.Split(value, "\n")
	for index := 1; index < len(lines); index++ {
		lines[index] = indent + lines[index]
	}
	return strings.Join(lines, "\n")
}

func loadGeneratedBodyFactories(testSectorCamelCase string) map[string]generatedBodyFactory {
	if generatedBodyFactoryCache == nil {
		generatedBodyFactoryCache = make(map[string]map[string]generatedBodyFactory)
	}
	if cached, ok := generatedBodyFactoryCache[testSectorCamelCase]; ok {
		return cached
	}
	filePath := testingToolkit.CurrPath() + "/comInternal/pageDoms/" + testSectorCamelCase + "/" + testSectorCamelCase + "GeneratedObjects.go"
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		empty := map[string]generatedBodyFactory{}
		generatedBodyFactoryCache[testSectorCamelCase] = empty
		return empty
	}
	functionPattern := regexp.MustCompile(`func\s+([A-Z][A-Za-z0-9_]*)\s*\((?s:(.*?))\)\s*map\[string\]any\s*\{`)
	matches := functionPattern.FindAllStringSubmatch(string(contentBytes), -1)
	factories := make(map[string]generatedBodyFactory)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := strings.TrimSpace(match[1])
		factories[name] = generatedBodyFactory{name: name, params: parseGeneratedBodyParams(match[2])}
	}
	generatedBodyFactoryCache[testSectorCamelCase] = factories
	return factories
}

func parseGeneratedBodyParams(paramsBlock string) []generatedBodyParam {
	lines := strings.Split(paramsBlock, "\n")
	params := make([]generatedBodyParam, 0)
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, ","))
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		typeName := strings.Join(fields[1:], " ")
		param := generatedBodyParam{name: name, typeName: typeName, baseType: typeName}
		if strings.HasPrefix(typeName, "Optional[") && strings.HasSuffix(typeName, "]") {
			param.isOptional = true
			param.baseType = strings.TrimSuffix(strings.TrimPrefix(typeName, "Optional["), "]")
		}
		params = append(params, param)
	}
	return params
}

func buildGeneratedBodyFactoryArgs(testSectorCamelCase string, factory generatedBodyFactory, sample map[string]any) ([]string, bool) {
	args := make([]string, 0, len(factory.params))
	for _, param := range factory.params {
		value, exists := sample[param.name]
		placeholder := goPlaceholderLiteralForParam(param)
		if param.isOptional {
			if placeholder != "" {
				args = append(args, testSectorCamelCase+".Include("+placeholder+")")
				continue
			}
			if !exists {
				args = append(args, testSectorCamelCase+".Omit["+param.baseType+"]()")
				continue
			}
			literal, ok := goLiteralFromSampleValue(value, param.baseType)
			if !ok {
				return nil, false
			}
			args = append(args, testSectorCamelCase+".Include("+literal+")")
			continue
		}
		if placeholder != "" {
			args = append(args, placeholder)
			continue
		}
		if !exists {
			return nil, false
		}
		literal, ok := goLiteralFromSampleValue(value, param.baseType)
		if !ok {
			return nil, false
		}
		args = append(args, literal)
	}
	return args, true
}

func goPlaceholderLiteralForParam(param generatedBodyParam) string {
	label := sanitizeCaseName(param.name)
	typeName := strings.TrimSpace(param.baseType)
	if typeName == "" {
		typeName = strings.TrimSpace(param.typeName)
	}
	marker := "/*" + strings.ToLower(label[:1]) + label[1:] + ":" + typeName + "*/"
	switch typeName {
	case "string":
		return strconv.Quote(marker)
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return marker + "0"
	case "float32", "float64":
		return marker + "0.0"
	case "bool":
		return marker + "false"
	case "any", "interface{}", "map[string]any":
		return marker + "any(nil)"
	}
	if strings.HasPrefix(typeName, "[]") {
		return marker + typeName + "{}"
	}
	if strings.HasPrefix(typeName, "map[") {
		return marker + typeName + "{}"
	}
	if strings.HasPrefix(typeName, "*") {
		return marker + "(" + typeName + ")(nil)"
	}
	if strings.HasPrefix(typeName, "chan ") || strings.HasPrefix(typeName, "<-chan ") || strings.HasPrefix(typeName, "chan<- ") || strings.HasPrefix(typeName, "func(") {
		return marker + "(" + typeName + ")(nil)"
	}
	return marker + typeName + "{}"
}

func goLiteralFromSampleValue(value any, expectedType string) (string, bool) {
	switch expectedType {
	case "string":
		return strconv.Quote(fmt.Sprintf("%v", value)), true
	case "int":
		switch v := value.(type) {
		case int:
			return strconv.Itoa(v), true
		case int64:
			return strconv.Itoa(int(v)), true
		case float64:
			return strconv.Itoa(int(v)), true
		case json.Number:
			i, err := strconv.Atoi(v.String())
			if err != nil {
				return "", false
			}
			return strconv.Itoa(i), true
		}
	case "int64":
		switch v := value.(type) {
		case int64:
			return strconv.FormatInt(v, 10), true
		case int:
			return strconv.FormatInt(int64(v), 10), true
		case float64:
			return strconv.FormatInt(int64(v), 10), true
		}
	case "float64":
		switch v := value.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), true
		case int:
			return strconv.FormatFloat(float64(v), 'f', -1, 64), true
		case int64:
			return strconv.FormatFloat(float64(v), 'f', -1, 64), true
		}
	case "float32":
		switch v := value.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 32), true
		case int:
			return strconv.FormatFloat(float64(v), 'f', -1, 32), true
		case int64:
			return strconv.FormatFloat(float64(v), 'f', -1, 32), true
		}
	case "bool":
		if v, ok := value.(bool); ok {
			return strconv.FormatBool(v), true
		}
	case "any":
		return fmt.Sprintf("%v", value), true
	}
	return "", false
}

func sanitizeGeneratedFunctionName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		if len(part) == 1 {
			builder.WriteString(strings.ToUpper(part))
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		builder.WriteString(part[1:])
	}
	return builder.String()
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return map[string]string{}
	}
	clone := make(map[string]string, len(source))
	for k, v := range source {
		clone[k] = v
	}
	return clone
}

func sanitizeCaseName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "value"
	}
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "{", "")
	name = strings.ReplaceAll(name, "}", "")
	name = strings.ReplaceAll(name, "[", "")
	name = strings.ReplaceAll(name, "]", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	return strings.ToUpper(name[:1]) + name[1:]
}

func scenarioReservationKey(kind, pathName, method, responseCode, qualifier string) string {
	parts := []string{strings.ToLower(strings.TrimSpace(kind)), strings.ToLower(strings.TrimSpace(pathName)), strings.ToLower(strings.TrimSpace(method)), strings.ToLower(strings.TrimSpace(responseCode)), strings.ToLower(strings.TrimSpace(qualifier))}
	return strings.Join(parts, "|")
}

func reserveTestCaseNumberForScenario(testSectorCamelCase, testSectorAbbr, scenarioKey string) (string, bool) {
	if generatedScenarioNumbers == nil {
		generatedScenarioNumbers = make(map[string]string)
	}
	if existing, ok := generatedScenarioNumbers[scenarioKey]; ok {
		return existing, false
	}
	next := NextTestCaseJSONNumber(testSectorCamelCase, testSectorAbbr)
	generatedScenarioNumbers[scenarioKey] = next
	return next, true
}

func testCaseJSONPath(testSectorCamelCase, testSectorAbbr, testCaseNumber, scenarioName string) string {
	return testingToolkit.CurrPath() + "/tests/" + testSectorCamelCase + "Tests/testcases/TC_" + testSectorAbbr + testCaseNumber + "_" + scenarioName + ".json"
}

func writeGeneratedTestCaseJSON(filePath, content string) {
	_, statErr := os.Stat(filePath)
	fileExisted := statErr == nil
	testingToolkit.PrintStringToFile(content, filePath)
	if activeGenerationReport != nil {
		return
	}
	if fileExisted {
		activeGenerationReport.JSONFilesOverwrote++
	} else {
		activeGenerationReport.JSONFilesCreated++
	}
}

func insertGeneratedTestCode(testSectorCamelCase, testCaseCode, scenario string) {
	if testCaseCode == "" {
		return
	}
	before, readErr := os.ReadFile(testingToolkit.CurrPath() + "/tests/" + testSectorCamelCase + "Tests/" + testSectorCamelCase + "Tests.go")
	err := InsertTestCodeIntoTestCodeFile(testSectorCamelCase, testCaseCode)
	if err != nil {
		log.Println("Error inserting test code for " + testSectorCamelCase + " scenario " + scenario + ": " + err.Error())
		reportError("code insertion failed for " + scenario + ": " + err.Error())
	}
	if activeGenerationReport == nil {
		return
	}
	after, readAfterErr := os.ReadFile(testingToolkit.CurrPath() + "/tests/" + testSectorCamelCase + "Tests/" + testSectorCamelCase + "Tests.go")
	if readErr == nil && readAfterErr == nil && string(before) == string(after) {
		activeGenerationReport.CodeSkips++
		return
	}
	activeGenerationReport.CodeInsertions++
}

func reportError(message string) {
	if activeGenerationReport == nil {
		return
	}
	activeGenerationReport.Errors = append(activeGenerationReport.Errors, message)
}

func reportWarning(message string) {
	if activeGenerationReport == nil {
		return
	}
	activeGenerationReport.Warnings = append(activeGenerationReport.Warnings, message)
}

func finalizeGenerationReport(testSectorCamelCase string) {
	if activeGenerationReport == nil {
		return
	}
	activeGenerationReport.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	activeGenerationReport.WarningSummary = summarizeMessages(activeGenerationReport.Warnings)
	activeGenerationReport.ErrorSummary = summarizeMessages(activeGenerationReport.Errors)
	reportPath := testingToolkit.CurrPath() + "/tests/" + testSectorCamelCase + "Tests/openApiGenerationReport" + testingToolkit.CurrentTimeForNaming() + ".json"
	activeGenerationReport.ReportPath = reportPath
	bytes, err := json.MarshalIndent(activeGenerationReport, "", "  ")
	if err != nil {
		log.Println("Error serializing generation report: " + err.Error())
		return
	}
	if err := os.WriteFile(reportPath, bytes, 0644); err != nil {
		log.Println("Error writing generation report to file: " + err.Error())
		return
	}
	logGenerationSummary(activeGenerationReport)
}

func summarizeMessages(messages []string) []messageSummaryItem {
	if len(messages) == 0 {
		return nil
	}
	type accumulator struct {
		count  int
		sample string
	}
	grouped := make(map[string]accumulator)
	for _, message := range messages {
		category := messageCategory(message)
		entry := grouped[category]
		entry.count++
		if entry.sample == "" {
			entry.sample = message
		}
		grouped[category] = entry
	}
	categories := make([]string, 0, len(grouped))
	for category := range grouped {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	summary := make([]messageSummaryItem, 0, len(categories))
	for _, category := range categories {
		entry := grouped[category]
		summary = append(summary, messageSummaryItem{
			Category: category,
			Count:    entry.count,
			Sample:   entry.sample,
		})
	}
	return summary
}

func messageCategory(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return "empty"
	}
	if index := strings.Index(message, ":"); index > 0 {
		return strings.TrimSpace(message[:index])
	}
	if index := strings.Index(message, " for "); index > 0 {
		return strings.TrimSpace(message[:index])
	}
	if index := strings.Index(message, " on "); index > 0 {
		return strings.TrimSpace(message[:index])
	}
	if len(message) > 80 {
		return message[:80]
	}
	return message
}

func logGenerationSummary(report *generationReport) {
	if report == nil {
		return
	}
	log.Printf(
		"OpenAPI generation summary: sector=%s, paths=%d, operations=%d, positive=%d, negative=%d, json(created=%d, overwritten=%d, code(inserted=%d, skipped=%d, errors=%d, report=%s",
		report.TestSector,
		report.PathsProcessed,
		report.OperationsProcessed,
		report.PositiveScenarios,
		report.NegativeScenarios,
		report.JSONFilesCreated,
		report.JSONFilesOverwrote,
		report.CodeInsertions,
		report.CodeSkips,
		len(report.Errors),
		report.ReportPath,
	)
}

func mergeParameters(pathParameters []Parameter, operationParameters []Parameter) []Parameter {
	if len(pathParameters) == 0 && len(operationParameters) == 0 {
		return nil
	}
	merged := make([]Parameter, 0, len(pathParameters)+len(operationParameters))
	indexByKey := make(map[string]int, len(pathParameters)+len(operationParameters))
	for _, parameter := range pathParameters {
		key := parameterKey(parameter)
		if _, exists := indexByKey[key]; exists {
			continue
		}
		indexByKey[key] = len(merged)
		merged = append(merged, parameter)
	}
	for _, parameter := range operationParameters {
		key := parameterKey(parameter)
		if index, exists := indexByKey[key]; exists {
			merged[index] = parameter
			continue
		}
		indexByKey[key] = len(merged)
		merged = append(merged, parameter)
	}
	return merged
}

func parameterKey(parameter Parameter) string {
	return strings.ToLower(strings.TrimSpace(parameter.In)) + "|" + strings.ToLower(strings.TrimSpace(parameter.Name))
}

func pathItemParameters(spec OpenAPISpec, pathName string) []Parameter {
	pathItemSchema, ok := spec.Paths[pathName]
	if !ok {
		return nil
	}
	bytes, err := json.Marshal(pathItemSchema)
	if err != nil {
		return nil
	}
	var item PathItem
	if err := json.Unmarshal(bytes, &item); err != nil {
		return nil
	}
	return item.Parameters
}

func parameterValueAsAString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(value)
	}
}

func mapStringLiteral(values map[string]string) string {
	if len(values) == 0 {
		return "map[string]string{}"
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, strconv.Quote(key)+": "+strconv.Quote(values[key]))
	}
	return "map[string]string{" + strings.Join(parts, ", ") + "}"
}

func buildResponseContractLiteral(spec OpenAPISpec, operation Operation, responseCode string, packageAlias string) string {
	response, ok := operation.Responses[responseCode]
	if !ok || len(response.Content) == 0 {
		return packageAlias + `.ResponseContract{Type: "any"}`
	}
	contentType, mediaType := selectRequestMediaType(response.Content)
	if !strings.Contains(contentType, "json") {
		return packageAlias + `.ResponseContract{Type: "any"}`
	}
	resolvedSchema := resolveSchemaReference(spec, mediaType.Schema, 0)
	return responseContractLiteralFromSchema(spec, resolvedSchema, 0, packageAlias)
}

func buildResponseContractReference(spec OpenAPISpec, operation Operation, responseCode string, packageAlias, testSectorCamelCase, responseObjectName string) string {
	literal := buildResponseContractLiteral(spec, operation, responseCode, packageAlias)
	if literal == packageAlias+`.ResponseContract{Type: "any"}` {
		return literal
	}
	helperName := responseContractHelperName(responseObjectName)
	if err := ensureResponseContractHelper(testSectorCamelCase, packageAlias, helperName, literal); err != nil {
		log.Println("Error ensuring response contract helper for " + testSectorCamelCase + ": " + err.Error())
		reportWarning("failed to generate response contract helper " + helperName + ": " + err.Error())
		return literal
	}
	return packageAlias + "." + helperName + "()"
}

func responseContractHelperName(responseObjectName string) string {
	name := strings.TrimSpace(responseObjectName)
	if name == "" || name == "." || strings.EqualFold(name, "No Response Body") {
		name = "Response"
	}
	return "Expected" + sanitizeExportedIdentifier(name) + "ResponseContract"
}

func ensureResponseContractHelper(testSectorCamelCase, packageAlias, helperName, literal string) error {
	if generatedResponseContractHelpers != nil && generatedResponseContractHelpers[helperName] {
		return nil
	}
	generatedObjectsPath := filepath.Join(testingToolkit.CurrPath(), "comInternal", "pageDoms", testSectorCamelCase, testSectorCamelCase+"GeneratedObjects.go")
	contentBytes, err := os.ReadFile(generatedObjectsPath)
	if err != nil {
		return err
	}
	signature := "func " + helperName + "() ResponseContract {"
	content := string(contentBytes)
	if strings.Contains(content, signature) {
		if generatedResponseContractHelpers != nil {
			generatedResponseContractHelpers[helperName] = true
		}
		return nil
	}
	localLiteral := strings.ReplaceAll(literal, packageAlias+".", "")
	helperText := "\nfunc " + helperName + "() ResponseContract {\n\treturn " + indentMultilineLiteral(localLiteral, "\t") + "\n}\n"
	updated := content + helperText
	if err := os.WriteFile(generatedObjectsPath, []byte(updated), 0644); err != nil {
		return err
	}
	if generatedResponseContractHelpers != nil {
		generatedResponseContractHelpers[helperName] = true
	}
	return nil
}

func responseContractLiteralFromSchema(spec OpenAPISpec, schema Schema, depth int, packageAlias string) string {
	return responseContractLiteralFromSchemaIndented(spec, schema, depth, packageAlias, 0)
}

func responseContractLiteralFromSchemaIndented(spec OpenAPISpec, schema Schema, depth int, packageAlias string, indentLevel int) string {
	if depth > 20 {
		return packageAlias + `.ResponseContract{Type: "any"}`
	}
	resolved := resolveSchemaReference(spec, schema, depth)
	typeName := inferResponseContractType(resolved)
	indent := strings.Repeat("\t", indentLevel)
	fieldIndent := strings.Repeat("\t", indentLevel+1)
	childIndent := strings.Repeat("\t", indentLevel+2)
	builder := strings.Builder{}
	builder.WriteString(packageAlias)
	builder.WriteString(".ResponseContract{\n")
	builder.WriteString(fieldIndent)
	builder.WriteString("Type: ")
	builder.WriteString(strconv.Quote(typeName))
	builder.WriteString(",")
	if resolved.Nullable {
		builder.WriteString("\n")
		builder.WriteString(fieldIndent)
		builder.WriteString("Nullable: true,")
	}
	if len(resolved.Properties) > 0 {
		required := append([]string{}, resolved.Required...)
		sort.Strings(required)
		builder.WriteString("\n")
		builder.WriteString(fieldIndent)
		builder.WriteString("Required: []string{")
		for index, field := range required {
			if index > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(strconv.Quote(field))
		}
		builder.WriteString("},")
	}
	if strings.EqualFold(typeName, "object") && len(resolved.Properties) > 0 {
		propertyNames := make([]string, 0, len(resolved.Properties))
		for propertyName := range resolved.Properties {
			propertyNames = append(propertyNames, propertyName)
		}
		sort.Strings(propertyNames)
		builder.WriteString("\n")
		builder.WriteString(fieldIndent)
		builder.WriteString("Properties: map[string]")
		builder.WriteString(packageAlias)
		builder.WriteString(".ResponseContract{\n")
		for _, propertyName := range propertyNames {
			builder.WriteString(childIndent)
			builder.WriteString(strconv.Quote(propertyName))
			builder.WriteString(": ")
			propertyLiteral := responseContractLiteralFromSchemaIndented(spec, resolved.Properties[propertyName], depth+1, packageAlias, indentLevel+2)
			builder.WriteString(elideQualifiedResponseContractType(propertyLiteral, packageAlias))
			builder.WriteString(",\n")
		}
		builder.WriteString(fieldIndent)
		builder.WriteString("},")
	}

	if strings.EqualFold(typeName, "object") {
		if additionalLiteral, ok := additionalPropertiesContractLiteral(spec, resolved.AdditionalProperties, depth+1, packageAlias); ok {
			builder.WriteString("\n")
			builder.WriteString(fieldIndent)
			builder.WriteString("AdditionalProperties: ")
			builder.WriteString(additionalLiteral)
			builder.WriteString(",")
		}
	}

	if strings.EqualFold(typeName, "array") && resolved.Items != nil {
		builder.WriteString("\n")
		builder.WriteString(fieldIndent)
		builder.WriteString("Items: &")
		builder.WriteString(responseContractLiteralFromSchemaIndented(spec, *resolved.Items, depth+1, packageAlias, indentLevel+1))
		builder.WriteString(",")
	}

	builder.WriteString("\n")
	builder.WriteString(indent)
	builder.WriteString("}")
	return builder.String()
}

func inferResponseContractType(schema Schema) string {
	if schema.Type != "" {
		return schema.Type
	}
	if len(schema.Properties) > 0 || schema.AdditionalProperties != nil {
		return "object"
	}
	if schema.Items != nil {
		return "array"
	}
	if len(schema.Enum) > 0 {
		if _, ok := schema.Enum[0].(string); ok {
			return "string"
		}
		if _, ok := schema.Enum[0].(bool); ok {
			return "boolean"
		}
		return "number"
	}
	return "any"
}

func additionalPropertiesContractLiteral(spec OpenAPISpec, additionalProperties any, depth int, packageAlias string) (string, bool) {
	if depth > 20 || additionalProperties == nil {
		return "", false
	}
	switch typed := additionalProperties.(type) {
	case bool:
		if typed {
			return packageAlias + `.ResponseContract{Type: "any"}`, true
		}
		return "", false
	case Schema:
		resolved := resolveSchemaReference(spec, typed, depth+1)
		return responseContractLiteralFromSchema(spec, resolved, depth+1, packageAlias), true
	case map[string]any:
		var schema Schema
		bytes, err := json.Marshal(typed)
		if err != nil {
			return packageAlias + `.ResponseContract{Type: "any"}`, true
		}
		if err := json.Unmarshal(bytes, &schema); err != nil {
			return packageAlias + `.ResponseContract{Type: "any"}`, true
		}
		resolved := resolveSchemaReference(spec, schema, depth+1)
		return responseContractLiteralFromSchema(spec, resolved, depth+1, packageAlias), true
	default:
		return packageAlias + `.ResponseContract{Type: "any"}`, true
	}
}

func elideQualifiedResponseContractType(literal, packageAlias string) string {
	prefix := packageAlias + ".ResponseContract"
	if strings.HasPrefix(literal, prefix) {
		return strings.TrimPrefix(literal, prefix)
	}
	return literal
}

func selectRequestMediaType(content map[string]MediaType) (string, MediaType) {
	priorities := []string{"application/json", "application/*+json", "text/json", "application/x-www-form-urlencoded", "multipart/form-data", "text/plain", "application/octet-stream", "*/*"}
	for _, priority := range priorities {
		if mediaType, ok := content[priority]; ok {
			return priority, mediaType
		}
	}
	keys := make([]string, 0, len(content))
	for key := range content {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return "", MediaType{}
	}
	selected := keys[0]
	return selected, content[selected]
}

func schemaDisplayName(schema Schema) string {
	if schema.Reference != "" {
		name := path.Base(schema.Reference)
		if name != "." {
			return name
		}
	}
	if schema.Type != "" {
		return "Inline " + strings.ToUpper(schema.Type) + " Body"
	}
	if len(schema.Properties) > 0 {
		return "Inline Object Body"
	}
	return "No Request Body"
}

func resolveSchemaReference(spec OpenAPISpec, schema Schema, depth int) Schema {
	if depth > 20 {
		return schema
	}
	resolved := schema
	if resolved.Reference != "" {
		if activeGenerationReport != nil {
			if referenceSchema, ok := activeSchemaResolver.ResolveSchemaReference(resolved.Reference); ok {
				resolved = referenceSchema
			}
		}
		if resolved.Reference != "" {
			name := path.Base(resolved.Reference)
			referenceSchema, ok := spec.Components.Schemas[name]
			if ok {
				resolved = referenceSchema
			}
		}
	}
	return resolveComposedSchema(spec, resolved, depth+1)
}

func newSchemaResolver(baseLocation string, spec OpenAPISpec) *schemaResolver {
	root := map[string]any{}
	bytes, err := json.Marshal(spec)
	if err == nil {
		if err := json.Unmarshal(bytes, &root); err != nil {
			root = map[string]any{}
		}
	}
	return &schemaResolver{
		baseLocation:    baseLocation,
		rootDocument:    root,
		externalDocs:    make(map[string]any),
		failedDocuments: make(map[string]bool),
	}
}

func (resolver *schemaResolver) ResolveSchemaReference(reference string) (Schema, bool) {
	documentReference, pointer := splitReference(reference)
	document, ok := resolver.documentForReference(documentReference)
	if !ok {
		return Schema{}, false
	}
	target := document
	if pointer != "" {
		resolvedTarget, ok := resolveJSONPointer(target, pointer)
		if !ok {
			return Schema{}, false
		}
		target = resolvedTarget
	}
	schema, ok := decodeSchema(target)
	if !ok {
		return Schema{}, false
	}
	return schema, true
}

func (resolver *schemaResolver) documentForReference(documentReference string) (any, bool) {
	if documentReference == "" {
		return resolver.rootDocument, true
	}
	resolvedLocation, err := resolveReferenceLocation(resolver.baseLocation, documentReference)
	if err != nil || resolvedLocation == "" {
		return nil, false
	}
	if resolver.failedDocuments[resolvedLocation] {
		return nil, false
	}
	if cached, ok := resolver.externalDocs[resolvedLocation]; ok {
		return cached, true
	}
	documentBytes, err := readReferenceDocumentBytes(resolvedLocation)
	if err != nil {
		resolver.failedDocuments[resolvedLocation] = true
		return nil, false
	}
	var document any
	if err := json.Unmarshal(documentBytes, &document); err != nil {
		resolver.failedDocuments[resolvedLocation] = true
		return nil, false
	}
	resolver.externalDocs[resolvedLocation] = document
	return document, true
}

func splitReference(reference string) (string, string) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return "", ""
	}
	if strings.HasPrefix(reference, "#") {
		return "", normalizePointer(strings.TrimPrefix(reference, "#"))
	}
	parts := strings.SplitN(reference, "#", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], normalizePointer(parts[1])
}

func normalizePointer(pointer string) string {
	pointer = strings.TrimSpace(pointer)
	if pointer == "" {
		return ""
	}
	if strings.HasPrefix(pointer, "/") {
		return pointer
	}
	return "/" + pointer
}

func resolveReferenceLocation(baseLocation, reference string) (string, error) {
	refURL, refErr := url.Parse(reference)
	if refErr == nil && refURL.IsAbs() {
		return refURL.String(), nil
	}
	baseURL, baseErr := url.Parse(baseLocation)
	if baseErr == nil && baseURL.IsAbs() {
		resolved := baseURL.ResolveReference(refURL)
		if resolved != nil && resolved.String() != "" {
			return resolved.String(), nil
		}
	}
	if filepath.IsAbs(reference) {
		return filepath.Clean(reference), nil
	}
	if strings.HasPrefix(baseLocation, "file://") {
		parsedBase, err := url.Parse(baseLocation)
		if err == nil {
			basePath := fileURLPath(parsedBase)
			if basePath != "" {
				return filepath.Clean(filepath.Join(filepath.Dir(basePath), filepath.FromSlash(reference))), nil
			}
		}
	}
	if baseLocation != "" && !strings.Contains(baseLocation, "://") {
		return filepath.Clean(filepath.Join(filepath.Dir(baseLocation), filepath.FromSlash(reference))), nil
	}
	return reference, nil
}

func readReferenceDocumentBytes(location string) ([]byte, error) {
	parsedLocation, err := url.Parse(location)
	if err == nil && parsedLocation.IsAbs() {
		switch strings.ToLower(parsedLocation.Scheme) {
		case "http", "https":
			response, err := http.Get(location)
			if err != nil {
				return nil, err
			}
			defer response.Body.Close()
			if response.StatusCode < 200 || response.StatusCode >= 300 {
				return nil, fmt.Errorf("failed to fetch reference %s: status %d", location, response.StatusCode)
			}
			return io.ReadAll(response.Body)
		case "file":
			return os.ReadFile(fileURLPath(parsedLocation))
		}
	}
	return os.ReadFile(location)
}

func fileURLPath(parsedURL *url.URL) string {
	if parsedURL == nil {
		return ""
	}
	filePath := parsedURL.Path
	if unescaped, err := url.PathUnescape(filePath); err == nil {
		filePath = unescaped
	}
	if len(filePath) >= 3 && filePath[0] == '/' && filePath[2] == ':' {
		filePath = filePath[1:]
	}
	if parsedURL.Host != "" && parsedURL.Host != "localhost" {
		filePath = `\\` + parsedURL.Host + filepath.FromSlash(filePath)
	}
	return filepath.FromSlash(filePath)
}

func resolveJSONPointer(document any, pointer string) (any, bool) {
	if pointer == "" {
		return document, true
	}
	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	current := document
	for _, rawPart := range parts {
		part := strings.ReplaceAll(strings.ReplaceAll(rawPart, "~1", "/"), "~0", "~")
		switch typed := current.(type) {
		case map[string]any:
			next, ok := typed[part]
			if !ok {
				return nil, false
			}
			current = next
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(typed) {
				return nil, false
			}
			current = typed[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func decodeSchema(value any) (Schema, bool) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return Schema{}, false
	}
	var schema Schema
	if err := json.Unmarshal(bytes, &schema); err != nil {
		return Schema{}, false
	}
	if schema.Type == "" && schema.Reference == "" && len(schema.Properties) == 0 && schema.Items == nil && len(schema.Enum) == 0 && len(schema.AllOf) == 0 && len(schema.AnyOf) == 0 && len(schema.OneOf) == 0 {
		return Schema{}, false
	}
	return schema, true
}

func resolveComposedSchema(spec OpenAPISpec, schema Schema, depth int) Schema {
	if depth > 20 {
		return schema
	}
	resolved := schema

	if resolved.Items != nil {
		item := resolveSchemaReference(spec, *resolved.Items, depth+1)
		resolved.Items = &item
	}
	if len(resolved.AllOf) > 0 {
		base := resolved
		base.AllOf = nil
		merged := base
		for _, part := range resolved.AllOf {
			merged = mergeSchemas(merged, resolveSchemaReference(spec, part, depth+1))
		}
		resolved = merged
	}

	if len(resolved.OneOf) > 0 {
		base := resolved
		base.OneOf = nil
		if discriminatorEnabled(resolved.Discriminator) {
			merged := base
			for _, part := range resolved.OneOf {
				merged = mergeSchemas(merged, resolveSchemaReference(spec, part, depth+1))
			}
			resolved = applyDiscriminatorSchemaHints(merged)
		} else {
			chosen := resolveSchemaReference(spec, resolved.OneOf[0], depth+1)
			resolved = mergeSchemas(base, chosen)
		}
	}
	if len(resolved.AnyOf) > 0 {
		base := resolved
		base.AnyOf = nil
		if discriminatorEnabled(resolved.Discriminator) {
			merged := base
			for _, part := range resolved.AnyOf {
				merged = mergeSchemas(merged, resolveSchemaReference(spec, part, depth+1))
			}
			resolved = applyDiscriminatorSchemaHints(merged)
		} else {
			chosen := resolveSchemaReference(spec, resolved.AnyOf[0], depth+1)
			resolved = mergeSchemas(base, chosen)
		}
	}
	return resolved
}

func mergeSchemas(primary Schema, secondary Schema) Schema {
	result := primary
	if result.Type == "" {
		result.Type = secondary.Type
	}
	if result.Description == "" {
		result.Description = secondary.Description
	}
	if result.Format == "" {
		result.Format = secondary.Format
	}
	if len(result.Enum) == 0 && len(secondary.Enum) > 0 {
		result.Enum = append([]any{}, secondary.Enum...)
	}
	if result.Items == nil && secondary.Items != nil {
		item := *secondary.Items
		result.Items = &item
	}
	if result.Properties == nil && secondary.Properties != nil {
		result.Properties = make(map[string]Schema, len(secondary.Properties))
	}
	for name, property := range secondary.Properties {
		if existing, ok := result.Properties[name]; ok {
			result.Properties[name] = mergeSchemas(existing, property)
			continue
		}
		result.Properties[name] = property
	}
	if result.Reference == "" {
		result.Reference = secondary.Reference
	}
	if result.AdditionalProperties == nil {
		result.AdditionalProperties = secondary.AdditionalProperties
	}
	if result.Discriminator == nil {
		result.Discriminator = secondary.Discriminator
	}
	result.OneOf = nil
	result.AnyOf = nil
	result.AllOf = nil
	return result
}

func discriminatorEnabled(discriminator *Discriminator) bool {
	return discriminator != nil && strings.TrimSpace(discriminator.PropertyName) != ""
}

func applyDiscriminatorSchemaHints(schema Schema) Schema {
	if !discriminatorEnabled(schema.Discriminator) {
		return schema
	}
	propertyName := strings.TrimSpace(schema.Discriminator.PropertyName)
	if schema.Type == "" {
		schema.Type = "object"
	}
	if !strings.EqualFold(schema.Type, "object") {
		return schema
	}
	if schema.Properties == nil {
		schema.Properties = make(map[string]Schema, 1)
	}
	if _, exists := schema.Properties[propertyName]; !exists {
		enums := make([]any, 0, len(schema.Discriminator.Mapping))
		for key := range schema.Discriminator.Mapping {
			enums = append(enums, key)
		}
		sort.Slice(enums, func(i, j int) bool {
			return fmt.Sprint(enums[i]) < fmt.Sprint(enums[j])
		})
		schema.Properties[propertyName] = Schema{
			Type: "string",
			Enum: enums,
		}
	}
	schema.Required = unionStrings(schema.Required, []string{propertyName})
	return schema
}

func unionStrings(left []string, right []string) []string {
	if len(left) == 0 && len(right) == 0 {
		return nil
	}
	combined := make(map[string]bool, len(left)+len(right))
	for _, value := range left {
		if value == "" {
			continue
		}
		combined[value] = true
	}
	for _, value := range right {
		if value == "" {
			continue
		}
		combined[value] = true
	}
	result := make([]string, 0, len(combined))
	for value := range combined {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func sampleValueFromSchema(spec OpenAPISpec, schema Schema, depth int) any {
	if depth > 20 {
		return nil
	}
	resolved := resolveSchemaReference(spec, schema, depth)
	if len(resolved.Enum) > 0 {
		return resolved.Enum[0]
	}
	typeName := resolved.Type
	if typeName == "" && len(resolved.Properties) > 0 {
		typeName = "object"
	}
	switch typeName {
	case "string":
		if resolved.Format == "date-time" {
			return time.Now().UTC().Format(time.RFC3339)
		}
		if resolved.Format == "date" {
			return time.Now().UTC().Format("2006-01-02")
		}
		return "string"
	case "integer":
		return 1
	case "number":
		return 1.0
	case "boolean":
		return true
	case "array":
		if resolved.Items == nil {
			return []any{}
		}
		itemValue := sampleValueFromSchema(spec, *resolved.Items, depth+1)
		if itemValue == nil {
			return []any{}
		}
		return []any{itemValue}
	case "object":
		propertyNames := make([]string, 0, len(resolved.Properties))
		for name := range resolved.Properties {
			propertyNames = append(propertyNames, name)
		}
		sort.Strings(propertyNames)
		result := make(map[string]any, len(propertyNames)+1)
		for _, propertyName := range propertyNames {
			propertySchema := resolved.Properties[propertyName]
			propertyValue := sampleValueFromSchema(spec, propertySchema, depth+1)
			if propertyValue != nil {
				result[propertyName] = propertyValue
			}
		}
		if discriminatorEnabled(resolved.Discriminator) {
			propertyName := strings.TrimSpace(resolved.Discriminator.PropertyName)
			if _, exists := result[propertyName]; !exists {
				result[propertyName] = sampleDiscriminatorValue(*resolved.Discriminator)
			}
		}
		if additionalSample, ok := sampleValueFromAdditionalProperties(spec, resolved.AdditionalProperties, depth+1); ok {
			result["additionalProperty1"] = additionalSample
		}
		return result
	default:
		return "value"
	}
}

func sampleDiscriminatorValue(discriminator Discriminator) string {
	if len(discriminator.Mapping) > 0 {
		keys := make([]string, 0, len(discriminator.Mapping))
		for key := range discriminator.Mapping {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		if len(keys) > 0 {
			return keys[0]
		}
	}
	return "type"
}

func sampleValueFromAdditionalProperties(spec OpenAPISpec, additionalProperties any, depth int) (any, bool) {
	if depth > 20 || additionalProperties == nil {
		return nil, false
	}
	switch typed := additionalProperties.(type) {
	case bool:
		if typed {
			return "value", true
		}
		return nil, false
	case Schema:
		value := sampleValueFromSchema(spec, typed, depth+1)
		if value == nil {
			return "value", true
		}
		return value, true
	case map[string]any:
		var schema Schema
		bytes, err := json.Marshal(typed)
		if err != nil {
			return "value", true
		}
		if err := json.Unmarshal(bytes, &schema); err != nil {
			return "value", true
		}
		value := sampleValueFromSchema(spec, schema, depth+1)
		if value == nil {
			return nil, false
		}
		return value, true
	default:
		return "value", true
	}
}

func GenTestCaseTitleFromPathAndMethod(pathName, method string) string {
	methodNameOperationToAllCapps := strings.ToUpper(method)
	pathNameParts := strings.Split(pathName, "/")
	title := methodNameOperationToAllCapps
	for _, part := range pathNameParts {
		if part == "" {
			continue
		}
		title += "_" + part
	}
	return title
}

func buildTestCaseTitle(method, pathName string) string {
	return strings.ToUpper(method) + " " + pathName
}

func buildNegativeTestCaseTitle(method, pathName, qualifier string) string {
	base := buildTestCaseTitle(method, pathName)
	qualifier = strings.TrimSpace(qualifier)
	if qualifier == "" {
		return base + " Negative"
	}
	return base + " " + qualifier
}

func negativeQualifierFileSegment(qualifier string) string {
	segment := strings.TrimSpace(qualifier)
	if segment == "" {
		return "Negative"
	}
	if index := strings.Index(segment, ":"); index >= 0 {
		segment = strings.TrimSpace(segment[:index])
	}
	tokens := strings.FieldsFunc(segment, func(r rune) bool {
		if r >= '0' && r <= '9' {
			return false
		}
		if r >= 'A' && r <= 'Z' {
			return false
		}
		if r >= 'a' && r <= 'z' {
			return false
		}
		return true
	})
	if len(tokens) == 0 {
		return "Negative"
	}
	for i := range tokens {
		tokens[i] = sanitizeCaseName(tokens[i])
	}
	return strings.Join(tokens, "_")
}

func GetOpenAPIPathNames(data OpenAPISpec) []string {
	pathsSection := data.Paths
	paths := make([]string, 0, len(pathsSection))
	for path := range pathsSection {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func NextTestCaseJSONNumber(testSectorCamelCase, testSectorAbbr string) string {
	files := testingToolkit.ListFilesInFolder(testingToolkit.CurrPath() + "/tests/" + testSectorCamelCase + "Tests/testCases")
	if len(files) == 0 {
		return "0001"
	}
	prefix := "TC_" + testSectorAbbr
	maxNumber := 0
	for _, file := range files {
		if !strings.HasPrefix(file, prefix) {
			continue
		}
		_, after, ok := strings.Cut(file, testSectorAbbr)
		if !ok || after == "" {
			continue
		}
		numPart, _, _ := strings.Cut(after, "_")
		if numPart == "" {
			continue
		}
		n := testingToolkit.ConvertStringToInt(numPart)
		if n > maxNumber {
			maxNumber = n
		}
	}
	next := maxNumber + 1
	nextStr := testingToolkit.ConvertIntToString(next)
	if len(nextStr) < 4 {
		nextStr = strings.Repeat("0", 4-len(nextStr)) + nextStr
	}
	return nextStr
}

func ReturnBodyKindIntFromContentType(contentType string) int {
	switch contentType {
	case "application/json":
		return 1
	case "application/x-www-form-urlencoded":
		return 5
	case "multipart/form-data":
		return 4
	case "application/octet-stream":
		return 2
	case "text/plain":
		return 2
	default:
		return 0
	}
}

func ArchivePreviousFiles(testSectorCamelCase string) error {
	testFolder := testingToolkit.CurrPath() + "/tests/" + testSectorCamelCase + "Tests"
	testCasesFolder := testFolder + "/testCases"
	archiveFolder := testFolder + "/archive"
	oldCodeFile := testFolder + "/" + testSectorCamelCase + "Tests.go"
	if !testingToolkit.VerifyFolderIsPresent(testFolder) {
		return nil
	}
	files := []string{}
	if testingToolkit.VerifyFolderIsPresent(testCasesFolder) {
		files = testingToolkit.ListFilesInFolder(testCasesFolder)
	}
	hasArchivableJSON := false
	for _, file := range files {
		if strings.HasPrefix(file, "openAPIJson") && strings.HasSuffix(file, ".json") {
			hasArchivableJSON = true
			break
		}
	}
	hasCodeFile := testingToolkit.VerifyFileIsPresent(oldCodeFile)
	if !hasArchivableJSON && !hasCodeFile {
		return nil
	}
	if err := os.MkdirAll(archiveFolder, 0755); err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasPrefix(file, "TC_") && strings.HasSuffix(file, ".json") {
			oldPath := testCasesFolder + "/" + file
			newPath := archiveFolder + "/" + file
			err := os.Rename(oldPath, newPath)
			if err != nil {
				log.Println("Unable to archive file: " + file + " Error: " + err.Error())
				continue
			}
		}
	}
	archiveCodeFile := archiveFolder + "/" + testSectorCamelCase + "Tests.go"
	if hasCodeFile {
		_ = os.Remove(archiveCodeFile)
		err := copyFileContents(oldCodeFile, archiveCodeFile)
		if err != nil {
			log.Println("Unable to archive code file: " + oldCodeFile + " Error: " + err.Error())
		}
	}
	return nil
}

func copyFileContents(sourcePath, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

func slicesContains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
