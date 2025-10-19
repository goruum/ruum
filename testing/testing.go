// Package testing provides utilities for testing Ruum applications.
package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/goruum/ruum/core"
)

// TestContext creates a test context
type TestContext struct {
	core.Context
	req       *http.Request
	res       *httptest.ResponseRecorder
	container core.Container
}

// NewTestContext creates a new test context
func NewTestContext(method, path string, body io.Reader) *TestContext {
	req := httptest.NewRequest(method, path, body)
	res := httptest.NewRecorder()
	container := core.NewContainer()

	ctx := core.NewContext(context.Background(), req, res, container)

	return &TestContext{
		Context:   ctx,
		req:       req,
		res:       res,
		container: container,
	}
}

// SetHeader sets a request header
func (tc *TestContext) SetHeader(key, value string) {
	tc.req.Header.Set(key, value)
}

// SetParam sets a URL parameter
func (tc *TestContext) SetParam(key, value string) {
	// This is a simplified version
	// In a real implementation, you'd use the router's context
}

// SetQuery sets a query parameter
func (tc *TestContext) SetQuery(key, value string) {
	q := tc.req.URL.Query()
	q.Set(key, value)
	tc.req.URL.RawQuery = q.Encode()
}

// RegisterProvider registers a provider in the test container
func (tc *TestContext) RegisterProvider(name string, provider interface{}, opts ...core.ProviderOption) error {
	return tc.container.Register(name, provider, opts...)
}

// RegisterValue registers a value in the test container
func (tc *TestContext) RegisterValue(name string, value interface{}) error {
	return tc.container.RegisterValue(name, value)
}

// GetResponse returns the response recorder
func (tc *TestContext) GetResponse() *httptest.ResponseRecorder {
	return tc.res
}

// GetStatusCode returns the response status code
func (tc *TestContext) GetStatusCode() int {
	return tc.res.Code
}

// GetBody returns the response body as string
func (tc *TestContext) GetBody() string {
	return tc.res.Body.String()
}

// GetBodyJSON parses the response body as JSON
func (tc *TestContext) GetBodyJSON(v interface{}) error {
	return json.Unmarshal(tc.res.Body.Bytes(), v)
}

// TestRequest builder for creating test requests
type TestRequest struct {
	method  string
	path    string
	headers map[string]string
	query   map[string]string
	body    interface{}
}

// NewTestRequest creates a new test request builder
func NewTestRequest(method, path string) *TestRequest {
	return &TestRequest{
		method:  method,
		path:    path,
		headers: make(map[string]string),
		query:   make(map[string]string),
	}
}

// Header adds a header to the request
func (tr *TestRequest) Header(key, value string) *TestRequest {
	tr.headers[key] = value
	return tr
}

// Query adds a query parameter
func (tr *TestRequest) Query(key, value string) *TestRequest {
	tr.query[key] = value
	return tr
}

// Body sets the request body
func (tr *TestRequest) Body(body interface{}) *TestRequest {
	tr.body = body
	return tr
}

// JSON sets the request body as JSON
func (tr *TestRequest) JSON(body interface{}) *TestRequest {
	tr.headers["Content-Type"] = "application/json"
	tr.body = body
	return tr
}

// Build builds the test context
func (tr *TestRequest) Build() *TestContext {
	var bodyReader io.Reader

	if tr.body != nil {
		switch v := tr.body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		default:
			jsonBody, _ := json.Marshal(v)
			bodyReader = bytes.NewReader(jsonBody)
		}
	}

	ctx := NewTestContext(tr.method, tr.path, bodyReader)

	// Set headers
	for key, value := range tr.headers {
		ctx.SetHeader(key, value)
	}

	// Set query params
	for key, value := range tr.query {
		ctx.SetQuery(key, value)
	}

	return ctx
}

// MockService is a generic mock service
type MockService struct {
	Calls      map[string][]interface{}
	Returns    map[string][]interface{}
	Errors     map[string]error
	callCounts map[string]int
}

// NewMockService creates a new mock service
func NewMockService() *MockService {
	return &MockService{
		Calls:      make(map[string][]interface{}),
		Returns:    make(map[string][]interface{}),
		Errors:     make(map[string]error),
		callCounts: make(map[string]int),
	}
}

// RecordCall records a method call
func (ms *MockService) RecordCall(method string, args ...interface{}) {
	ms.Calls[method] = append(ms.Calls[method], args)
	ms.callCounts[method]++
}

// SetReturn sets the return value for a method
func (ms *MockService) SetReturn(method string, returns ...interface{}) {
	ms.Returns[method] = returns
}

// SetError sets the error for a method
func (ms *MockService) SetError(method string, err error) {
	ms.Errors[method] = err
}

// GetCallCount returns the number of times a method was called
func (ms *MockService) GetCallCount(method string) int {
	return ms.callCounts[method]
}

// WasCalled returns true if a method was called
func (ms *MockService) WasCalled(method string) bool {
	return ms.callCounts[method] > 0
}

// GetLastCall returns the last call arguments for a method
func (ms *MockService) GetLastCall(method string) []interface{} {
	calls := ms.Calls[method]
	if len(calls) == 0 {
		return nil
	}
	return calls[len(calls)-1].([]interface{})
}

// Reset resets all mock data
func (ms *MockService) Reset() {
	ms.Calls = make(map[string][]interface{})
	ms.Returns = make(map[string][]interface{})
	ms.Errors = make(map[string]error)
	ms.callCounts = make(map[string]int)
}

// TestApp wraps an application for testing
type TestApp struct {
	app       core.Application
	container core.Container
}

// NewTestApp creates a new test app
func NewTestApp(app core.Application) *TestApp {
	return &TestApp{
		app:       app,
		container: app.GetContainer(),
	}
}

// Request creates a test request
func (ta *TestApp) Request(method, path string) *TestRequest {
	return NewTestRequest(method, path)
}

// Get creates a GET request
func (ta *TestApp) Get(path string) *TestRequest {
	return NewTestRequest("GET", path)
}

// Post creates a POST request
func (ta *TestApp) Post(path string) *TestRequest {
	return NewTestRequest("POST", path)
}

// Put creates a PUT request
func (ta *TestApp) Put(path string) *TestRequest {
	return NewTestRequest("PUT", path)
}

// Delete creates a DELETE request
func (ta *TestApp) Delete(path string) *TestRequest {
	return NewTestRequest("DELETE", path)
}

// Patch creates a PATCH request
func (ta *TestApp) Patch(path string) *TestRequest {
	return NewTestRequest("PATCH", path)
}

// MockContainer is a mock DI container for testing
type MockContainer struct {
	providers map[string]interface{}
}

// NewMockContainer creates a new mock container
func NewMockContainer() *MockContainer {
	return &MockContainer{
		providers: make(map[string]interface{}),
	}
}

// Register registers a mock provider
func (mc *MockContainer) Register(name string, provider interface{}, _ ...core.ProviderOption) error {
	mc.providers[name] = provider
	return nil
}

// RegisterFactory registers a mock factory
func (mc *MockContainer) RegisterFactory(name string, factory interface{}, _ ...core.ProviderOption) error {
	mc.providers[name] = factory
	return nil
}

// RegisterValue registers a mock value
func (mc *MockContainer) RegisterValue(name string, value interface{}) error {
	mc.providers[name] = value
	return nil
}

// Resolve resolves a mock provider
func (mc *MockContainer) Resolve(name string) (interface{}, error) {
	if provider, exists := mc.providers[name]; exists {
		return provider, nil
	}
	return nil, core.NotFoundException("provider not found: " + name)
}

// ResolveByType resolves by type (not implemented for mock)
func (mc *MockContainer) ResolveByType(_ interface{}) (interface{}, error) {
	return nil, core.NotFoundException("not implemented in mock")
}

// Has checks if a provider exists
func (mc *MockContainer) Has(name string) bool {
	_, exists := mc.providers[name]
	return exists
}

// GetAll returns all providers
func (mc *MockContainer) GetAll() map[string]interface{} {
	return mc.providers
}

// AssertionHelper provides assertion helpers
type AssertionHelper struct {
	t interface{} // testing.T or similar
}

// NewAssertionHelper creates a new assertion helper
func NewAssertionHelper(t interface{}) *AssertionHelper {
	return &AssertionHelper{t: t}
}

// ExpectStatus asserts the status code
func (ah *AssertionHelper) ExpectStatus(ctx *TestContext, expected int) bool {
	actual := ctx.GetStatusCode()
	return actual == expected
}

// ExpectHeader asserts a response header
func (ah *AssertionHelper) ExpectHeader(ctx *TestContext, key, expected string) bool {
	actual := ctx.GetResponse().Header().Get(key)
	return actual == expected
}

// ExpectBody asserts the response body
func (ah *AssertionHelper) ExpectBody(ctx *TestContext, expected string) bool {
	actual := ctx.GetBody()
	return actual == expected
}

// ExpectBodyContains asserts the response body contains a string
func (ah *AssertionHelper) ExpectBodyContains(ctx *TestContext, substr string) bool {
	actual := ctx.GetBody()
	return bytes.Contains([]byte(actual), []byte(substr))
}

// ExpectJSON asserts the response is valid JSON
func (ah *AssertionHelper) ExpectJSON(ctx *TestContext, v interface{}) bool {
	return ctx.GetBodyJSON(v) == nil
}
