package main

import "testing"

func Test_newModAuditCmd_exists(t *testing.T) { if newModAuditCmd() == nil { t.Fatal("nil") } }

