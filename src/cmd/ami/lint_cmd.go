package main

import (
    "github.com/spf13/cobra"
)

// newLintCmd creates the `ami lint` subcommand (Stage A).
func newLintCmd() *cobra.Command {
    var jsonOut bool
    var strict bool
    var stageB bool
    var rUnknown bool
    var rUnused bool
    var rImports bool
    var rDup bool
    var rMemsafe bool
    var rRAII bool
    var rules []string
    var maxWarn int
    var failFast bool
    var compatCodes bool
    cmd := &cobra.Command{
        Use:   "lint",
        Short: "Lint the workspace and sources (Stage A)",
        RunE: func(cmd *cobra.Command, args []string) error {
            verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
            if lg := getRootLogger(); lg != nil {
                lg.Info("lint.start", map[string]any{"json": jsonOut, "verbose": verbose})
            }
            // Apply lint options
            setLintOptions(LintOptions{Rules: rules, MaxWarn: maxWarn, FailFast: failFast, CompatCodes: compatCodes})
            // Apply rule toggles (Stage B still a no-op until frontend integration)
            setRuleToggles(RuleToggles{
                StageB:       stageB,
                UnknownIdent: rUnknown,
                Unused:       rUnused,
                ImportExist:  rImports,
                Duplicates:   rDup,
                MemorySafety: rMemsafe,
                RAIIHint:     rRAII,
            })
            return runLint(cmd.OutOrStdout(), ".", jsonOut, verbose, strict)
        },
    }
    cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-parsable JSON diagnostics")
    cmd.Flags().BoolVar(&strict, "strict", false, "elevate warnings to errors")
    cmd.Flags().BoolVar(&stageB, "stage-b", false, "enable parser-backed rules (no-op until frontend is available)")
    cmd.Flags().BoolVar(&rUnknown, "rule-unknown", false, "enable unknown identifiers rule (Stage B)")
    cmd.Flags().BoolVar(&rUnused, "rule-unused", false, "enable unused detection (Stage B)")
    cmd.Flags().BoolVar(&rImports, "rule-imports", false, "enable import existence/versioning (Stage B)")
    cmd.Flags().BoolVar(&rDup, "rule-duplicates", false, "enable duplicate/alias checks (Stage B)")
    cmd.Flags().BoolVar(&rMemsafe, "rule-memsafe", false, "enable memory-safety diagnostics (Stage B)")
    cmd.Flags().BoolVar(&rRAII, "rule-raii", false, "enable RAII hint diagnostics (Stage B)")
    cmd.Flags().StringSliceVar(&rules, "rules", nil, "filter diagnostics by code pattern (supports substring, glob, and re:<regex> or /regex/)")
    cmd.Flags().IntVar(&maxWarn, "max-warn", -1, "fail when warnings exceed this count (-1 = no limit)")
    cmd.Flags().BoolVar(&failFast, "failfast", false, "exit non-zero if any warning or error is found")
    cmd.Flags().BoolVar(&compatCodes, "compat-codes", false, "add LINT_* compat code alias in JSON 'data.compatCode'")
    return cmd
}
