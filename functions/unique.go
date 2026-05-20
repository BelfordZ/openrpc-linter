package functions

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/open-rpc/openrpc-linter/types"
)

type UniqueRule struct{}

func (r *UniqueRule) RunRule(value interface{}, context types.RuleFunctionContext) []types.RuleFunctionResult {
	items, ok := uniqueItems(value, context.Path)
	if !ok {
		return []types.RuleFunctionResult{{
			Message: "unique function requires array input",
			Path:    resultPath(context.Path),
		}}
	}

	ignoreMissing := true
	if context.Rule.Then.FunctionOptions != nil {
		rawIgnoreMissing, exists := context.Rule.Then.FunctionOptions["ignoreMissing"]
		if exists {
			parsed, ok := rawIgnoreMissing.(bool)
			if !ok {
				return []types.RuleFunctionResult{{
					Message: "unique function option ignoreMissing must be a boolean",
					Path:    resultPath(context.Path),
				}}
			}
			ignoreMissing = parsed
		}
	}

	seen := make(map[string]struct{})
	var results []types.RuleFunctionResult

	for _, item := range items {
		compareValue := item.Value
		comparePath := item.Path
		if context.TargetField != "" {
			fieldValue, exists := objectFieldValue(item.Value, context.TargetField)
			if !exists && !isObject(item.Value) {
				results = append(results, types.RuleFunctionResult{
					Message: fmt.Sprintf("unique function expected object items to read field '%s'", context.TargetField),
					Path:    resultPath(item.Path),
				})
				continue
			}
			if !exists && ignoreMissing {
				continue
			}
			if !exists {
				fieldValue = nil
			}
			compareValue = fieldValue
			comparePath = fieldPath(item.Path, context.TargetField)
		}

		key, displayValue, supported := comparableKey(compareValue)
		if !supported {
			message := "unique function does not support non-primitive values"
			if context.TargetField != "" {
				message = fmt.Sprintf("unique function does not support non-primitive value for field '%s'", context.TargetField)
			}
			results = append(results, types.RuleFunctionResult{
				Message: message,
				Path:    resultPath(comparePath),
			})
			continue
		}

		if _, found := seen[key]; found {
			message := fmt.Sprintf("Duplicate value: %s", displayValue)
			if context.TargetField != "" {
				message = fmt.Sprintf("Duplicate value for field '%s': %s", context.TargetField, displayValue)
			}
			results = append(results, types.RuleFunctionResult{
				Message: message,
				Path:    resultPath(comparePath),
			})
			continue
		}

		seen[key] = struct{}{}
	}

	return results
}

type uniqueItem struct {
	Value interface{}
	Path  string
}

func uniqueItems(value interface{}, basePath string) ([]uniqueItem, bool) {
	switch v := value.(type) {
	case []interface{}:
		items := make([]uniqueItem, 0, len(v))
		for i, item := range v {
			items = append(items, uniqueItem{
				Value: item,
				Path:  fmt.Sprintf("%s[%d]", basePath, i),
			})
		}
		return items, true
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		items := make([]uniqueItem, 0, len(v))
		for _, key := range keys {
			items = append(items, uniqueItem{
				Value: v[key],
				Path:  basePath + pathNameSegment(key),
			})
		}
		return items, true
	default:
		return nil, false
	}
}

func comparableKey(value interface{}) (key string, displayValue string, supported bool) {
	switch v := value.(type) {
	case nil:
		return "null", "null", true
	case string:
		return "string:" + v, strconv.Quote(v), true
	case bool:
		if v {
			return "bool:true", "true", true
		}
		return "bool:false", "false", true
	case float64:
		return fmt.Sprintf("number:%g", v), fmt.Sprintf("%g", v), true
	default:
		return "", "", false
	}
}
