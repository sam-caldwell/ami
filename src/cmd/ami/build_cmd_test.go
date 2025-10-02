package main

import "testing"

func Test_newBuildCmd_exists(t *testing.T) { if newBuildCmd() == nil { t.Fatal("nil") } }

