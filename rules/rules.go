package rules

import (
	"fmt"
	"strings"

	"github.com/open-rpc/openrpc-linter/functions"
	"github.com/open-rpc/openrpc-linter/types"

	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
	"gopkg.in/yaml.v3"
)

func ExecuteRule(rule *types.Rule, context types.RuleFunctionContext) ([]types.RuleFunctionResult, error) {
	if rule.Then == nil {
		return []types.RuleFunctionResult{}, nil
	}

	ruleFunc := functions.FunctionRegistry[rule.Then.Function]
	if ruleFunc == nil {
		return nil, fmt.Errorf("unknown function: %s", rule.Then.Function)
	}

	path, err := jsonpath.Parse(rule.Given)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON path: %w", err)
	}

	document := context.Document
	if context.ResolvedDocument != nil {
		document = context.ResolvedDocument
	}

	var allResults []types.RuleFunctionResult

	if parentPath, fieldName, ok := terminalFieldParentPath(path, rule.Given); ok {
		for _, parent := range parentPath.SelectLocated(document) {
			var value interface{}
			if parentMap, ok := parent.Node.(map[string]interface{}); ok {
				value = parentMap[fieldName]
			}

			fieldPath := append(spec.NormalizedPath{}, parent.Path...)
			fieldPath = append(fieldPath, spec.Name(fieldName))

			itemContext := context
			itemContext.Parent = parent.Node
			itemContext.ParentPath = parent.Path.String()
			itemContext.Path = fieldPath.String()

			results := ruleFunc.RunRule(value, itemContext)
			for _, result := range results {
				if result.Message == "" {
					continue
				}
				if len(result.Path) == 0 {
					result.Path = []string{itemContext.Path}
				}
				allResults = append(allResults, result)
			}
		}
		return allResults, nil
	}

	for _, node := range path.SelectLocated(document) {
		itemContext := context
		itemContext.Path = node.Path.String()
		results := ruleFunc.RunRule(node.Node, itemContext)
		for _, result := range results {
			if result.Message == "" {
				continue
			}
			if len(result.Path) == 0 {
				result.Path = []string{itemContext.Path}
			}
			allResults = append(allResults, result)
		}
	}

	return allResults, nil
}

func terminalFieldParentPath(path *jsonpath.Path, given string) (*jsonpath.Path, string, bool) {
	query := path.Query()
	segments := query.Segments()
	if len(segments) == 0 {
		return nil, "", false
	}

	lastSegment := segments[len(segments)-1]
	if strings.HasPrefix(lastSegment.String(), "..") {
		return nil, "", false
	}

	selectors := lastSegment.Selectors()
	if len(selectors) != 1 {
		return nil, "", false
	}

	name, ok := selectors[0].(spec.Name)
	if !ok {
		return nil, "", false
	}

	parentQuery := spec.Query(strings.HasPrefix(given, "$"), segments[:len(segments)-1]...)
	return jsonpath.New(parentQuery), string(name), true
}

func GetFieldFromNode(node *yaml.Node, field string) *yaml.Node {
	for i, n := range node.Content {
		if n.Value == field {
			return node.Content[i+1]
		}
	}
	return nil
}
