package utils

import (
	"github.com/zeromicro/go-zero/core/mapping"
	"github.com/zeromicro/go-zero/rest/pathvar"
	"io"
	"net/http"
	"net/textproto"
	"strings"
)

const (
	formKey           = "form"
	pathKey           = "path"
	queryKey          = "query"
	headerKey         = "header"
	maxMemory         = 32 << 20 // 32MB
	maxBodyLen        = 8 << 20  // 8MB
	separator         = ";"
	tokensInAttribute = 2

	// ApplicationJson stands for application/json.
	ApplicationJson = "application/json"
	// ContentType is the header key for Content-Type.
	ContentType = "Content-Type"
	// JsonContentType is the content type for JSON.
	JsonContentType = "application/json; charset=utf-8"
)

var (
	formUnmarshaler   = mapping.NewUnmarshaler(formKey, mapping.WithStringValues())
	pathUnmarshaler   = mapping.NewUnmarshaler(pathKey, mapping.WithStringValues())
	queryUnmarshaler  = mapping.NewUnmarshaler(queryKey, mapping.WithStringValues())
	headerUnmarshaler = mapping.NewUnmarshaler(headerKey, mapping.WithStringValues(),
		mapping.WithCanonicalKeyFunc(textproto.CanonicalMIMEHeaderKey))
)

// Parse parses the request.
func Parse(r *http.Request, v interface{}) error {
	if err := ParsePath(r, v); err != nil {
		return err
	}

	if err := ParseForm(r, v); err != nil {
		return err
	}

	if err := ParseQuery(r, v); err != nil {
		return err
	}

	if err := ParseHeaders(r, v); err != nil {
		return err
	}

	return ParseJsonBody(r, v)
}

// ParsePath parses the symbols reside in url path.
// Like http://localhost/bag/:name
func ParsePath(r *http.Request, v interface{}) error {
	vars := pathvar.Vars(r)
	m := make(map[string]interface{}, len(vars))
	for k, v := range vars {
		m[k] = v
	}

	return pathUnmarshaler.Unmarshal(m, v)
}

// ParseForm parses the form request.
func ParseForm(r *http.Request, v interface{}) error {
	params, err := GetFormValues(r)
	if err != nil {
		return err
	}

	return formUnmarshaler.Unmarshal(params, v)
}

// ParseQuery parses the query request.
// Like http://localhost/bag?name=123
func ParseQuery(r *http.Request, v interface{}) error {
	query := r.URL.Query()
	m := make(map[string]interface{}, len(query))
	for name := range query {
		value := query.Get(name)
		m[name] = value
	}

	return queryUnmarshaler.Unmarshal(m, v)
}

// ParseHeaders parses the headers request.
func ParseHeaders(r *http.Request, v interface{}) error {
	header := r.Header
	m := map[string]interface{}{}
	for k, v := range header {
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			m[k] = v
		}
	}

	return headerUnmarshaler.Unmarshal(m, v)
}

// ParseJsonBody parses the post request which contains json in body.
func ParseJsonBody(r *http.Request, v interface{}) error {
	if withJsonBody(r) {
		reader := io.LimitReader(r.Body, maxBodyLen)
		return mapping.UnmarshalJsonReader(reader, v)
	}

	return mapping.UnmarshalJsonMap(nil, v)
}

func withJsonBody(r *http.Request) bool {
	return r.ContentLength > 0 && strings.Contains(r.Header.Get(ContentType), ApplicationJson)
}

// GetFormValues returns the form values.
func GetFormValues(r *http.Request) (map[string]interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if err != http.ErrNotMultipart {
			return nil, err
		}
	}

	params := make(map[string]interface{}, len(r.Form))
	for name := range r.Form {
		formValue := r.Form.Get(name)
		if len(formValue) > 0 {
			params[name] = formValue
		}
	}

	return params, nil
}
