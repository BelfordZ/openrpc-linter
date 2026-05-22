package reporters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/open-rpc/openrpc-linter/types"
)

func TestTextReporterIncludesPathWhenPresent(t *testing.T) {
	var output bytes.Buffer
	reporter := &TextReporter{}

	err := reporter.Format([]types.RuleFunctionResult{
		{
			RuleID:   "method-description",
			Message:  "missing required field 'description'",
			Path:     []string{"$['methods'][0]['description']"},
			PathLabels: types.PathLabels{Method: "eth_getLogs"},
			Severity: types.SeverityError,
		},
	}, 1, &output)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, `method "eth_getLogs"`) {
		t.Fatalf("expected method label in output, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "methods[0].description") {
		t.Fatalf("expected friendly path in output, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "missing required field 'description'") {
		t.Fatalf("expected message on detail line, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "❌ 1 error(s) found in 1 rule") {
		t.Fatalf("expected error summary, got:\n%s", outputStr)
	}
}

func TestTextReporterWarnPrefixAndSummary(t *testing.T) {
	var output bytes.Buffer
	reporter := &TextReporter{}

	err := reporter.Format([]types.RuleFunctionResult{
		{
			RuleID:     "info-license",
			Message:    "missing required field 'license'",
			Path:       []string{"$['info']['license']"},
			PathLabels: types.PathLabels{Section: "info"},
			Severity:   types.SeverityWarn,
		},
	}, 1, &output)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "⚠️ info-license") {
		t.Fatalf("expected warning prefix, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "info.license") {
		t.Fatalf("expected friendly path, got:\n%s", outputStr)
	}
	if strings.Contains(outputStr, "❌") {
		t.Fatalf("expected no error prefix for warn severity, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "⚠️ 1 warning(s) found in 1 rule") {
		t.Fatalf("expected warning summary, got:\n%s", outputStr)
	}
}

func TestTextReporterSchemaLabel(t *testing.T) {
	var output bytes.Buffer
	reporter := &TextReporter{}

	err := reporter.Format([]types.RuleFunctionResult{
		{
			RuleID:     "schema-title",
			Message:    "missing required field 'title'",
			Path:       []string{"$['components']['schemas']['Pet']['title']"},
			PathLabels: types.PathLabels{Section: "components", Schema: "Pet"},
			Severity:   types.SeverityWarn,
		},
	}, 1, &output)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, `schema "Pet"`) {
		t.Fatalf("expected schema label, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "components.schemas.Pet.title") {
		t.Fatalf("expected friendly path, got:\n%s", outputStr)
	}
}

func TestTextReporterMixedSeveritySummary(t *testing.T) {
	var output bytes.Buffer
	reporter := &TextReporter{}

	err := reporter.Format([]types.RuleFunctionResult{
		{
			RuleID:   "method-description",
			Message:  "Missing description",
			Severity: types.SeverityError,
		},
		{
			RuleID:   "info-license",
			Message:  "Missing license",
			Severity: types.SeverityWarn,
		},
		{
			RuleID:   "info-license",
			Message:  "Also missing something else",
			Severity: types.SeverityWarn,
		},
	}, 3, &output)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	expected := "❌ 1 error(s), ⚠️ 2 warning(s) found in 2 rules"
	if !strings.Contains(output.String(), expected) {
		t.Fatalf("expected mixed summary %q, got:\n%s", expected, output.String())
	}
}

func TestTextReporterInfoPrefix(t *testing.T) {
	var output bytes.Buffer
	reporter := &TextReporter{}

	err := reporter.Format([]types.RuleFunctionResult{
		{
			RuleID:   "hint-rule",
			Message:  "Consider adding examples",
			Severity: types.SeverityInfo,
		},
	}, 1, &output)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if !strings.Contains(output.String(), "ℹ️ hint-rule: Consider adding examples") {
		t.Fatalf("expected info prefix, got:\n%s", output.String())
	}
	if !strings.Contains(output.String(), "ℹ️ 1 info found in 1 rule") {
		t.Fatalf("expected info summary, got:\n%s", output.String())
	}
}

func TestTextReporterDefaultsMissingSeverityToError(t *testing.T) {
	var output bytes.Buffer
	reporter := &TextReporter{}

	err := reporter.Format([]types.RuleFunctionResult{
		{
			RuleID:  "legacy-rule",
			Message: "Something failed",
		},
	}, 1, &output)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if !strings.Contains(output.String(), "❌ legacy-rule: Something failed") {
		t.Fatalf("expected default error prefix, got:\n%s", output.String())
	}
}
