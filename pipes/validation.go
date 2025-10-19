package pipes

import (
	"fmt"
	"reflect"

	"github.com/goruum/ruum/core"
)

// ValidationPipe validates and transforms input data
type ValidationPipe struct {
	validators []Validator
}

// Validator is a function that validates a value
type Validator func(value interface{}) error

// NewValidationPipe creates a new validation pipe
func NewValidationPipe(validators ...Validator) *ValidationPipe {
	return &ValidationPipe{
		validators: validators,
	}
}

func (p *ValidationPipe) Transform(value interface{}, metadata *core.ArgumentMetadata) (interface{}, error) {
	for _, validator := range p.validators {
		if err := validator(value); err != nil {
			return nil, core.BadRequestException(err.Error())
		}
	}
	return value, nil
}

// ParseIntPipe parses a string to int
type ParseIntPipe struct{}

func NewParseIntPipe() *ParseIntPipe {
	return &ParseIntPipe{}
}

func (p *ParseIntPipe) Transform(value interface{}, metadata *core.ArgumentMetadata) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, core.BadRequestException("Value must be a string")
	}
	
	var result int
	_, err := fmt.Sscanf(str, "%d", &result)
	if err != nil {
		return nil, core.BadRequestException("Invalid integer value")
	}
	
	return result, nil
}

// ParseBoolPipe parses a string to bool
type ParseBoolPipe struct{}

func NewParseBoolPipe() *ParseBoolPipe {
	return &ParseBoolPipe{}
}

func (p *ParseBoolPipe) Transform(value interface{}, metadata *core.ArgumentMetadata) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, core.BadRequestException("Value must be a string")
	}
	
	switch str {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return nil, core.BadRequestException("Invalid boolean value")
	}
}

// DefaultValuePipe provides a default value if the input is nil
type DefaultValuePipe struct {
	defaultValue interface{}
}

func NewDefaultValuePipe(defaultValue interface{}) *DefaultValuePipe {
	return &DefaultValuePipe{
		defaultValue: defaultValue,
	}
}

func (p *DefaultValuePipe) Transform(value interface{}, metadata *core.ArgumentMetadata) (interface{}, error) {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return p.defaultValue, nil
	}
	return value, nil
}

// Common validators
func Required() Validator {
	return func(value interface{}) error {
		if value == nil {
			return fmt.Errorf("value is required")
		}
		
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return fmt.Errorf("value is required")
		}
		
		if v.Kind() == reflect.String && v.Len() == 0 {
			return fmt.Errorf("value is required")
		}
		
		return nil
	}
}

func MinLength(min int) Validator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value must be a string")
		}
		
		if len(str) < min {
			return fmt.Errorf("value must be at least %d characters", min)
		}
		
		return nil
	}
}

func MaxLength(max int) Validator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value must be a string")
		}
		
		if len(str) > max {
			return fmt.Errorf("value must be at most %d characters", max)
		}
		
		return nil
	}
}

func Min(min float64) Validator {
	return func(value interface{}) error {
		var num float64
		
		switch v := value.(type) {
		case int:
			num = float64(v)
		case float64:
			num = v
		default:
			return fmt.Errorf("value must be a number")
		}
		
		if num < min {
			return fmt.Errorf("value must be at least %f", min)
		}
		
		return nil
	}
}

func Max(max float64) Validator {
	return func(value interface{}) error {
		var num float64
		
		switch v := value.(type) {
		case int:
			num = float64(v)
		case float64:
			num = v
		default:
			return fmt.Errorf("value must be a number")
		}
		
		if num > max {
			return fmt.Errorf("value must be at most %f", max)
		}
		
		return nil
	}
}

