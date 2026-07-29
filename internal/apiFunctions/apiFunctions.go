package apiFunctions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

/* Example Usage
status, body, headers, err := DoRequest(RequestOptions{
	Method: http.MethodPost,
	URL: "https://api.example.com/endpoint",
	AuthMode: AuthBearer,
	Token: "your-jwt",
	Headers: map[string]string{"X-Corelation-ID": "12345"},
	BodyKind: BodyJSON,
	JSONBody: map[string]any{"key": "value"},
	Timeout: 10 * time.Second,
})
*/

func DoRequest(opts RequestOptions) (int, []byte, map[string][]string, error) {
	if opts.Method == "" {
		opts.Method = http.MethodGet
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	var body io.Reader
	var contentType string
	switch opts.BodyKind {
	case BodyNone:
	case BodyJSON:
		var b []byte
		switch v := opts.JSONBody.(type) {
		case nil:
			b = nil
		case []byte:
			b = v
		case string:
			b = []byte(v)
		default:
			x, err := json.Marshal(v)
			if err != nil {
				return 0, nil, nil, err
			}
			b = x
		}
		body = bytes.NewReader(b)
		contentType = "application/json"
	case BodyRaw:
		body = bytes.NewReader(opts.RawBody)
	case BodyFormURLEncoded:
		keys := make([]string, 0, len(opts.FormBody))
		for k := range opts.FormBody {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		vals := url.Values{}
		for _, k := range keys {
			vals.Set(k, opts.FormBody[k])
		}
		body = strings.NewReader(vals.Encode())
		contentType = "application/x-www-form-urlencoded"
	case BodyMultipartFile:
		buf := &bytes.Buffer{}
		w := multipart.NewWriter(buf)
		field := opts.FileField
		if field == "" {
			field = "file"
		}
		f, err := os.Open(opts.FilePath)
		if err != nil {
			return 0, nil, nil, err
		}
		defer f.Close()
		part, err := createMultipartFilePart(w, field, filepath.Base(opts.FilePath), opts.FileContentType)
		if err != nil {
			return 0, nil, nil, err
		}
		_, err = io.Copy(part, f)
		if err != nil {
			return 0, nil, nil, err
		}
		err = w.Close()
		if err != nil {
			return 0, nil, nil, err
		}
		body = buf
		contentType = w.FormDataContentType()
	case BodyMultipartMixed:
		buf := &bytes.Buffer{}
		w := multipart.NewWriter(buf)
		for k, v := range opts.FormFields {
			if err := w.WriteField(k, v); err != nil {
				return 0, nil, nil, err
			}
		}
		field := opts.FileField
		if field == "" {
			field = "file"
		}
		f, err := os.Open(opts.FilePath)
		if err != nil {
			return 0, nil, nil, err
		}
		defer f.Close()
		part, err := createMultipartFilePart(w, field, filepath.Base(opts.FilePath), opts.FileContentType)
		if err != nil {
			return 0, nil, nil, err
		}
		_, err = io.Copy(part, f)
		if err != nil {
			return 0, nil, nil, err
		}
		err = w.Close()
		if err != nil {
			return 0, nil, nil, err
		}
		body = buf
		contentType = w.FormDataContentType()
	}
	finalURL := opts.URL
	if len(opts.QueryParams) > 0 {
		var err error
		finalURL, err = BuildWithQueryParams(opts.URL, opts.QueryParams)
		if err != nil {
			return 0, nil, nil, err
		}
	}
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, opts.Method, finalURL, body)
	if err != nil {
		return 0, nil, nil, err
	}
	switch opts.AuthMode {
	case AuthBearer:
		if opts.Token != "" {
			req.Header.Set("Authorization", "Bearer "+opts.Token)
		}
	case AuthBasic:
		req.SetBasicAuth(opts.Username, opts.Password)
	}
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}
	if contentType != "" {
		if !hasHeader(opts.Headers, "Content-Type") {
			req.Header.Set("Content-Type", contentType)
		}
	}
	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: opts.Timeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, err
	}
	return resp.StatusCode, b, resp.Header, nil
}

type AuthMode int

const (
	AuthNone AuthMode = iota
	AuthBearer
	AuthBasic
)

type BodyKind int

const (
	//BodyKind are ints:
	// 0: BodyNone
	// 1: JSON body (object, string, or []byte)
	// 2: raw []byte body
	// 3: multipart form with single file
	// 4: multipart form with file and fields
	// 5: application/x-www-form-urlencoded form
	BodyNone BodyKind = iota
	BodyJSON
	BodyRaw
	BodyMultipartFile
	BodyMultipartMixed
	BodyFormURLEncoded
)

type RequestOptions struct {
	Method          string
	URL             string
	QueryParams     map[string]string
	AuthMode        AuthMode
	Token           string
	Username        string
	Password        string
	Headers         map[string]string
	BodyKind        BodyKind
	JSONBody        any
	Context         context.Context
	RawBody         []byte
	FilePath        string
	FileField       string
	FileContentType string
	FormFields      map[string]string
	FormBody        map[string]string
	//Optional: provide a preconfigured client (for example with a cookie jar)
	Client  *http.Client
	Timeout time.Duration
}

func BuildQuryParams(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals := url.Values{}
	for _, k := range keys {
		vals.Set(k, params[k])
	}
	return "?" + vals.Encode()
}

func hasHeader(headers map[string]string, headerName string) bool {
	for k := range headers {
		if strings.EqualFold(k, headerName) {
			return true
		}
	}
	return false
}

// BuildURLWithQueryParams lets callers build a final URL externally, while
// DoRequest can also use the same behavior through RequestOptions.QueryParams.
func BuildWithQueryParams(baseURL string, params map[string]string) (string, error) {
	if len(params) == 0 {
		return baseURL, nil
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	q := u.Query()
	for _, k := range keys {
		q.Set(k, params[k])
	}
	u.RawQuery = q.Encode()
	log.Println("Built URL with query params:", u.String())
	return u.String(), nil
}

func createMultipartFilePart(w *multipart.Writer, fieldName, fileName, contentType string) (io.Writer, error) {
	if contentType == "" {
		return w.CreateFormFile(fieldName, fileName)
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName))
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}
