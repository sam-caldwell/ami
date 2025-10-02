package driver

type sourcesUnit struct {
    Schema          string         `json:"schema"`
    Package         string         `json:"package"`
    Unit            string         `json:"unit"`
    ImportsDetailed []importDetail `json:"importsDetailed"`
    Pragmas         []pragmaDetail `json:"pragmas,omitempty"`
}
// writer and detail types moved to separate files to satisfy single-declaration rule
