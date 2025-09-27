package edge

import "testing"

func TestKind_ValuesDistinct(t *testing.T) {
    if KindFIFO == KindLIFO || KindLIFO == KindPipeline || KindPipeline == KindMultiPath {
        t.Fatalf("edge kind constants should be distinct")
    }
}

