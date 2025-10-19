package main

import (
	"log"

	"github.com/goruum/ruum/core"
	"github.com/goruum/ruum/factory"
	"github.com/goruum/ruum/guards"
	ruumhttp "github.com/goruum/ruum/http"
	"github.com/goruum/ruum/interceptors"
	"github.com/goruum/ruum/logger"
	"github.com/goruum/ruum/middleware"
)

// Product represents a product model
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

// ProductService handles business logic for products
type ProductService struct {
	products map[string]*Product
}

func NewProductService() *ProductService {
	return &ProductService{
		products: map[string]*Product{
			"1": {ID: "1", Name: "Laptop", Description: "High-end laptop", Price: 1299.99},
			"2": {ID: "2", Name: "Mouse", Description: "Wireless mouse", Price: 29.99},
		},
	}
}

func (s *ProductService) FindAll() []*Product {
	products := make([]*Product, 0, len(s.products))
	for _, p := range s.products {
		products = append(products, p)
	}
	return products
}

func (s *ProductService) FindOne(id string) (*Product, error) {
	product, exists := s.products[id]
	if !exists {
		return nil, core.NotFoundException("Product not found")
	}
	return product, nil
}

func (s *ProductService) Create(product *Product) (*Product, error) {
	s.products[product.ID] = product
	return product, nil
}

func (s *ProductService) Update(id string, product *Product) (*Product, error) {
	if _, exists := s.products[id]; !exists {
		return nil, core.NotFoundException("Product not found")
	}
	product.ID = id
	s.products[id] = product
	return product, nil
}

func (s *ProductService) Delete(id string) error {
	if _, exists := s.products[id]; !exists {
		return core.NotFoundException("Product not found")
	}
	delete(s.products, id)
	return nil
}

// ProductController handles product-related requests
type ProductController struct {
	*ruumhttp.BaseController
	productService *ProductService
}

func NewProductController(productService *ProductService) *ProductController {
	ctrl := &ProductController{
		BaseController: ruumhttp.NewBaseController("/products"),
		productService: productService,
	}

	// Add controller-level guards (authentication required for all routes)
	ctrl.UseGuards(guards.NewAuthGuard())

	// Register routes
	ctrl.Get("", ctrl.FindAll)
	ctrl.Get("/{id}", ctrl.FindOne)
	ctrl.Post("", ctrl.Create)
	ctrl.Put("/{id}", ctrl.Update)
	ctrl.Delete("/{id}", ctrl.Delete)

	return ctrl
}

func (c *ProductController) FindAll(ctx core.Context) error {
	products := c.productService.FindAll()
	return ctx.JSON(200, map[string]interface{}{
		"data": products,
	})
}

func (c *ProductController) FindOne(ctx core.Context) error {
	id := ctx.Param("id")

	product, err := c.productService.FindOne(id)
	if err != nil {
		return err
	}

	return ctx.JSON(200, map[string]interface{}{
		"data": product,
	})
}

func (c *ProductController) Create(ctx core.Context) error {
	var product Product
	if err := ctx.Body(&product); err != nil {
		return core.BadRequestException("Invalid request body")
	}

	created, err := c.productService.Create(&product)
	if err != nil {
		return err
	}

	return ctx.JSON(201, map[string]interface{}{
		"message": "Product created successfully",
		"data":    created,
	})
}

func (c *ProductController) Update(ctx core.Context) error {
	id := ctx.Param("id")

	var product Product
	if err := ctx.Body(&product); err != nil {
		return core.BadRequestException("Invalid request body")
	}

	updated, err := c.productService.Update(id, &product)
	if err != nil {
		return err
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Product updated successfully",
		"data":    updated,
	})
}

func (c *ProductController) Delete(ctx core.Context) error {
	id := ctx.Param("id")

	if err := c.productService.Delete(id); err != nil {
		return err
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Product deleted successfully",
	})
}

// PublicController handles public endpoints (no authentication)
type PublicController struct {
	*ruumhttp.BaseController
}

func NewPublicController() *PublicController {
	ctrl := &PublicController{
		BaseController: ruumhttp.NewBaseController("/public"),
	}

	ctrl.Get("/info", ctrl.Info)

	return ctrl
}

func (c *PublicController) Info(ctx core.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"app":     "Ruum Advanced Example",
		"version": "1.0.0",
		"status":  "running",
	})
}

func main() {
	// Create logger
	ruumLogger := logger.NewDefaultLogger()

	// Create services
	productService := NewProductService()

	// Create module
	appModule := core.NewModuleBuilder().
		Controllers(
			NewProductController(productService),
			NewPublicController(),
		).
		Provider("productService", func() *ProductService {
			return productService
		}, core.ScopeSingleton, true).
		Build()

	// Create application
	app, err := factory.CreateApplication(
		appModule,
		factory.WithLogger(ruumLogger),
		factory.WithGlobalPrefix("/api/v1"),
		factory.WithCORS("http://localhost:3000", "http://localhost:4200"),
		factory.WithShutdownHooks(true),
	)

	if err != nil {
		log.Fatal("Failed to create application:", err)
	}

	// Use global middleware
	app.Use(middleware.Recovery(ruumLogger))
	app.Use(middleware.Logger(ruumLogger))

	// Use global interceptors
	app.UseGlobalInterceptors(interceptors.NewLoggingInterceptor(ruumLogger))

	// Use global exception filters
	app.UseGlobalFilters(core.NewDefaultExceptionFilter())

	// Start server
	ruumLogger.Info("🚀 Starting advanced Ruum application...", map[string]interface{}{
		"port": 3000,
		"env":  "development",
	})

	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
