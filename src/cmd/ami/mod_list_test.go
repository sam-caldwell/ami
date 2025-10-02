package main

import "testing"

func Test_newModListCmd_exists(t *testing.T) { if newModListCmd() == nil { t.Fatal("nil") } }

