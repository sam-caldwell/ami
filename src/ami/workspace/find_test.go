package workspace

import "testing"

func testFindPackage_FindsByKey(t *testing.T) {
	w := Workspace{
		Version: "1.0.0",
		Packages: PackageList{
			{Key: "main", Package: Package{Name: "app", Version: "0.0.1", Root: "./src"}},
			{Key: "util", Package: Package{Name: "util", Version: "1.2.3", Root: "./util"}},
		},
	}
	if p := w.FindPackage("util"); p == nil || p.Name != "util" || p.Version != "1.2.3" {
		t.Fatalf("expected util package; got %#v", p)
	}
}

func testFindPackage_NotFoundReturnsNil(t *testing.T) {
	w := Workspace{Version: "1.0.0", Packages: PackageList{{Key: "main", Package: Package{}}}}
	if p := w.FindPackage("missing"); p != nil {
		t.Fatalf("expected nil; got %#v", p)
	}
}
