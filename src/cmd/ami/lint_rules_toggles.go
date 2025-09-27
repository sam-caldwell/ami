package main

// RuleToggles controls parser-backed lint rules (Stage B).
type RuleToggles struct {
    StageB       bool
    UnknownIdent bool
    Unused       bool
    ImportExist  bool
    Duplicates   bool
    MemorySafety bool
    RAIIHint     bool
}

