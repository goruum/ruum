package core

// HandlerFunc is the function signature for route handlers
type HandlerFunc func(Context) error

// MiddlewareFunc is the function signature for middleware
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Guard checks if a request should be allowed to proceed
type Guard interface {
	CanActivate(ctx Context) (bool, error)
}

// Interceptor can intercept and modify requests/responses
type Interceptor interface {
	Intercept(ctx Context, next HandlerFunc) error
}

// Pipe transforms and validates input data
type Pipe interface {
	Transform(value interface{}, metadata *ArgumentMetadata) (interface{}, error)
}

// ArgumentMetadata contains metadata about an argument
type ArgumentMetadata struct {
	Type     string
	Metatype interface{}
	Data     string
}

// ExceptionFilter handles exceptions
type ExceptionFilter interface {
	Catch(err error, ctx Context) error
}

// Logger provides logging capabilities
type Logger interface {
	Log(level string, message string, fields map[string]interface{})
	Debug(message string, fields map[string]interface{})
	Info(message string, fields map[string]interface{})
	Warn(message string, fields map[string]interface{})
	Error(message string, fields map[string]interface{})
}

// ConfigService provides configuration management
type ConfigService interface {
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	Set(key string, value interface{})
	Has(key string) bool
}

// Module represents a feature module
type Module interface {
	Configure(container Container) error
	GetControllers() []interface{}
	GetProviders() []interface{}
	GetImports() []Module
	GetExports() []string
}

// OnModuleInit is called when a module is initialized
type OnModuleInit interface {
	OnModuleInit() error
}

// OnModuleDestroy is called when a module is destroyed
type OnModuleDestroy interface {
	OnModuleDestroy() error
}

// OnApplicationBootstrap is called when the application starts
type OnApplicationBootstrap interface {
	OnApplicationBootstrap() error
}

// OnApplicationShutdown is called when the application shuts down
type OnApplicationShutdown interface {
	OnApplicationShutdown() error
}

