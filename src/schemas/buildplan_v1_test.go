package schemas

import (
	"encoding/json"
	"testing"
)

func TestBuildPlanV1_ValidateAndMarshal(t *testing.T) {
	bp := &BuildPlanV1{Schema: "buildplan.v1", Workspace: ".", Toolchain: ToolchainV1{AmiVersion: "v0.1.0", GoVersion: "1.25"}}
	if err := bp.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	data, err := json.Marshal(bp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got BuildPlanV1
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Schema != "buildplan.v1" {
		t.Fatalf("unexpected schema: %s", got.Schema)
	}
}
