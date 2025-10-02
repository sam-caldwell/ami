package main

import "testing"

func Test_newErrorsCmd_exists(t *testing.T) { if newErrorsCmd() == nil { t.Fatal("nil") } }

