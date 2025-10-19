package core

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Context provides access to request/response and dependency injection container
type Context interface {
	context.Context

	// HTTP related
	Request() *http.Request
	Response() http.ResponseWriter

	// Request data
	Param(key string) string
	Query(key string) string
	QueryDefault(key, defaultValue string) string
	QueryInt(key string) (int, error)
	QueryIntDefault(key string, defaultValue int) int
	QueryBool(key string) (bool, error)
	QueryBoolDefault(key string, defaultValue bool) bool
	Body(v interface{}) error
	BindJSON(v interface{}) error
	BindXML(v interface{}) error
	BindQuery(v interface{}) error
	Header(key string) string
	Headers() http.Header

	// Cookies
	Cookie(name string) (*http.Cookie, error)
	SetCookie(cookie *http.Cookie)
	Cookies() []*http.Cookie

	// File uploads
	FormFile(name string) (*multipart.FileHeader, error)
	MultipartForm() (*multipart.Form, error)
	FormValue(name string) string
	PostForm(key string) string

	// Response methods
	JSON(code int, v interface{}) error
	XML(code int, v interface{}) error
	String(code int, s string) error
	HTML(code int, html string) error
	Redirect(code int, location string) error
	Stream(code int, contentType string, reader io.Reader) error
	File(filepath string) error
	NoContent(code int) error
	Status(code int)
	SetHeader(key, value string)
	AddHeader(key, value string)

	// Request info
	Method() string
	Path() string
	Host() string
	Protocol() string
	RemoteAddr() string
	ContentType() string
	IsWebSocket() bool
	IsAjax() bool
	IsTLS() bool

	// Context data
	Get(key string) interface{}
	Set(key string, value interface{})
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	MustGet(key string) interface{}

	// DI Container access
	Container() Container
}

// DefaultContext implements Context interface
type DefaultContext struct {
	context.Context
	req       *http.Request
	res       http.ResponseWriter
	container Container
	data      map[string]interface{}
	params    map[string]string
}

// NewContext creates a new Context
func NewContext(ctx context.Context, req *http.Request, res http.ResponseWriter, container Container) Context {
	return &DefaultContext{
		Context:   ctx,
		req:       req,
		res:       res,
		container: container,
		data:      make(map[string]interface{}),
		params:    make(map[string]string),
	}
}

// Request returns the HTTP request.
func (c *DefaultContext) Request() *http.Request {
	return c.req
}

// Response returns the HTTP response writer.
func (c *DefaultContext) Response() http.ResponseWriter {
	return c.res
}

// Param returns the URL parameter value.
func (c *DefaultContext) Param(key string) string {
	return c.params[key]
}

// setParam sets a URL parameter value (internal use only).
func (c *DefaultContext) setParam(key, value string) {
	c.params[key] = value
}

// Query returns the query parameter value.
func (c *DefaultContext) Query(key string) string {
	return c.req.URL.Query().Get(key)
}

// QueryDefault returns the query parameter value or default if not found.
func (c *DefaultContext) QueryDefault(key, defaultValue string) string {
	if value := c.Query(key); value != "" {
		return value
	}
	return defaultValue
}

// QueryInt returns the query parameter as int.
func (c *DefaultContext) QueryInt(key string) (int, error) {
	value := c.Query(key)
	if value == "" {
		return 0, fmt.Errorf("query parameter '%s' not found", key)
	}
	return strconv.Atoi(value)
}

// QueryIntDefault returns the query parameter as int or default.
func (c *DefaultContext) QueryIntDefault(key string, defaultValue int) int {
	value, err := c.QueryInt(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// QueryBool returns the query parameter as bool.
func (c *DefaultContext) QueryBool(key string) (bool, error) {
	value := c.Query(key)
	if value == "" {
		return false, fmt.Errorf("query parameter '%s' not found", key)
	}
	return strconv.ParseBool(value)
}

// QueryBoolDefault returns the query parameter as bool or default.
func (c *DefaultContext) QueryBoolDefault(key string, defaultValue bool) bool {
	value, err := c.QueryBool(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// Body parses the request body into the provided interface.
func (c *DefaultContext) Body(v interface{}) error {
	return json.NewDecoder(c.req.Body).Decode(v)
}

// BindJSON binds the request body as JSON.
func (c *DefaultContext) BindJSON(v interface{}) error {
	if c.req.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	return json.NewDecoder(c.req.Body).Decode(v)
}

// BindXML binds the request body as XML.
func (c *DefaultContext) BindXML(v interface{}) error {
	if c.req.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	return xml.NewDecoder(c.req.Body).Decode(v)
}

// BindQuery binds query parameters to struct.
func (c *DefaultContext) BindQuery(v interface{}) error {
	values := c.req.URL.Query()
	return bindData(v, values, "query")
}

// Header returns the request header value.
func (c *DefaultContext) Header(key string) string {
	return c.req.Header.Get(key)
}

// Headers returns all request headers.
func (c *DefaultContext) Headers() http.Header {
	return c.req.Header
}

// Cookie returns a cookie by name.
func (c *DefaultContext) Cookie(name string) (*http.Cookie, error) {
	return c.req.Cookie(name)
}

// SetCookie sets a cookie.
func (c *DefaultContext) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.res, cookie)
}

// Cookies returns all cookies.
func (c *DefaultContext) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

// FormFile returns the first file for the provided form key.
func (c *DefaultContext) FormFile(name string) (*multipart.FileHeader, error) {
	_, fileHeader, err := c.req.FormFile(name)
	return fileHeader, err
}

// MultipartForm returns the multipart form.
func (c *DefaultContext) MultipartForm() (*multipart.Form, error) {
	err := c.req.ParseMultipartForm(32 << 20) // 32 MB
	if err != nil {
		return nil, err
	}
	return c.req.MultipartForm, nil
}

// FormValue returns the form value.
func (c *DefaultContext) FormValue(name string) string {
	return c.req.FormValue(name)
}

// PostForm returns the form value from POST body.
func (c *DefaultContext) PostForm(key string) string {
	return c.req.PostFormValue(key)
}

// JSON sends a JSON response with the specified status code.
func (c *DefaultContext) JSON(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/json")
	c.res.WriteHeader(code)
	return json.NewEncoder(c.res).Encode(v)
}

// XML sends an XML response.
func (c *DefaultContext) XML(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/xml")
	c.res.WriteHeader(code)
	return xml.NewEncoder(c.res).Encode(v)
}

// String sends a plain text response.
func (c *DefaultContext) String(code int, s string) error {
	c.res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.res.WriteHeader(code)
	_, err := c.res.Write([]byte(s))
	return err
}

// HTML sends an HTML response.
func (c *DefaultContext) HTML(code int, html string) error {
	c.res.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.res.WriteHeader(code)
	_, err := c.res.Write([]byte(html))
	return err
}

// Redirect redirects to the specified URL.
func (c *DefaultContext) Redirect(code int, location string) error {
	if code < 300 || code > 308 {
		return fmt.Errorf("invalid redirect status code: %d", code)
	}
	http.Redirect(c.res, c.req, location, code)
	return nil
}

// Stream sends a streaming response.
func (c *DefaultContext) Stream(code int, contentType string, reader io.Reader) error {
	c.res.Header().Set("Content-Type", contentType)
	c.res.WriteHeader(code)
	_, err := io.Copy(c.res, reader)
	return err
}

// File serves a file.
func (c *DefaultContext) File(filepath string) error {
	http.ServeFile(c.res, c.req, filepath)
	return nil
}

// NoContent sends a response with no content.
func (c *DefaultContext) NoContent(code int) error {
	c.res.WriteHeader(code)
	return nil
}

// Status sets the HTTP status code.
func (c *DefaultContext) Status(code int) {
	c.res.WriteHeader(code)
}

// SetHeader sets a response header.
func (c *DefaultContext) SetHeader(key, value string) {
	c.res.Header().Set(key, value)
}

// AddHeader adds a response header.
func (c *DefaultContext) AddHeader(key, value string) {
	c.res.Header().Add(key, value)
}

// Method returns the request method.
func (c *DefaultContext) Method() string {
	return c.req.Method
}

// Path returns the request path.
func (c *DefaultContext) Path() string {
	return c.req.URL.Path
}

// Host returns the request host.
func (c *DefaultContext) Host() string {
	return c.req.Host
}

// Protocol returns the request protocol.
func (c *DefaultContext) Protocol() string {
	return c.req.Proto
}

// RemoteAddr returns the remote address.
func (c *DefaultContext) RemoteAddr() string {
	return c.req.RemoteAddr
}

// ContentType returns the content type.
func (c *DefaultContext) ContentType() string {
	return c.req.Header.Get("Content-Type")
}

// IsWebSocket checks if the request is a WebSocket upgrade.
func (c *DefaultContext) IsWebSocket() bool {
	upgrade := c.req.Header.Get("Upgrade")
	return strings.EqualFold(upgrade, "websocket")
}

// IsAjax checks if the request is an AJAX request.
func (c *DefaultContext) IsAjax() bool {
	return c.req.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// IsTLS checks if the connection is TLS.
func (c *DefaultContext) IsTLS() bool {
	return c.req.TLS != nil
}

// Get returns a value from the context.
func (c *DefaultContext) Get(key string) interface{} {
	return c.data[key]
}

// Set stores a value in the context.
func (c *DefaultContext) Set(key string, value interface{}) {
	c.data[key] = value
}

// GetString returns a string value from the context.
func (c *DefaultContext) GetString(key string) string {
	if value, ok := c.data[key].(string); ok {
		return value
	}
	return ""
}

// GetBool returns a bool value from the context.
func (c *DefaultContext) GetBool(key string) bool {
	if value, ok := c.data[key].(bool); ok {
		return value
	}
	return false
}

// GetInt returns an int value from the context.
func (c *DefaultContext) GetInt(key string) int {
	if value, ok := c.data[key].(int); ok {
		return value
	}
	return 0
}

// MustGet returns a value or panics if not found.
func (c *DefaultContext) MustGet(key string) interface{} {
	if value, exists := c.data[key]; exists {
		return value
	}
	panic(fmt.Sprintf("key '%s' does not exist", key))
}

// Container returns the dependency injection container.
func (c *DefaultContext) Container() Container {
	return c.container
}

// bindData binds data to struct using reflection
func bindData(ptr interface{}, data url.Values, tag string) error {
	// This is a simplified implementation
	// A full implementation would use reflection to bind all fields
	return nil
}
