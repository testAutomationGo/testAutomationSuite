package openApiInternal

type OpenAPISpec struct {
	OpenApi    string                          `json:"openapi"`
	Info       OpenAPIInfo                     `json:"info"`
	Paths      map[string]map[string]Operation `json:"paths"`
	Security   []map[string][]string           `json:"security,omitempty"`
	Components OpenAPIComponents               `json:"components"`
}

type OpenAPIInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

type PathItem struct {
	Get        *Operation  `json:"get,omitempty"`
	Post       *Operation  `json:"post,omitempty"`
	Put        *Operation  `json:"put,omitempty"`
	Delete     *Operation  `json:"delete,omitempty"`
	Patch      *Operation  `json:"patch,omitempty"`
	Options    *Operation  `json:"options,omitempty"`
	Head       *Operation  `json:"head,omitempty"`
	Trace      *Operation  `json:"trace,omitempty"`
	Parameters []Parameter `json:"parameters,omitempty"`
}

type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required,omitempty"`
	Schema   Schema `json:"schema,omitempty"`
}

type RequestBody struct {
	Content map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Description          string            `json:"description,omitempty"`
	Type                 string            `json:"type,omitempty"`
	Reference            string            `json:"$ref,omitempty"`
	Enum                 []any             `json:"enum,omitempty"`
	Discriminator        *Discriminator    `json:"discriminator,omitempty"`
	OneOf                []Schema          `json:"oneOf,omitempty"`
	AnyOf                []Schema          `json:"anyOf,omitempty"`
	AllOf                []Schema          `json:"allOf,omitempty"`
	Properties           map[string]Schema `json:"properties,omitempty"`
	Format               string            `json:"format,omitempty"`
	Items                *Schema           `json:"items,omitempty"`
	Nullable             bool              `json:"nullable,omitempty"`
	AdditionalProperties any               `json:"additionalProperties,omitempty"`
	Required             []string          `json:"required,omitempty"`
	Style                string            `json:"style,omitempty"`
}

type Discriminator struct {
	PropertyName string            `json:"propertyName,omitempty"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type OpenAPIComponents struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type,omitempty"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}
