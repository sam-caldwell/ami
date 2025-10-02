package main

import "testing"

func test_updateSumEntry_objectForm(t *testing.T) {
	m := map[string]any{"schema": "ami.sum/v1", "packages": map[string]any{"pkg": map[string]any{"version": "1.0.0"}}}
	if !updateSumEntry(m, "pkg", "1.0.0", "abc", "") {
		t.Fatal("update failed")
	}
}

func test_updateSumEntry_arrayForm(t *testing.T) {
	m := map[string]any{"schema": "ami.sum/v1", "packages": []any{map[string]any{"name": "pkg", "version": "1.0.0"}}}
	if !updateSumEntry(m, "pkg", "1.0.0", "abc", "src") {
		t.Fatal("update failed")
	}
}
