package reporters

import (
	"fmt"
	"strings"

	"github.com/open-rpc/openrpc-linter/location"
	"github.com/open-rpc/openrpc-linter/types"
)

func formatPathLabels(labels types.PathLabels) string {
	var parts []string
	if labels.Method != "" {
		parts = append(parts, fmt.Sprintf(`method %q`, labels.Method))
	}
	if labels.Param != "" {
		parts = append(parts, fmt.Sprintf(`param %q`, labels.Param))
	}
	if labels.Schema != "" {
		parts = append(parts, fmt.Sprintf(`schema %q`, labels.Schema))
	}
	if labels.Descriptor != "" {
		parts = append(parts, fmt.Sprintf(`descriptor %q`, labels.Descriptor))
	}
	if labels.Tag != "" {
		parts = append(parts, fmt.Sprintf(`tag %q`, labels.Tag))
	}
	if labels.Section != "" && labels.Method == "" && labels.Schema == "" && labels.Descriptor == "" {
		parts = append(parts, labels.Section)
	}
	return strings.Join(parts, "  ")
}

func violationLocation(result types.RuleFunctionResult) string {
	var parts []string
	if labelStr := formatPathLabels(result.PathLabels); labelStr != "" {
		parts = append(parts, labelStr)
	}
	if len(result.Path) > 0 {
		parts = append(parts, location.FriendlyPath(result.Path[0]))
	}
	return strings.Join(parts, "  ")
}
