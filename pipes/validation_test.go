package pipes

import (
	"testing"

	"github.com/goruum/ruum/core"
)

func TestNewValidationPipe(t *testing.T) {
	validator := func(value interface{}) error {
		return nil
	}

	pipe := NewValidationPipe(validator)
	if pipe == nil {
		t.Fatal("NewValidationPipe() returned nil")
	}

	if len(pipe.validators) != 1 {
		t.Errorf("validators length = %d, want 1", len(pipe.validators))
	}
}

func TestValidationPipe_Transform_Success(t *testing.T) {
	validator := func(value interface{}) error {
		return nil
	}

	pipe := NewValidationPipe(validator)
	result, err := pipe.Transform("test-value", nil)

	if err != nil {
		t.Errorf("Transform() error = %v", err)
	}

	if result != "test-value" {
		t.Errorf("Transform() = %v, want 'test-value'", result)
	}
}

func TestValidationPipe_Transform_ValidationError(t *testing.T) {
	validator := func(value interface{}) error {
		return core.BadRequestException("validation failed")
	}

	pipe := NewValidationPipe(validator)
	_, err := pipe.Transform("test-value", nil)

	if err == nil {
		t.Error("Expected validation error, got nil")
	}
}

func TestNewParseIntPipe(t *testing.T) {
	pipe := NewParseIntPipe()
	if pipe == nil {
		t.Fatal("NewParseIntPipe() returned nil")
	}
}

func TestParseIntPipe_Transform_Success(t *testing.T) {
	pipe := NewParseIntPipe()

	tests := []struct {
		input    string
		expected int
	}{
		{"42", 42},
		{"0", 0},
		{"-10", -10},
		{"999", 999},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := pipe.Transform(tt.input, nil)
			if err != nil {
				t.Errorf("Transform() error = %v", err)
			}

			intResult, ok := result.(int)
			if !ok {
				t.Error("Transform() did not return int")
			}

			if intResult != tt.expected {
				t.Errorf("Transform() = %v, want %v", intResult, tt.expected)
			}
		})
	}
}

func TestParseIntPipe_Transform_InvalidString(t *testing.T) {
	pipe := NewParseIntPipe()

	_, err := pipe.Transform("not-a-number", nil)
	if err == nil {
		t.Error("Expected error for invalid int string, got nil")
	}
}

func TestParseIntPipe_Transform_NonStringInput(t *testing.T) {
	pipe := NewParseIntPipe()

	_, err := pipe.Transform(42, nil)
	if err == nil {
		t.Error("Expected error for non-string input, got nil")
	}
}

func TestNewParseBoolPipe(t *testing.T) {
	pipe := NewParseBoolPipe()
	if pipe == nil {
		t.Fatal("NewParseBoolPipe() returned nil")
	}
}

func TestParseBoolPipe_Transform_TrueValues(t *testing.T) {
	pipe := NewParseBoolPipe()

	trueValues := []string{"true", "1", "yes"}

	for _, val := range trueValues {
		t.Run(val, func(t *testing.T) {
			result, err := pipe.Transform(val, nil)
			if err != nil {
				t.Errorf("Transform() error = %v", err)
			}

			boolResult, ok := result.(bool)
			if !ok {
				t.Error("Transform() did not return bool")
			}

			if !boolResult {
				t.Errorf("Transform('%s') = false, want true", val)
			}
		})
	}
}

func TestParseBoolPipe_Transform_FalseValues(t *testing.T) {
	pipe := NewParseBoolPipe()

	falseValues := []string{"false", "0", "no"}

	for _, val := range falseValues {
		t.Run(val, func(t *testing.T) {
			result, err := pipe.Transform(val, nil)
			if err != nil {
				t.Errorf("Transform() error = %v", err)
			}

			boolResult, ok := result.(bool)
			if !ok {
				t.Error("Transform() did not return bool")
			}

			if boolResult {
				t.Errorf("Transform('%s') = true, want false", val)
			}
		})
	}
}

func TestParseBoolPipe_Transform_InvalidValue(t *testing.T) {
	pipe := NewParseBoolPipe()

	_, err := pipe.Transform("invalid", nil)
	if err == nil {
		t.Error("Expected error for invalid bool string, got nil")
	}
}

func TestParseBoolPipe_Transform_NonStringInput(t *testing.T) {
	pipe := NewParseBoolPipe()

	_, err := pipe.Transform(true, nil)
	if err == nil {
		t.Error("Expected error for non-string input, got nil")
	}
}

func TestNewDefaultValuePipe(t *testing.T) {
	defaultValue := "default"
	pipe := NewDefaultValuePipe(defaultValue)

	if pipe == nil {
		t.Fatal("NewDefaultValuePipe() returned nil")
	}

	if pipe.defaultValue != defaultValue {
		t.Error("defaultValue not set correctly")
	}
}

func TestDefaultValuePipe_Transform_NilValue(t *testing.T) {
	pipe := NewDefaultValuePipe("default-value")

	result, err := pipe.Transform(nil, nil)
	if err != nil {
		t.Errorf("Transform() error = %v", err)
	}

	if result != "default-value" {
		t.Errorf("Transform() = %v, want 'default-value'", result)
	}
}

func TestDefaultValuePipe_Transform_NonNilValue(t *testing.T) {
	pipe := NewDefaultValuePipe("default-value")

	result, err := pipe.Transform("actual-value", nil)
	if err != nil {
		t.Errorf("Transform() error = %v", err)
	}

	if result != "actual-value" {
		t.Errorf("Transform() = %v, want 'actual-value'", result)
	}
}

func TestRequired(t *testing.T) {
	validator := Required()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"non-nil string", "test", false},
		{"nil value", nil, true},
		{"empty string", "", true},
		{"non-empty slice", []string{"a"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Required() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	validator := MinLength(5)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"exact length", "12345", false},
		{"longer", "123456", false},
		{"shorter", "1234", true},
		{"non-string", 12345, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("MinLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	validator := MaxLength(5)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"exact length", "12345", false},
		{"shorter", "1234", false},
		{"longer", "123456", true},
		{"non-string", 12345, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("MaxLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMin(t *testing.T) {
	validator := Min(10.0)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"exact value", 10, false},
		{"greater int", 15, false},
		{"greater float", 15.5, false},
		{"less int", 5, true},
		{"less float", 5.5, true},
		{"non-number", "10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Min() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMax(t *testing.T) {
	validator := Max(10.0)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"exact value", 10, false},
		{"less int", 5, false},
		{"less float", 5.5, false},
		{"greater int", 15, true},
		{"greater float", 15.5, true},
		{"non-number", "10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Max() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationPipe_MultipleValidators(t *testing.T) {
	pipe := NewValidationPipe(
		Required(),
		MinLength(3),
		MaxLength(10),
	)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid", "hello", false},
		{"too short", "hi", true},
		{"too long", "this is too long", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pipe.Transform(tt.value, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transform() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
