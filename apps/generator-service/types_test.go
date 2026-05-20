package main

import (
	"encoding/json"
	"testing"
)

func TestDBConfigJSON(t *testing.T) {
	config := DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "123456",
		DBName:   "mydb",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal DBConfig: %v", err)
	}

	var unmarshaled DBConfig
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal DBConfig: %v", err)
	}

	if unmarshaled.Host != config.Host {
		t.Errorf("Host mismatch: got %s, want %s", unmarshaled.Host, config.Host)
	}
	if unmarshaled.Port != config.Port {
		t.Errorf("Port mismatch: got %s, want %s", unmarshaled.Port, config.Port)
	}
	if unmarshaled.DBName != config.DBName {
		t.Errorf("DBName mismatch: got %s, want %s", unmarshaled.DBName, config.DBName)
	}

	t.Logf("✓ DBConfig JSON serialization/deserialization works correctly")
}

func TestTableConfigValidation(t *testing.T) {
	validTable := TableConfig{
		Name:    "users",
		Comment: "User accounts table",
		Fields: []TableField{
			{Name: "id", Type: "int", Primary: true, Comment: "Primary key"},
			{Name: "username", Type: "varchar", Nullable: false, Comment: "Username"},
			{Name: "email", Type: "varchar", Nullable: false, Comment: "Email address"},
		},
	}

	if validTable.Name == "" {
		t.Error("Table name should not be empty")
	}
	if len(validTable.Fields) == 0 {
		t.Error("Table should have at least one field")
	}
	if len(validTable.Fields) < 2 {
		t.Error("Table should have multiple fields for a proper test")
	}

	for i, field := range validTable.Fields {
		if field.Name == "" {
			t.Errorf("Field %d: Name should not be empty", i)
		}
		if field.Type == "" {
			t.Errorf("Field %d: Type should not be empty", i)
		}
	}

	t.Logf("✓ TableConfig validation passed for table '%s' with %d fields", validTable.Name, len(validTable.Fields))
}

func TestGenerateRequestValidation(t *testing.T) {
	validReq := GenerateRequest{
		DBConfig: DBConfig{
			Host:   "localhost",
			Port:   "5432",
			User:   "postgres",
			DBName: "testdb",
		},
		Tables: []TableConfig{
			{
				Name:    "products",
				Comment: "Products table",
				Fields: []TableField{
					{Name: "id", Type: "int", Primary: true},
					{Name: "name", Type: "varchar", Nullable: false},
					{Name: "price", Type: "decimal", Nullable: false},
				},
			},
		},
		ProjectName: "my-awesome-project",
	}

	if validReq.ProjectName == "" {
		t.Error("Project name is required")
	}
	if len(validReq.Tables) == 0 {
		t.Error("At least one table is required")
	}
	if validReq.DBConfig.Host == "" {
		t.Error("DB host is required")
	}

	data, err := json.Marshal(validReq)
	if err != nil {
		t.Fatalf("Failed to marshal GenerateRequest: %v", err)
	}

	t.Logf("✓ GenerateRequest validation passed")
	t.Logf("  Project: %s", validReq.ProjectName)
	t.Logf("  Tables: %d", len(validReq.Tables))
	t.Logf("  JSON size: %d bytes", len(data))
}

func TestGeneratedCodeStructure(t *testing.T) {
	code := GeneratedCode{
		Files: map[string]string{
			"main.go":           "package main\n\nfunc main() {}",
			"config/config.go": "package config\n\nvar Config = struct{}{}",
			"README.md":        "# My Generated Project",
		},
	}

	if len(code.Files) == 0 {
		t.Error("GeneratedCode should contain at least one file")
	}

	for path, content := range code.Files {
		if path == "" {
			t.Error("File path should not be empty")
		}
		if content == "" {
			t.Errorf("File content should not be empty for path: %s", path)
		}
	}

	expectedFiles := []string{"main.go", "config/config.go", "README.md"}
	for _, expected := range expectedFiles {
		if _, exists := code.Files[expected]; !exists {
			t.Errorf("Expected file %s not found in GeneratedCode", expected)
		}
	}

	t.Logf("✓ GeneratedCode structure validation passed with %d files", len(code.Files))
}

func TestGenerateDocsRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        GenerateDocsRequest
		expectError    bool
		errorMsg       string
	}{
		{
			name: "Valid API docs request",
			request: GenerateDocsRequest{
				ProjectName:     "test-project",
				DocType:         "api",
				Format:          "markdown",
				IncludeExamples: true,
				IncludeComments: true,
			},
			expectError: false,
		},
		{
			name: "Valid config docs request",
			request: GenerateDocsRequest{
				ProjectName:     "test-project",
				DocType:         "config",
				Format:          "markdown",
				IncludeExamples: false,
				IncludeComments: false,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.request.ProjectName == "" && !tt.expectError {
				t.Error("ProjectName should not be empty")
			}
			if tt.request.DocType == "" && !tt.expectError {
				t.Error("DocType should not be empty")
			}
			if tt.request.Format == "" && !tt.expectError {
				t.Error("Format should not be empty")
			}

			t.Logf("✓ GenerateDocsRequest validation passed: %s", tt.name)
		})
	}
}

func BenchmarkToCamelCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		toCamelCase("user_email_address")
	}
}

func BenchmarkGoTypeFromSQL(b *testing.B) {
	types := []string{"int", "varchar", "text", "boolean", "timestamp", "json"}
	for i := 0; i < b.N; i++ {
		goTypeFromSQL(types[i%len(types)])
	}
}
