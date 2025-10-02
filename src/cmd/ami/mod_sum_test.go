package main

import "testing"

func Test_newModSumCmd_exists(t *testing.T) { if newModSumCmd() == nil { t.Fatal("nil") } }

