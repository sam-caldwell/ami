package sem

import "testing"

func testFindGenericArityMismatchDeep_NestedOwned(t *testing.T) {
	exp := "slice<Owned<T>>"
	act := "slice<Owned<int,string>>"
	if m, base, w, g := findGenericArityMismatchDeep(exp, act); !m || base != "Owned" || w != 1 || g != 2 {
		t.Fatalf("expected mismatch at Owned 1 vs 2; got m=%v base=%s w=%d g=%d", m, base, w, g)
	}
}

func testFindGenericArityMismatchDeepPath_NestedOwned(t *testing.T) {
	exp := "slice<Owned<T>>"
	act := "slice<Owned<int,string>>"
	if m, path, pathIdx, base, w, g := findGenericArityMismatchDeepPath(exp, act); !m || base != "Owned" || w != 1 || g != 2 {
		t.Fatalf("expected mismatch at Owned 1 vs 2; got m=%v base=%s w=%d g=%d", m, base, w, g)
	} else {
		if len(path) < 2 || path[0] != "slice" || path[1] != "Owned" {
			t.Fatalf("unexpected path: %#v", path)
		}
		if len(pathIdx) < 1 || pathIdx[0] != 0 {
			t.Fatalf("unexpected pathIdx: %#v", pathIdx)
		}
	}
}

func testFindGenericArityMismatchWithFields_StructPath(t *testing.T) {
	exp := "Struct{a:slice<Owned<T>>}"
	act := "Struct{a:slice<Owned<int,string>>}"
	if m, path, idx, fieldPath, base, w, g := findGenericArityMismatchWithFields(exp, act); !m || base != "Owned" || w != 1 || g != 2 {
		t.Fatalf("expected mismatch at Owned 1 vs 2; got m=%v base=%s w=%d g=%d", m, base, w, g)
	} else {
		if len(fieldPath) < 1 || fieldPath[0] != "a" {
			t.Fatalf("unexpected fieldPath: %#v", fieldPath)
		}
		if len(path) < 2 || path[0] != "slice" || path[1] != "Owned" {
			t.Fatalf("unexpected path: %#v", path)
		}
		if len(idx) < 1 || idx[0] != 0 {
			t.Fatalf("unexpected pathIdx: %#v", idx)
		}
	}
}
