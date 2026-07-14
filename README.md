# MakeLint 🔧

A comprehensive Makefile linter and analyzer CLI tool written in Go. MakeLint parses Makefiles and checks for best practices, common issues, security concerns, and generates quality scores with actionable suggestions.

## Features

- **15 built-in linting rules** covering best practices, syntax, reliability, safety, and documentation
- **Score & Grade**: 0-100 score with A+ to F grading
- **JSON output**: CI/CD-friendly machine-readable reports
- **Severity filtering**: Filter issues by error, warning, or info
- **Target analysis**: Extracts targets, dependencies, and phony declarations
- **Rule listing**: View all available linting rules with `--rules`

## Linting Rules

| ID | Name | Severity | Category | Description |
|----|------|----------|----------|-------------|
| ML001 | Missing PHONY | warning | best-practice | Targets that don't produce files should be declared .PHONY |
| ML002 | Tab vs Spaces | error | syntax | Recipe lines must use tabs, not spaces |
| ML003 | Undefined Variable | warning | reliability | Variable used but never defined in the Makefile |
| ML004 | Hardcoded Path | info | portability | Hardcoded absolute paths in recipes |
| ML005 | Missing .DEFAULT_GOAL | info | best-practice | No .DEFAULT_GOAL specified |
| ML006 | Silent Command | info | debugging | Command prefixed with @ may hide errors |
| ML007 | Recursive Make | warning | performance | Recursive make calls detected |
| ML008 | Duplicate Target | warning | reliability | Target defined multiple times |
| ML009 | Missing .SUFFIXES | info | best-practice | No .SUFFIXES declaration for pattern rules |
| ML010 | Dangerous rm -rf | warning | safety | Use with caution: rm -rf in recipes |
| ML011 | No Clean Target | info | best-practice | No clean or distclean target found |
| ML012 | No Install Target | info | best-practice | No install target found (for library projects) |
| ML013 | Shell Injection | error | safety | Unquoted variable in shell command |
| ML014 | Circular Dependency | error | reliability | Potential circular dependency detected |
| ML015 | Missing Shebang | info | documentation | No comment header or shebang at top of file |

## Installation

```bash
go install github.com/EdgarOrtegaRamirez/makelint@latest
```

Or build from source:

```bash
git clone https://github.com/EdgarOrtegaRamirez/makelint.git
cd makelint
go build -o makelint
```

## Usage

```bash
# Lint the current directory's Makefile
makelint

# Lint a specific file
makelint -f Makefile.test

# Output as JSON (CI/CD friendly)
makelint --json

# List all linting rules
makelint --rules

# Filter by severity
makelint --severity error

# Show help
makelint --help
```

## Example Output

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔧 MakeLint Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📄 File: Makefile
📊 Score: 85/100  Grade: B

📈 Summary:
   Errors:   0
   Warnings: 2
   Info:     3
   Total:    5 issues (from 15 rules)

🎯 Targets (4):
   Line 1: all [PHONY]
   Line 5: build
   Line 10: test [PHONY]
   Line 15: clean [PHONY]

🚨 Issues:

   ⚠️  [ML001] Target 'build' should be declared .PHONY
      💡 Add '.PHONY: build' to the Makefile

   ⚠️  [ML007] Recursive make call detected
      💡 Consider using .WAIT or restructuring to avoid recursive make

   ℹ️  [ML005] No .DEFAULT_GOAL specified
      💡 Add '.DEFAULT_GOAL := all' to set a default target
```

## License

MIT - see [LICENSE](LICENSE) for details.
