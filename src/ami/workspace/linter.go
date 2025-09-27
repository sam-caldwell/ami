package workspace

// Linter holds linter-specific options and rule severity overrides.
// Rules map codes (e.g., "W_IMPORT_ORDER") to one of: "off", "info", "warn", "error".
type Linter struct {
    Options []string          `yaml:"options"`
    Rules   map[string]string `yaml:"rules,omitempty"`
}
