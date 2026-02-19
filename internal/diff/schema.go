package diff

// SchemaSnapshot represents the full schema of a database.
type SchemaSnapshot struct {
	Tables  []Table  `json:"tables" yaml:"tables"`
	Indexes []Index  `json:"indexes" yaml:"indexes"`
}

// Table represents a database table.
type Table struct {
	Name        string       `json:"name" yaml:"name"`
	Schema      string       `json:"schema,omitempty" yaml:"schema,omitempty"`
	Columns     []Column     `json:"columns" yaml:"columns"`
	PrimaryKey  []string     `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	ForeignKeys []ForeignKey `json:"foreignKeys,omitempty" yaml:"foreignKeys,omitempty"`
}

// Column represents a table column.
type Column struct {
	Name         string `json:"name" yaml:"name"`
	DataType     string `json:"dataType" yaml:"dataType"`
	Nullable     bool   `json:"nullable" yaml:"nullable"`
	DefaultValue string `json:"defaultValue,omitempty" yaml:"defaultValue,omitempty"`
}

// Index represents a database index.
type Index struct {
	Name      string   `json:"name" yaml:"name"`
	Table     string   `json:"table" yaml:"table"`
	Columns   []string `json:"columns" yaml:"columns"`
	Unique    bool     `json:"unique" yaml:"unique"`
}

// ForeignKey represents a foreign key constraint.
type ForeignKey struct {
	Name            string   `json:"name" yaml:"name"`
	Columns         []string `json:"columns" yaml:"columns"`
	ReferencedTable string   `json:"referencedTable" yaml:"referencedTable"`
	ReferencedCols  []string `json:"referencedColumns" yaml:"referencedColumns"`
}
