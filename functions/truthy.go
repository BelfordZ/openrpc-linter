package functions

import (
	"github.com/open-rpc/openrpc-linter/types"
)

type TruthyRule struct{}

// RunRule evaluates one Target. The selector already decided what the
// target is — value mode (Field == "") or field mode (Field != "") — so
// truthy only has to answer "is this present and truthy?" and emit the
// right diagnostic. No JSONPath introspection lives here anymore.
func (r *TruthyRule) RunRule(value interface{}, context types.RuleFunctionContext) []types.RuleFunctionResult {
	t := context.Target
	if t == nil {
		// Direct invocation without a target (used by some tests). Fall
		// back to "is the supplied value truthy?" — the old behavior for
		// the no-context case.
		if !truthyValue(value) {
			return []types.RuleFunctionResult{{
				Message: "Field must have a truthy value",
				Path:    resultPath(context.Path),
			}}
		}
		return nil
	}

	if t.Field != "" && !t.Exists {
		return []types.RuleFunctionResult{{
			Message: "Missing required field '" + t.Field + "' at " + t.PathString(),
			Path:    resultPath(t.PathString()),
		}}
	}

	if !truthyValue(value) {
		return []types.RuleFunctionResult{{
			Message: "Field must have a truthy value",
			Path:    resultPath(t.PathString()),
		}}
	}

	return nil
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
