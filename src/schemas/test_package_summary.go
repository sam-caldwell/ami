package schemas

// TestPackageSummary aggregates results for a single package.
type TestPackageSummary struct {
    Package string `json:"package"`
    Pass    int    `json:"pass"`
    Fail    int    `json:"fail"`
    Skip    int    `json:"skip"`
    Cases   int    `json:"cases"`
}

