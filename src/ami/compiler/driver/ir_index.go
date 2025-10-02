package driver

type irIndex struct {
    Schema  string        `json:"schema"`
    Package string        `json:"package"`
    Units   []irIndexUnit `json:"units"`
}
// unit type and writer moved to dedicated files
