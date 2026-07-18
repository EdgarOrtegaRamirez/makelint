package main

import (
	"testing"
)

func TestScoreToGrade(t *testing.T) {
	tests := []struct {
		score  float64
		expect string
	}{
		{100, "A+"},
		{90, "A"},
		{80, "B"},
		{70, "C"},
		{60, "D"},
		{0, "F"},
		{-10, "F"},
	}
	for _, tt := range tests {
		got := scoreToGrade(tt.score)
		if got != tt.expect {
			t.Errorf("scoreToGrade(%v) = %v, want %v", tt.score, got, tt.expect)
		}
	}
}

func TestCalculateScore(t *testing.T) {
	// No issues = perfect score
	result := AnalysisResult{
		Targets: []Target{
			{Name: "build", Line: 1, Dependencies: []string{"all"}, Commands: []string{"echo hello"}},
		},
	}
	score := calculateScore(result.Issues, result.Targets)
	if score != 100 {
		t.Errorf("expected score 100 with no issues, got %v", score)
	}

	// One error issue reduces score
	result2 := AnalysisResult{
		Issues: []Issue{
			{Severity: "error", RuleID: "test-rule"},
		},
		Targets: []Target{
			{Name: "build", Line: 1, Dependencies: []string{"all"}},
		},
	}
	score2 := calculateScore(result2.Issues, result2.Targets)
	if score2 >= 100 {
		t.Errorf("expected score < 100 with error issue, got %v", score2)
	}
}

func TestRuleStruct(t *testing.T) {
	rule := Rule{
		ID:          "test-rule",
		Name:        "Test Rule",
		Description: "A test rule",
		Severity:    "error",
		Category:    "best-practice",
	}
	if rule.ID != "test-rule" {
		t.Errorf("expected ID test-rule, got %s", rule.ID)
	}
	if rule.Severity != "error" {
		t.Errorf("expected severity error, got %s", rule.Severity)
	}
}

func TestIssueStruct(t *testing.T) {
	issue := Issue{
		RuleID:   "test-rule",
		Line:     5,
		Column:   10,
		Severity: "warning",
		Message:  "Test issue",
	}
	if issue.RuleID != "test-rule" {
		t.Errorf("expected RuleID test-rule, got %s", issue.RuleID)
	}
	if issue.Line != 5 {
		t.Errorf("expected Line 5, got %d", issue.Line)
	}
}

func TestAnalysisResultStruct(t *testing.T) {
	result := AnalysisResult{
		File:      "Makefile",
		Score:     85.0,
		Grade:     "B",
		Issues:    []Issue{},
		Targets:   []Target{},
		RuleCount: 10,
	}
	if result.File != "Makefile" {
		t.Errorf("expected File Makefile, got %s", result.File)
	}
	if result.Score != 85.0 {
		t.Errorf("expected Score 85.0, got %v", result.Score)
	}
	if result.Grade != "B" {
		t.Errorf("expected Grade B, got %s", result.Grade)
	}
}

func TestTargetStruct(t *testing.T) {
	target := Target{
		Name:         "test",
		Line:         10,
		Dependencies: []string{"all"},
		Commands:     []string{"@echo test"},
		IsPhony:      true,
	}
	if target.Name != "test" {
		t.Errorf("expected Name test, got %s", target.Name)
	}
	if len(target.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(target.Dependencies))
	}
	if !target.IsPhony {
		t.Error("expected IsPhony to be true")
	}
}

func TestPhonyTargets(t *testing.T) {
	// Common phony targets should be recognized
	if !phonyTargets["all"] {
		t.Error("expected 'all' to be a phony target")
	}
	if !phonyTargets["clean"] {
		t.Error("expected 'clean' to be a phony target")
	}
	if !phonyTargets["build"] {
		t.Error("expected 'build' to be a phony target")
	}
	// Random target should not be phony
	if phonyTargets["xyzzy"] {
		t.Error("expected 'xyzzy' to NOT be a phony target")
	}
}
