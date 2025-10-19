package strutil

import (
	"testing"
)

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{
			name:     "join with comma",
			strs:     []string{"a", "b", "c"},
			sep:      ",",
			expected: "a,b,c",
		},
		{
			name:     "join with space",
			strs:     []string{"hello", "world"},
			sep:      " ",
			expected: "hello world",
		},
		{
			name:     "empty slice",
			strs:     []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "single element",
			strs:     []string{"alone"},
			sep:      ",",
			expected: "alone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinStrings(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("JoinStrings() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "contains value",
			slice:    []string{"a", "b", "c"},
			value:    "b",
			expected: true,
		},
		{
			name:     "does not contain value",
			slice:    []string{"a", "b", "c"},
			value:    "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			value:    "a",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.slice, tt.value)
			if result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "simple map",
			input: map[string]interface{}{
				"key": "value",
			},
			wantErr: false,
		},
		{
			name:    "string",
			input:   "test",
			wantErr: false,
		},
		{
			name:    "number",
			input:   42,
			wantErr: false,
		},
		{
			name: "struct",
			input: struct {
				Name string `json:"name"`
			}{Name: "test"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("ToJSON() returned empty string")
			}
		})
	}
}

func TestFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		target  interface{}
		wantErr bool
	}{
		{
			name:    "valid json",
			json:    `{"name":"test"}`,
			target:  &map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    `{invalid}`,
			target:  &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FromJSON(tt.json, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTernary(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		trueVal   interface{}
		falseVal  interface{}
		expected  interface{}
	}{
		{
			name:      "condition true",
			condition: true,
			trueVal:   "yes",
			falseVal:  "no",
			expected:  "yes",
		},
		{
			name:      "condition false",
			condition: false,
			trueVal:   "yes",
			falseVal:  "no",
			expected:  "no",
		},
		{
			name:      "numeric values",
			condition: true,
			trueVal:   1,
			falseVal:  0,
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Ternary(tt.condition, tt.trueVal, tt.falseVal)
			if result != tt.expected {
				t.Errorf("Ternary() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStringPtr(t *testing.T) {
	s := "test"
	ptr := StringPtr(s)
	if ptr == nil {
		t.Fatal("StringPtr() returned nil")
	}
	if *ptr != s {
		t.Errorf("StringPtr() = %v, want %v", *ptr, s)
	}
}

func TestIntPtr(t *testing.T) {
	i := 42
	ptr := IntPtr(i)
	if ptr == nil {
		t.Fatal("IntPtr() returned nil")
	}
	if *ptr != i {
		t.Errorf("IntPtr() = %v, want %v", *ptr, i)
	}
}

func TestBoolPtr(t *testing.T) {
	b := true
	ptr := BoolPtr(b)
	if ptr == nil {
		t.Fatal("BoolPtr() returned nil")
	}
	if *ptr != b {
		t.Errorf("BoolPtr() = %v, want %v", *ptr, b)
	}
}

func TestCoalesce(t *testing.T) {
	tests := []struct {
		name     string
		values   []interface{}
		expected interface{}
	}{
		{
			name:     "first non-nil",
			values:   []interface{}{nil, "first", "second"},
			expected: "first",
		},
		{
			name:     "all nil",
			values:   []interface{}{nil, nil, nil},
			expected: nil,
		},
		{
			name:     "first is non-nil",
			values:   []interface{}{"first", nil, "second"},
			expected: "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Coalesce(tt.values...)
			if result != tt.expected {
				t.Errorf("Coalesce() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMap(t *testing.T) {
	input := []interface{}{1, 2, 3}
	double := func(v interface{}) interface{} {
		return v.(int) * 2
	}

	result := Map(input, double)
	if len(result) != 3 {
		t.Errorf("Map() length = %v, want 3", len(result))
	}
	if result[0] != 2 || result[1] != 4 || result[2] != 6 {
		t.Errorf("Map() = %v, want [2 4 6]", result)
	}
}

func TestFilter(t *testing.T) {
	input := []interface{}{1, 2, 3, 4, 5}
	isEven := func(v interface{}) bool {
		return v.(int)%2 == 0
	}

	result := Filter(input, isEven)
	if len(result) != 2 {
		t.Errorf("Filter() length = %v, want 2", len(result))
	}
}

func TestReduce(t *testing.T) {
	input := []interface{}{1, 2, 3, 4}
	sum := func(acc, v interface{}) interface{} {
		return acc.(int) + v.(int)
	}

	result := Reduce(input, sum, 0)
	if result != 10 {
		t.Errorf("Reduce() = %v, want 10", result)
	}
}

