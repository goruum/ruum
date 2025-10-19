# 🏛️ Ruum

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square&logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/goruum/ruum?style=flat-square)](https://goreportcard.com/report/github.com/goruum/ruum)
[![Documentation](https://img.shields.io/badge/Documentation-pkg.go.dev-007d9c?style=flat-square&logo=go&logoColor=white)](https://pkg.go.dev/github.com/goruum/ruum)
[![Release](https://img.shields.io/github/v/release/goruum/ruum?style=flat-square&label=Release&color=blue&logo=github&logoColor=white)](https://github.com/goruum/ruum/releases)
[![Tests](https://img.shields.io/github/actions/workflow/status/goruum/ruum/test.yml?branch=main&style=flat-square&label=Tests&logo=githubactions&logoColor=white)](https://github.com/goruum/ruum/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/codecov/c/github/goruum/ruum?style=flat-square&label=Coverage&logo=codecov&logoColor=white)](https://codecov.io/gh/goruum/ruum)
[![Quality Gate](https://img.shields.io/sonar/quality_gate/goruum_ruum?server=https%3A%2F%2Fsonarcloud.io&style=flat-square&label=Quality%20Gate&logo=sonarcloud&logoColor=white)](https://sonarcloud.io/summary/new_code?id=goruum_ruum)
[![Security](https://img.shields.io/sonar/security_rating/goruum_ruum?server=https%3A%2F%2Fsonarcloud.io&style=flat-square&label=Security&logo=sonarcloud&logoColor=white)](https://sonarcloud.io/summary/new_code?id=goruum_ruum)
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

- 🏗️ **Modular Architecture** - Organize code into independent, reusable modules
- 💉 **Dependency Injection** - Built-in DI container with multiple scopes (singleton, transient, request)
- 🎯 **Type-Safe** - Leverage Go's type system for compile-time safety
- 🛡️ **Guards** - Declarative route protection with authentication and authorization
- 🔄 **Interceptors** - Transform requests and responses with powerful interceptor chains
- 🚰 **Pipes** - Validate and transform input data before it reaches handlers
- ⚠️ **Exception Filters** - Centralized, structured error handling
- 🪵 **Advanced Logging** - Built-in logger with colored output and multiple levels
- ⚙️ **Configuration Management** - Load config from files or environment variables
- 🌐 **Powerful Routing** - Clean, declarative HTTP routing with gorilla/mux
- 🔌 **Middleware Chain** - Flexible middleware system for request/response processing
- 🧪 **Testable** - Designed with testing in mind from the ground up

## 📦 Installation

```bash
go get github.com/goruum/ruum
```

**Requirements:** Go 1.21 or higher

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
return core.NewHttpExceptionWithDetails(
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

Process requests globally:

```go
import "github.com/goruum/ruum/middleware"

// Built-in middleware
app.Use(middleware.Recovery(logger))        // Panic recovery
app.Use(middleware.Logger(logger))          // Request logging
app.Use(middleware.CORS(middleware.DefaultCORSConfig()))

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
- ✅ Tests pass (≥70% coverage)
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
