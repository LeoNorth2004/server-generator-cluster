package main

type DBConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

type TableField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Primary  bool   `json:"primary"`
	Comment  string `json:"comment"`
}

type TableConfig struct {
	Name    string       `json:"name"`
	Comment string       `json:"comment"`
	Fields  []TableField `json:"fields"`
}

type GenerateRequest struct {
	DBConfig    DBConfig      `json:"db_config"`
	Tables      []TableConfig `json:"tables"`
	ProjectName string        `json:"project_name"`
}

type GeneratedCode struct {
	Files map[string]string `json:"files"`
}

type GenerateDocsRequest struct {
	ProjectName     string        `json:"project_name" binding:"required"`
	DocType         string        `json:"doc_type" binding:"required"`
	Format          string        `json:"format" binding:"required"`
	IncludeExamples bool          `json:"include_examples"`
	IncludeComments bool          `json:"include_comments"`
	Tables          []TableConfig `json:"tables"`
}
