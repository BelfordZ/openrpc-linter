package reporters

import (
	"fmt"
	"io"
	"strings"

	"github.com/open-rpc/openrpc-linter/types"
)

type TextReporter struct{}

func severityPrefix(severity types.Severity) string {
	switch severity {
	case types.SeverityWarn:
		return "⚠️"
	case types.SeverityInfo:
		return "ℹ️"
	default:
		return "❌"
	}
}

func effectiveSeverity(severity types.Severity) types.Severity {
	if severity == "" {
		return types.SeverityError
	}
	return severity
}

func (r *TextReporter) Format(results []types.RuleFunctionResult, totalRules int, output io.Writer) error {
	errorCount := 0
	warnCount := 0
	infoCount := 0
	ruleViolations := make(map[string][]types.RuleFunctionResult)

	for _, result := range results {
		if result.Message == "" {
			continue
		}
		ruleViolations[result.RuleID] = append(ruleViolations[result.RuleID], result)
		switch effectiveSeverity(result.Severity) {
		case types.SeverityWarn:
			warnCount++
		case types.SeverityInfo:
			infoCount++
		default:
			errorCount++
		}
	}

	for ruleID, ruleResults := range ruleViolations {
		for _, result := range ruleResults {
			prefix := severityPrefix(effectiveSeverity(result.Severity))
			if len(result.Path) == 0 {
				if _, err := fmt.Fprintf(output, "%s %s: %s\n", prefix, ruleID, result.Message); err != nil {
					return err
				}
				continue
			}

			if _, err := fmt.Fprintf(output, "%s %s at %s: %s\n", prefix, ruleID, result.Path[0], result.Message); err != nil {
				return err
			}
		}
	}

	violationCount := errorCount + warnCount + infoCount
	if violationCount == 0 {
		if _, err := fmt.Fprintf(output, "\n✅ All %d rules passed!\n", totalRules); err != nil {
			return err
		}
		return nil
	}

	rulesWithViolations := len(ruleViolations)
	summary := formatViolationSummary(errorCount, warnCount, infoCount, rulesWithViolations)
	if _, err := fmt.Fprintf(output, "\n%s\n", summary); err != nil {
		return err
	}

	return nil
}

func formatViolationSummary(errorCount, warnCount, infoCount, rulesWithViolations int) string {
	var parts []string
	if errorCount > 0 {
		parts = append(parts, fmt.Sprintf("❌ %d error(s)", errorCount))
	}
	if warnCount > 0 {
		parts = append(parts, fmt.Sprintf("⚠️ %d warning(s)", warnCount))
	}
	if infoCount > 0 {
		parts = append(parts, fmt.Sprintf("ℹ️ %d info", infoCount))
	}

	ruleLabel := "rules"
	if rulesWithViolations == 1 {
		ruleLabel = "rule"
	}
	return fmt.Sprintf("%s found in %d %s", strings.Join(parts, ", "), rulesWithViolations, ruleLabel)
}
