package main

import "testing"

func Test_setRuleToggles_setsGlobal(t *testing.T) {
    before := currentRuleToggles
    setRuleToggles(RuleToggles{StageB: true})
    if !currentRuleToggles.StageB { t.Fatal("not set") }
    currentRuleToggles = before
}

