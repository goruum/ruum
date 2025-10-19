# Advanced Ruum Example

Demonstrates advanced Ruum features including guards, interceptors, dependency injection, and service layer architecture.

## Features

- Service layer with business logic
- Dependency injection
- Authentication guard
- Logging interceptor
- Global exception filters
- Protected and public routes
- CORS configuration
- Global prefix

## Running the Example

```bash
cd examples/advanced
go run main.go
```

The server will start on `http://localhost:3000`

## API Endpoints

### Products (Protected)

All product endpoints require authentication:

- `GET /api/v1/products` - Get all products
- `GET /api/v1/products/{id}` - Get product by ID
- `POST /api/v1/products` - Create a new product
- `PUT /api/v1/products/{id}` - Update product
- `DELETE /api/v1/products/{id}` - Delete product

### Public

- `GET /api/v1/public/info` - Public app information

## Authentication

Products endpoints are protected by an authentication guard. You need to include an Authorization header:

```
Authorization: Bearer your-token-here
```

## Example Requests

### Get public info (no auth required)

```bash
curl http://localhost:3000/api/v1/public/info
```

Response:
```json
{
  "app": "Ruum Advanced Example",
  "version": "1.0.0",
  "status": "running"
}
```

### Get all products (auth required)

```bash
curl http://localhost:3000/api/v1/products \
  -H "Authorization: Bearer my-secret-token"
```

Response:
```json
{
  "data": [
    {
      "id": "1",
      "name": "Laptop",
      "description": "High-end laptop",
      "price": 1299.99
    },
    {
      "id": "2",
      "name": "Mouse",
      "description": "Wireless mouse",
      "price": 29.99
    }
  ]
}
```

### Get product by ID

```bash
curl http://localhost:3000/api/v1/products/1 \
  -H "Authorization: Bearer my-secret-token"
```

### Create a new product

```bash
curl -X POST http://localhost:3000/api/v1/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-token" \
  -d '{
    "id": "3",
    "name": "Keyboard",
    "description": "Mechanical keyboard",
    "price": 89.99
  }'
```

### Update a product

```bash
curl -X PUT http://localhost:3000/api/v1/products/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-token" \
  -d '{
    "name": "Laptop Pro",
    "description": "High-end laptop updated",
    "price": 1499.99
  }'
```

### Delete a product

```bash
curl -X DELETE http://localhost:3000/api/v1/products/1 \
  -H "Authorization: Bearer my-secret-token"
```

### Without authentication (returns 401)

```bash
curl http://localhost:3000/api/v1/products
```

Response:
```json
{
  "statusCode": 401,
  "message": "Missing authorization header",
  "error": "Unauthorized"
}
```

## Code Structure

```
advanced/
├── main.go           # Application entry point
└── README.md         # This file
```

The example demonstrates:
- Service layer for business logic
- Dependency injection of services
- Controller-level guards (authentication)
- Logging interceptors
- Global exception filters
- Protected vs public routes
- Global route prefix
- CORS with specific origins
- Error responses

## Key Concepts

### Service Layer

The `ProductService` encapsulates business logic:
- Data management
- Error handling
- Business rules

### Dependency Injection

Services are registered as providers and injected into controllers:

```go
module.Provider("productService", func() *ProductService {
    return productService
}, core.ScopeSingleton, true)
```

### Guards

The `AuthGuard` protects routes:

```go
ctrl.UseGuards(guards.NewAuthGuard())
```

### Interceptors

Logging interceptor tracks all requests:

```go
app.UseGlobalInterceptors(interceptors.NewLoggingInterceptor(logger))
```

### Exception Filters

Global filter handles all errors consistently:

```go
app.UseGlobalFilters(core.NewDefaultExceptionFilter())
```

