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

	if rule.Then.Function == "truthy" {
		if parentPath, targetField, ok := terminalFieldParentPath(path, rule.Given); ok {
			return executeRuleOnTerminalField(ruleFunc, parentPath, targetField, document, context), nil
		}
	}
	if rule.Then.Function == "unique" {
		if collectionPath, targetField, ok := uniqueCollectionPath(path, rule.Given); ok {
			return executeRuleOnCollection(ruleFunc, collectionPath, targetField, document, context), nil
		}
	}

	var allResults []types.RuleFunctionResult
	for _, node := range path.SelectLocated(document) {
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

func executeRuleOnTerminalField(ruleFunc types.RuleFunction, parentPath *jsonpath.Path, targetField string, document interface{}, context types.RuleFunctionContext) []types.RuleFunctionResult {
	var allResults []types.RuleFunctionResult
	for _, parent := range parentPath.SelectLocated(document) {
		itemContext := context
		itemContext.Parent = parent.Node
		itemContext.ParentPath = parent.Path.String()
		itemContext.TargetField = targetField
		itemContext.Path = itemContext.ParentPath + pathNameSegment(targetField)

		for _, result := range ruleFunc.RunRule(parent.Node, itemContext) {
			if result.Message == "" {
				continue
			}
			if len(result.Path) == 0 {
				result.Path = []string{itemContext.Path}
			}
			allResults = append(allResults, result)
		}
	}

	return allResults
}

func executeRuleOnCollection(ruleFunc types.RuleFunction, collectionPath *jsonpath.Path, targetField string, document interface{}, context types.RuleFunctionContext) []types.RuleFunctionResult {
	var allResults []types.RuleFunctionResult
	for _, collection := range collectionPath.SelectLocated(document) {
		itemContext := context
		itemContext.Path = collection.Path.String()
		itemContext.TargetField = targetField

		for _, result := range ruleFunc.RunRule(collection.Node, itemContext) {
			if result.Message == "" {
				continue
			}
			if len(result.Path) == 0 {
				result.Path = []string{itemContext.Path}
			}
			allResults = append(allResults, result)
		}
	}

	return allResults
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

func uniqueCollectionPath(path *jsonpath.Path, given string) (*jsonpath.Path, string, bool) {
	query := path.Query()
	segments := query.Segments()
	if len(segments) < 1 {
		return nil, "", false
	}

	root := strings.HasPrefix(given, "$")
	lastSegment := segments[len(segments)-1]
	lastSelectors := lastSegment.Selectors()
	if len(lastSelectors) == 1 && isWildcardSegment(lastSegment) {
		if len(segments) == 1 {
			return jsonpath.New(spec.Query(root)), "", true
		}
		return jsonpath.New(spec.Query(root, segments[:len(segments)-1]...)), "", true
	}

	if len(segments) < 2 {
		return nil, "", false
	}

	name, ok := singleNameSelector(lastSegment)
	if !ok || !isWildcardSegment(segments[len(segments)-2]) {
		return nil, "", false
	}

	collectionSegments := segments[:len(segments)-2]
	return jsonpath.New(spec.Query(root, collectionSegments...)), name, true
}

func singleNameSelector(segment *spec.Segment) (string, bool) {
	if strings.HasPrefix(segment.String(), "..") {
		return "", false
	}

	selectors := segment.Selectors()
	if len(selectors) != 1 {
		return "", false
	}

	name, ok := selectors[0].(spec.Name)
	if !ok {
		return "", false
	}

	return string(name), true
}

func isWildcardSegment(segment *spec.Segment) bool {
	if strings.HasPrefix(segment.String(), "..") {
		return false
	}
	selectors := segment.Selectors()
	return len(selectors) == 1 && selectors[0].String() == "*"
}

func pathNameSegment(name string) string {
	escaped := strings.ReplaceAll(name, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return "['" + escaped + "']"
}

func GetFieldFromNode(node *yaml.Node, field string) *yaml.Node {
	for i, n := range node.Content {
		if n.Value == field {
			return node.Content[i+1]
		}
	}
	return nil
}
