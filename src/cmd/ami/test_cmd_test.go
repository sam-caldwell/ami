package main

import "testing"

func Test_newTestCmd_exists(t *testing.T) { if newTestCmd() == nil { t.Fatal("nil") } }

