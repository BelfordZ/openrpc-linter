package functions

import "github.com/open-rpc/openrpc-linter/types"

type TruthyRule struct{}

func (r *TruthyRule) RunRule(value interface{}, context types.RuleFunctionContext) []types.RuleFunctionResult {
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
