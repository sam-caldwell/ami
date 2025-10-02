package main

import "testing"

func Test_newPipelineStatsCmd_exists(t *testing.T) { if newPipelineStatsCmd() == nil { t.Fatal("nil") } }

