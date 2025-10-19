package main

import (
	"log"

	"github.com/goruum/ruum/core"
	"github.com/goruum/ruum/factory"
	ruumhttp "github.com/goruum/ruum/http"
	"github.com/goruum/ruum/logger"
	"github.com/goruum/ruum/middleware"
)

// User represents a user model
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserController handles user-related requests
type UserController struct {
	*ruumhttp.BaseController
}

// NewUserController creates a new user controller
func NewUserController() *UserController {
	ctrl := &UserController{
		BaseController: ruumhttp.NewBaseController("/users"),
	}
	
	// Register routes
	ctrl.Get("", ctrl.GetUsers)
	ctrl.Get("/{id}", ctrl.GetUser)
	ctrl.Post("", ctrl.CreateUser)
	ctrl.Put("/{id}", ctrl.UpdateUser)
	ctrl.Delete("/{id}", ctrl.DeleteUser)
	
	return ctrl
}

func (c *UserController) GetUsers(ctx core.Context) error {
	users := []User{
		{ID: "1", Name: "John Doe", Email: "john@example.com"},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
	}
	
	return ctx.JSON(200, map[string]interface{}{
		"data": users,
	})
}

func (c *UserController) GetUser(ctx core.Context) error {
	id := ctx.Param("id")
	
	user := User{
		ID:    id,
		Name:  "John Doe",
		Email: "john@example.com",
	}
	
	return ctx.JSON(200, map[string]interface{}{
		"data": user,
	})
}

func (c *UserController) CreateUser(ctx core.Context) error {
	var user User
	if err := ctx.Body(&user); err != nil {
		return core.BadRequestException("Invalid request body")
	}
	
	// Simulate user creation
	user.ID = "3"
	
	return ctx.JSON(201, map[string]interface{}{
		"message": "User created successfully",
		"data":    user,
	})
}

func (c *UserController) UpdateUser(ctx core.Context) error {
	id := ctx.Param("id")
	
	var user User
	if err := ctx.Body(&user); err != nil {
		return core.BadRequestException("Invalid request body")
	}
	
	user.ID = id
	
	return ctx.JSON(200, map[string]interface{}{
		"message": "User updated successfully",
		"data":    user,
	})
}

func (c *UserController) DeleteUser(ctx core.Context) error {
	id := ctx.Param("id")
	
	return ctx.JSON(200, map[string]interface{}{
		"message": "User deleted successfully",
		"id":      id,
	})
}

// HealthController handles health check requests
type HealthController struct {
	*ruumhttp.BaseController
}

func NewHealthController() *HealthController {
	ctrl := &HealthController{
		BaseController: ruumhttp.NewBaseController("/health"),
	}
	
	ctrl.Get("", ctrl.Check)
	
	return ctrl
}

func (c *HealthController) Check(ctx core.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"status": "ok",
		"time":   "2024-01-01T00:00:00Z",
	})
}

func main() {
	// Create logger
	logger := logger.NewDefaultLogger()
	
	// Create application module
	appModule := core.NewModuleBuilder().
		Controllers(
			NewUserController(),
			NewHealthController(),
		).
		Build()
	
	// Create application
	app, err := factory.CreateApplication(
		appModule,
		factory.WithLogger(logger),
		factory.WithCORS("*"),
		factory.WithShutdownHooks(true),
	)
	
	if err != nil {
		log.Fatal("Failed to create application:", err)
	}
	
	// Use middleware
	app.Use(middleware.Recovery(logger))
	app.Use(middleware.Logger(logger))
	app.Use(middleware.CORS(middleware.DefaultCORSConfig()))
	
	// Start server
	logger.Info("🚀 Starting Ruum application...", nil)
	
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

