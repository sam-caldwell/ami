package workspace

import "gopkg.in/yaml.v3"

// PackageList preserves the YAML shape used by SPEC: a sequence of
// single-entry maps, e.g.,
//   packages:
//     - main:
//         name: ...
// This type marshals/unmarshals between that form and a simple slice.
type PackageList []PackageEntry

// MarshalYAML implements YAML marshalling to a sequence of single-entry maps.
func (l PackageList) MarshalYAML() (interface{}, error) {
    arr := make([]map[string]Package, 0, len(l))
    for _, e := range l {
        arr = append(arr, map[string]Package{e.Key: e.Package})
    }
    return arr, nil
}

// UnmarshalYAML implements YAML unmarshalling from a sequence of single-entry maps.
func (l *PackageList) UnmarshalYAML(value *yaml.Node) error {
    if value.Kind != yaml.SequenceNode {
        return &yaml.TypeError{Errors: []string{"packages: expected sequence"}}
    }
    var out PackageList
    for _, item := range value.Content {
        if item.Kind != yaml.MappingNode || len(item.Content) != 2 {
            return &yaml.TypeError{Errors: []string{"packages: expected mapping with one entry"}}
        }
        keyNode := item.Content[0]
        valNode := item.Content[1]
        var pkg Package
        if err := valNode.Decode(&pkg); err != nil {
            return err
        }
        out = append(out, PackageEntry{Key: keyNode.Value, Package: pkg})
    }
    *l = out
    return nil
}
