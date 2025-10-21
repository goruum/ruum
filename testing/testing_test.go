package testing

import (
	"reflect"
	"strings"
	"testing"

	"github.com/goruum/ruum/core"
)

func TestNewTestContext(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)

	if ctx == nil {
		t.Fatal("NewTestContext() returned nil")
	}
}

func TestTestContext_SetHeader(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)
	ctx.SetHeader("X-Test", "value")

	if ctx.req.Header.Get("X-Test") != "value" {
		t.Error("Header not set correctly")
	}
}

func TestTestContext_SetQuery(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)
	ctx.SetQuery("page", "1")

	if ctx.req.URL.Query().Get("page") != "1" {
		t.Error("Query not set correctly")
	}
}

func TestTestContext_SetParam(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)
	ctx.SetParam("id", "123")
	// SetParam is a placeholder for now, so just ensure it doesn't panic
}

func TestTestContext_GetStatusCode(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)
	ctx.res.WriteHeader(200)

	if ctx.GetStatusCode() != 200 {
		t.Errorf("GetStatusCode() = %d, want 200", ctx.GetStatusCode())
	}
}

func TestTestContext_GetBody(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)
	_, _ = ctx.res.WriteString("test body")

	if ctx.GetBody() != "test body" {
		t.Errorf("GetBody() = %v, want 'test body'", ctx.GetBody())
	}
}

func TestTestContext_GetBodyJSON(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)
	_, _ = ctx.res.WriteString(`{"key":"value"}`)

	var result map[string]string
	err := ctx.GetBodyJSON(&result)

	if err != nil {
		t.Errorf("GetBodyJSON() error = %v", err)
	}

	if result["key"] != "value" {
		t.Error("JSON not parsed correctly")
	}
}

func TestNewTestRequest(t *testing.T) {
	req := NewTestRequest("POST", "/test")

	if req == nil {
		t.Fatal("NewTestRequest() returned nil")
	}

	if req.method != "POST" {
		t.Errorf("Method = %v, want POST", req.method)
	}

	if req.path != "/test" {
		t.Errorf("Path = %v, want /test", req.path)
	}
}

func TestTestRequest_Header(t *testing.T) {
	req := NewTestRequest("GET", "/test").Header("X-Test", "value")

	if req.headers["X-Test"] != "value" {
		t.Error("Header not set")
	}
}

func TestTestRequest_Query(t *testing.T) {
	req := NewTestRequest("GET", "/test").Query("page", "1")

	if req.query["page"] != "1" {
		t.Error("Query not set")
	}
}

func TestTestRequest_Body(t *testing.T) {
	req := NewTestRequest("POST", "/test").Body("test body")

	if req.body != "test body" {
		t.Error("Body not set")
	}
}

func TestTestRequest_JSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	req := NewTestRequest("POST", "/test").JSON(data)

	if req.body == nil {
		t.Error("JSON body not set")
	}

	if req.headers["Content-Type"] != "application/json" {
		t.Error("Content-Type header not set")
	}
}

func TestTestRequest_Build(t *testing.T) {
	ctx := NewTestRequest("POST", "/test").
		Header("X-Test", "value").
		Query("page", "1").
		Body("test body").
		Build()

	if ctx == nil {
		t.Fatal("Build() returned nil")
	}
}

func TestTestRequest_Build_WithByteSlice(t *testing.T) {
	ctx := NewTestRequest("POST", "/test").
		Body([]byte("test body")).
		Build()

	if ctx == nil {
		t.Fatal("Build() with byte slice returned nil")
	}
}

func TestTestRequest_Build_WithJSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	ctx := NewTestRequest("POST", "/test").
		JSON(data).
		Build()

	if ctx == nil {
		t.Fatal("Build() with JSON returned nil")
	}

	if ctx.req.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header not set for JSON")
	}
}

func TestNewMockService(t *testing.T) {
	mock := NewMockService()

	if mock == nil {
		t.Fatal("NewMockService() returned nil")
	}
}

func TestMockService_RecordCall(t *testing.T) {
	mock := NewMockService()
	mock.RecordCall("GetUser", 1)

	if len(mock.Calls["GetUser"]) != 1 {
		t.Error("RecordCall() did not record call")
	}
}

func TestMockService_SetReturn(t *testing.T) {
	mock := NewMockService()
	mock.SetReturn("GetUser", "user1", nil)

	if len(mock.Returns["GetUser"]) != 2 {
		t.Error("SetReturn() did not set return values")
	}
}

func TestMockService_SetError(t *testing.T) {
	mock := NewMockService()
	testErr := &CustomError{msg: "test error"}
	mock.SetError("GetUser", testErr)

	if mock.Errors["GetUser"] == nil {
		t.Error("SetError() did not set error")
	}
}

type CustomError struct {
	msg string
}

func (e *CustomError) Error() string {
	return e.msg
}

func TestMockService_GetCallCount(t *testing.T) {
	mock := NewMockService()
	mock.RecordCall("Test")
	mock.RecordCall("Test")

	if mock.GetCallCount("Test") != 2 {
		t.Errorf("GetCallCount() = %d, want 2", mock.GetCallCount("Test"))
	}
}

func TestMockService_WasCalled(t *testing.T) {
	mock := NewMockService()
	mock.RecordCall("Test")

	if !mock.WasCalled("Test") {
		t.Error("WasCalled() should return true")
	}

	if mock.WasCalled("NotCalled") {
		t.Error("WasCalled() should return false for not called method")
	}
}

func TestMockService_GetLastCall(t *testing.T) {
	mock := NewMockService()
	mock.RecordCall("Test", 1, 2)

	lastCall := mock.GetLastCall("Test")
	if lastCall == nil {
		t.Error("GetLastCall() should return last call args")
	}
}

func TestMockService_GetLastCall_Empty(t *testing.T) {
	mock := NewMockService()

	lastCall := mock.GetLastCall("NonExistent")
	if lastCall != nil {
		t.Error("GetLastCall() should return nil for non-existent method")
	}
}

func TestMockService_Reset(t *testing.T) {
	mock := NewMockService()
	mock.RecordCall("Test")
	mock.SetReturn("Test", "value")

	mock.Reset()

	if len(mock.Calls) != 0 || len(mock.Returns) != 0 {
		t.Error("Reset() should clear all data")
	}
}

func TestNewAssertionHelper(t *testing.T) {
	helper := NewAssertionHelper(t)

	if helper == nil {
		t.Fatal("NewAssertionHelper() returned nil")
	}
}

func TestAssertionHelper_ExpectStatus(t *testing.T) {
	helper := NewAssertionHelper(t)
	ctx := NewTestContext("GET", "/test", nil)
	ctx.res.WriteHeader(200)

	if !helper.ExpectStatus(ctx, 200) {
		t.Error("ExpectStatus() should return true for matching status")
	}

	if helper.ExpectStatus(ctx, 404) {
		t.Error("ExpectStatus() should return false for different status")
	}
}

func TestAssertionHelper_ExpectHeader(t *testing.T) {
	helper := NewAssertionHelper(t)
	ctx := NewTestContext("GET", "/test", nil)
	ctx.res.Header().Set("Content-Type", "application/json")

	if !helper.ExpectHeader(ctx, "Content-Type", "application/json") {
		t.Error("ExpectHeader() should return true for matching header")
	}

	if helper.ExpectHeader(ctx, "Content-Type", "text/plain") {
		t.Error("ExpectHeader() should return false for different header")
	}
}

func TestAssertionHelper_ExpectBody(t *testing.T) {
	helper := NewAssertionHelper(t)
	ctx := NewTestContext("GET", "/test", nil)
	_, _ = ctx.res.WriteString("test body")

	if !helper.ExpectBody(ctx, "test body") {
		t.Error("ExpectBody() should return true for matching body")
	}

	if helper.ExpectBody(ctx, "other") {
		t.Error("ExpectBody() should return false for different body")
	}
}

func TestAssertionHelper_ExpectBodyContains(t *testing.T) {
	helper := NewAssertionHelper(t)
	ctx := NewTestContext("GET", "/test", nil)
	_, _ = ctx.res.WriteString("this is a test body")

	if !helper.ExpectBodyContains(ctx, "test") {
		t.Error("ExpectBodyContains() should return true when substring exists")
	}

	if helper.ExpectBodyContains(ctx, "notfound") {
		t.Error("ExpectBodyContains() should return false when substring doesn't exist")
	}
}

func TestAssertionHelper_ExpectJSON(t *testing.T) {
	helper := NewAssertionHelper(t)
	ctx := NewTestContext("GET", "/test", nil)
	_, _ = ctx.res.WriteString(`{"key":"value"}`)

	var result map[string]string
	if !helper.ExpectJSON(ctx, &result) {
		t.Error("ExpectJSON() should return true for valid JSON")
	}

	if result["key"] != "value" {
		t.Error("ExpectJSON() should parse JSON correctly")
	}
}

func TestTestContext_WithBody(t *testing.T) {
	body := strings.NewReader("test body")
	ctx := NewTestContext("POST", "/test", body)

	if ctx == nil {
		t.Error("NewTestContext() with body should not return nil")
	}
}

func TestNewMockContainer(t *testing.T) {
	container := NewMockContainer()

	if container == nil {
		t.Fatal("NewMockContainer() returned nil")
	}
}

func TestMockContainer_Register(t *testing.T) {
	container := NewMockContainer()
	err := container.Register("test", "value")

	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	if !container.Has("test") {
		t.Error("Register() should add provider")
	}
}

func TestMockContainer_RegisterValue(t *testing.T) {
	container := NewMockContainer()
	err := container.RegisterValue("test", "value")

	if err != nil {
		t.Errorf("RegisterValue() error = %v", err)
	}

	val, _ := container.Resolve("test")
	if val != "value" {
		t.Error("RegisterValue() should set value correctly")
	}
}

func TestMockContainer_Resolve(t *testing.T) {
	container := NewMockContainer()
	_ = container.RegisterValue("test", "value")

	val, err := container.Resolve("test")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	if val != "value" {
		t.Errorf("Resolve() = %v, want value", val)
	}
}

func TestMockContainer_Resolve_NotFound(t *testing.T) {
	container := NewMockContainer()

	_, err := container.Resolve("notfound")
	if err == nil {
		t.Error("Resolve() should return error for not found provider")
	}
}

func TestTestContext_RegisterProvider(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)
	err := ctx.RegisterProvider("test", func() string { return "value" })

	if err != nil {
		t.Errorf("RegisterProvider() error = %v", err)
	}
}

func TestTestContext_GetResponse(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)
	res := ctx.GetResponse()

	if res == nil {
		t.Error("GetResponse() should not return nil")
	}
}

func TestTestContext_RegisterValue(t *testing.T) {
	ctx := NewTestContext("GET", "/test", nil)
	err := ctx.RegisterValue("testKey", "testValue")

	if err != nil {
		t.Errorf("RegisterValue() error = %v", err)
	}
}

func TestMockContainer_RegisterFactory(t *testing.T) {
	container := NewMockContainer()
	err := container.RegisterFactory("test", func() string { return "value" })

	if err != nil {
		t.Errorf("RegisterFactory() error = %v", err)
	}
}

func TestMockContainer_GetAll(t *testing.T) {
	container := NewMockContainer()
	_ = container.RegisterValue("test1", "value1")
	_ = container.RegisterValue("test2", "value2")

	all := container.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() returned %d items, want 2", len(all))
	}
}

func TestMockContainer_ResolveByType(t *testing.T) {
	container := NewMockContainer()

	var s string
	_, err := container.ResolveByType(reflect.TypeOf(s))
	if err == nil {
		t.Error("ResolveByType() should return error (not implemented)")
	}
}

func TestNewTestApp(t *testing.T) {
	// Create a simple mock application
	container := NewMockContainer()
	mockApp := &mockApplication{container: container}

	testApp := NewTestApp(mockApp)

	if testApp == nil {
		t.Fatal("NewTestApp() returned nil")
	}

	if testApp.app == nil {
		t.Error("TestApp should have an app instance")
	}
}

func TestTestApp_Request(t *testing.T) {
	container := NewMockContainer()
	mockApp := &mockApplication{container: container}
	testApp := NewTestApp(mockApp)

	req := testApp.Request("GET", "/test")

	if req == nil {
		t.Fatal("Request() returned nil")
	}

	if req.method != "GET" {
		t.Errorf("Request method = %s, want GET", req.method)
	}

	if req.path != "/test" {
		t.Errorf("Request path = %s, want /test", req.path)
	}
}

func TestTestApp_Get(t *testing.T) {
	container := NewMockContainer()
	mockApp := &mockApplication{container: container}
	testApp := NewTestApp(mockApp)

	req := testApp.Get("/users")

	if req == nil {
		t.Fatal("Get() returned nil")
	}

	if req.method != "GET" {
		t.Errorf("Get method = %s, want GET", req.method)
	}

	if req.path != "/users" {
		t.Errorf("Get path = %s, want /users", req.path)
	}
}

func TestTestApp_Post(t *testing.T) {
	container := NewMockContainer()
	mockApp := &mockApplication{container: container}
	testApp := NewTestApp(mockApp)

	req := testApp.Post("/users")

	if req == nil {
		t.Fatal("Post() returned nil")
	}

	if req.method != "POST" {
		t.Errorf("Post method = %s, want POST", req.method)
	}

	if req.path != "/users" {
		t.Errorf("Post path = %s, want /users", req.path)
	}
}

func TestTestApp_Put(t *testing.T) {
	container := NewMockContainer()
	mockApp := &mockApplication{container: container}
	testApp := NewTestApp(mockApp)

	req := testApp.Put("/users/1")

	if req == nil {
		t.Fatal("Put() returned nil")
	}

	if req.method != "PUT" {
		t.Errorf("Put method = %s, want PUT", req.method)
	}

	if req.path != "/users/1" {
		t.Errorf("Put path = %s, want /users/1", req.path)
	}
}

func TestTestApp_Delete(t *testing.T) {
	container := NewMockContainer()
	mockApp := &mockApplication{container: container}
	testApp := NewTestApp(mockApp)

	req := testApp.Delete("/users/1")

	if req == nil {
		t.Fatal("Delete() returned nil")
	}

	if req.method != "DELETE" {
		t.Errorf("Delete method = %s, want DELETE", req.method)
	}

	if req.path != "/users/1" {
		t.Errorf("Delete path = %s, want /users/1", req.path)
	}
}

func TestTestApp_Patch(t *testing.T) {
	container := NewMockContainer()
	mockApp := &mockApplication{container: container}
	testApp := NewTestApp(mockApp)

	req := testApp.Patch("/users/1")

	if req == nil {
		t.Fatal("Patch() returned nil")
	}

	if req.method != "PATCH" {
		t.Errorf("Patch method = %s, want PATCH", req.method)
	}

	if req.path != "/users/1" {
		t.Errorf("Patch path = %s, want /users/1", req.path)
	}
}

// mockApplication is a mock implementation for testing
type mockApplication struct {
	container core.Container
}

func (m *mockApplication) Use(middleware ...core.MiddlewareFunc)                  {}
func (m *mockApplication) UseGlobalGuards(guards ...core.Guard)                   {}
func (m *mockApplication) UseGlobalInterceptors(interceptors ...core.Interceptor) {}
func (m *mockApplication) UseGlobalPipes(pipes ...core.Pipe)                      {}
func (m *mockApplication) UseGlobalFilters(filters ...core.ExceptionFilter)       {}
func (m *mockApplication) Get(path string, handler core.HandlerFunc)              {}
func (m *mockApplication) Post(path string, handler core.HandlerFunc)             {}
func (m *mockApplication) Put(path string, handler core.HandlerFunc)              {}
func (m *mockApplication) Delete(path string, handler core.HandlerFunc)           {}
func (m *mockApplication) Patch(path string, handler core.HandlerFunc)            {}
func (m *mockApplication) Listen(addr string) error                               { return nil }
func (m *mockApplication) Close() error                                           { return nil }
func (m *mockApplication) GetContainer() core.Container                           { return m.container }
func (m *mockApplication) GetLogger() core.Logger                                 { return nil }
func (m *mockApplication) SetLogger(logger core.Logger)                           {}
