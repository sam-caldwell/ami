package exec

import (
    "encoding/json"
    "os"
)

func extractWorkersFromPipelines(path, pipeline string) ([]string, bool) {
    type pipeList struct{ Pipelines []struct{ Name string; Steps []struct{ Name string; Args []string } } }
    var pl pipeList
    b, err := os.ReadFile(path)
    if err != nil { return nil, false }
    if err := json.Unmarshal(b, &pl); err != nil { return nil, false }
    for _, pe := range pl.Pipelines {
        if pe.Name != pipeline { continue }
        var ws []string
        for _, s := range pe.Steps { if s.Name == "Transform" && len(s.Args) > 0 { ws = append(ws, s.Args[0]) } }
        return ws, true
    }
    return nil, false
}

