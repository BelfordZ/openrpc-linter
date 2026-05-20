package functions

import (
	"reflect"

	"github.com/open-rpc/openrpc-linter/types"
	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
)

type TruthyRule struct{}

func (r *TruthyRule) RunRule(value interface{}, context types.RuleFunctionContext) []types.RuleFunctionResult {
	if context.Path == "" && context.GivenPath != nil {
		if results, ok := runInferredTruthy(context); ok {
			return results
		}
		if value == nil {
			return nil
		}
	}

	var results []types.RuleFunctionResult

	isTruthy := true

	if value == nil {
		isTruthy = false
	} else if str, ok := value.(string); ok && str == "" {
		isTruthy = false
	} else if str, ok := value.(string); ok && str == "null" {
		isTruthy = false
	}

	if !isTruthy {
		results = append(results, types.RuleFunctionResult{
			Message: "Field must have a truthy value",
			Path:    resultPath(context.Path),
		})
	}

	return results
}

func runInferredTruthy(context types.RuleFunctionContext) ([]types.RuleFunctionResult, bool) {
	fieldName, parentPath, ok := inferTruthyField(context.GivenPath)
	if !ok {
		return nil, false
	}

	document := context.Document
	if context.ResolvedDocument != nil {
		document = context.ResolvedDocument
	}

	var results []types.RuleFunctionResult
	for _, parent := range parentPath.SelectLocated(document) {
		fieldValue, exists := mapField(parent.Node, fieldName)
		resultFieldPath := fieldPath(parent.Path.String(), fieldName)

		if !exists {
			results = append(results, types.RuleFunctionResult{
				Message: "Missing required field '" + fieldName + "' at " + resultFieldPath,
				Path:    resultPath(resultFieldPath),
			})
			continue
		}

		if !truthyValue(fieldValue) {
			results = append(results, types.RuleFunctionResult{
				Message: "Field must have a truthy value",
				Path:    resultPath(resultFieldPath),
			})
		}
	}

	return results, true
}

func inferTruthyField(path *jsonpath.Path) (string, *jsonpath.Path, bool) {
	query := path.Query()
	segments := query.Segments()
	if len(segments) == 0 {
		return "", nil, false
	}

	lastSegment := segments[len(segments)-1]
	if lastSegment.IsDescendant() {
		return "", nil, false
	}

	selectors := lastSegment.Selectors()
	if len(selectors) != 1 {
		return "", nil, false
	}

	fieldName, ok := selectors[0].(spec.Name)
	if !ok || fieldName == "" {
		return "", nil, false
	}

	parentPath := jsonpath.New(spec.Query(true, segments[:len(segments)-1]...))
	return string(fieldName), parentPath, true
}

func mapField(value interface{}, fieldName string) (interface{}, bool) {
	if itemMap, ok := value.(map[string]interface{}); ok {
		fieldValue, exists := itemMap[fieldName]
		return fieldValue, exists
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil, false
	}
	if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String {
		return nil, false
	}

	fieldValue := v.MapIndex(reflect.ValueOf(fieldName))
	if !fieldValue.IsValid() {
		return nil, false
	}

	return fieldValue.Interface(), true
}

func truthyValue(value interface{}) bool {
	if value == nil {
		return false
	}

	if str, ok := value.(string); ok && (str == "" || str == "null") {
		return false
	}

	return true
}
