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
    // StrictMDPOverride overrides workspace strict_merge_dedup_partition when set.
    // HasStrictMDPOverride indicates whether the CLI explicitly set the override.
    StrictMDPOverride   bool
    HasStrictMDPOverride bool
}

// currentRuleToggles is the active Stage B rule configuration.
var currentRuleToggles RuleToggles
