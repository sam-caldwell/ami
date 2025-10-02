package main

import "testing"

func Test_newModGetCmd_exists(t *testing.T) { if newModGetCmd() == nil { t.Fatal("nil") } }

