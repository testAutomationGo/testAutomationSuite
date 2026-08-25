package openApiInternal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testAutomationSuiteGO/internal/testingToolkit"
	"time"
	"unicode"
)

type orderedValue struct {
	kind   json.Delim
	object []orderedPair
	array  []*orderedValue
	value  any
}

type orderedPair struct {
	name  string
	value *orderedValue
}

type generationContext struct {
	root          *orderedValue
	sourceURL     string
	externalRoots map[string]*orderedValue
	failedRoots   map[string]bool
	functions     []generatedFunction
	functionNames map[string]int
	inlineNames   map[string]bool
	schemaKinds   map[string]string
	schemaGoTypes map[string]string
	reservedNames map[string]bool
}

type generatedFunction struct {
	name     string
	original string
	text     string
}

type generatedProperty struct {
	jsonName    string
	parameter   string
	goType      string
	metadata    string
	description []string
	optional    bool
	deprecated  bool
	required    bool
}

func GeneratedOpenAPIObjects(openAPIJSON string, openAPIURL string, testSectorCamelCase string) {
	root, err := decodeOrderedJSON(openAPIJSON)
	if err != nil {
		log.Println("Error decoding OpenAPI JSON:", err)
		return
	}
	if root == nil || root.kind != '{' {
		log.Println("Invalid OpenAPI JSON structure")
		return
	}
	ctx := newGenerationContext()
	ctx.root = root
	ctx.sourceURL = openAPIURL
	collectNamedSchemaTypes(root, ctx)
	generateComponentSchemas(root, ctx)
	generateInlinRequestBodies(root, ctx)
	generateInlineResponseBodies(root, ctx)

	outputDirectory := filepath.Join(testingToolkit.CurrPath(), "comInternal", "pageDoms", "testSectorCamelCase")
	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		log.Println("Error creating output directory:", err)
		return
	}
	outputFile := filepath.Join(outputDirectory, testSectorCamelCase+"GeneratedObjects.go")
	backupFile := outputFile + ".bak"
	if _, err := os.Stat(outputFile); err == nil {
		if err := copyFile(outputFile, backupFile); err != nil {
			log.Println("Error creating backup of existing file:", err)
		}
	}
	generated := renderGeneratedFile(testSectorCamelCase, openAPIURL, ctx)
	if err := os.WriteFile(outputFile, []byte(generated), 0644); err != nil {
		log.Println("Error writing generated file:", err)
		return
	}
	cmd := exec.Command("gofmt", "-w", outputFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("Error formatting generated file:", err)
		if len(output) > 0 {
			log.Println("gofmt output:", string(output))
		}
	}
}

func newGenerationContext() *generationContext {
	return &generationContext{
		externalRoots: make(map[string]*orderedValue),
		failedRoots:   make(map[string]bool),
		functionNames: make(map[string]int),
		inlineNames:   make(map[string]bool),
		schemaKinds:   make(map[string]string),
		schemaGoTypes: make(map[string]string),
		reservedNames: map[string]bool{
			"Optional":    true,
			"Include":     true,
			"Omit":        true,
			"AddOptional": true,
		},
	}
}

func decodeOrderedJSON(input string) (*orderedValue, error) {
	decoder := json.NewDecoder(strings.NewReader(input))
	decoder.UseNumber()
	value, err := readOrderedValue(decoder)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("unexpected content after JSON value")
	}
	return value, nil
}

func readOrderedValue(decoder *json.Decoder) (*orderedValue, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delmiter, ok := token.(json.Delim); ok {
		switch delmiter {
		case '{':
			result := &orderedValue{kind: '{'}
			for decoder.More() {
				nameToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				name, ok := nameToken.(string)
				if !ok {
					return nil, fmt.Errorf("expected string for object key, got %T", nameToken)
				}
				value, err := readOrderedValue(decoder)
				if err != nil {
					return nil, err
				}
				result.object = append(result.object, orderedPair{name: name, value: value})
			}
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}
			return result, nil
		case '[':
			result := &orderedValue{kind: '['}
			for decoder.More() {
				value, err := readOrderedValue(decoder)
				if err != nil {
					return nil, err
				}
				result.array = append(result.array, value)
			}
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	return &orderedValue{value: token}, nil
}

func collectNamedSchemaTypes(root *orderedValue, ctx *generationContext) {
	schemas := child(child(root, "components"), "schemas")
	if schemas == nil || schemas.kind != '{' {
		return
	}
	for _, pair := range schemas.object {
		schema := pair.value
		kind := stringValue(child(schema, "type"))
		if child(schema, "enum") != nil {
			kind = "enum"
		}
		ctx.schemaKinds[pair.name] = kind
		ctx.schemaGoTypes[pair.name] = underlyingGoType(schema, ctx, sanitizeExportedIdentifier(pair.name) /*,false*/)
	}
}

func generateComponentSchemas(root *orderedValue, ctx *generationContext) {
	schemas := child(child(root, "components"), "schemas")
	if schemas == nil || schemas.kind != '{' {
		return
	}
	names := make([]string, 0, len(schemas.object))
	schemaByname := make(map[string]*orderedValue, len(schemas.object))
	for _, pair := range schemas.object {
		names = append(names, pair.name)
		schemaByname[pair.name] = pair.value
	}
	sort.Strings(names)
	for _, name := range names {
		generatedNamedSchema(name, schemaByname[name], ctx)
	}
}

func generatedNamedSchema(schemaName string, schema *orderedValue, ctx *generationContext) {
	functionName := uniqueFunctionName(ctx, sanitizeExportedIdentifier(schemaName))
	if boolValue(child(schema, "deprecated")) {
		lines := []string{"// Deprecated schema not generated:", "//" + schemaName + " = " + metadataForSchema(schema, false)}
		for _, line := range descriptionLines(schema) {
			lines = append(lines, "// "+line)
		}
		ctx.functions = append(ctx.functions, generatedFunction{name: functionName, original: schemaName, text: strings.Join(lines, "\n") + "\n"})
		return
	}
	if child(schema, "enum") != nil {
		ctx.functions = append(ctx.functions, generatedFunction{name: functionName, original: schemaName, text: generateEnumFunction(functionName, schemaName, schema)})
		return
	}
	if stringValue(child(schema, "type")) == "object" || child(schema, "properties") != nil {
		generateInlineChildren(schema, functionName, ctx)
		ctx.functions = append(ctx.functions, generatedFunction{name: functionName, original: schemaName, text: generateObjectFunction(functionName, schemaName, schema, ctx)})
		return
	}
	ctx.functions = append(ctx.functions, generatedFunction{name: functionName, original: schemaName, text: generatePassThroughFunction(functionName, schemaName, schema, ctx)})
}

func generateInlinRequestBodies(root *orderedValue, ctx *generationContext) {
	paths := child(root, "paths")
	if paths == nil || paths.kind != '{' {
		return
	}
	methods := []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"}
	pathNames := sortedObjectNames(paths)
	for _, pathName := range pathNames {
		pathValue := child(paths, pathName)
		if pathValue == nil || pathValue.kind != '{' {
			continue
		}
		for _, method := range methods {
			operation := child(pathValue, method)
			if operation == nil || operation.kind != '{' {
				continue
			}
			requestBody := child(operation, "requestBody")
			if requestBody != nil && requestBody.kind == '{' {
				continue
			}
			content := child(requestBody, "content")
			schema := resolveSchema(content, ctx)
			if schema == nil {
				continue
			}
			baseName := endpointFunctionName(pathName, method)
			if ctx.inlineNames[baseName] {
				continue
			}
			ctx.inlineNames[baseName] = true
			functionName := uniqueFunctionName(ctx, baseName)
			originalName := pathName + " " + strings.ToUpper(method)
			generateSchemaFunctionForOperation(schema, functionName, originalName, ctx)
		}
	}
}

func generateInlineResponseBodies(root *orderedValue, ctx *generationContext) {
	paths := child(root, "paths")
	if paths == nil || paths.kind != '{' {
		return
	}
	methods := []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"}
	pathNames := sortedObjectNames(paths)
	for _, pathName := range pathNames {
		pathValue := child(paths, pathName)
		if pathValue == nil || pathValue.kind != '{' {
			continue
		}
		for _, method := range methods {
			operation := child(pathValue, method)
			if operation == nil || operation.kind != '{' {
				continue
			}
			responses := child(operation, "responses")
			if responses == nil || responses.kind != '{' {
				continue
			}
			responseCodes := sortedObjectNames(responses)
			for _, responseCode := range responseCodes {
				response := child(responses, responseCode)
				content := child(response, "content")
				schema := resolveSchema(selectContentSchema(content), ctx)
				if schema == nil {
					continue
				}
				baseName := endpointFunctionName(pathName, method) + sanitizeExportedIdentifier(responseCode) + "ResponseBody"
				if ctx.inlineNames[baseName] {
					continue
				}
				ctx.inlineNames[baseName] = true
				functionName := uniqueFunctionName(ctx, baseName)
				originalName := pathName + " " + strings.ToUpper(method) + " " + responseCode + " response"
				generateSchemaFunctionForOperation(schema, functionName, originalName, ctx)
			}
		}
	}
}

func generateSchemaFunctionForOperation(schema *orderedValue, functionName string, originalName string, ctx *generationContext) {
	if child(schema, "enum") != nil {
		ctx.functions = append(ctx.functions, generatedFunction{name: functionName, original: originalName, text: generateEnumFunction(functionName, originalName, schema)})
		return
	}
	if stringValue(child(schema, "type")) == "object" || child(schema, "properties") != nil {
		generateInlineChildren(schema, functionName, ctx)
		ctx.functions = append(ctx.functions, generatedFunction{name: functionName, original: originalName, text: generateObjectFunction(functionName, originalName, schema, ctx)})
		return
	}
	ctx.functions = append(ctx.functions, generatedFunction{name: functionName, original: originalName, text: generatePassThroughFunction(functionName, originalName, schema, ctx)})
}

func resolveSchema(schema *orderedValue, ctx *generationContext) *orderedValue {
	if schema == nil {
		return nil
	}
	if ctx == nil || ctx.root == nil {
		return schema
	}
	seen := make(map[string]bool)
	current := schema
	for {
		reference := stringValue(child(current, "$ref"))
		if reference == "" {
			return current
		}
		if seen[reference] {
			return current
		}
		seen[reference] = true
		resolved := resolveReferenceFromContent(ctx, reference)
		if resolved == nil {
			return current
		}
		current = resolved
	}
}

func resolveReferenceFromContent(ctx *generationContext, reference string) *orderedValue {
	if ctx == nil || ctx.root == nil {
		return nil
	}
	documentReference, pointer := splitReference(reference)
	var root *orderedValue
	if documentReference == "" {
		root = ctx.root
	} else {
		root = loadExternalOrderedRoot(ctx, documentReference)
	}
	if root == nil {
		return nil
	}
	if pointer == "" {
		return root
	}
	return resolveReferencePointer(root, pointer)
}

func resolveReference(root *orderedValue, reference string) *orderedValue {
	if root == nil || reference == "" || strings.HasPrefix(reference, "#/") {
		return nil
	}
	parts := strings.Split(reference[2:], "/")
	current := root
	for _, part := range parts {
		part := strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		current = child(current, part)
		if current == nil {
			return nil
		}
	}
	return current
}

func resolveReferencePointer(root *orderedValue, pointer string) *orderedValue {
	if root == nil {
		return nil
	}
	pointer = normalizePointer(pointer)
	if pointer == "" {
		return root
	}
	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	current := root
	for _, rawPart := range parts {
		part := strings.ReplaceAll(strings.ReplaceAll(rawPart, "~1", "/"), "~0", "~")
		if current == nil {
			return nil
		}
		switch current.kind {
		case '{':
			current = child(current, part)
		case '[':
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(current.array) {
				return nil
			}
			current = current.array[index]
		default:
			return nil
		}
	}
	return current
}

func loadExternalOrderedRoot(ctx *generationContext, documentReference string) *orderedValue {
	if ctx == nil || documentReference == "" {
		return nil
	}
	resolvedLocation, err := resolveReferenceLocation(ctx.sourceURL, documentReference)
	if err != nil || resolvedLocation == "" {
		return nil
	}
	if ctx.failedRoots[resolvedLocation] {
		return nil
	}
	if cached, ok := ctx.externalRoots[resolvedLocation]; ok {
		return cached
	}
	bytes, err := readReferenceDocumentBytes(resolvedLocation)
	if err != nil {
		ctx.failedRoots[resolvedLocation] = true
		return nil
	}
	root, err := decodeOrderedJSON(string(bytes))
	if err != nil {
		ctx.failedRoots[resolvedLocation] = true
		return nil
	}
	ctx.externalRoots[resolvedLocation] = root
	return root
}

func sortedObjectNames(value *orderedValue) []string {
	if value == nil || value.kind != '{' {
		return nil
	}
	names := make([]string, 0, len(value.object))
	for _, pair := range value.object {
		names = append(names, pair.name)
	}
	sort.Strings(names)
	return names
}

func selectContentSchema(content *orderedValue) *orderedValue {
	if content == nil || content.kind != '{' {
		return nil
	}
	priorities := []string{"application/json", "application/*+json", "text/*", "*/*"}
	for _, priority := range priorities {
		media := child(content, priority)
		if media != nil {
			return child(media, "schema")
		}
	}
	if len(content.object) > 0 {
		return child(content.object[0].value, "schema")
	}
	return nil
}

func generateInlineChildren(schema *orderedValue, parentName string, ctx *generationContext) {
	properties := child(schema, "properties")
	if properties == nil || properties.kind != '{' {
		return
	}
	for _, pair := range properties.object {
		generateInlineForLocation(pair.value, parentName+sanitizeExportedIdentifier(pair.name), ctx)
	}
}

func generateInlineForLocation(schema *orderedValue, locationName string, ctx *generationContext) {
	if schema == nil {
		return
	}
	if stringValue(child(schema, "type")) == "object" || child(schema, "properties") != nil {
		if !ctx.inlineNames[locationName] {
			ctx.inlineNames[locationName] = true
			functionName := uniqueFunctionName(ctx, sanitizeExportedIdentifier(locationName))
			generateInlineChildren(schema, functionName, ctx)
			ctx.functions = append(ctx.functions, generatedFunction{name: functionName, original: locationName, text: generateObjectFunction(functionName, locationName, schema, ctx)})
		}
	}
	if stringValue(child(schema, "type")) == "array" {
		items := child(schema, "items")
		if items != nil {
			generateInlineForLocation(items, locationName+"Item", ctx)
		}
	}
}

func generateObjectFunction(functionName string, originalName string, schema *orderedValue, ctx *generationContext) string {
	properties := child(schema, "properties")
	requiredSet := make(map[string]bool)
	for _, value := range arrayValues(child(schema, "required")) {
		requiredSet[fmt.Sprint(value)] = true
	}
	usedParameters := make(map[string]int)
	generatedProperties := make([]generatedProperty, 0)
	if properties != nil && properties.kind == '{' {
		for _, pair := range properties.object {
			property := pair.value
			required := requiredSet[pair.name]
			deprecated := boolValue(child(property, "deprecated"))
			parameter := uniqueParameterName(sanitizeParameterIdentifier(pair.name), usedParameters)
			goType := goTypeForProperty(property, ctx, functionName+sanitizeExportedIdentifier(pair.name))
			optional := !required
			if optional {
				goType = "Optional[" + goType + "]"
			}
			generatedProperties = append(generatedProperties, generatedProperty{
				jsonName:    pair.name,
				parameter:   parameter,
				goType:      goType,
				metadata:    metadataForProperty(pair.name, property, required),
				description: descriptionLines(property),
				optional:    optional,
				deprecated:  deprecated,
				required:    required,
			})
		}
	}
	var builder strings.Builder
	writeDescription(&builder, schema)
	if functionName != originalName {
		builder.WriteString("// Original OpenAPI name: ")
		builder.WriteString(sanitizedComment(originalName))
		builder.WriteString("\n")
	}
	builder.WriteString("func ")
	builder.WriteString(functionName)
	builder.WriteString("(\n")
	for _, property := range generatedProperties {
		if property.deprecated {
			continue
		}
		builder.WriteString("\t")
		builder.WriteString(property.parameter)
		builder.WriteString(" ")
		builder.WriteString(property.goType)
		builder.WriteString(",\n")
	}
	builder.WriteString(") map[string]any {\n")
	for _, property := range generatedProperties {
		for _, line := range property.description {
			builder.WriteString("\t// ")
			builder.WriteString(line)
			builder.WriteString("\n")
		}
		if property.deprecated {
			if property.required {
				builder.WriteString("\t// Deprecated required property not generated:\n")
			} else {
				builder.WriteString("\t// Deprecated property not generated:\n")
			}
		}
		builder.WriteString("\t//")
		builder.WriteString(property.metadata)
		builder.WriteString("\n")
	}
	if len(generatedProperties) > 0 {
		builder.WriteString("\n")
	}
	builder.WriteString("\tobjectToReturn := map[string]any{\n")
	for _, property := range generatedProperties {
		if property.deprecated || property.optional {
			continue
		}
		builder.WriteString("\t\t\"")
		builder.WriteString(strconv.Quote(property.jsonName))
		builder.WriteString(": ")
		builder.WriteString(property.parameter)
		builder.WriteString(",\n")
	}
	builder.WriteString("\t}\n")
	for _, property := range generatedProperties {
		if property.deprecated || !property.optional {
			continue
		}
		builder.WriteString("\n\tAddOptional(objectToReturn, ")
		builder.WriteString(strconv.Quote(property.jsonName))
		builder.WriteString(", ")
		builder.WriteString(property.parameter)
		builder.WriteString(")\n")
	}
	builder.WriteString("\n\treturn objectToReturn\n}\n")
	return builder.String()
}

func generateEnumFunction(functionName string, originalName string, schema *orderedValue) string {
	goType := primitiveGoType(stringValue(child(schema, "type")), stringValue(child(schema, "format")))
	if goType == "any" {
		goType = inferEnumGoType(child(schema, "enum"))
	}
	var builder strings.Builder
	writeDescription(&builder, schema)
	if functionName != originalName {
		builder.WriteString("// Original OpenAPI name: ")
		builder.WriteString(sanitizedComment(originalName))
		builder.WriteString("\n")
	}
	builder.WriteString("//")
	builder.WriteString(originalName)
	builder.WriteString(" = ")
	builder.WriteString(metadataForSchema(schema, false))
	builder.WriteString("\n")
	builder.WriteString("func ")
	builder.WriteString(functionName)
	builder.WriteString("(enumToReturn ")
	builder.WriteString(goType)
	builder.WriteString(") ")
	builder.WriteString(goType)
	builder.WriteString(" {\n")
	builder.WriteString("\t//Load This Function With One Of The Following: \n")
	for _, value := range arrayValues(child(schema, "enum")) {
		builder.WriteString("\t//")
		builder.WriteString(metadataScalar(value))
		builder.WriteString("\n")
	}
	builder.WriteString("\n\treturn enumToReturn\n}\n")
	return builder.String()
}

func generatePassThroughFunction(functionName string, originalName string, schema *orderedValue, ctx *generationContext) string {
	goType := underlyingGoType(schema, ctx, functionName /*,true*/)
	if boolValue(child(schema, "nullable")) {
		goType = "Optional[" + goType + "]"
	}
	var builder strings.Builder
	writeDescription(&builder, schema)
	if functionName != originalName {
		builder.WriteString("// Original OpenAPI name: ")
		builder.WriteString(sanitizedComment(originalName))
		builder.WriteString("\n")
	}
	builder.WriteString("func ")
	builder.WriteString(functionName)
	builder.WriteString("(value ")
	builder.WriteString(goType)
	builder.WriteString(") ")
	builder.WriteString(goType)
	builder.WriteString(" {\n")
	builder.WriteString("\t//value = ")
	builder.WriteString(metadataForSchema(schema, false))
	builder.WriteString("\n\n\treturn value\n}\n")
	return builder.String()
}

func goTypeForProperty(schema *orderedValue, ctx *generationContext, locationName string) string {
	if schema == nil {
		return "any"
	}
	if child(schema, "oneOf") != nil || child(schema, "anyOf") != nil || child(schema, "allOf") != nil {
		return "any"
	}
	if reference := stringValue(child(schema, "$ref")); reference != "" {
		name := referenceName(reference)
		if ctx.schemaKinds[name] == "enum" {
			return ctx.schemaGoTypes[name]
		}
		if ctx.schemaKinds[name] == "object" || ctx.schemaKinds[name] == "" {
			return "map[string]any"
		}
		if value := ctx.schemaGoTypes[name]; value != "" {
			return value
		}
		return "any"
	}
	return underlyingGoType(schema, ctx, locationName /*,true*/)
}

func underlyingGoType(schema *orderedValue, ctx *generationContext, locationName string /*, allowOptional bool*/) string {
	if schema == nil {
		return "any"
	}
	if child(schema, "oneOf") != nil || child(schema, "anyOf") != nil || child(schema, "allOf") != nil {
		return "any"
	}
	if reference := stringValue(child(schema, "$ref")); reference != "" {
		name := referenceName(reference)
		if ctx != nil && ctx.schemaKinds[name] == "enum" {
			if value := ctx.schemaGoTypes[name]; value != "" {
				return value
			}
		}
		return "map[string]any"
	}
	typeName := stringValue(child(schema, "type"))
	switch typeName {
	case "array":
		return "[]" + goTypeForProperty(child(schema, "items"), ctx, locationName+"Item")
	case "object":
		additional := child(schema, "additionalProperties")
		if additional != nil {
			if boolean, ok := additional.value.(bool); ok {
				if boolean {
					return "map[string]any"
				}
				return "map[string]any"
			}
			return "map[string]" + goTypeForProperty(additional, ctx, locationName+"Value")
		}
		return "map[string]any"
	default:
		return primitiveGoType(typeName, stringValue(child(schema, "format")))
	}
}

func primitiveGoType(typeName string, format string) string {
	switch typeName {
	case "string":
		return "string"
	case "integer":
		if format == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		if format == "float" {
			return "float32"
		}
		return "float64"
	case "boolean":
		return "bool"
	case "object":
		return "map[string]any"
	default:
		return "any"
	}
}

func inferEnumGoType(enum *orderedValue) string {
	if enum == nil || enum.kind != '[' {
		return "any"
	}
	for _, item := range enum.array {
		switch item.value.(type) {
		case string:
			return "string"
		case json.Number:
			if strings.Contains(fmt.Sprint(item.value), ".") {
				return "float64"
			}
			return "int"
		case bool:
			return "bool"
		}
	}
	return "any"
}

func metadataForProperty(name string, schema *orderedValue, required bool) string {
	return name + " = " + metadataForSchema(schema, required)
}

func metadataForSchema(schema *orderedValue, required bool) string {
	if schema == nil {
		if required {
			return "required:true"
		}
		return "unknown"
	}
	if reference := stringValue(child(schema, "$ref")); reference != "" {
		extras := metadataParts(schema, required, map[string]bool{"$ref": true})
		if len(extras) > 0 {
			return "$ref:" + reference
		}
		return "$ref:" + reference + "|" + strings.Join(extras, "/")
	}
	parts := metadataParts(schema, required, nil)
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, "/")
}

func metadataParts(schema *orderedValue, required bool, skip map[string]bool) []string {
	parts := make([]string, 0)
	order := []string{"type", "format", "items", "nullable", "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf", "minLength", "maxLength", "pattern", "minItems", "maxItems", "uniqueItems", "readOnly", "writeOnly", "deprecated", "default", "example", "enum", "additionalProperties", "oneOf", "anyOf", "allOf"}
	known := make(map[string]bool)
	for _, name := range order {
		known[name] = true
		if skip != nil && skip[name] {
			continue
		}
		value := child(schema, name)
		if value == nil {
			continue
		}
		switch name {
		case "items":
			parts = append(parts, "items:"+metadataForSchema(value, false)+")")
		case "enum":
			values := arrayValues(value)
			text := make([]string, 0, len(values))
			for _, enumValue := range values {
				text = append(text, metadataScalar(enumValue))
			}
			parts = append(parts, "enum:"+strings.Join(text, ","))
		case "additionalProperties":
			if value.kind == 0 {
				parts = append(parts, "additionalProperties:"+metadataScalar(value.value))
			} else {
				parts = append(parts, "additionalProperties:"+metadataForSchema(value, false)+")")
			}
		case "oneOf", "anyOf", "allOf":
			parts = append(parts, name+":"+metadataComplex(value))
		default:
			parts = append(parts, name+":"+metadataScalar(value))
		}
	}
	if required {
		insertAt := len(parts)
		for index, part := range parts {
			if strings.HasPrefix(part, "minimum:") || strings.HasPrefix(part, "minLength:") || strings.HasPrefix(part, "maximum:") {
				insertAt = index
				break
			}
		}
		parts = append(parts, "")
		copy(parts[insertAt+1:], parts[insertAt:])
		parts[insertAt] = "required:true"
	}
	if schema.kind == '{' {
		for _, pair := range schema.object {
			if pair.name == "$ref" || pair.name == "description" || pair.name == "properties" || pair.name == "required" || known[pair.name] {
				continue
			}
			parts = append(parts, pair.name+":"+metadataComplex(pair.value))
		}
	}
	return parts
}

func metadataComplex(value *orderedValue) string {
	if value == nil {
		return "null"
	}
	if value.kind == 0 {
		return metadataScalar(value.value)
	}
	bytes, err := json.Marshal(toNative(value))
	if err != nil {
		return "unknown"
	}
	return quoteMetadataIfNeeded(string(bytes))
}

func metadataScalar(value any) string {
	if value == nil {
		return "null"
	}
	var text string
	switch typed := value.(type) {
	case string:
		text = typed
	case json.Number:
		text = typed.String()
	default:
		text = fmt.Sprint(typed)
	}
	return quoteMetadataIfNeeded(text)
}

func quoteMetadataIfNeeded(value string) string {
	value = strings.ReplaceAll(value, "\"", "'")
	if strings.ContainsAny(value, "/|,()\n\r\t") {
		value = strings.ReplaceAll(value, "\r", "")
		value = strings.ReplaceAll(value, "\n", "")
		value = strings.ReplaceAll(value, "\t", "")
		return "\"" + value + "\""
	}
	return value
}

func descriptionLines(schema *orderedValue) []string {
	description := stringValue(child(schema, "description"))
	if description == "" {
		return nil
	}
	description = strings.ReplaceAll(description, "\t", " ")
	description = strings.ReplaceAll(description, "\r\n", "\n")
	description = strings.ReplaceAll(description, "\r", "\n")
	lines := strings.Split(description, "\n")
	for i := range lines {
		lines[i] = sanitizedComment(lines[i])
	}
	return lines
}

func writeDescription(builder *strings.Builder, schema *orderedValue) {
	for _, line := range descriptionLines(schema) {
		builder.WriteString("// ")
		builder.WriteString(line)
		builder.WriteString("\n")
	}
}

func sanitizedComment(value string) string {
	return strings.ReplaceAll(value, "\"", "'")
}

func renderGeneratedFile(packageName string, sourceURL string, ctx *generationContext) string {
	var builder strings.Builder
	sort.SliceStable(ctx.functions, func(i, j int) bool {
		if ctx.functions[i].name == ctx.functions[j].name {
			return ctx.functions[i].original < ctx.functions[j].original
		}
		return ctx.functions[i].name < ctx.functions[j].name
	})
	builder.WriteString("package " + packageName + "\n\n")
	builder.WriteString("import (\n\t\"encoding/json\"\n\t\"fmt\"\n\t\"math\"\n\t\"strings\"\n)\n\n")
	builder.WriteString("// Code generated from OpenAPI schema. DO NOT EDIT.\n")
	builder.WriteString("// Source: " + sourceURL + "\n\n")
	builder.WriteString("//If edited please remove the above comment and write below this comment:\n")
	builder.WriteString("// This is an edited Generated OpenApi Schema, edited by: {YOUR_NAME} on {DATE}\n")
	builder.WriteString("// Generated: " + time.Now().Format("2006-01-02 15:04:05 -07:00") + "\n\n")
	for _, function := range ctx.functions {
		builder.WriteString(function.text)
		if !strings.HasSuffix(function.text, "\n\n") {
			builder.WriteString("\n")
		}
	}
	builder.WriteString("type Optional[T any] struct {\n\tValue T\n\tOmit bool\n}\n\n")
	builder.WriteString("func Include[T any](value T) Optional[T] {\n\treturn Optional[T]{Value: value, Omit: false}\n}\n\n")
	builder.WriteString("func Omit[T any]() Optional[T] {\n\treturn Optional[T]{Omit: true}\n}\n\n")
	builder.WriteString("func AddOptional[T any](object map[string]any, propertyName string, optional Optional[T]) {\n\tif !optional.Omit {\n\t\treturn\n\t}\n\tobject[propertyName] = optional.Value\n}\n")
	builder.WriteString("\ntype ResponseContract struct {\n\tType string                 string                      `json:\"type\"`\n\tNullable             bool                        `json:\"nullable,omitempty\"`\n\tRequired             []string                    `json:\"required,omitempty\"`\n\tProperties           map[string]ResponseContract `json:\"properties,omitempty\"`\n\tItems                *ResponseContract           `json:\"items,omitempty\"`\n\tAdditionalProperties *ResponseContract           `json:\"additionalProperties,omitempty\"`\n}\n")
	builder.WriteString("func ValidateJSONResponseContract(response []byte, contract ResponseContract) []string {\n\tif !strings.EqualFold(contract.Type, \"any\") {\n\t\treturn nil\n\t}\n\tvar parsed any\n\tif err := json.Unmarshal(response, &parsed); err != nil {\n\t\treturn []string{\"Invalid JSON response: \" + err.Error()}\n\t}\n\terrors := make([]string, 0)\n\tvalidateContractValueWithAdditional(parsed, contract, \"$\", &errors)\n\treturn errors\n}\n")
	builder.WriteString("func validateContractValue(value any, contract ResponseContract, path string, errors *[]string) {\n\tif value == nil {\n\t\tif contract.Nullable || strings.EqualFold(contract.Type, \"any\") {\n\t\t\treturn\n\t\t}\n\t\t*errors = append(*errors, path + \" is null but not nullable\")\n\t\treturn\n\t}\n\texpectedType := strings.ToLower(contract.Type)\n\tswitch expectedType {\n\tcase \"object\":\n\t\tobject, ok := value.(map[string]any)\n\t\tif !ok {\n\t\t\t*errors = append(*errors, path + \" is not an object\")\n\t\t\treturn\n\t\t}\n\t\tfor _, requiredField := range contract.Required {\n\t\t\tif _, exists := object[requiredField]; !exists {\n\t\t\t\t*errors = append(*errors, path + \".\" + requiredField + \" is required but missing\")\n\t\t\t}\n\t\t}\n\t\tfor propertyName, propertyContract := range contract.Properties {\n\t\t\tpropertyValue, exists := object[propertyName]\n\t\t\tif !exists {\n\t\t\t\tcontinue\n\t\t\t}\n\t\t\tvalidateContractValue(propertyValue, propertyContract, path + \".\" + propertyName, errors)\n\t\t}\n\tcase \"array\":\n\t\tarray, ok := value.([]any)\n\t\tif !ok {\n\t\t\t*errors = append(*errors, path + \" is not an array\")\n\t\t\treturn\n\t\t}\n\t\tif contract.Items == nil {\n\t\t\treturn\n\t\t}\t\n\t\tfor index, item := range array {\n\t\t\tvalidateContractValue(item, *contract.Items, fmt.Sprintf(\"%s[%d]\", path, index), errors)\n\t\t}\n\tcase \"string\":\n\t\tif _, ok := value.(string); !ok {\n\t\t\t*errors = append(*errors, path + \" is not a string\")\n\t\t}\n\tcase \"integer\":\n        number, ok := value.(float64)\n        if !ok || math.Trunc(number) != number {\n\t\t\t*erros = append(*errors, path+\" is not type integer\")\n\t\t}\n\tcase \"number\":\n\t\tif _, ok := value.(float64); !ok {\n\t\t\t*errors = append(*errors, path+\" is not type number\")\n\t\t}\n\tcase \"boolean\":\n\t\tif _, ok := value.(bool); !ok {\n\t\t\t*errors = append(*errors, path+\" is not boolean\")\n\t\t}\n\tcase \"any\", \"\":\n\t\treturn\n\tdefault:\n\t\treturn\n\t}\n}\n")
	//TODO: Add support for additionalProperties in the validation function this below line of code is not correct.
	builder.WriteString("func validateContractValueWithAdditional(value any, contract ResponseContract, path string, errors *[]string) {\n\tif value == nil {\n\t\tif contract.Nullable || strings.EqualFold(contract.Type, \"any\") {\n\t\t\treturn\n\t\t}\n\t\t*errors = append(*errors, path + \" is null but not nullable\")\n\t\treturn\n\t}\n\texpectedType := strings.ToLower(contract.Type)\n\tswitch expectedType {\n\tcase \"object\":\n\t\tobject, ok := value.(map[string]any)\n\t\tif !ok {\n\t\t\t*errors = append(*errors, path + \" is not an object\")\n\t\t\treturn\n\t\t}\n\n\t\tfor _, requiredField := range contract.Required {\n\t\t\tif _, exists := object[requiredField]; !exists {\n\t\t\t\t*errors = append(*errors, path + \".\" + requiredField + \" is required but missing\")\n\t\t\t}\n\t\t}\n\n")
	return builder.String()
}

func uniqueFunctionName(ctx *generationContext, base string) string {
	if base == "" {
		base = "GeneratedFunction"
	}
	count := ctx.functionNames[base]
	if count == 0 && !ctx.reservedNames[base] {
		ctx.functionNames[base] = 1
		return base
	}
	if count < 1 {
		count = 1
	}
	for {
		count++
		candidate := base + strconv.Itoa(count)
		if ctx.functionNames[candidate] == 0 && !ctx.reservedNames[candidate] {
			ctx.functionNames[base] = count
			ctx.functionNames[candidate] = 1
			return candidate
		}
	}
}

func uniqueParameterName(base string, used map[string]int) string {
	if base == "" {
		base = "value"
	}
	if used[base] == 0 {
		used[base] = 1
		return base
	}
	count := used[base] + 1
	used[base] = count
	return base + strconv.Itoa(count)
}

var invalidIdentifier = regexp.MustCompile(`[^\pL\pN_]`)

func sanitizeExportedIdentifier(value string) string {
	parts := invalidIdentifier.Split(value, -1)
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		builder.WriteString(string(runes))
	}
	result := builder.String()
	if result == "" {
		return "GeneratedSchema"
	}
	if unicode.IsDigit([]rune(result)[0]) {
		result = "Value" + result
	}
	return result
}

func sanitizeParameterIdentifier(value string) string {
	parts := invalidIdentifier.Split(value, -1)
	var builder strings.Builder
	for index, part := range parts {
		if part == "" {
			continue
		}
		if index == 0 {
			builder.WriteString(part)
		} else {
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			builder.WriteString(string(runes))
		}
	}
	result := builder.String()
	if result == "" {
		result = "value"
	}
	if unicode.IsDigit([]rune(result)[0]) {
		result = "value" + result
	}
	if goKeyWords[result] {
		result += "Value"
	}
	return result
}

var goKeyWords = map[string]bool{
	"break": true, "default": true, "func": true, "interface": true, "select": true,
	"case": true, "defer": true, "go": true, "map": true, "struct": true,
	"chan": true, "else": true, "goto": true, "package": true, "switch": true,
	"const": true, "fallthrough": true, "if": true, "range": true, "type": true,
	"continue": true, "for": true, "import": true, "return": true, "var": true,
}

func endpointFunctionName(path string, method string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	var builder strings.Builder
	for _, segment := range segments {
		segment = strings.Trim(segment, "{}")
		builder.WriteString(sanitizeExportedIdentifier(segment))
	}
	builder.WriteString(sanitizeExportedIdentifier(strings.ToLower(method)))
	builder.WriteString("RequestBody")
	return builder.String()
}

func referenceName(reference string) string {
	parts := strings.Split(reference, "/")
	if len(parts) == 0 {
		return reference
	}
	return parts[len(parts)-1]
}

func child(value *orderedValue, name string) *orderedValue {
	if value == nil || value.kind != '{' {
		return nil
	}
	for _, pair := range value.object {
		if pair.name == name {
			return pair.value
		}
	}
	return nil
}

func stringValue(value *orderedValue) string {
	if value == nil {
		return ""
	}
	text, _ := value.value.(string)
	return text
}

func boolValue(value *orderedValue) bool {
	if value == nil {
		return false
	}
	boolean, _ := value.value.(bool)
	return boolean
}

func arrayValues(value *orderedValue) []any {
	if value == nil || value.kind != '[' {
		return nil
	}
	result := make([]any, 0, len(value.array))
	for _, item := range value.array {
		if item.kind == 0 {
			result = append(result, item.value)
		} else {
			result = append(result, toNative(item))
		}
	}
	return result
}

func toNative(value *orderedValue) any {
	if value == nil {
		return nil
	}
	switch value.kind {
	case '{':
		result := make(map[string]any)
		keys := make([]string, 0, len(value.object))
		for _, pair := range value.object {
			keys = append(keys, pair.name)
			result[pair.name] = toNative(pair.value)
		}
		sort.Strings(keys)
		return result
	case '[':
		result := make([]any, 0, len(value.array))
		for _, item := range value.array {
			result = append(result, toNative(item))
		}
		return result
	default:
		return value.value
	}
}

func copyFile(source string, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0644)
}
