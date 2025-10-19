# Basic Ruum Example

A simple REST API demonstrating the core features of Ruum.

## Features

- Basic CRUD operations
- Controller with multiple routes
- JSON request/response handling
- Error handling
- Middleware (CORS, Logger, Recovery)
- Health check endpoint

## Running the Example

```bash
cd examples/basic
go run main.go
```

The server will start on `http://localhost:3000`

## API Endpoints

### Users

- `GET /users` - Get all users
- `GET /users/{id}` - Get user by ID
- `POST /users` - Create a new user
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user

### Health

- `GET /health` - Health check

## Example Requests

### Get all users

```bash
curl http://localhost:3000/users
```

Response:
```json
{
  "data": [
    {
      "id": "1",
      "name": "John Doe",
      "email": "john@example.com"
    },
    {
      "id": "2",
      "name": "Jane Smith",
      "email": "jane@example.com"
    }
  ]
}
```

### Get user by ID

```bash
curl http://localhost:3000/users/1
```

### Create a new user

```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bob Johnson",
    "email": "bob@example.com"
  }'
```

### Update a user

```bash
curl -X PUT http://localhost:3000/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated",
    "email": "john.updated@example.com"
  }'
```

### Delete a user

```bash
curl -X DELETE http://localhost:3000/users/1
```

### Health check

```bash
curl http://localhost:3000/health
```

## Code Structure

```
basic/
├── main.go           # Application entry point
└── README.md         # This file
```

The example demonstrates:
- Creating controllers with route handlers
- Building a module with controllers
- Creating an application with factory
- Using middleware
- Handling HTTP methods (GET, POST, PUT, DELETE)
- Path parameters and request body parsing
- JSON responses

