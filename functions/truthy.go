package functions

import (
	"fmt"

	"github.com/open-rpc/openrpc-linter/types"
)

type TruthyRule struct{}

func (r *TruthyRule) RunRule(value interface{}, context types.RuleFunctionContext) []types.RuleFunctionResult {
	var results []types.RuleFunctionResult

	fieldName, fieldPath, fieldConfigResults := truthyField(context)
	if len(fieldConfigResults) > 0 {
		return fieldConfigResults
	}
	if fieldName != "" {
		itemMap, ok := value.(map[string]interface{})
		if !ok {
			return []types.RuleFunctionResult{{
				Message: fmt.Sprintf("truthy function expected object to read field '%s'", fieldName),
				Path:    resultPath(context.Path),
			}}
		}
		value = itemMap[fieldName]
	}

	isTruthy := true

	if value == nil {
		isTruthy = false
	} else if str, ok := value.(string); ok && str == "" {
		isTruthy = false
	} else if str, ok := value.(string); ok && str == "null" {
		isTruthy = false
	}

	if !isTruthy {
		var message string
		path := resultPath(context.Path)
		if fieldName != "" {
			message = "Missing required field '" + fieldName + "' at " + fieldPath
			path = resultPath(fieldPath)
		} else {
			message = "Field must have a truthy value"
		}

		results = append(results, types.RuleFunctionResult{
			Message: message,
			Path:    path,
		})
	}

	return results
}

func truthyField(context types.RuleFunctionContext) (string, string, []types.RuleFunctionResult) {
	if context.Rule == nil || context.Rule.Then == nil || context.Rule.Then.FunctionOptions == nil {
		return "", "", nil
	}

	rawField, exists := context.Rule.Then.FunctionOptions["field"]
	if !exists {
		return "", "", nil
	}

	fieldName, ok := rawField.(string)
	if !ok || fieldName == "" {
		return "", "", []types.RuleFunctionResult{{
			Message: "truthy function option field must be a non-empty string",
			Path:    resultPath(context.Path),
		}}
	}

	return fieldName, fieldPath(context.Path, fieldName), nil
}
