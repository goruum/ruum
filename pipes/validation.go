// Package pipes provides data transformation and validation pipes.
package pipes

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/goruum/ruum/core"
)

const (
	// ErrValueRequired is the error message for required validation
	ErrValueRequired = "value is required"
)

var (
	// ErrRequired is the error returned when a value is required
	ErrRequired = errors.New(ErrValueRequired)
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

// Transform validates the value using the configured validators.
func (p *ValidationPipe) Transform(value interface{}, _ *core.ArgumentMetadata) (interface{}, error) {
	for _, validator := range p.validators {
		if err := validator(value); err != nil {
			return nil, core.BadRequestException(err.Error())
		}
	}
	return value, nil
}

// ParseIntPipe parses a string to int
type ParseIntPipe struct{}

// NewParseIntPipe creates a new integer parsing pipe.
func NewParseIntPipe() *ParseIntPipe {
	return &ParseIntPipe{}
}

// Transform parses a string value to an integer.
func (p *ParseIntPipe) Transform(value interface{}, _ *core.ArgumentMetadata) (interface{}, error) {
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

// NewParseBoolPipe creates a new boolean parsing pipe.
func NewParseBoolPipe() *ParseBoolPipe {
	return &ParseBoolPipe{}
}

// Transform parses a string value to a boolean.
func (p *ParseBoolPipe) Transform(value interface{}, _ *core.ArgumentMetadata) (interface{}, error) {
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

// NewDefaultValuePipe creates a pipe that provides default values.
func NewDefaultValuePipe(defaultValue interface{}) *DefaultValuePipe {
	return &DefaultValuePipe{
		defaultValue: defaultValue,
	}
}

// Transform returns the default value if the input is nil.
func (p *DefaultValuePipe) Transform(value interface{}, _ *core.ArgumentMetadata) (interface{}, error) {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return p.defaultValue, nil
	}
	return value, nil
}

// Required validates that a value is not nil or empty.
func Required() Validator {
	return func(value interface{}) error {
		if value == nil {
			return ErrRequired
		}

		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return ErrRequired
		}

		if v.Kind() == reflect.String && v.Len() == 0 {
			return ErrRequired
		}

		return nil
	}
}

// MinLength validates that a string has at least minLen characters.
func MinLength(minLen int) Validator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value must be a string")
		}

		if len(str) < minLen {
			return fmt.Errorf("value must be at least %d characters", minLen)
		}

		return nil
	}
}

// MaxLength validates that a string has at most maxLen characters.
func MaxLength(maxLen int) Validator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value must be a string")
		}

		if len(str) > maxLen {
			return fmt.Errorf("value must be at most %d characters", maxLen)
		}

		return nil
	}
}

// Min validates that a number is at least minVal.
func Min(minVal float64) Validator {
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

		if num < minVal {
			return fmt.Errorf("value must be at least %f", minVal)
		}

		return nil
	}
}

// Max validates that a number is at most maxVal.
func Max(maxVal float64) Validator {
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

		if num > maxVal {
			return fmt.Errorf("value must be at most %f", maxVal)
		}

		return nil
	}
}
