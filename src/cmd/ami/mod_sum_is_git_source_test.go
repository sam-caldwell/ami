package main

import "testing"

func Test_isGitSource(t *testing.T) {
    if !isGitSource("git+ssh://example.com/repo.git") { t.Fatal("git+ssh should be true") }
    if !isGitSource("file+git:///abs/path/repo.git") { t.Fatal("file+git should be true") }
    if isGitSource("https://example.com/repo.git") { t.Fatal("https should be false") }
    if isGitSource("") { t.Fatal("empty should be false") }
}

