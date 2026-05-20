package main

import (
	"testing"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_name", "UserName"},
		{"id", "Id"},
		{"created_at", "CreatedAt"},
		{"email_address", "EmailAddress"},
		{"phone", "Phone"},
		{"", ""},
	}

	for _, tt := range tests {
		result := toCamelCase(tt.input)
		if result != tt.expected {
			t.Errorf("toCamelCase(%s) = %s, expected %s", tt.input, result, tt.expected)
		} else {
			t.Logf("✓ toCamelCase(%s) = %s", tt.input, result)
		}
	}
}

func TestToLowerCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_name", "userName"},
		{"id", "id"},
		{"created_at", "createdAt"},
		{"email_address", "emailAddress"},
		{"", ""},
	}

	for _, tt := range tests {
		result := toLowerCamelCase(tt.input)
		if result != tt.expected {
			t.Errorf("toLowerCamelCase(%s) = %s, expected %s", tt.input, result, tt.expected)
		} else {
			t.Logf("✓ toLowerCamelCase(%s) = %s", tt.input, result)
		}
	}
}

func TestGoTypeFromSQL(t *testing.T) {
	tests := []struct {
		sqlType  string
		expected string
	}{
		{"int", "int"},
		{"integer", "int"},
		{"bigint", "int64"},
		{"varchar", "string"},
		{"text", "string"},
		{"boolean", "bool"},
		{"bool", "bool"},
		{"float", "float64"},
		{"double", "float64"},
		{"timestamp", "time.Time"},
		{"datetime", "time.Time"},
		{"json", "datatypes.JSON"},
		{"uuid", "string"},
		{"unknown_type", "string"},
	}

	for _, tt := range tests {
		result := goTypeFromSQL(tt.sqlType)
		if result != tt.expected {
			t.Errorf("goTypeFromSQL(%s) = %s, expected %s", tt.sqlType, result, tt.expected)
		} else {
			t.Logf("✓ goTypeFromSQL(%s) = %s", tt.sqlType, result)
		}
	}
}

func TestGormTagFromField(t *testing.T) {
	tests := []struct {
		field    TableField
		contains string
	}{
		{
			field:    TableField{Name: "id", Type: "int", Primary: true},
			contains: "primaryKey",
		},
		{
			field:    TableField{Name: "name", Type: "varchar", Nullable: false},
			contains: "not null",
		},
		{
			field:    TableField{Name: "email", Type: "varchar", Comment: "Email address"},
			contains: "comment:Email address",
		},
		{
			field:    TableField{Name: "age", Type: "int", Nullable: true},
			contains: "column:age",
		},
	}

	for i, tt := range tests {
		result := gormTagFromField(tt.field)
		if !containsString(result, tt.contains) {
			t.Errorf("Test %d: gormTagFromField() = %s, should contain %s", i+1, result, tt.contains)
		} else {
			t.Logf("✓ Test %d: gormTag contains '%s'", i+1, tt.contains)
		}
	}
}

func TestSafeSprintf(t *testing.T) {
	tests := []struct {
		format   string
		args     []interface{}
		expected string
	}{
		{"Hello %s", []interface{}{"World"}, "Hello World"},
		{"Number %d", []interface{}{42}, "Number 42"},
		{"No args", []interface{}{}, "No args"},
		{"Multiple %s and %s", []interface{}{"foo", "bar"}, "Multiple foo and bar"},
	}

	for i, tt := range tests {
		result := safeSprintf(tt.format, tt.args...)
		if result != tt.expected {
			t.Errorf("Test %d: safeSprintf() = %s, expected %s", i+1, result, tt.expected)
		} else {
			t.Logf("✓ Test %d: safeSprintf() = %s", i+1, result)
		}
	}
}

func TestExtractPort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"8080", "8080"},
		{":8080", "8080"},
		{"localhost:3000", "3000"},
		{"", ""},
		{"invalid", ""},
	}

	for _, tt := range tests {
		result := extractPort(tt.input)
		if result != tt.expected {
			t.Errorf("extractPort(%s) = %s, expected %s", tt.input, result, tt.expected)
		} else {
			t.Logf("✓ extractPort(%s) = %s", tt.input, result)
		}
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		} else {
			t.Logf("✓ min(%d, %d) = %d", tt.a, tt.b, result)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
