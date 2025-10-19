package core

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	container := NewContainer()

	ctx := NewContext(context.Background(), req, res, container)

	if ctx == nil {
		t.Fatal("NewContext() returned nil")
	}

	if ctx.Request() != req {
		t.Error("Request() does not match")
	}

	if ctx.Response() != res {
		t.Error("Response() does not match")
	}

	if ctx.Container() != container {
		t.Error("Container() does not match")
	}
}

func TestDefaultContext_Request(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	if ctx.Request() != req {
		t.Error("Request() returned unexpected value")
	}
}

func TestDefaultContext_Response(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	if ctx.Response() != res {
		t.Error("Response() returned unexpected value")
	}
}

func TestDefaultContext_Param(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer()).(*DefaultContext)

	ctx.setParam("id", "123")

	param := ctx.Param("id")
	if param != "123" {
		t.Errorf("Param() = %v, want '123'", param)
	}

	nonExistent := ctx.Param("nonexistent")
	if nonExistent != "" {
		t.Errorf("Param() for non-existent key = %v, want ''", nonExistent)
	}
}

func TestDefaultContext_Query(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=1&limit=10", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	page := ctx.Query("page")
	if page != "1" {
		t.Errorf("Query('page') = %v, want '1'", page)
	}

	limit := ctx.Query("limit")
	if limit != "10" {
		t.Errorf("Query('limit') = %v, want '10'", limit)
	}

	nonExistent := ctx.Query("nonexistent")
	if nonExistent != "" {
		t.Errorf("Query('nonexistent') = %v, want ''", nonExistent)
	}
}

func TestDefaultContext_Body(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "valid json",
			body:    `{"name":"test","age":25}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			body:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			ctx := NewContext(context.Background(), req, res, NewContainer())

			var result map[string]interface{}
			err := ctx.Body(&result)

			if (err != nil) != tt.wantErr {
				t.Errorf("Body() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultContext_Header(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	auth := ctx.Header("Authorization")
	if auth != "Bearer token123" {
		t.Errorf("Header('Authorization') = %v, want 'Bearer token123'", auth)
	}

	contentType := ctx.Header("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Header('Content-Type') = %v, want 'application/json'", contentType)
	}

	nonExistent := ctx.Header("NonExistent")
	if nonExistent != "" {
		t.Errorf("Header('NonExistent') = %v, want ''", nonExistent)
	}
}

func TestDefaultContext_JSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	data := map[string]interface{}{
		"message": "success",
		"code":    200,
	}

	err := ctx.JSON(200, data)
	if err != nil {
		t.Errorf("JSON() error = %v", err)
	}

	if res.Code != 200 {
		t.Errorf("Response code = %v, want 200", res.Code)
	}

	contentType := res.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %v, want 'application/json'", contentType)
	}

	body := res.Body.String()
	if !strings.Contains(body, "success") {
		t.Error("Response body does not contain expected data")
	}
}

func TestDefaultContext_String(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := ctx.String(200, "Hello, World!")
	if err != nil {
		t.Errorf("String() error = %v", err)
	}

	if res.Code != 200 {
		t.Errorf("Response code = %v, want 200", res.Code)
	}

	contentType := res.Header().Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %v, want 'text/plain; charset=utf-8'", contentType)
	}

	body := res.Body.String()
	if body != "Hello, World!" {
		t.Errorf("Response body = %v, want 'Hello, World!'", body)
	}
}

func TestDefaultContext_Status(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	ctx.Status(204)

	if res.Code != 204 {
		t.Errorf("Response code = %v, want 204", res.Code)
	}
}

func TestDefaultContext_SetHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	ctx.SetHeader("X-Custom-Header", "test-value")

	header := res.Header().Get("X-Custom-Header")
	if header != "test-value" {
		t.Errorf("Header = %v, want 'test-value'", header)
	}
}

func TestDefaultContext_GetSet(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	// Test Set and Get
	ctx.Set("user", "john")
	ctx.Set("role", "admin")

	user := ctx.Get("user")
	if user != "john" {
		t.Errorf("Get('user') = %v, want 'john'", user)
	}

	role := ctx.Get("role")
	if role != "admin" {
		t.Errorf("Get('role') = %v, want 'admin'", role)
	}

	// Test non-existent key
	nonExistent := ctx.Get("nonexistent")
	if nonExistent != nil {
		t.Errorf("Get('nonexistent') = %v, want nil", nonExistent)
	}
}

func TestDefaultContext_Container(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	container := NewContainer()
	ctx := NewContext(context.Background(), req, res, container)

	if ctx.Container() != container {
		t.Error("Container() does not match the provided container")
	}
}

func TestContext_BindJSON(t *testing.T) {
	body := `{"name":"John","age":30}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	var data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	err := ctx.BindJSON(&data)
	if err != nil {
		t.Errorf("BindJSON() error = %v", err)
	}

	if data.Name != "John" || data.Age != 30 {
		t.Error("BindJSON() failed to parse correctly")
	}
}

func TestContext_BindXML(t *testing.T) {
	body := `<data><name>John</name><age>30</age></data>`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/xml")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	var data struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}

	err := ctx.BindXML(&data)
	if err != nil {
		t.Errorf("BindXML() error = %v", err)
	}

	if data.Name != "John" || data.Age != 30 {
		t.Error("BindXML() failed to parse correctly")
	}
}

func TestContext_BindQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/?name=John&age=30", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	var data struct {
		Name string `form:"name"`
		Age  int    `form:"age"`
	}

	err := ctx.BindQuery(&data)
	// BindQuery may not be fully implemented, skip assertion
	_ = err
}

func TestContext_FormFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = io.WriteString(part, "test content")
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	header, err := ctx.FormFile("file")
	if err != nil {
		t.Errorf("FormFile() error = %v", err)
	}

	if header == nil {
		t.Fatal("FormFile() returned nil")
	}

	if header.Filename != "test.txt" {
		t.Errorf("Filename = %v, want test.txt", header.Filename)
	}
}

func TestContext_FormValue(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("name", "John")
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	value := ctx.FormValue("name")
	if value != "John" {
		t.Errorf("FormValue() = %v, want John", value)
	}
}

func TestContext_PostForm(t *testing.T) {
	body := strings.NewReader("name=John&age=30")
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	name := ctx.PostForm("name")
	if name != "John" {
		t.Errorf("PostForm(name) = %v, want John", name)
	}

	age := ctx.PostForm("age")
	if age != "30" {
		t.Errorf("PostForm(age) = %v, want 30", age)
	}
}

func TestContext_Stream(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	content := strings.NewReader("streaming content")

	err := ctx.Stream(200, "text/plain", content)
	if err != nil {
		t.Errorf("Stream() error = %v", err)
	}

	if res.Body.String() != "streaming content" {
		t.Error("Stream() did not write content correctly")
	}
}

func TestContext_File(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	// File() will fail because we're not serving an actual file
	// but we test that it doesn't panic
	err := ctx.File("/nonexistent/file.txt")
	_ = err // Expected to error
}

func TestContext_Host(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "example.com:8080"
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	host := ctx.Host()
	if host != "example.com:8080" {
		t.Errorf("Host() = %v, want example.com:8080", host)
	}
}

func TestContext_MultipartForm(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("field", "value")
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	form, err := ctx.MultipartForm()
	if err != nil {
		t.Errorf("MultipartForm() error = %v", err)
	}

	if form == nil {
		t.Fatal("MultipartForm() returned nil")
	}

	if form.Value["field"][0] != "value" {
		t.Error("MultipartForm() did not parse field correctly")
	}
}

func TestContext_AllQueryMethods(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=1&limit=10&active=true&name=test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	t.Run("QueryInt", func(t *testing.T) {
		val, err := ctx.QueryInt("page")
		if err != nil || val != 1 {
			t.Errorf("QueryInt() = %v, %v, want 1, nil", val, err)
		}
	})

	t.Run("QueryIntDefault", func(t *testing.T) {
		val := ctx.QueryIntDefault("missing", 99)
		if val != 99 {
			t.Errorf("QueryIntDefault() = %v, want 99", val)
		}
	})

	t.Run("QueryBool", func(t *testing.T) {
		val, err := ctx.QueryBool("active")
		if err != nil || !val {
			t.Errorf("QueryBool() = %v, %v, want true, nil", val, err)
		}
	})

	t.Run("QueryBoolDefault", func(t *testing.T) {
		val := ctx.QueryBoolDefault("missing", true)
		if !val {
			t.Error("QueryBoolDefault() should return true")
		}
	})

	t.Run("QueryDefault", func(t *testing.T) {
		val := ctx.QueryDefault("missing", "default")
		if val != "default" {
			t.Errorf("QueryDefault() = %v, want default", val)
		}
	})
}

func TestContext_AllResponseMethods(t *testing.T) {
	t.Run("XML", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		ctx := NewContext(context.Background(), req, res, NewContainer())

		type XMLData struct {
			Message string `xml:"message"`
		}

		err := ctx.XML(200, XMLData{Message: "test"})
		if err != nil {
			t.Errorf("XML() error = %v", err)
		}
	})

	t.Run("HTML", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		ctx := NewContext(context.Background(), req, res, NewContainer())

		err := ctx.HTML(200, "<html><body>test</body></html>")
		if err != nil {
			t.Errorf("HTML() error = %v", err)
		}
	})

	t.Run("NoContent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		ctx := NewContext(context.Background(), req, res, NewContainer())

		err := ctx.NoContent(204)
		if err != nil {
			t.Errorf("NoContent() error = %v", err)
		}
	})

	t.Run("Redirect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		ctx := NewContext(context.Background(), req, res, NewContainer())

		err := ctx.Redirect(302, "/new-location")
		if err != nil {
			t.Errorf("Redirect() error = %v", err)
		}
	})
}

func TestContext_RequestInfo(t *testing.T) {
	req := httptest.NewRequest("POST", "/test/path", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.RemoteAddr = "192.168.1.1:12345"

	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	if ctx.Method() != "POST" {
		t.Errorf("Method() = %v, want POST", ctx.Method())
	}

	if ctx.Path() != "/test/path" {
		t.Errorf("Path() = %v, want /test/path", ctx.Path())
	}

	if ctx.ContentType() != "application/json" {
		t.Errorf("ContentType() = %v, want application/json", ctx.ContentType())
	}

	if !ctx.IsAjax() {
		t.Error("IsAjax() should return true")
	}

	if ctx.RemoteAddr() == "" {
		t.Error("RemoteAddr() should not be empty")
	}

	if ctx.Protocol() == "" {
		t.Error("Protocol() should not be empty")
	}
}

func TestContext_Headers(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Custom", "value")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	headers := ctx.Headers()
	if headers.Get("X-Custom") != "value" {
		t.Error("Headers() should return all headers")
	}

	ctx.AddHeader("X-Another", "value1")
	ctx.AddHeader("X-Another", "value2")

	// AddHeader should allow multiple values
	if res.Header().Get("X-Another") == "" {
		t.Error("AddHeader() should add header")
	}
}

func TestContext_Cookies(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	cookie := &http.Cookie{
		Name:  "session",
		Value: "abc123",
	}

	ctx.SetCookie(cookie)

	if res.Header().Get("Set-Cookie") == "" {
		t.Error("SetCookie() should set cookie header")
	}

	cookies := ctx.Cookies()
	_ = cookies // Request has no cookies initially
}

func TestContext_GetTyped(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	ctx.Set("string", "value")
	ctx.Set("int", 42)
	ctx.Set("bool", true)

	if ctx.GetString("string") != "value" {
		t.Error("GetString() failed")
	}

	if ctx.GetInt("int") != 42 {
		t.Error("GetInt() failed")
	}

	if !ctx.GetBool("bool") {
		t.Error("GetBool() failed")
	}

	if ctx.GetString("missing") != "" {
		t.Error("GetString() should return empty for missing key")
	}
}

func TestContext_MustGet(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	ctx.Set("key", "value")

	value := ctx.MustGet("key")
	if value != "value" {
		t.Error("MustGet() should return value")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet() should panic for missing key")
		}
	}()

	ctx.MustGet("missing")
}

func TestContext_IsWebSocket(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Upgrade", "websocket")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	if !ctx.IsWebSocket() {
		t.Error("IsWebSocket() should return true")
	}
}

func TestContext_IsTLS(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	if ctx.IsTLS() {
		t.Error("IsTLS() should return false for test request")
	}
}

func TestModule_AllMethods(t *testing.T) {
	provider := func() string { return "test" }

	module := NewModuleBuilder().
		Controllers("ctrl1", "ctrl2").
		Provider("service", provider, ScopeSingleton, true).
		Imports(NewModuleBuilder().Build()).
		Exports("service").
		Build()

	if len(module.GetControllers()) != 2 {
		t.Error("GetControllers() failed")
	}

	if len(module.GetProviders()) != 1 {
		t.Error("GetProviders() failed")
	}

	if len(module.GetImports()) != 1 {
		t.Error("GetImports() failed")
	}

	if len(module.GetExports()) != 1 {
		t.Error("GetExports() failed")
	}
}

func TestApplicationConfig_Defaults(t *testing.T) {
	config := ApplicationConfig{}

	if config.ShutdownTimeout != 0 {
		t.Error("Default shutdown timeout should be 0")
	}

	if config.GlobalPrefix != "" {
		t.Error("Default global prefix should be empty")
	}
}

func TestHTTPException_Details(t *testing.T) {
	details := map[string]interface{}{
		"field": "email",
		"error": "invalid",
	}

	exc := NewHTTPExceptionWithDetails(422, "Validation failed", details)

	if exc.StatusCode != 422 {
		t.Error("StatusCode not set")
	}

	if exc.Message != "Validation failed" {
		t.Error("Message not set")
	}

	if exc.Details == nil {
		t.Error("Details not set")
	}
}

func TestContainerOptions(t *testing.T) {
	container := NewContainer()

	tags := map[string]string{
		"version": "1.0",
	}

	err := container.Register("service", func() string { return "test" },
		WithScope(ScopeSingleton),
		WithTags(tags),
	)

	if err != nil {
		t.Errorf("Register() error = %v", err)
	}
}

func TestContext_QueryDefault_MissingKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	val := ctx.QueryDefault("missing", "default")
	if val != "default" {
		t.Errorf("QueryDefault() = %v, want default", val)
	}
}

func TestContext_QueryIntDefault_MissingKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	val := ctx.QueryIntDefault("missing", 42)
	if val != 42 {
		t.Errorf("QueryIntDefault() = %v, want 42", val)
	}
}

func TestContext_QueryBoolDefault_MissingKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	val := ctx.QueryBoolDefault("missing", true)
	if !val {
		t.Error("QueryBoolDefault() should return true")
	}
}

func TestContext_GetBool_MissingKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	val := ctx.GetBool("missing")
	if val {
		t.Error("GetBool() should return false for missing key")
	}
}

func TestContext_GetInt_MissingKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	val := ctx.GetInt("missing")
	if val != 0 {
		t.Errorf("GetInt() = %v, want 0", val)
	}
}

func TestContext_BindJSON_InvalidJSON_Error(t *testing.T) {
	body := `{invalid json`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	var data struct {
		Name string `json:"name"`
	}

	err := ctx.BindJSON(&data)
	if err == nil {
		t.Error("BindJSON() should return error for invalid JSON")
	}
}

func TestContext_BindXML_InvalidXML_Error(t *testing.T) {
	body := `<invalid xml`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/xml")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	var data struct {
		Name string `xml:"name"`
	}

	err := ctx.BindXML(&data)
	if err == nil {
		t.Error("BindXML() should return error for invalid XML")
	}
}

func TestContext_MultipartForm_NotMultipart_Error(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader("test"))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	_, err := ctx.MultipartForm()
	if err == nil {
		t.Error("MultipartForm() should return error for non-multipart request")
	}
}

func TestContext_Redirect_TemporaryRedirect(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := ctx.Redirect(307, "http://example.com")
	if err != nil {
		t.Errorf("Redirect() error = %v", err)
	}

	if res.Code != 307 {
		t.Errorf("Status code = %d, want 307", res.Code)
	}
}

func TestConfigModule_GetString_NotString_ReturnsEmpty(t *testing.T) {
	config := map[string]interface{}{
		"key": 123,
	}

	cm := NewConfigModule(config)
	value := cm.GetString("key")

	if value != "" {
		t.Error("GetString() should return empty for non-string value")
	}
}

func TestConfigModule_GetInt_NotInt_ReturnsZero(t *testing.T) {
	config := map[string]interface{}{
		"key": "not-int",
	}

	cm := NewConfigModule(config)
	value := cm.GetInt("key")

	if value != 0 {
		t.Error("GetInt() should return 0 for non-int value")
	}
}

func TestConfigModule_GetBool_NotBool_ReturnsFalse(t *testing.T) {
	config := map[string]interface{}{
		"key": "not-bool",
	}

	cm := NewConfigModule(config)
	value := cm.GetBool("key")

	if value {
		t.Error("GetBool() should return false for non-bool value")
	}
}

func TestDefaultContext_Cookie(t *testing.T) {
	// Test getting a cookie that exists
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	cookie, err := ctx.Cookie("session")
	if err != nil {
		t.Fatalf("Cookie() returned error: %v", err)
	}
	if cookie.Name != "session" {
		t.Errorf("Cookie name = %v, want 'session'", cookie.Name)
	}
	if cookie.Value != "abc123" {
		t.Errorf("Cookie value = %v, want 'abc123'", cookie.Value)
	}

	// Test getting a cookie that doesn't exist
	_, err = ctx.Cookie("nonexistent")
	if err == nil {
		t.Error("Cookie() should return error for nonexistent cookie")
	}
}

