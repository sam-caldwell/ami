package main

import "testing"

func Test_newEventsCmd_exists(t *testing.T) { if newEventsCmd() == nil { t.Fatal("nil") } }

