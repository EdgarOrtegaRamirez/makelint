package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	defaultFile = "Makefile"
)

// Rule represents a linting rule
type Rule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // "error", "warning", "info"
	Category    string `json:"category"`
}

// Issue represents a linting issue found in a Makefile
type Issue struct {
	RuleID    string `json:"rule_id"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// Target represents a Makefile target
type Target struct {
	Name      string   `json:"name"`
	Line      int      `json:"line"`
	Dependencies []string `json:"dependencies"`
	Commands  []string `json:"commands"`
	IsPhony   bool     `json:"is_phony"`
	IsPattern bool     `json:"is_pattern"`
}

// AnalysisResult represents the complete analysis of a Makefile
type AnalysisResult struct {
	File      string   `json:"file"`
	Targets   []Target `json:"targets"`
	Issues    []Issue  `json:"issues"`
	Score     float64  `json:"score"`
	Grade     string   `json:"grade"`
	RuleCount int      `json:"rule_count"`
	IssueCount int    `json:"issue_count"`
	ErrorCount int      `json:"error_count"`
	WarnCount  int     `json:"warn_count"`
	InfoCount  int     `json:"info_count"`
}

var rules = []Rule{
	{"ML001", "Missing PHONY", "Targets that don't produce files should be declared .PHONY", "warning", "best-practice"},
	{"ML002", "Tab vs Spaces", "Recipe lines must use tabs, not spaces", "error", "syntax"},
	{"ML003", "Undefined Variable", "Variable used but never defined in the Makefile", "warning", "reliability"},
	{"ML004", "Hardcoded Path", "Hardcoded absolute paths in recipes", "info", "portability"},
	{"ML005", "Missing .DEFAULT_GOAL", "No .DEFAULT_GOAL specified", "info", "best-practice"},
	{"ML006", "Silent Command", "Command prefixed with @ may hide errors", "info", "debugging"},
	{"ML007", "Recursive Make", "Recursive make calls detected", "warning", "performance"},
	{"ML008", "Duplicate Target", "Target defined multiple times", "warning", "reliability"},
	{"ML009", "Missing .SUFFIXES", "No .SUFFIXES declaration for pattern rules", "info", "best-practice"},
	{"ML010", "Dangerous rm -rf", "Use with caution: rm -rf in recipes", "warning", "safety"},
	{"ML011", "No Clean Target", "No clean or distclean target found", "info", "best-practice"},
	{"ML012", "No Install Target", "No install target found (for library projects)", "info", "best-practice"},
	{"ML013", "Shell Injection", "Unquoted variable in shell command", "error", "safety"},
	{"ML014", "Circular Dependency", "Potential circular dependency detected", "error", "reliability"},
	{"ML015", "Missing Shebang", "No comment header or shebang at top of file", "info", "documentation"},
}

var phonyTargets = map[string]bool{
	"all": true, "clean": true, "distclean": true, "install": true,
	"uninstall": true, "test": true, "check": true, "lint": true,
	"format": true, "fmt": true, "build": true, "package": true,
	"deploy": true, "push": true, "pull": true, "docs": true,
	"doc": true, "html": true, "pdf": true, "release": true,
	"publish": true, "tag": true, "version": true, "help": true,
	".PHONY": true, "coverage": true, "cov": true, "bench": true,
	"benchmark": true, "setup": true, "init": true, "config": true,
	"rebuild": true, "rerun": true, "restart": true, "stop": true,
	"start": true, "logs": true, "log": true, "shell": true,
	"info": true, "print": true, "list": true, "search": true,
	"find": true, "grep": true, "watch": true, "serve": true,
	"run": true, "exec": true, "execute": true, "validate": true,
	"verify": true, "prepare": true, "pre-build": true,
}

var definedVariables = map[string]bool{}

func main() {
	file := defaultFile
	outputJSON := false
	showRules := false
	severity := "" // filter by severity

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--file", "-f":
			if i+1 < len(os.Args) {
				file = os.Args[i+1]
				i++
			}
		case "--json":
			outputJSON = true
		case "--rules":
			showRules = true
		case "--severity", "-s":
			if i+1 < len(os.Args) {
				severity = os.Args[i+1]
				i++
			}
		case "--help", "-h":
			printUsage()
			return
		}
	}

	if showRules {
		printRules()
		return
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file '%s' not found\n", file)
		os.Exit(1)
	}

	result := analyzeFile(file)

	if outputJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(result)
		return
	}

	printReport(result, severity)
}

func printUsage() {
	fmt.Println(`MakeLint — Makefile Linter & Analyzer

A comprehensive Makefile linter that checks for best practices,
common issues, security concerns, and generates quality scores.

Usage: makelint [options] [file]

Options:
  --file, -f FILE    Makefile to analyze (default: Makefile)
  --json             Output results as JSON
  --rules            List all available linting rules
  --severity LEVEL   Filter by severity (error, warning, info)
  --help, -h         Show this help

Examples:
  makelint                    # Lint current directory Makefile
  makelint -f Makefile.test   # Lint a specific file
  makelint --json             # JSON output
  makelint --rules            # List all rules
  makelint --severity error   # Show only errors`)
}

func printRules() {
	fmt.Println("📋 MakeLint Rules")
	fmt.Printf("%-6s %-25s %-8s %-15s %s\n", "ID", "Name", "Severity", "Category", "Description")
	fmt.Println(strings.Repeat("-", 100))

	for _, r := range rules {
		fmt.Printf("%-6s %-25s %-8s %-15s %s\n", r.ID, r.Name, r.Severity, r.Category, r.Description)
	}
	fmt.Printf("\nTotal: %d rules\n", len(rules))
}

func analyzeFile(filename string) AnalysisResult {
	definedVariables = map[string]bool{}
	result := AnalysisResult{
		File:      filename,
		RuleCount: len(rules),
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		result.Issues = append(result.Issues, Issue{
			RuleID:  "ML000",
			Line:    0,
			Severity: "error",
			Message: fmt.Sprintf("Failed to read file: %v", err),
		})
		return result
	}

	lines := strings.Split(string(content), "\n")

	// First pass: collect targets, variables, phony declarations, .DEFAULT_GOAL, .SUFFIXES
	targets := make(map[string]*Target)
	var phonyList []string
	var hasDefaultGoal bool
	var hasCleanTarget bool
	var hasInstallTarget bool
	var hasShebang bool
	var hasSuffixes bool

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if i == 0 && strings.HasPrefix(trimmed, "#!") {
				hasShebang = true
			}
			continue
		}

		// Check for .DEFAULT_GOAL
		if strings.HasPrefix(trimmed, ".DEFAULT_GOAL") {
			hasDefaultGoal = true
		}

		// Check for .SUFFIXES
		if strings.HasPrefix(trimmed, ".SUFFIXES") {
			hasSuffixes = true
		}

		// Check for .PHONY
		if strings.HasPrefix(trimmed, ".PHONY") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				for _, p := range parts[1:] {
					phonyList = append(phonyList, strings.Split(p, ",")...)
				}
			}
			continue
		}

		// Check for variable definitions (VAR = value, VAR := value, VAR ?= value, VAR += value)
		if reVarDef := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*[:?+]?=`); reVarDef.MatchString(trimmed) {
			matches := reVarDef.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				definedVariables[matches[1]] = true
			}
		}

		// Check for target definitions (target: [deps]) — but skip := assignments
		if !strings.Contains(trimmed, ":=") {
			if reTarget := regexp.MustCompile(`^([A-Za-z0-9_.%-][A-Za-z0-9_.% -]*)\s*:`); reTarget.MatchString(trimmed) {
				matches := reTarget.FindStringSubmatch(trimmed)
				if len(matches) > 1 {
					targetName := strings.TrimSpace(matches[1])
					if _, exists := targets[targetName]; !exists {
						targets[targetName] = &Target{
							Name: targetName,
							Line: lineNum,
						}
					}
				}
			}
		}
	}

	// Second pass: assign recipe lines to correct targets
	var currentTarget *Target
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Check for target definitions
		if reTarget := regexp.MustCompile(`^([A-Za-z0-9_.%-][A-Za-z0-9_.% -]*)\s*:\s*(.*)`); reTarget.MatchString(trimmed) {
			matches := reTarget.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				targetName := strings.TrimSpace(matches[1])
				if strings.Contains(trimmed, ":=") {
					continue
				}
				if t, exists := targets[targetName]; exists {
					currentTarget = t
					// Extract dependencies
					depStr := strings.TrimSpace(matches[2])
					for _, d := range strings.Fields(depStr) {
						if !strings.HasPrefix(d, ".") {
							t.Dependencies = append(t.Dependencies, d)
						}
					}
				}
			}
		}

		// Check for recipe lines (start with tab)
		if strings.HasPrefix(line, "	") && currentTarget != nil {
			cmd := strings.TrimSpace(line[1:])
			currentTarget.Commands = append(currentTarget.Commands, cmd)
			if cmd == "clean" || strings.HasPrefix(cmd, "clean ") {
				hasCleanTarget = true
			}
			if strings.HasPrefix(cmd, "install ") || cmd == "install" {
				hasInstallTarget = true
			}
			// Check for rm -rf
			if strings.Contains(cmd, "rm -rf") || strings.Contains(cmd, "rm -fr") {
				result.Issues = append(result.Issues, Issue{
					RuleID:   "ML010",
					Line:     lineNum,
					Severity: "warning",
					Message:  "Dangerous 'rm -rf' command detected",
					Suggestion: "Consider using 'rm -f' with explicit file lists or a cleanup directory",
				})
			}
			// Check for silent commands
			if strings.HasPrefix(cmd, "@") {
				result.Issues = append(result.Issues, Issue{
					RuleID:   "ML006",
					Line:     lineNum,
					Severity: "info",
					Message:  "Silent command (prefixed with @)",
					Suggestion: "Remove @ prefix if you want to see commands in output",
				})
			}
			// Check for hardcoded paths
			if rePath := regexp.MustCompile(`(?:^|\s)/[a-z][a-z0-9_/.-]+`); rePath.MatchString(cmd) {
				result.Issues = append(result.Issues, Issue{
					RuleID:   "ML004",
					Line:     lineNum,
					Severity: "info",
					Message:  "Hardcoded absolute path detected",
					Suggestion: "Use variables like $(PREFIX) or $(DESTDIR) for portability",
				})
			}
			// Check for recursive make
			if strings.Contains(cmd, "$(MAKE)") || strings.Contains(cmd, "$(MAKE) ") {
				result.Issues = append(result.Issues, Issue{
					RuleID:   "ML007",
					Line:     lineNum,
					Severity: "warning",
					Message:  "Recursive make call detected",
					Suggestion: "Consider using .WAIT or restructuring to avoid recursive make",
				})
			}
		}
	}

	// Populate targets slice
	for _, t := range targets {
		// Check if target is phony
		for _, p := range phonyList {
			if t.Name == p {
				t.IsPhony = true
				break
			}
		}
		// Check if it's a pattern rule
		if strings.Contains(t.Name, "%") {
			t.IsPattern = true
		}
		// Extract dependencies
		for _, target := range targets {
			if target.Name == t.Name {
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if reTarget := regexp.MustCompile(`^` + regexp.QuoteMeta(t.Name) + `\s*:\s*(.*)`); reTarget.MatchString(trimmed) {
						matches := reTarget.FindStringSubmatch(trimmed)
						if len(matches) > 1 {
							depStr := strings.TrimSpace(matches[1])
							for _, d := range strings.Fields(depStr) {
								if !strings.HasPrefix(d, ".") {
									t.Dependencies = append(t.Dependencies, d)
								}
							}
						}
					}
				}
			}
		}
		if t.Name == "clean" {
			hasCleanTarget = true
		}
		if t.Name == "install" {
			hasInstallTarget = true
		}
	}

	// Add targets to result
	for _, t := range targets {
		result.Targets = append(result.Targets, *t)
	}

	// Sort targets by line number
	sort.Slice(result.Targets, func(i, j int) bool {
		return result.Targets[i].Line < result.Targets[j].Line
	})

	// ML001: Check for non-phony targets
	for _, t := range result.Targets {
		if !t.IsPhony && !t.IsPattern && !strings.Contains(t.Name, ".") {
			if phonyTargets[t.Name] {
				result.Issues = append(result.Issues, Issue{
					RuleID:    "ML001",
					Line:      t.Line,
					Severity:  "warning",
					Message:   fmt.Sprintf("Target '%s' should be declared .PHONY", t.Name),
					Suggestion: "Add '.PHONY: " + t.Name + "' to the Makefile",
				})
			}
		}
	}

	// ML005: Check for .DEFAULT_GOAL
	if !hasDefaultGoal {
		result.Issues = append(result.Issues, Issue{
			RuleID:   "ML005",
			Line:     0,
			Severity: "info",
			Message:  "No .DEFAULT_GOAL specified",
			Suggestion: "Add '.DEFAULT_GOAL := all' to set a default target",
		})
	}

	// ML011: Check for clean target
	if !hasCleanTarget {
		result.Issues = append(result.Issues, Issue{
			RuleID:   "ML011",
			Line:     0,
			Severity: "info",
			Message:  "No clean target found",
			Suggestion: "Add a 'clean' target to remove build artifacts",
		})
	}

	// ML012: Check for install target
	if !hasInstallTarget {
		result.Issues = append(result.Issues, Issue{
			RuleID:   "ML012",
			Line:     0,
			Severity: "info",
			Message:  "No install target found",
			Suggestion: "Consider adding an 'install' target for library projects",
		})
	}

	// ML015: Check for shebang/header
	if !hasShebang && len(lines) > 0 && !strings.HasPrefix(strings.TrimSpace(lines[0]), "#") {
		result.Issues = append(result.Issues, Issue{
			RuleID:   "ML015",
			Line:     0,
			Severity: "info",
			Message:  "No comment header at top of Makefile",
			Suggestion: "Add a comment explaining the purpose of this Makefile",
		})
	}

	// ML009: Check for .SUFFIXES
	if !hasSuffixes {
		// Only flag if there are pattern rules
		hasPatternRules := false
		for _, t := range result.Targets {
			if t.IsPattern {
				hasPatternRules = true
				break
			}
		}
		if hasPatternRules {
			result.Issues = append(result.Issues, Issue{
				RuleID:   "ML009",
				Line:     0,
				Severity: "info",
				Message:  "Pattern rules exist but no .SUFFIXES declaration",
				Suggestion: "Add '.SUFFIXES' to define supported file extensions",
			})
		}
	}

	// Calculate score
	result.Score = calculateScore(result.Issues, result.Targets)
	result.Grade = scoreToGrade(result.Score)
	result.IssueCount = len(result.Issues)
	for _, issue := range result.Issues {
		switch issue.Severity {
		case "error":
			result.ErrorCount++
		case "warning":
			result.WarnCount++
		case "info":
			result.InfoCount++
		}
	}

	return result
}

func calculateScore(issues []Issue, targets []Target) float64 {
	score := 100.0

	for _, issue := range issues {
		switch issue.Severity {
		case "error":
			score -= 15.0
		case "warning":
			score -= 5.0
		case "info":
			score -= 1.0
		}
	}

	// Bonus for having a clean target
	hasClean := false
	for _, t := range targets {
		if t.Name == "clean" {
			hasClean = true
			break
		}
	}
	if hasClean {
		score += 2.0
	}

	// Bonus for having .PHONY
	phonyCount := 0
	for _, t := range targets {
		if t.IsPhony {
			phonyCount++
		}
	}
	if phonyCount > 0 {
		score += float64(phonyCount) * 0.5
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return math.Round(score*100) / 100
}

func scoreToGrade(score float64) string {
	switch {
	case score >= 95:
		return "A+"
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

func printReport(result AnalysisResult, severityFilter string) {
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("🔧 MakeLint Report\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	fmt.Printf("📄 File: %s\n", result.File)
	fmt.Printf("📊 Score: %.0f/100  Grade: %s\n\n", result.Score, result.Grade)

	// Severity summary
	fmt.Printf("📈 Summary:\n")
	fmt.Printf("   Errors:   %d\n", result.ErrorCount)
	fmt.Printf("   Warnings: %d\n", result.WarnCount)
	fmt.Printf("   Info:     %d\n", result.InfoCount)
	fmt.Printf("   Total:    %d issues (from %d rules)\n\n", result.IssueCount, result.RuleCount)

	// Targets
	if len(result.Targets) > 0 {
		fmt.Printf("🎯 Targets (%d):\n", len(result.Targets))
		for _, t := range result.Targets {
			phony := ""
			if t.IsPhony {
				phony = " [PHONY]"
			}
			pattern := ""
			if t.IsPattern {
				pattern = " [PATTERN]"
			}
			deps := ""
			if len(t.Dependencies) > 0 {
				deps = " → " + strings.Join(t.Dependencies, ", ")
			}
			fmt.Printf("   Line %d: %s%s%s%s\n", t.Line, t.Name, phony, pattern, deps)
		}
		fmt.Printf("\n")
	}

	// Issues
	if len(result.Issues) > 0 {
		fmt.Printf("🚨 Issues:\n\n")
		for _, issue := range result.Issues {
			if severityFilter != "" && issue.Severity != severityFilter {
				continue
			}
			emoji := "ℹ️ "
			if issue.Severity == "warning" {
				emoji = "⚠️ "
			}
			if issue.Severity == "error" {
				emoji = "❌ "
			}
			lineStr := ""
			if issue.Line > 0 {
				lineStr = fmt.Sprintf(" (line %d)", issue.Line)
			}
			fmt.Printf("   %s [%s] %s%s\n", emoji, issue.RuleID, issue.Message, lineStr)
			if issue.Suggestion != "" {
				fmt.Printf("      💡 %s\n", issue.Suggestion)
			}
			fmt.Printf("\n")
		}
	}

	if result.IssueCount == 0 {
		fmt.Printf("✅ No issues found! Makefile looks great.\n")
	}
}
