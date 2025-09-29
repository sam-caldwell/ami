package workspace

// Linter holds linter-specific options and rule severity overrides.
// Rules map codes (e.g., "W_IMPORT_ORDER") to one of: "off", "info", "warn", "error".
type Linter struct {
    Options []string          `yaml:"options"`
    Rules   map[string]string `yaml:"rules,omitempty"`
    // Suppress maps a path prefix (workspace-relative) to a list of rule codes to suppress.
    // Example:
    //   suppress:
    //     "./legacy": ["W_IDENT_UNDERSCORE", "W_IMPORT_ORDER"]
    Suppress map[string][]string `yaml:"suppress,omitempty"`
    // DecoratorsDisabled lists decorator names that are disabled for analysis.
    // Example:
    //   decorators_disabled: ["metrics", "deprecated"]
    DecoratorsDisabled []string `yaml:"decorators_disabled,omitempty"`
}
