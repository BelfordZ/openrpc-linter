package rules

import (
	"fmt"

	"github.com/open-rpc/openrpc-linter/functions"
	"github.com/open-rpc/openrpc-linter/types"

	"github.com/theory/jsonpath"
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
	context.GivenPath = path

	var allResults []types.RuleFunctionResult
	selectedNodes := path.SelectLocated(document)

	if rule.Then.Function == "truthy" {
		itemContext := context
		itemContext.Path = ""
		for _, result := range ruleFunc.RunRule(nil, itemContext) {
			if result.Message == "" {
				continue
			}
			allResults = append(allResults, result)
		}
		if len(allResults) > 0 {
			return allResults, nil
		}
	}

	for _, node := range selectedNodes {
		valueToValidate := node.Node
		itemContext := context
		itemContext.Path = node.Path.String()

		for _, result := range ruleFunc.RunRule(valueToValidate, itemContext) {
			if result.Message == "" {
				continue
			}
			if len(result.Path) == 0 {
				result.Path = []string{node.Path.String()}
			}
			allResults = append(allResults, result)
		}
	}

	return allResults, nil
}

func GetFieldFromNode(node *yaml.Node, field string) *yaml.Node {
	for i, n := range node.Content {
		if n.Value == field {
			return node.Content[i+1]
		}
	}
	return nil
}
