package testing

import (
	"strings"
	"testing"
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

	_, err := container.ResolveByType("string")
	if err == nil {
		t.Error("ResolveByType() should return error (not implemented)")
	}
}
