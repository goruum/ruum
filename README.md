# 🏛️ Ruum

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square&logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/goruum/ruum.svg)](https://pkg.go.dev/github.com/goruum/ruum)
[![Release](https://img.shields.io/github/v/release/goruum/ruum?style=flat-square&label=Release&color=blue&logo=github&logoColor=white)](https://github.com/goruum/ruum/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/goruum/ruum/test.yml?branch=main&style=flat-square&label=CI&logo=githubactions&logoColor=white)](https://github.com/goruum/ruum/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/codecov/c/github/goruum/ruum?style=flat-square&label=Coverage&logo=codecov&logoColor=white)](https://codecov.io/gh/goruum/ruum)
[![Sonar](https://img.shields.io/sonar/quality_gate/goruum_ruum?server=https%3A%2F%2Fsonarcloud.io&style=flat-square&label=Sonar&logo=sonarqubecloud&logoColor=white)](https://sonarcloud.io/summary/new_code?id=goruum_ruum)
[![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-brightgreen?style=flat-square&logo=github&logoColor=white)](https://github.com/goruum/ruum/pulls)

**A progressive, modular web framework for Go**

Build scalable and maintainable server-side applications with elegant architecture patterns

[Features](#-features) •
[Installation](#-installation) •
[Quick Start](#-quick-start) •
[Examples](#-examples) •
[Documentation](#-documentation)

</div>

---

## ✨ Features

### Core Features
- 🏗️ **Modular Architecture** - Organize code into independent, reusable modules
- 💉 **Dependency Injection** - Built-in DI container with multiple scopes (singleton, transient, request)
- 🎯 **Type-Safe** - Leverage Go's type system for compile-time safety
- 🔄 **Dynamic Modules** - Support for dynamic, configurable modules with ForRoot/ForFeature patterns
- 🔁 **Lifecycle Hooks** - Comprehensive lifecycle management with multiple hooks

### HTTP & API
- 🌐 **Powerful Context** - Rich context with cookies, file uploads, redirects, and more
- 🛡️ **Guards** - Declarative route protection with authentication and authorization
- 🔄 **Interceptors** - Transform requests and responses with powerful interceptor chains
- 🚰 **Pipes** - Validate and transform input data before it reaches handlers
- ⚠️ **Exception Filters** - Centralized error handling with stack traces
- 🌐 **WebSocket Support** - Full-featured WebSocket implementation with rooms and events

### Validation & Data
- ✅ **Struct Validation** - Powerful tag-based validation similar to class-validator
- 📝 **Data Binding** - Automatic binding of JSON, XML, query params, and forms
- 🔍 **Type Conversion** - Smart type conversion for query parameters

### Middleware
- ⚡ **Circuit Breaker** (NEW!) - Prevent cascading failures with automatic service recovery
- ⏱️ **Timeout** (NEW!) - Request timeout control with graceful handling
- 🚦 **Rate Limiting** - Configurable request rate limiting
- 🗜️ **Compression** - Automatic gzip compression with smart content detection
- 💾 **Caching** - Response caching with TTL and smart invalidation
- 🏥 **Health Checks** - Built-in health check endpoints with custom indicators
- 🔢 **Request ID** - Automatic request ID generation and tracking
- 🪵 **Advanced Logging** - Request/response logging with correlation IDs
- 🔄 **CORS** - Flexible CORS configuration

### Advanced Features
- 📅 **Task Scheduling** - Cron-like task scheduling with multiple schedule types
- 📡 **Event Emitter** - Event-driven architecture with sync/async event handling
- 🧪 **Testing Utilities** - Comprehensive testing framework with mocks and helpers
- ⚙️ **Configuration** - Load config from files or environment variables
- 📊 **Stack Traces** - Detailed error stack traces for debugging

## 📦 Installation

```bash
go get github.com/goruum/ruum
```

**Requirements:** Go 1.23 or higher

## 🚀 Quick Start

Create a simple REST API in minutes:

```go
package main

import (
    "log"
    
    "github.com/goruum/ruum/core"
    "github.com/goruum/ruum/factory"
    "github.com/goruum/ruum/http"
    "github.com/goruum/ruum/logger"
    "github.com/goruum/ruum/middleware"
)

// Define your controller
type UserController struct {
    *http.BaseController
}

func NewUserController() *UserController {
    ctrl := &UserController{
        BaseController: http.NewBaseController("/users"),
    }
    
    ctrl.Get("", ctrl.GetAll)
    ctrl.Get("/{id}", ctrl.GetOne)
    ctrl.Post("", ctrl.Create)
    
    return ctrl
}

func (c *UserController) GetAll(ctx core.Context) error {
    users := []map[string]interface{}{
        {"id": "1", "name": "Alice"},
        {"id": "2", "name": "Bob"},
    }
    return ctx.JSON(200, map[string]interface{}{"data": users})
}

func (c *UserController) GetOne(ctx core.Context) error {
    id := ctx.Param("id")
    return ctx.JSON(200, map[string]interface{}{
        "id": id,
        "name": "Alice",
    })
}

func (c *UserController) Create(ctx core.Context) error {
    var body map[string]interface{}
    if err := ctx.Body(&body); err != nil {
        return core.BadRequestException("Invalid request body")
    }
    return ctx.JSON(201, map[string]interface{}{
        "message": "User created",
        "data": body,
    })
}

func main() {
    // Create logger
    logger := logger.NewDefaultLogger()
    
    // Build your module
    appModule := core.NewModuleBuilder().
        Controllers(NewUserController()).
        Build()
    
    // Create application
    app, err := factory.CreateApplication(
        appModule,
        factory.WithLogger(logger),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Add middleware
    app.Use(middleware.Recovery(logger))
    app.Use(middleware.Logger(logger))
    
    // Start server
    logger.Info("🚀 Server running on http://localhost:3000", nil)
    if err := app.Listen(":3000"); err != nil {
        log.Fatal(err)
    }
}
```

Run your application:

```bash
go run main.go
```

Test it:

```bash
# Get all users
curl http://localhost:3000/users

# Get specific user
curl http://localhost:3000/users/1

# Create user
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```

## 📚 Core Concepts

### Modules

Organize your application into cohesive feature modules:

```go
module := core.NewModuleBuilder().
    Controllers(
        NewUserController(),
        NewProductController(),
    ).
    Provider("userService", NewUserService, core.ScopeSingleton, true).
    Provider("productService", NewProductService, core.ScopeSingleton, true).
    Imports(
        NewDatabaseModule(),
        NewAuthModule(),
    ).
    Exports("userService").
    Build()
```

### Dependency Injection

Automatic dependency resolution with lifecycle management:

```go
// Register provider
container.Register("database", func() *Database {
    return NewDatabase("localhost:5432")
}, core.ScopeSingleton)

// Register with dependencies
container.Register("userService", func(db *Database) *UserService {
    return NewUserService(db)
}, core.ScopeSingleton)

// Ruum automatically resolves and injects dependencies
```

**Scopes:**
- **Singleton**: One instance for the entire application
- **Transient**: New instance on every resolution
- **Request**: One instance per HTTP request

### Controllers & Routing

Clean, declarative routing:

```go
type ProductController struct {
    *http.BaseController
    service *ProductService
}

func NewProductController(service *ProductService) *ProductController {
    ctrl := &ProductController{
        BaseController: http.NewBaseController("/products"),
        service:        service,
    }
    
    ctrl.Get("", ctrl.FindAll)
    ctrl.Get("/{id}", ctrl.FindOne)
    ctrl.Post("", ctrl.Create)
    ctrl.Put("/{id}", ctrl.Update)
    ctrl.Delete("/{id}", ctrl.Delete)
    
    return ctrl
}

func (c *ProductController) FindAll(ctx core.Context) error {
    products := c.service.GetAll()
    return ctx.JSON(200, map[string]interface{}{"data": products})
}
```

### Guards

Protect routes with guards:

```go
import "github.com/goruum/ruum/guards"

// Apply to entire controller
ctrl.UseGuards(guards.NewAuthGuard())

// Apply to specific route
ctrl.Get("/admin", handler, http.WithGuards(guards.NewAuthGuard()))

// Create custom guard
type AdminGuard struct{}

func (g *AdminGuard) CanActivate(ctx core.Context) (bool, error) {
    user := ctx.Get("user")
    if user == nil {
        return false, core.UnauthorizedException("Not authenticated")
    }
    
    if !user.IsAdmin {
        return false, core.ForbiddenException("Admin access required")
    }
    
    return true, nil
}
```

### Interceptors

Transform requests and responses:

```go
import "github.com/goruum/ruum/interceptors"

// Use global interceptor
app.UseGlobalInterceptors(interceptors.NewLoggingInterceptor(logger))

// Custom interceptor
type CacheInterceptor struct {
    cache map[string]interface{}
}

func (i *CacheInterceptor) Intercept(ctx core.Context, next core.HandlerFunc) error {
    key := ctx.Request().URL.Path
    
    // Check cache
    if cached, exists := i.cache[key]; exists {
        return ctx.JSON(200, cached)
    }
    
    // Call handler
    err := next(ctx)
    
    // Cache result
    // ... implementation
    
    return err
}
```

### Pipes

Validate and transform input:

```go
import "github.com/goruum/ruum/pipes"

// Built-in pipes
parseIntPipe := pipes.NewParseIntPipe()
parseBoolPipe := pipes.NewParseBoolPipe()

// Validation pipe
validationPipe := pipes.NewValidationPipe(
    pipes.Required(),
    pipes.MinLength(3),
    pipes.MaxLength(50),
)

// Custom pipe
type UpperCasePipe struct{}

func (p *UpperCasePipe) Transform(value interface{}, metadata *core.ArgumentMetadata) (interface{}, error) {
    str, ok := value.(string)
    if !ok {
        return nil, core.BadRequestException("Value must be a string")
    }
    return strings.ToUpper(str), nil
}
```

### Exception Handling

Structured error responses:

```go
// Use built-in exceptions
return core.BadRequestException("Invalid input")
return core.UnauthorizedException("Not authenticated")
return core.ForbiddenException("Access denied")
return core.NotFoundException("Resource not found")
return core.InternalServerErrorException("Server error")

// Custom exception with details
return core.NewHTTPExceptionWithDetails(
    422,
    "Validation failed",
    map[string]interface{}{
        "fields": []string{"email", "password"},
        "errors": []string{"Invalid format", "Too short"},
    },
)

// Global exception filter
app.UseGlobalFilters(core.NewDefaultExceptionFilter())
```

### Middleware

Process requests globally with powerful middleware:

```go
import "github.com/goruum/ruum/middleware"

// Built-in middleware
app.Use(middleware.Recovery(logger))        // Panic recovery
app.Use(middleware.Logger(logger))          // Request logging
app.Use(middleware.CORS(middleware.DefaultCORSConfig()))

// Circuit Breaker (NEW!) - Prevent cascading failures
app.Use(middleware.CircuitBreaker(middleware.CircuitBreakerConfig{
    MaxRequests: 3,
    Timeout:     60 * time.Second,
    ReadyToTrip: func(counts middleware.Counts) bool {
        return counts.ConsecutiveFailures > 5
    },
}))

// Timeout (NEW!) - Request timeout control
app.Use(middleware.Timeout(middleware.TimeoutConfig{
    Timeout: 30 * time.Second,
    OnTimeout: func(ctx core.Context) {
        logger.Warn("Request timed out", map[string]interface{}{
            "path": ctx.Path(),
        })
    },
}))

// Rate limiting
app.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
    RequestsPerWindow: 100,
    Window:            time.Minute,
}))

// Compression
app.Use(middleware.Compression(middleware.DefaultCompressionConfig()))

// Request ID tracking
app.Use(middleware.RequestID(middleware.DefaultRequestIDConfig()))

// Response caching
app.Use(middleware.Cache(middleware.CacheConfig{
    TTL: 5 * time.Minute,
    Methods: []string{"GET"},
}))

// Health checks
app.Use(middleware.HealthCheckMiddleware(middleware.HealthCheckConfig{
    Path: "/health",
    Checks: map[string]func() bool{
        "database": func() bool { return db.Ping() == nil },
    },
}))

// Custom middleware
func AuthMiddleware(next core.HandlerFunc) core.HandlerFunc {
    return func(ctx core.Context) error {
        token := ctx.Header("Authorization")
        if token == "" {
            return core.UnauthorizedException("Missing token")
        }
        
        user := validateToken(token)
        ctx.Set("user", user)
        
        return next(ctx)
    }
}

app.Use(AuthMiddleware)
```

### Validation

Struct validation with tags (similar to class-validator):

```go
import "github.com/goruum/ruum/validation"

type CreateUserDTO struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,minLength=8"`
    Age      int    `json:"age" validate:"required,min=18,max=100"`
    Username string `json:"username" validate:"required,alphanumeric,minLength=3,maxLength=20"`
    Website  string `json:"website" validate:"url"`
}

func (c *UserController) Create(ctx core.Context) error {
    var dto CreateUserDTO
    if err := ctx.BindJSON(&dto); err != nil {
        return core.BadRequestException("Invalid JSON")
    }
    
    // Validate
    if err := validation.Validate(&dto); err != nil {
        return core.ValidationError("Validation failed", map[string]interface{}{
            "errors": err,
        })
    }
    
    // Process...
    return ctx.JSON(201, dto)
}
```

### WebSocket

Full-featured WebSocket support:

```go
import "github.com/goruum/ruum/websocket"

// Create hub
hub := websocket.NewHub()
go hub.Run()

// Register event handlers
hub.On("chat:message", func(client *websocket.Client, msg websocket.Message) error {
    // Broadcast to room
    hub.BroadcastToRoom("lobby", "chat:message", msg.Data)
    return nil
})

hub.On("room:join", func(client *websocket.Client, msg websocket.Message) error {
    room := msg.Data.(string)
    client.JoinRoom(room)
    return nil
})

// Add WebSocket route
app.Get("/ws", websocket.Handler(websocket.HandlerConfig{
    Hub: hub,
    OnConnect: func(client *websocket.Client) error {
        log.Printf("Client connected: %s", client.ID)
        return nil
    },
    OnDisconnect: func(client *websocket.Client) {
        log.Printf("Client disconnected: %s", client.ID)
    },
}))
```

### Event Emitter

Event-driven architecture:

```go
import "github.com/goruum/ruum/events"

// Create emitter
emitter := events.NewEventEmitter()

// Register listeners
emitter.On("user.created", func(event events.Event) error {
    user := event.Payload.(User)
    log.Printf("User created: %s", user.Email)
    // Send welcome email
    return sendWelcomeEmail(user)
})

emitter.OnAsync("user.created", func(event events.Event) {
    user := event.Payload.(User)
    // Update analytics (async)
    analytics.Track(user.ID, "user_created")
})

// Emit events
emitter.Emit("user.created", newUser)
```

### Task Scheduling

Cron-like task scheduling:

```go
import "github.com/goruum/ruum/scheduler"

// Create scheduler
sched := scheduler.NewScheduler()
sched.SetLogger(logger)

// Add tasks
sched.AddTask("cleanup", "Database Cleanup", 
    scheduler.Daily(2, 0), // Every day at 2:00 AM
    func(ctx context.Context) error {
        return cleanupOldRecords()
    },
)

sched.AddTask("backup", "Daily Backup",
    scheduler.Every(24 * time.Hour),
    func(ctx context.Context) error {
        return performBackup()
    },
)

sched.AddTask("health-check", "Health Check",
    scheduler.Every(5 * time.Minute),
    func(ctx context.Context) error {
        return checkSystemHealth()
    },
)

// Start scheduler
sched.Start()
```

### Testing

Comprehensive testing utilities:

```go
import (
    "testing"
    "github.com/goruum/ruum/testing"
)

func TestUserController(t *testing.T) {
    // Create test context
    ctx := testing.NewTestRequest("POST", "/users").
        JSON(map[string]interface{}{
            "email": "test@example.com",
            "password": "password123",
        }).
        Build()
    
    // Mock service
    mockService := testing.NewMockService()
    mockService.SetReturn("CreateUser", &User{ID: "1"}, nil)
    ctx.RegisterValue("userService", mockService)
    
    // Create controller and test
    ctrl := NewUserController()
    err := ctrl.Create(ctx)
    
    // Assertions
    assert := testing.NewAssertionHelper(t)
    assert.ExpectStatus(ctx, 201)
    assert.ExpectJSON(ctx, &result)
    
    // Verify mock was called
    if !mockService.WasCalled("CreateUser") {
        t.Error("CreateUser was not called")
    }
}
```

## 🎯 Examples

### Basic REST API

Check out `examples/basic/` for a simple CRUD API with:
- Multiple controllers
- JSON request/response
- Error handling
- Middleware

```bash
cd examples/basic
go run main.go
```

### Advanced Application

Check out `examples/advanced/` for a full-featured app with:
- Service layer with business logic
- Dependency injection
- Authentication guards
- Logging interceptors
- Global exception filters
- Protected and public routes

```bash
cd examples/advanced
go run main.go
```

## 🔧 Configuration

### Load from JSON

```go
import "github.com/goruum/ruum/config"

configService := config.NewConfigService()
configService.LoadFromFile("./config.json")

port := configService.GetInt("port")
dbHost := configService.GetString("database.host")
```

### Environment Variables

```go
configService := config.NewConfigService()

// Access directly from env
port := configService.GetInt("PORT")          // from PORT env var
dbUrl := configService.GetString("DATABASE_URL")
```

## 🏗️ Project Structure

Recommended structure for Ruum applications:

```
my-app/
├── main.go                    # Application entry point
├── go.mod
├── modules/
│   ├── user/
│   │   ├── user.controller.go
│   │   ├── user.service.go
│   │   ├── user.module.go
│   │   └── user.model.go
│   ├── product/
│   │   ├── product.controller.go
│   │   ├── product.service.go
│   │   ├── product.module.go
│   │   └── product.model.go
│   └── auth/
│       ├── auth.controller.go
│       ├── auth.service.go
│       ├── auth.module.go
│       └── auth.guard.go
├── common/
│   ├── guards/
│   ├── interceptors/
│   ├── pipes/
│   └── filters/
└── config/
    └── config.json
```

## 🧪 Testing

Ruum is designed for testability:

```go
func TestUserController(t *testing.T) {
    // Create test container
    container := core.NewContainer()
    container.RegisterValue("userService", NewMockUserService())
    
    // Create controller
    ctrl := NewUserController()
    
    // Create test context
    req := httptest.NewRequest("GET", "/users", nil)
    res := httptest.NewRecorder()
    ctx := core.NewContext(context.Background(), req, res, container)
    
    // Execute handler
    err := ctrl.GetAll(ctx)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 200, res.Code)
}
```

## 📖 API Reference

### Application

```go
// Create application
app, err := factory.CreateApplication(rootModule, ...options)

// Configuration options
factory.WithLogger(logger)
factory.WithGlobalPrefix("/api/v1")
factory.WithCORS("*")
factory.WithShutdownHooks(true)

// Add global components
app.Use(middleware...)
app.UseGlobalGuards(guards...)
app.UseGlobalInterceptors(interceptors...)
app.UseGlobalPipes(pipes...)
app.UseGlobalFilters(filters...)

// Start server
app.Listen(":3000")
```

### Context Methods

```go
// Request
ctx.Request() *http.Request
ctx.Param("id") string
ctx.Query("page") string
ctx.Body(&dto) error
ctx.Header("Authorization") string

// Response
ctx.JSON(200, data) error
ctx.String(200, "text") error
ctx.Status(204)
ctx.SetHeader("X-Custom", "value")

// Data
ctx.Get("key") interface{}
ctx.Set("key", value)
ctx.Container() core.Container
```

### Module Builder

```go
module := core.NewModuleBuilder().
    Controllers(ctrl1, ctrl2).
    Provider("service", factory, scope, export).
    Imports(module1, module2).
    Exports("service").
    Build()
```

## 🤝 Contributing

We use **Conventional Commits** for automatic versioning with **strict quality gates**. 

### Quality Guaranteed Releases 🛡️

Every merge to `main` triggers automatic release **ONLY IF** all checks pass:
- ✅ Tests pass (≥80% coverage)
- ✅ Linting passes
- ✅ Security checks pass
- ✅ Build succeeds

Run `make ci` locally to verify before pushing!

**Commit Types:**
- `feat:` → Minor version (1.0.0 → 1.1.0)
- `fix:` → Patch version (1.0.0 → 1.0.1)
- `feat!:` → Major version (1.0.0 → 2.0.0)

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with ❤️ by the Ruum community. Special thanks to all contributors!

## 📞 Support

- 🐛 Issues: [GitHub Issues](https://github.com/goruum/ruum/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/goruum/ruum/discussions)

---

<div align="center">

**[⬆ back to top](#ruum)**

Made with ❤️ using Go

</div>
