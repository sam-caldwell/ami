package driver

type ssaUnit struct {
    Schema   string     `json:"schema"`
    Package  string     `json:"package"`
    Unit     string     `json:"unit"`
    Functions []ssaFunc `json:"functions"`
}
// helpers and writer moved to separate files
