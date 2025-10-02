package driver

type irTypesIndex struct {
    Schema  string             `json:"schema"`
    Package string             `json:"package"`
    Units   []irTypesIndexUnit `json:"units"`
}
// collectors, writer, and util moved to separate files to satisfy single-declaration rule
