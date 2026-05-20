package functions

import (
	"fmt"
	"reflect"

	"github.com/open-rpc/openrpc-linter/types"
)

type TruthyRule struct{}

func (r *TruthyRule) RunRule(value interface{}, context types.RuleFunctionContext) []types.RuleFunctionResult {
	var results []types.RuleFunctionResult

	fieldName := context.TargetField
	if fieldName != "" {
		parent := value
		if context.Parent != nil {
			parent = context.Parent
		}

		fieldValue, ok := objectFieldValue(parent, fieldName)
		if !ok && !isObject(parent) {
			return []types.RuleFunctionResult{{
				Message: fmt.Sprintf("truthy function expected object to read field '%s'", fieldName),
				Path:    resultPath(context.Path),
			}}
		}
		value = fieldValue
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
			message = "Missing required field '" + fieldName + "' at " + context.Path
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

func objectFieldValue(value interface{}, field string) (interface{}, bool) {
	if obj, ok := value.(map[string]interface{}); ok {
		val, exists := obj[field]
		return val, exists
	}

	obj := reflect.ValueOf(value)
	if obj.Kind() == reflect.Map && obj.Type().Key().Kind() == reflect.String {
		val := obj.MapIndex(reflect.ValueOf(field))
		if val.Kind() != reflect.Invalid {
			return val.Interface(), true
		}
	}

	return nil, false
}

func isObject(value interface{}) bool {
	if _, ok := value.(map[string]interface{}); ok {
		return true
	}

	obj := reflect.ValueOf(value)
	return obj.Kind() == reflect.Map && obj.Type().Key().Kind() == reflect.String
}
