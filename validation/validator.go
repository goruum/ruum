// Package validation provides struct validation using tags, similar to NestJS class-validator.
package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Validator validates structs using tags
type Validator struct {
	tagName string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		tagName: "validate",
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field      string                 `json:"field"`
	Tag        string                 `json:"tag"`
	Message    string                 `json:"message"`
	Value      interface{}            `json:"value,omitempty"`
	Constraint map[string]interface{} `json:"constraint,omitempty"`
}

// Error implements error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements error interface
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validate validates a struct
func (v *Validator) Validate(s interface{}) error {
	return v.validate(reflect.ValueOf(s), "")
}

func (v *Validator) validate(val reflect.Value, prefix string) error {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	var errors ValidationErrors
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		fieldName := typeField.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		// Get validation tag
		tag := typeField.Tag.Get(v.tagName)
		if tag == "" || tag == "-" {
			// Check nested structs
			if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
				if err := v.validate(field, fieldName); err != nil {
					if ve, ok := err.(ValidationErrors); ok {
						errors = append(errors, ve...)
					}
				}
			}
			continue
		}

		// Parse and validate tags
		tags := strings.Split(tag, ",")
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if err := v.validateField(field, typeField, fieldName, t); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func (v *Validator) validateField(field reflect.Value, typeField reflect.StructField, fieldName, tag string) *ValidationError {
	parts := strings.SplitN(tag, "=", 2)
	validator := parts[0]
	var param string
	if len(parts) > 1 {
		param = parts[1]
	}

	switch validator {
	case "required":
		return v.validateRequired(field, fieldName)
	case "min":
		return v.validateMin(field, fieldName, param)
	case "max":
		return v.validateMax(field, fieldName, param)
	case "minLength":
		return v.validateMinLength(field, fieldName, param)
	case "maxLength":
		return v.validateMaxLength(field, fieldName, param)
	case "email":
		return v.validateEmail(field, fieldName)
	case "url":
		return v.validateURL(field, fieldName)
	case "alpha":
		return v.validateAlpha(field, fieldName)
	case "alphanumeric":
		return v.validateAlphanumeric(field, fieldName)
	case "numeric":
		return v.validateNumeric(field, fieldName)
	case "pattern":
		return v.validatePattern(field, fieldName, param)
	case "eq":
		return v.validateEquals(field, fieldName, param)
	case "ne":
		return v.validateNotEquals(field, fieldName, param)
	case "gt":
		return v.validateGreaterThan(field, fieldName, param)
	case "gte":
		return v.validateGreaterOrEqual(field, fieldName, param)
	case "lt":
		return v.validateLessThan(field, fieldName, param)
	case "lte":
		return v.validateLessOrEqual(field, fieldName, param)
	case "oneof":
		return v.validateOneOf(field, fieldName, param)
	case "isArray":
		return v.validateIsArray(field, fieldName)
	case "isBoolean":
		return v.validateIsBoolean(field, fieldName)
	case "isNumber":
		return v.validateIsNumber(field, fieldName)
	case "isString":
		return v.validateIsString(field, fieldName)
	}

	return nil
}

func (v *Validator) validateRequired(field reflect.Value, fieldName string) *ValidationError {
	if isZero(field) {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "required",
			Message: fmt.Sprintf("%s is required", fieldName),
		}
	}
	return nil
}

func (v *Validator) validateMin(field reflect.Value, fieldName, param string) *ValidationError {
	min, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	var value float64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(field.Uint())
	case reflect.Float32, reflect.Float64:
		value = field.Float()
	default:
		return nil
	}

	if value < min {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "min",
			Message: fmt.Sprintf("%s must be at least %s", fieldName, param),
			Value:   value,
			Constraint: map[string]interface{}{
				"min": min,
			},
		}
	}
	return nil
}

func (v *Validator) validateMax(field reflect.Value, fieldName, param string) *ValidationError {
	max, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	var value float64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(field.Uint())
	case reflect.Float32, reflect.Float64:
		value = field.Float()
	default:
		return nil
	}

	if value > max {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "max",
			Message: fmt.Sprintf("%s must be at most %s", fieldName, param),
			Value:   value,
			Constraint: map[string]interface{}{
				"max": max,
			},
		}
	}
	return nil
}

func (v *Validator) validateMinLength(field reflect.Value, fieldName, param string) *ValidationError {
	minLen, err := strconv.Atoi(param)
	if err != nil {
		return nil
	}

	var length int
	switch field.Kind() {
	case reflect.String:
		length = len(field.String())
	case reflect.Slice, reflect.Array, reflect.Map:
		length = field.Len()
	default:
		return nil
	}

	if length < minLen {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "minLength",
			Message: fmt.Sprintf("%s must be at least %d characters/items long", fieldName, minLen),
			Value:   length,
			Constraint: map[string]interface{}{
				"minLength": minLen,
			},
		}
	}
	return nil
}

func (v *Validator) validateMaxLength(field reflect.Value, fieldName, param string) *ValidationError {
	maxLen, err := strconv.Atoi(param)
	if err != nil {
		return nil
	}

	var length int
	switch field.Kind() {
	case reflect.String:
		length = len(field.String())
	case reflect.Slice, reflect.Array, reflect.Map:
		length = field.Len()
	default:
		return nil
	}

	if length > maxLen {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "maxLength",
			Message: fmt.Sprintf("%s must be at most %d characters/items long", fieldName, maxLen),
			Value:   length,
			Constraint: map[string]interface{}{
				"maxLength": maxLen,
			},
		}
	}
	return nil
}

func (v *Validator) validateEmail(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return nil
	}

	email := field.String()
	if email == "" {
		return nil // Let 'required' handle empty values
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "email",
			Message: fmt.Sprintf("%s must be a valid email address", fieldName),
			Value:   email,
		}
	}
	return nil
}

func (v *Validator) validateURL(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return nil
	}

	url := field.String()
	if url == "" {
		return nil
	}

	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "url",
			Message: fmt.Sprintf("%s must be a valid URL", fieldName),
			Value:   url,
		}
	}
	return nil
}

func (v *Validator) validateAlpha(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return nil
	}

	str := field.String()
	if str == "" {
		return nil
	}

	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(str) {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "alpha",
			Message: fmt.Sprintf("%s must contain only letters", fieldName),
			Value:   str,
		}
	}
	return nil
}

func (v *Validator) validateAlphanumeric(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return nil
	}

	str := field.String()
	if str == "" {
		return nil
	}

	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphanumericRegex.MatchString(str) {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "alphanumeric",
			Message: fmt.Sprintf("%s must contain only letters and numbers", fieldName),
			Value:   str,
		}
	}
	return nil
}

func (v *Validator) validateNumeric(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return nil
	}

	str := field.String()
	if str == "" {
		return nil
	}

	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	if !numericRegex.MatchString(str) {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "numeric",
			Message: fmt.Sprintf("%s must contain only numbers", fieldName),
			Value:   str,
		}
	}
	return nil
}

func (v *Validator) validatePattern(field reflect.Value, fieldName, pattern string) *ValidationError {
	if field.Kind() != reflect.String {
		return nil
	}

	str := field.String()
	if str == "" {
		return nil
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	if !regex.MatchString(str) {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "pattern",
			Message: fmt.Sprintf("%s must match pattern %s", fieldName, pattern),
			Value:   str,
			Constraint: map[string]interface{}{
				"pattern": pattern,
			},
		}
	}
	return nil
}

func (v *Validator) validateEquals(field reflect.Value, fieldName, param string) *ValidationError {
	value := fmt.Sprintf("%v", field.Interface())
	if value != param {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "eq",
			Message: fmt.Sprintf("%s must be equal to %s", fieldName, param),
			Value:   value,
		}
	}
	return nil
}

func (v *Validator) validateNotEquals(field reflect.Value, fieldName, param string) *ValidationError {
	value := fmt.Sprintf("%v", field.Interface())
	if value == param {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "ne",
			Message: fmt.Sprintf("%s must not be equal to %s", fieldName, param),
			Value:   value,
		}
	}
	return nil
}

func (v *Validator) validateGreaterThan(field reflect.Value, fieldName, param string) *ValidationError {
	target, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	var value float64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(field.Uint())
	case reflect.Float32, reflect.Float64:
		value = field.Float()
	default:
		return nil
	}

	if value <= target {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "gt",
			Message: fmt.Sprintf("%s must be greater than %s", fieldName, param),
			Value:   value,
		}
	}
	return nil
}

func (v *Validator) validateGreaterOrEqual(field reflect.Value, fieldName, param string) *ValidationError {
	target, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	var value float64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(field.Uint())
	case reflect.Float32, reflect.Float64:
		value = field.Float()
	default:
		return nil
	}

	if value < target {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "gte",
			Message: fmt.Sprintf("%s must be greater than or equal to %s", fieldName, param),
			Value:   value,
		}
	}
	return nil
}

func (v *Validator) validateLessThan(field reflect.Value, fieldName, param string) *ValidationError {
	target, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	var value float64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(field.Uint())
	case reflect.Float32, reflect.Float64:
		value = field.Float()
	default:
		return nil
	}

	if value >= target {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "lt",
			Message: fmt.Sprintf("%s must be less than %s", fieldName, param),
			Value:   value,
		}
	}
	return nil
}

func (v *Validator) validateLessOrEqual(field reflect.Value, fieldName, param string) *ValidationError {
	target, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	var value float64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(field.Uint())
	case reflect.Float32, reflect.Float64:
		value = field.Float()
	default:
		return nil
	}

	if value > target {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "lte",
			Message: fmt.Sprintf("%s must be less than or equal to %s", fieldName, param),
			Value:   value,
		}
	}
	return nil
}

func (v *Validator) validateOneOf(field reflect.Value, fieldName, param string) *ValidationError {
	value := fmt.Sprintf("%v", field.Interface())
	options := strings.Split(param, " ")

	for _, option := range options {
		if value == option {
			return nil
		}
	}

	return &ValidationError{
		Field:   fieldName,
		Tag:     "oneof",
		Message: fmt.Sprintf("%s must be one of [%s]", fieldName, param),
		Value:   value,
		Constraint: map[string]interface{}{
			"options": options,
		},
	}
}

func (v *Validator) validateIsArray(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "isArray",
			Message: fmt.Sprintf("%s must be an array", fieldName),
		}
	}
	return nil
}

func (v *Validator) validateIsBoolean(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.Bool {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "isBoolean",
			Message: fmt.Sprintf("%s must be a boolean", fieldName),
		}
	}
	return nil
}

func (v *Validator) validateIsNumber(field reflect.Value, fieldName string) *ValidationError {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return nil
	default:
		return &ValidationError{
			Field:   fieldName,
			Tag:     "isNumber",
			Message: fmt.Sprintf("%s must be a number", fieldName),
		}
	}
}

func (v *Validator) validateIsString(field reflect.Value, fieldName string) *ValidationError {
	if field.Kind() != reflect.String {
		return &ValidationError{
			Field:   fieldName,
			Tag:     "isString",
			Message: fmt.Sprintf("%s must be a string", fieldName),
		}
	}
	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// DefaultValidator is the default validator instance
var DefaultValidator = NewValidator()

// Validate validates a struct using the default validator
func Validate(s interface{}) error {
	return DefaultValidator.Validate(s)
}
