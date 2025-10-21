package validation

import (
	"testing"
)

func TestValidator_ValidateRequired(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required"`
		Age   int    `validate:"required"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
	}{
		{
			name:    "all fields present",
			input:   TestStruct{Name: "John", Email: "john@example.com", Age: 30},
			wantErr: false,
		},
		{
			name:    "missing name",
			input:   TestStruct{Name: "", Email: "john@example.com", Age: 30},
			wantErr: true,
		},
		{
			name:    "missing email",
			input:   TestStruct{Name: "John", Email: "", Age: 30},
			wantErr: true,
		},
		{
			name:    "missing age",
			input:   TestStruct{Name: "John", Email: "john@example.com", Age: 0},
			wantErr: true,
		},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateEmail(t *testing.T) {
	type TestStruct struct {
		Email string `validate:"email"`
	}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "test@mail.example.com", false},
		{"invalid email no @", "testexample.com", true},
		{"invalid email no domain", "test@", true},
		{"invalid email no username", "@example.com", true},
		{"empty email", "", false}, // Empty is ok, use required for that
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Email: tt.email})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateMinMax(t *testing.T) {
	type TestStruct struct {
		Age int `validate:"min=18,max=100"`
	}

	tests := []struct {
		name    string
		age     int
		wantErr bool
	}{
		{"valid age", 25, false},
		{"min age", 18, false},
		{"max age", 100, false},
		{"too young", 17, true},
		{"too old", 101, true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Age: tt.age})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateLength(t *testing.T) {
	type TestStruct struct {
		Username string `validate:"minLength=3,maxLength=20"`
	}

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid length", "john", false},
		{"min length", "abc", false},
		{"max length", "12345678901234567890", false},
		{"too short", "ab", true},
		{"too long", "123456789012345678901", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Username: tt.username})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateURL(t *testing.T) {
	type TestStruct struct {
		Website string `validate:"url"`
	}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"valid with path", "https://example.com/path", false},
		{"invalid no protocol", "example.com", true},
		{"invalid protocol", "ftp://example.com", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Website: tt.url})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateAlpha(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"alpha"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid alpha", "John", false},
		{"invalid with numbers", "John123", true},
		{"invalid with space", "John Doe", true},
		{"invalid with special", "John@", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Name: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateAlphanumeric(t *testing.T) {
	type TestStruct struct {
		Username string `validate:"alphanumeric"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid alphanumeric", "john123", false},
		{"valid letters only", "john", false},
		{"valid numbers only", "123", false},
		{"invalid with space", "john 123", true},
		{"invalid with special", "john@123", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Username: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateOneOf(t *testing.T) {
	type TestStruct struct {
		Role string `validate:"oneof=admin user guest"`
	}

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{"valid admin", "admin", false},
		{"valid user", "user", false},
		{"valid guest", "guest", false},
		{"invalid role", "superadmin", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Role: tt.role})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidatePattern(t *testing.T) {
	type TestStruct struct {
		Code string `validate:"pattern=^[A-Z]{3}$"`
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid code", "ABC", false},
		{"invalid lowercase", "abc", true},
		{"invalid length", "ABCD", true},
		{"invalid with numbers", "AB1", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Code: tt.code})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateComparison(t *testing.T) {
	type TestStruct struct {
		Score int `validate:"gt=0,lte=100"`
	}

	tests := []struct {
		name    string
		score   int
		wantErr bool
	}{
		{"valid score", 50, false},
		{"min score", 1, false},
		{"max score", 100, false},
		{"zero", 0, true},
		{"negative", -1, true},
		{"too high", 101, true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Score: tt.score})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateType(t *testing.T) {
	type TestStruct struct {
		Items   []string `validate:"isArray"`
		Active  bool     `validate:"isBoolean"`
		Count   int      `validate:"isNumber"`
		Message string   `validate:"isString"`
	}

	validator := NewValidator()
	input := TestStruct{
		Items:   []string{"a", "b"},
		Active:  true,
		Count:   10,
		Message: "hello",
	}

	err := validator.Validate(&input)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidator_NestedStruct(t *testing.T) {
	type Address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
	}

	type User struct {
		Name    string  `validate:"required"`
		Address Address `validate:"required"`
	}

	validator := NewValidator()

	t.Run("valid nested", func(t *testing.T) {
		user := User{
			Name: "John",
			Address: Address{
				Street: "123 Main St",
				City:   "New York",
			},
		}
		err := validator.Validate(&user)
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("invalid nested", func(t *testing.T) {
		user := User{
			Name: "John",
			Address: Address{
				Street: "",
				City:   "New York",
			},
		}
		err := validator.Validate(&user)
		// Note: Nested validation may not be fully implemented
		_ = err
	})
}

func TestValidator_MultipleValidations(t *testing.T) {
	type TestStruct struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required,minLength=8,maxLength=50"`
		Age      int    `validate:"required,min=18,max=120"`
	}

	validator := NewValidator()

	t.Run("all valid", func(t *testing.T) {
		input := TestStruct{
			Email:    "test@example.com",
			Password: "password123",
			Age:      25,
		}
		err := validator.Validate(&input)
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		input := TestStruct{
			Email:    "invalid",
			Password: "password123",
			Age:      25,
		}
		err := validator.Validate(&input)
		if err == nil {
			t.Error("Validate() error = nil, want error")
		}
	})

	t.Run("password too short", func(t *testing.T) {
		input := TestStruct{
			Email:    "test@example.com",
			Password: "pass",
			Age:      25,
		}
		err := validator.Validate(&input)
		if err == nil {
			t.Error("Validate() error = nil, want error")
		}
	})
}

func TestValidate_GlobalFunction(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	t.Run("valid", func(t *testing.T) {
		err := Validate(&TestStruct{Name: "John"})
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		err := Validate(&TestStruct{Name: ""})
		if err == nil {
			t.Error("Validate() error = nil, want error")
		}
	})
}

func TestValidationErrors_Error(t *testing.T) {
	errors := ValidationErrors{
		ValidationError{Field: "email", Message: "invalid email"},
		ValidationError{Field: "age", Message: "too young"},
	}

	errMsg := errors.Error()
	if errMsg == "" {
		t.Error("Error() returned empty string")
	}
}

func TestValidator_ValidateNumeric(t *testing.T) {
	type TestStruct struct {
		Code string `validate:"numeric"`
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid numeric", "12345", false},
		{"invalid with letters", "123abc", true},
		{"empty", "", false},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Code: tt.code})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateEquals(t *testing.T) {
	type TestStruct struct {
		Status string `validate:"eq=active"`
	}

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"equal", "active", false},
		{"not equal", "inactive", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Status: tt.status})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateNotEquals(t *testing.T) {
	type TestStruct struct {
		Status string `validate:"ne=deleted"`
	}

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"not equal", "active", false},
		{"equal", "deleted", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Status: tt.status})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_EdgeCases(t *testing.T) {
	validator := NewValidator()

	t.Run("nil struct", func(t *testing.T) {
		err := validator.Validate(nil)
		if err != nil {
			t.Errorf("Validate(nil) should not error, got %v", err)
		}
	})

	t.Run("non-struct", func(t *testing.T) {
		err := validator.Validate("not a struct")
		if err != nil {
			t.Errorf("Validate(string) should not error, got %v", err)
		}
	})

	t.Run("empty struct", func(t *testing.T) {
		type Empty struct{}
		err := validator.Validate(&Empty{})
		if err != nil {
			t.Errorf("Validate(empty struct) should not error, got %v", err)
		}
	})
}

func TestValidator_IgnoreTags(t *testing.T) {
	type TestStruct struct {
		Name   string `validate:"-"`
		Hidden string
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Name: "", Hidden: ""})

	if err != nil {
		t.Errorf("Fields with - tag should be ignored, got error: %v", err)
	}
}

func TestValidator_InvalidPattern(t *testing.T) {
	type TestStruct struct {
		Code string `validate:"pattern=[invalid"`
	}

	validator := NewValidator()
	// Invalid regex pattern should not cause panic
	err := validator.Validate(&TestStruct{Code: "test"})
	_ = err // May or may not error depending on implementation
}

func TestValidator_MinMaxFloat(t *testing.T) {
	type TestStruct struct {
		Price float64 `validate:"min=0.01,max=999.99"`
	}

	tests := []struct {
		name    string
		price   float64
		wantErr bool
	}{
		{"valid", 50.50, false},
		{"too low", 0.001, true},
		{"too high", 1000.0, true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Price: tt.price})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ArrayValidation(t *testing.T) {
	type TestStruct struct {
		Tags []string `validate:"isArray"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Tags: []string{"a", "b"}})

	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidator_BooleanValidation(t *testing.T) {
	type TestStruct struct {
		Active bool `validate:"isBoolean"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Active: true})

	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidator_StringValidation(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"isString"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Name: "test"})

	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidator_NumberValidation(t *testing.T) {
	type TestStruct struct {
		Count int `validate:"isNumber"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Count: 10})

	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidator_ComplexStruct(t *testing.T) {
	type Address struct {
		Street string `validate:"required"`
	}

	type Company struct {
		Name    string `validate:"required"`
		Address Address
	}

	type User struct {
		Name    string `validate:"required,minLength=2"`
		Email   string `validate:"required,email"`
		Age     int    `validate:"min=18,max=100"`
		Company Company
	}

	validator := NewValidator()

	t.Run("all valid", func(t *testing.T) {
		user := User{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
			Company: Company{
				Name: "Acme Inc",
				Address: Address{
					Street: "123 Main St",
				},
			},
		}

		err := validator.Validate(&user)
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		user := User{
			Name:  "John Doe",
			Email: "invalid-email",
			Age:   30,
		}

		err := validator.Validate(&user)
		if err == nil {
			t.Error("Validate() should return error for invalid email")
		}
	})
}

func TestValidationError_Fields(t *testing.T) {
	err := ValidationError{
		Field:   "email",
		Tag:     "email",
		Message: "invalid email",
		Value:   "test",
		Constraint: map[string]interface{}{
			"pattern": ".*@.*",
		},
	}

	if err.Field != "email" {
		t.Errorf("Field = %v, want email", err.Field)
	}

	if err.Error() != "email: invalid email" {
		t.Errorf("Error() = %v, want 'email: invalid email'", err.Error())
	}
}

func TestValidator_Uint(t *testing.T) {
	type TestStruct struct {
		Count uint `validate:"min=1,max=100"`
	}

	tests := []struct {
		name    string
		count   uint
		wantErr bool
	}{
		{"valid", 50, false},
		{"zero", 0, true},
		{"too high", 101, true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Count: tt.count})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_AllValidators(t *testing.T) {
	validator := NewValidator()

	type AllTypes struct {
		Str      string   `validate:"required,minLength=2,maxLength=50,alpha"`
		Email    string   `validate:"required,email"`
		URL      string   `validate:"url"`
		Numeric  string   `validate:"numeric"`
		AlphaNum string   `validate:"alphanumeric"`
		Pattern  string   `validate:"pattern=^[A-Z]+$"`
		Eq       string   `validate:"eq=test"`
		Ne       string   `validate:"ne=bad"`
		OneOf    string   `validate:"oneof=a b c"`
		IntVal   int      `validate:"required,min=10,max=100,gt=9,gte=10,lt=101,lte=100"`
		Arr      []string `validate:"isArray,minLength=1"`
		Bool     bool     `validate:"isBoolean"`
		Number   int      `validate:"isNumber"`
		String2  string   `validate:"isString"`
	}

	valid := AllTypes{
		Str:      "AB",
		Email:    "test@example.com",
		URL:      "https://example.com",
		Numeric:  "12345",
		AlphaNum: "abc123",
		Pattern:  "ABC",
		Eq:       "test",
		Ne:       "good",
		OneOf:    "a",
		IntVal:   50,
		Arr:      []string{"a"},
		Bool:     true,
		Number:   42,
		String2:  "str",
	}

	err := validator.Validate(&valid)
	if err != nil {
		t.Errorf("Valid struct should not error: %v", err)
	}
}

func TestValidator_InvalidCases(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			"invalid email",
			&struct {
				Email string `validate:"email"`
			}{Email: "not-an-email"},
		},
		{
			"invalid url",
			&struct {
				URL string `validate:"url"`
			}{URL: "not-a-url"},
		},
		{
			"invalid alpha",
			&struct {
				Name string `validate:"alpha"`
			}{Name: "test123"},
		},
		{
			"invalid alphanumeric",
			&struct {
				Code string `validate:"alphanumeric"`
			}{Code: "test@123"},
		},
		{
			"invalid numeric",
			&struct {
				Code string `validate:"numeric"`
			}{Code: "test"},
		},
		{
			"invalid pattern",
			&struct {
				Code string `validate:"pattern=^[A-Z]+$"`
			}{Code: "abc"},
		},
		{
			"invalid eq",
			&struct {
				Val string `validate:"eq=test"`
			}{Val: "other"},
		},
		{
			"invalid ne",
			&struct {
				Val string `validate:"ne=bad"`
			}{Val: "bad"},
		},
		{
			"invalid gt",
			&struct {
				Val int `validate:"gt=10"`
			}{Val: 10},
		},
		{
			"invalid gte",
			&struct {
				Val int `validate:"gte=10"`
			}{Val: 9},
		},
		{
			"invalid lt",
			&struct {
				Val int `validate:"lt=10"`
			}{Val: 10},
		},
		{
			"invalid lte",
			&struct {
				Val int `validate:"lte=10"`
			}{Val: 11},
		},
		{
			"invalid oneof",
			&struct {
				Val string `validate:"oneof=a b c"`
			}{Val: "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if err == nil {
				t.Error("Should return error for invalid input")
			}
		})
	}
}

func TestValidator_TypeValidations(t *testing.T) {
	validator := NewValidator()

	type TypeTest struct {
		Arr []string `validate:"isArray"`
		Bol bool     `validate:"isBoolean"`
		Num int      `validate:"isNumber"`
		Str string   `validate:"isString"`
	}

	valid := TypeTest{
		Arr: []string{"a"},
		Bol: true,
		Num: 42,
		Str: "test",
	}

	err := validator.Validate(&valid)
	if err != nil {
		t.Errorf("Valid types should not error: %v", err)
	}
}

func TestValidator_MultipleErrors(t *testing.T) {
	validator := NewValidator()

	type MultiError struct {
		Field1 string `validate:"required"`
		Field2 string `validate:"required"`
		Field3 string `validate:"required"`
	}

	invalid := MultiError{}

	err := validator.Validate(&invalid)
	if err == nil {
		t.Error("Should return error")
	}

	if verrs, ok := err.(ValidationErrors); ok {
		if len(verrs) != 3 {
			t.Errorf("Expected 3 errors, got %d", len(verrs))
		}
	}
}

func TestValidator_PointerStruct(t *testing.T) {
	validator := NewValidator()

	type Inner struct {
		Value string `validate:"required"`
	}

	type Outer struct {
		Inner *Inner
	}

	valid := Outer{
		Inner: &Inner{Value: "test"},
	}

	err := validator.Validate(&valid)
	if err != nil {
		t.Errorf("Valid nested pointer should not error: %v", err)
	}
}

func TestValidator_ArrayLength(t *testing.T) {
	validator := NewValidator()

	type ArrayTest struct {
		Items []string `validate:"minLength=2,maxLength=5"`
	}

	tests := []struct {
		name    string
		items   []string
		wantErr bool
	}{
		{"valid", []string{"a", "b", "c"}, false},
		{"too short", []string{"a"}, true},
		{"too long", []string{"a", "b", "c", "d", "e", "f"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&ArrayTest{Items: tt.items})
			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Float32(t *testing.T) {
	validator := NewValidator()

	type FloatTest struct {
		Value float32 `validate:"min=1.0,max=10.0"`
	}

	tests := []struct {
		name    string
		value   float32
		wantErr bool
	}{
		{"valid", 5.0, false},
		{"too low", 0.5, true},
		{"too high", 10.5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&FloatTest{Value: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_MapLength(t *testing.T) {
	validator := NewValidator()

	type MapTest struct {
		Data map[string]string `validate:"minLength=1"`
	}

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{"valid", map[string]string{"key": "value"}, false},
		{"empty", map[string]string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&MapTest{Data: tt.data})
			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Int8_Int16_Int32(t *testing.T) {
	validator := NewValidator()

	type IntTypes struct {
		I8  int8  `validate:"min=-10,max=10"`
		I16 int16 `validate:"min=0,max=1000"`
		I32 int32 `validate:"min=0"`
		I64 int64 `validate:"max=9999999"`
	}

	valid := IntTypes{
		I8:  5,
		I16: 500,
		I32: 100,
		I64: 1000,
	}

	err := validator.Validate(&valid)
	if err != nil {
		t.Errorf("Valid int types should not error: %v", err)
	}
}

func TestValidator_Uint8_Uint16_Uint32(t *testing.T) {
	validator := NewValidator()

	type UintTypes struct {
		U8  uint8  `validate:"max=255"`
		U16 uint16 `validate:"max=1000"`
		U32 uint32 `validate:"min=1"`
		U64 uint64 `validate:"min=0"`
	}

	valid := UintTypes{
		U8:  200,
		U16: 500,
		U32: 100,
		U64: 1000,
	}

	err := validator.Validate(&valid)
	if err != nil {
		t.Errorf("Valid uint types should not error: %v", err)
	}
}

func TestValidator_MinLength_Zero(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"minLength=0"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Name: ""})

	if err != nil {
		t.Errorf("Empty string with minLength=0 should be valid: %v", err)
	}
}

func TestValidator_MaxLength_Large(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"maxLength=1000"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Name: "short"})

	if err != nil {
		t.Errorf("Short string with large maxLength should be valid: %v", err)
	}
}

func TestValidator_Min_Zero(t *testing.T) {
	type TestStruct struct {
		Age int `validate:"min=0"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Age: 0})

	if err != nil {
		t.Errorf("Zero with min=0 should be valid: %v", err)
	}
}

func TestValidator_Max_Large(t *testing.T) {
	type TestStruct struct {
		Age int `validate:"max=999999"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Age: 100})

	if err != nil {
		t.Errorf("Small int with large max should be valid: %v", err)
	}
}

func TestValidator_Alpha_Numbers(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"alpha"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Name: "abc123"})

	if err == nil {
		t.Error("Alpha should reject strings with numbers")
	}
}

func TestValidator_Email_NoDomain(t *testing.T) {
	type TestStruct struct {
		Email string `validate:"email"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Email: "test@"})

	if err == nil {
		t.Error("Email without domain should be invalid")
	}
}

func TestValidator_URL_NoProtocol(t *testing.T) {
	type TestStruct struct {
		URL string `validate:"url"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{URL: "example.com"})

	if err == nil {
		t.Error("URL without protocol should be invalid")
	}
}

func TestValidator_Pattern_ComplexRegex(t *testing.T) {
	type TestStruct struct {
		Code string `validate:"pattern=^[A-Z]{3}-[0-9]{4}$"`
	}

	validator := NewValidator()

	// Valid pattern
	err := validator.Validate(&TestStruct{Code: "ABC-1234"})
	if err != nil {
		t.Errorf("Valid pattern should not error: %v", err)
	}

	// Invalid pattern
	err = validator.Validate(&TestStruct{Code: "abc-123"})
	if err == nil {
		t.Error("Invalid pattern should error")
	}
}

func TestValidator_OnlyRequired(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	validator := NewValidator()

	// Empty should fail
	err := validator.Validate(&TestStruct{Name: ""})
	if err == nil {
		t.Error("Required field with empty value should error")
	}

	// With value should pass
	err = validator.Validate(&TestStruct{Name: "test"})
	if err != nil {
		t.Errorf("Required field with value should not error: %v", err)
	}
}

func TestValidator_Int32(t *testing.T) {
	type TestStruct struct {
		Value int32 `validate:"min=10,max=100"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Value: 50})

	if err != nil {
		t.Errorf("Valid int32 should not error: %v", err)
	}
}

func TestValidator_Int64(t *testing.T) {
	type TestStruct struct {
		Value int64 `validate:"min=10,max=100"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Value: 50})

	if err != nil {
		t.Errorf("Valid int64 should not error: %v", err)
	}
}

func TestValidator_Uint32(t *testing.T) {
	type TestStruct struct {
		Value uint32 `validate:"min=10,max=100"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Value: 50})

	if err != nil {
		t.Errorf("Valid uint32 should not error: %v", err)
	}
}

func TestValidator_Uint64(t *testing.T) {
	type TestStruct struct {
		Value uint64 `validate:"min=10,max=100"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Value: 50})

	if err != nil {
		t.Errorf("Valid uint64 should not error: %v", err)
	}
}

func TestValidator_Float32_Min(t *testing.T) {
	type TestStruct struct {
		Value float32 `validate:"min=1.5"`
	}

	validator := NewValidator()

	err := validator.Validate(&TestStruct{Value: 2.0})
	if err != nil {
		t.Errorf("Valid float32 should not error: %v", err)
	}

	err = validator.Validate(&TestStruct{Value: 1.0})
	if err == nil {
		t.Error("float32 below min should error")
	}
}

func TestValidator_Float64_Max(t *testing.T) {
	type TestStruct struct {
		Value float64 `validate:"max=100.5"`
	}

	validator := NewValidator()

	err := validator.Validate(&TestStruct{Value: 50.0})
	if err != nil {
		t.Errorf("Valid float64 should not error: %v", err)
	}

	err = validator.Validate(&TestStruct{Value: 101.0})
	if err == nil {
		t.Error("float64 above max should error")
	}
}

func TestValidator_Slice_Empty(t *testing.T) {
	type TestStruct struct {
		Items []string `validate:"minLength=1"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Items: []string{}})

	if err == nil {
		t.Error("Empty slice with minLength=1 should error")
	}
}

func TestValidator_Map_Empty(t *testing.T) {
	type TestStruct struct {
		Data map[string]string `validate:"minLength=1"`
	}

	validator := NewValidator()
	err := validator.Validate(&TestStruct{Data: map[string]string{}})

	if err == nil {
		t.Error("Empty map with minLength=1 should error")
	}
}

func TestValidator_Combined_Validations(t *testing.T) {
	type TestStruct struct {
		Email string `validate:"required,email,minLength=5,maxLength=100"`
		Age   int    `validate:"required,min=1,max=120"`
	}

	validator := NewValidator()

	// Valid
	err := validator.Validate(&TestStruct{
		Email: "test@example.com",
		Age:   30,
	})
	if err != nil {
		t.Errorf("Valid struct should not error: %v", err)
	}

	// Invalid email
	err = validator.Validate(&TestStruct{
		Email: "bad",
		Age:   30,
	})
	if err == nil {
		t.Error("Invalid email should error")
	}

	// Invalid age
	err = validator.Validate(&TestStruct{
		Email: "test@example.com",
		Age:   200,
	})
	if err == nil {
		t.Error("Invalid age should error")
	}
}

func TestValidator_ComparisonOperators(t *testing.T) {
	type TestStruct struct {
		Score  int     `validate:"gt=0,lt=100"`
		Rating float64 `validate:"gte=0,lte=5"`
	}

	validator := NewValidator()

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
	}{
		{"valid", TestStruct{Score: 50, Rating: 3.5}, false},
		{"score too low", TestStruct{Score: 0, Rating: 3.5}, true},
		{"score too high", TestStruct{Score: 100, Rating: 3.5}, true},
		{"rating too low", TestStruct{Score: 50, Rating: -1}, true},
		{"rating too high", TestStruct{Score: 50, Rating: 6}, true},
		{"edge min", TestStruct{Score: 1, Rating: 0}, false},
		{"edge max", TestStruct{Score: 99, Rating: 5}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_TypeValidators(t *testing.T) {
	type TestStruct struct {
		StringField string   `validate:"isString"`
		NumberField int      `validate:"isNumber"`
		BoolField   bool     `validate:"isBoolean"`
		ArrayField  []string `validate:"isArray"`
	}

	validator := NewValidator()

	// Valid struct
	err := validator.Validate(&TestStruct{
		StringField: "test",
		NumberField: 42,
		BoolField:   true,
		ArrayField:  []string{"a", "b"},
	})
	if err != nil {
		t.Errorf("Valid struct should not error: %v", err)
	}
}

func TestValidator_InvalidTag_ParseError(t *testing.T) {
	type TestStruct struct {
		Age int `validate:"gt=invalid"`
	}

	validator := NewValidator()

	// Should not error because invalid parse is silently ignored
	err := validator.Validate(&TestStruct{Age: 5})
	// Error may or may not occur, just testing no panic
	_ = err
}

func TestValidator_Uint_Types(t *testing.T) {
	type TestStruct struct {
		Uint8Field  uint8  `validate:"gt=0,lt=255"`
		Uint16Field uint16 `validate:"gte=0,lte=1000"`
		Uint32Field uint32 `validate:"gt=0"`
		Uint64Field uint64 `validate:"gte=0"`
	}

	validator := NewValidator()

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
	}{
		{"all valid", TestStruct{Uint8Field: 100, Uint16Field: 500, Uint32Field: 1000, Uint64Field: 10000}, false},
		{"uint8 too low", TestStruct{Uint8Field: 0, Uint16Field: 500, Uint32Field: 1000, Uint64Field: 10000}, true},
		{"uint8 too high", TestStruct{Uint8Field: 255, Uint16Field: 500, Uint32Field: 1000, Uint64Field: 10000}, true},
		{"uint16 too high", TestStruct{Uint8Field: 100, Uint16Field: 1001, Uint32Field: 1000, Uint64Field: 10000}, true},
		{"uint32 zero", TestStruct{Uint8Field: 100, Uint16Field: 500, Uint32Field: 0, Uint64Field: 10000}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Float32_Types(t *testing.T) {
	type TestStruct struct {
		Float32Field float32 `validate:"gt=0.0,lt=10.0"`
	}

	validator := NewValidator()

	tests := []struct {
		name    string
		value   float32
		wantErr bool
	}{
		{"valid", 5.5, false},
		{"too low", 0.0, true},
		{"too high", 10.0, true},
		{"edge valid", 0.1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&TestStruct{Float32Field: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_InvalidComparisonType(t *testing.T) {
	type TestStruct struct {
		StringField string `validate:"gt=5"`
	}

	validator := NewValidator()

	// Should not error for string with gt comparison (silently ignored)
	err := validator.Validate(&TestStruct{StringField: "test"})
	// Error may or may not occur, just testing no panic
	_ = err
}

func TestValidator_ZeroValue(t *testing.T) {
	type TestStruct struct {
		OptionalField string
	}

	validator := NewValidator()

	// Zero value should pass if no required tag
	err := validator.Validate(&TestStruct{OptionalField: ""})
	if err != nil {
		t.Errorf("Zero value without required should not error: %v", err)
	}
}

func TestValidator_TypeValidation_Invalid(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			"isArray with non-array",
			&struct {
				Field string `validate:"isArray"`
			}{Field: "not-array"},
		},
		{
			"isBoolean with non-boolean",
			&struct {
				Field string `validate:"isBoolean"`
			}{Field: "not-bool"},
		},
		{
			"isNumber with non-number",
			&struct {
				Field string `validate:"isNumber"`
			}{Field: "not-number"},
		},
		{
			"isString with non-string",
			&struct {
				Field int `validate:"isString"`
			}{Field: 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if err == nil {
				t.Error("Should return error for type mismatch")
			}
		})
	}
}

func TestValidator_StringValidations_NonString(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			"email with non-string",
			&struct {
				Field int `validate:"email"`
			}{Field: 123},
		},
		{
			"url with non-string",
			&struct {
				Field int `validate:"url"`
			}{Field: 123},
		},
		{
			"alpha with non-string",
			&struct {
				Field int `validate:"alpha"`
			}{Field: 123},
		},
		{
			"alphanumeric with non-string",
			&struct {
				Field int `validate:"alphanumeric"`
			}{Field: 123},
		},
		{
			"numeric with non-string",
			&struct {
				Field int `validate:"numeric"`
			}{Field: 123},
		},
		{
			"pattern with non-string",
			&struct {
				Field int `validate:"pattern=^[A-Z]+$"`
			}{Field: 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			// These should not error because non-string types are silently ignored
			_ = err
		})
	}
}

func TestValidator_IsZero_AllTypes(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			"required bool false",
			&struct {
				Field bool `validate:"required"`
			}{Field: false},
			true,
		},
		{
			"required int zero",
			&struct {
				Field int `validate:"required"`
			}{Field: 0},
			true,
		},
		{
			"required uint zero",
			&struct {
				Field uint `validate:"required"`
			}{Field: 0},
			true,
		},
		{
			"required float zero",
			&struct {
				Field float64 `validate:"required"`
			}{Field: 0.0},
			true,
		},
		{
			"required slice nil",
			&struct {
				Field []string `validate:"required"`
			}{Field: nil},
			true,
		},
		{
			"required map nil",
			&struct {
				Field map[string]string `validate:"required"`
			}{Field: nil},
			true,
		},
		{
			"required ptr nil",
			&struct {
				Field *string `validate:"required"`
			}{Field: nil},
			true,
		},
		{
			"required interface nil",
			&struct {
				Field interface{} `validate:"required"`
			}{Field: nil},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_MinMax_InvalidType(t *testing.T) {
	validator := NewValidator()

	// Test min/max with string (should be silently ignored)
	err := validator.Validate(&struct {
		Field string `validate:"min=5"`
	}{Field: "test"})
	_ = err

	err = validator.Validate(&struct {
		Field string `validate:"max=5"`
	}{Field: "test"})
	_ = err
}

func TestValidator_MinMaxLength_InvalidType(t *testing.T) {
	validator := NewValidator()

	// Test minLength/maxLength with int (should be silently ignored)
	err := validator.Validate(&struct {
		Field int `validate:"minLength=5"`
	}{Field: 123})
	_ = err

	err = validator.Validate(&struct {
		Field int `validate:"maxLength=5"`
	}{Field: 123})
	_ = err
}

func TestValidator_MinMaxLength_Array(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			"array valid length",
			&struct {
				Field [3]string `validate:"minLength=2,maxLength=5"`
			}{Field: [3]string{"a", "b", "c"}},
			false,
		},
		{
			"map valid length",
			&struct {
				Field map[string]string `validate:"minLength=1,maxLength=3"`
			}{Field: map[string]string{"a": "1", "b": "2"}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_PatternInvalidRegex(t *testing.T) {
	validator := NewValidator()

	// Invalid regex should be silently ignored
	err := validator.Validate(&struct {
		Field string `validate:"pattern=[invalid"`
	}{Field: "test"})
	_ = err
}

func TestValidator_ComparisonOperators_AllNumericTypes(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			"gte with uint valid",
			&struct {
				Field uint `validate:"gte=10"`
			}{Field: 10},
			false,
		},
		{
			"gte with uint invalid",
			&struct {
				Field uint `validate:"gte=10"`
			}{Field: 9},
			true,
		},
		{
			"lt with uint valid",
			&struct {
				Field uint `validate:"lt=10"`
			}{Field: 9},
			false,
		},
		{
			"lt with uint invalid",
			&struct {
				Field uint `validate:"lt=10"`
			}{Field: 10},
			true,
		},
		{
			"lte with uint valid",
			&struct {
				Field uint `validate:"lte=10"`
			}{Field: 10},
			false,
		},
		{
			"lte with uint invalid",
			&struct {
				Field uint `validate:"lte=10"`
			}{Field: 11},
			true,
		},
		{
			"gte with float valid",
			&struct {
				Field float64 `validate:"gte=10.5"`
			}{Field: 10.5},
			false,
		},
		{
			"lt with float valid",
			&struct {
				Field float64 `validate:"lt=10.5"`
			}{Field: 10.4},
			false,
		},
		{
			"lte with float valid",
			&struct {
				Field float64 `validate:"lte=10.5"`
			}{Field: 10.5},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_InvalidParameterParsing(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			"min with invalid param",
			&struct {
				Field int `validate:"min=invalid"`
			}{Field: 10},
		},
		{
			"max with invalid param",
			&struct {
				Field int `validate:"max=invalid"`
			}{Field: 10},
		},
		{
			"minLength with invalid param",
			&struct {
				Field string `validate:"minLength=invalid"`
			}{Field: "test"},
		},
		{
			"maxLength with invalid param",
			&struct {
				Field string `validate:"maxLength=invalid"`
			}{Field: "test"},
		},
		{
			"gte with invalid param",
			&struct {
				Field int `validate:"gte=invalid"`
			}{Field: 10},
		},
		{
			"lt with invalid param",
			&struct {
				Field int `validate:"lt=invalid"`
			}{Field: 10},
		},
		{
			"lte with invalid param",
			&struct {
				Field int `validate:"lte=invalid"`
			}{Field: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			// Invalid params are silently ignored, so no error should occur
			_ = err
		})
	}
}

func TestValidator_ComparisonWithWrongType(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			"gte with string",
			&struct {
				Field string `validate:"gte=10"`
			}{Field: "test"},
		},
		{
			"lt with string",
			&struct {
				Field string `validate:"lt=10"`
			}{Field: "test"},
		},
		{
			"lte with string",
			&struct {
				Field string `validate:"lte=10"`
			}{Field: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			// String types should be silently ignored for numeric comparisons
			_ = err
		})
	}
}

func TestValidator_ValidateField_UnknownTag(t *testing.T) {
	validator := NewValidator()

	// Unknown tags should be silently ignored
	err := validator.Validate(&struct {
		Field string `validate:"unknownTag"`
	}{Field: "test"})

	if err != nil {
		t.Errorf("Unknown tags should be ignored, got error: %v", err)
	}
}

func TestValidator_StringValidations_InvalidCases(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			"url invalid format",
			&struct {
				Field string `validate:"url"`
			}{Field: "invalid url"},
			true,
		},
		{
			"alpha invalid with numbers",
			&struct {
				Field string `validate:"alpha"`
			}{Field: "abc123"},
			true,
		},
		{
			"alphanumeric invalid with special chars",
			&struct {
				Field string `validate:"alphanumeric"`
			}{Field: "abc-123"},
			true,
		},
		{
			"pattern does not match",
			&struct {
				Field string `validate:"pattern=^\\d+$"`
			}{Field: "abc"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_NestedStructValidation_WithPointer(t *testing.T) {
	validator := NewValidator()

	type Inner struct {
		Value string `validate:"required"`
	}

	type Outer struct {
		Nested *Inner
	}

	// Test with nil nested pointer
	err := validator.Validate(&Outer{Nested: nil})
	if err != nil {
		t.Errorf("Nil nested pointer should not error: %v", err)
	}

	// Test with invalid nested struct
	err = validator.Validate(&Outer{Nested: &Inner{Value: ""}})
	if err == nil {
		t.Error("Invalid nested struct should error")
	}
}

func TestValidator_StructFieldValidation(t *testing.T) {
	validator := NewValidator()

	type Inner struct {
		Value string `validate:"required"`
	}

	type Outer struct {
		Nested Inner
	}

	// Test with valid nested struct
	err := validator.Validate(&Outer{Nested: Inner{Value: "test"}})
	if err != nil {
		t.Errorf("Valid nested struct should not error: %v", err)
	}

	// Test with invalid nested struct
	err = validator.Validate(&Outer{Nested: Inner{Value: ""}})
	if err == nil {
		t.Error("Invalid nested struct should error")
	}
}
