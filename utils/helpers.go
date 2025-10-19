// Package utils provides utility functions for the framework.
package utils

import (
	"encoding/json"
	"strings"
)

// JoinStrings joins strings with a separator
func JoinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// Contains checks if a slice contains a value
func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// ToJSON converts a value to JSON string
func ToJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON parses JSON string to a value
func FromJSON(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// Ternary implements ternary operator
func Ternary(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// Coalesce returns the first non-nil value
func Coalesce(values ...interface{}) interface{} {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

// Map applies a function to each element of a slice
func Map(slice []interface{}, fn func(interface{}) interface{}) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter filters a slice based on a predicate
func Filter(slice []interface{}, fn func(interface{}) bool) []interface{} {
	result := make([]interface{}, 0)
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce reduces a slice to a single value
func Reduce(slice []interface{}, fn func(interface{}, interface{}) interface{}, initial interface{}) interface{} {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}
