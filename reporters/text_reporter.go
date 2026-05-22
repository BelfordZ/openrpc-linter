package reporters

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/open-rpc/openrpc-linter/location"
	"github.com/open-rpc/openrpc-linter/types"
)

// TextReporter renders violations grouped by their most-specific document
// anchor (method, components schema/descriptor/tag, etc.). SourceFile is
// optional and, when set, prints once as a header before the rows.
type TextReporter struct {
	SourceFile string
}

func effectiveSeverity(severity types.Severity) types.Severity {
	if severity == "" {
		return types.SeverityError
	}
	return severity
}

// row holds the precomputed cells for one violation so we can size columns
// before printing.
type row struct {
	path       string
	severity   types.Severity
	message    string
	ruleID     string
	pathLabels types.PathLabels
	groupKind  string
	groupName  string
}

// groupOrder ranks group kinds for stable rendering: methods first, then
// components/* (schema, descriptor, tag), then other sections, general last.
var groupOrder = map[string]int{
	groupKindMethod:     0,
	groupKindSchema:     1,
	groupKindDescriptor: 2,
	groupKindTag:        3,
	groupKindSection:    4,
	groupKindGeneral:    5,
}

type bucket struct {
	kind     string
	name     string
	rows     []row
	methodIx int // smallest methods[N] index across the rows (groupKindMethod only)
}

func (r *TextReporter) Format(results []types.RuleFunctionResult, totalRules int, output io.Writer) error {
	colorEnabled := supportsColor(output)

	type pending struct {
		row       row
		methodIdx int
	}

	pendings := make([]pending, 0, len(results))
	errorCount, warnCount, infoCount := 0, 0, 0
	ruleViolations := make(map[string]struct{})

	for _, res := range results {
		if res.Message == "" {
			continue
		}
		ruleViolations[res.RuleID] = struct{}{}
		switch effectiveSeverity(res.Severity) {
		case types.SeverityWarn:
			warnCount++
		case types.SeverityInfo:
			infoCount++
		default:
			errorCount++
		}

		friendlyPath := ""
		mIdx := -1
		if len(res.Path) > 0 {
			friendlyPath = truncate(location.FriendlyPath(res.Path[0]), colPath)
			mIdx = methodIndexFromPath(res.Path[0])
		}
		kind, name := resolveGroup(res.PathLabels)
		pendings = append(pendings, pending{
			row: row{
				path:       friendlyPath,
				severity:   res.Severity,
				message:    truncate(res.Message, colMessage),
				ruleID:     truncate(res.RuleID, colRule),
				pathLabels: res.PathLabels,
				groupKind:  kind,
				groupName:  name,
			},
			methodIdx: mIdx,
		})
	}

	if len(pendings) == 0 {
		if _, err := fmt.Fprintf(output, "All %d rules passed\n", totalRules); err != nil {
			return err
		}
		return nil
	}

	// Bucket rows by (kind, name), tracking the smallest method index per
	// bucket so method groups can be ordered by document position.
	buckets := make(map[string]*bucket)
	keys := make([]string, 0)
	for _, p := range pendings {
		key := p.row.groupKind + "\x00" + p.row.groupName
		b, ok := buckets[key]
		if !ok {
			b = &bucket{kind: p.row.groupKind, name: p.row.groupName, methodIx: -1}
			buckets[key] = b
			keys = append(keys, key)
		}
		b.rows = append(b.rows, p.row)
		if p.methodIdx >= 0 && (b.methodIx < 0 || p.methodIdx < b.methodIx) {
			b.methodIx = p.methodIdx
		}
	}

	// Sort group keys by (kind rank, secondary key).
	sort.SliceStable(keys, func(i, j int) bool {
		bi, bj := buckets[keys[i]], buckets[keys[j]]
		ri, rj := groupOrder[bi.kind], groupOrder[bj.kind]
		if ri != rj {
			return ri < rj
		}
		if bi.kind == groupKindMethod {
			// Both -1 fall through to name; otherwise smaller doc index first.
			if bi.methodIx != bj.methodIx {
				if bi.methodIx < 0 {
					return false
				}
				if bj.methodIx < 0 {
					return true
				}
				return bi.methodIx < bj.methodIx
			}
		}
		return bi.name < bj.name
	})

	// Stable within-group sort: path → ruleID → message.
	for _, k := range keys {
		rows := buckets[k].rows
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].path != rows[j].path {
				return rows[i].path < rows[j].path
			}
			if rows[i].ruleID != rows[j].ruleID {
				return rows[i].ruleID < rows[j].ruleID
			}
			return rows[i].message < rows[j].message
		})
	}

	// Compute path column width across ALL rows so every group lines up.
	allPaths := make([]string, 0, len(pendings))
	for _, p := range pendings {
		allPaths = append(allPaths, p.row.path)
	}
	pathW := colWidth(allPaths, colPath)
	// Header word "path" is 4 chars; never shrink below that or the header
	// row would overflow its cell.
	if pathW < len("path") {
		pathW = len("path")
	}

	summary := formatViolationSummary(errorCount, warnCount, infoCount, len(ruleViolations))

	if r.SourceFile != "" {
		if _, err := fmt.Fprintf(output, "%s\n\n", r.SourceFile); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(output, "%s\n\n", summary); err != nil {
		return err
	}

	if _, err := io.WriteString(output, formatColumnHeaderRow(pathW)); err != nil {
		return err
	}

	for _, k := range keys {
		b := buckets[k]
		if _, err := fmt.Fprintf(output, "\n%s\n", formatGroupHeader(b.kind, b.name)); err != nil {
			return err
		}
		for _, row := range b.rows {
			pathCell := styledCell(row.path, pathW, ansiDarkGrey, colorEnabled)
			sevCell := formatSeverityCol(row.severity, colorEnabled)
			msgCell := styledCell(row.message, colMessage, ansiLightGrey, colorEnabled)
			ruleCell := colorize(row.ruleID, ansiDarkGrey, colorEnabled)
			if _, err := io.WriteString(output, formatViolationLine(pathCell, sevCell, msgCell, ruleCell)); err != nil {
				return err
			}
			for _, sec := range formatSecondaryLabels(row.pathLabels, row.groupKind) {
				styled := colorize(truncate(sec, colMessage), ansiDarkGrey, colorEnabled)
				if _, err := fmt.Fprintf(output, "    %s\n", styled); err != nil {
					return err
				}
			}
		}
	}

	if _, err := fmt.Fprintf(output, "\n%s\n", summary); err != nil {
		return err
	}
	return nil
}

func formatViolationSummary(errorCount, warnCount, infoCount, rulesWithViolations int) string {
	var parts []string
	if errorCount > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", errorCount, plural(errorCount, "error", "errors")))
	}
	if warnCount > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", warnCount, plural(warnCount, "warning", "warnings")))
	}
	if infoCount > 0 {
		parts = append(parts, fmt.Sprintf("%d info", infoCount))
	}

	ruleLabel := "rules"
	if rulesWithViolations == 1 {
		ruleLabel = "rule"
	}
	return fmt.Sprintf("%s found in %d %s", strings.Join(parts, ", "), rulesWithViolations, ruleLabel)
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return singular
	}
	return pluralForm
}
