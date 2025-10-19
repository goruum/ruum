package core

import (
	"context"
	"encoding/json"
	"net/http"
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
	Body(v interface{}) error
	Header(key string) string

	// Response methods
	JSON(code int, v interface{}) error
	String(code int, s string) error
	Status(code int)
	SetHeader(key, value string)

	// Context data
	Get(key string) interface{}
	Set(key string, value interface{})

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

// Body parses the request body into the provided interface.
func (c *DefaultContext) Body(v interface{}) error {
	return json.NewDecoder(c.req.Body).Decode(v)
}

// Header returns the request header value.
func (c *DefaultContext) Header(key string) string {
	return c.req.Header.Get(key)
}

// JSON sends a JSON response with the specified status code.
func (c *DefaultContext) JSON(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/json")
	c.res.WriteHeader(code)
	return json.NewEncoder(c.res).Encode(v)
}

func (c *DefaultContext) String(code int, s string) error {
	c.res.Header().Set("Content-Type", "text/plain")
	c.res.WriteHeader(code)
	_, err := c.res.Write([]byte(s))
	return err
}

// Status sets the HTTP status code.
func (c *DefaultContext) Status(code int) {
	c.res.WriteHeader(code)
}

// SetHeader sets a response header.
func (c *DefaultContext) SetHeader(key, value string) {
	c.res.Header().Set(key, value)
}

// Get returns a value from the context.
func (c *DefaultContext) Get(key string) interface{} {
	return c.data[key]
}

// Set stores a value in the context.
func (c *DefaultContext) Set(key string, value interface{}) {
	c.data[key] = value
}

// Container returns the dependency injection container.
func (c *DefaultContext) Container() Container {
	return c.container
}
