package functions

import (
	"reflect"
	"testing"

	"github.com/open-rpc/openrpc-linter/selector"
	"github.com/open-rpc/openrpc-linter/types"
	"github.com/theory/jsonpath/spec"
)

func TestUniqueRule(t *testing.T) {
	methodName := func(idx int) spec.NormalizedPath {
		return spec.NormalizedPath{spec.Name("methods"), spec.Index(idx), spec.Name("name")}
	}
	method := func(idx int) spec.NormalizedPath {
		return spec.NormalizedPath{spec.Name("methods"), spec.Index(idx)}
	}
	schemaPath := spec.NormalizedPath{spec.Name("methods"), spec.Index(0), spec.Name("schema")}

	tests := []struct {
		name            string
		targets         []selector.Target
		functionOptions map[string]interface{}
		expected        []string
	}{
		{
			name: "emits no diagnostic on first occurrence",
			targets: []selector.Target{
				valueTargetAt("ping", methodName(0)),
			},
		},
		{
			name: "emits duplicate on second occurrence",
			targets: []selector.Target{
				valueTargetAt("ping", methodName(0)),
				valueTargetAt("ping", methodName(1)),
			},
			expected: []string{`Duplicate value "ping" (first seen at $['methods'][0]['name'])`},
		},
		{
			name: "ignores missing fields by default",
			targets: []selector.Target{
				missingFieldTarget("summary", method(0)),
				missingFieldTarget("summary", method(1)),
			},
		},
		{
			name:            "treats missing fields as null when ignoreMissing is false",
			functionOptions: map[string]interface{}{"ignoreMissing": false},
			targets: []selector.Target{
				missingFieldTarget("summary", method(0)),
				missingFieldTarget("summary", method(1)),
			},
			expected: []string{"Duplicate value null (first seen at $['methods'][0]['summary'])"},
		},
		{
			name:            "rejects non boolean ignoreMissing option",
			functionOptions: map[string]interface{}{"ignoreMissing": "false"},
			targets: []selector.Target{
				valueTargetAt("ping", methodName(0)),
			},
			expected: []string{"unique function option ignoreMissing must be a boolean"},
		},
		{
			name: "rejects non primitive values",
			targets: []selector.Target{
				valueTargetAt(map[string]interface{}{"type": "string"}, schemaPath),
			},
			expected: []string{"unique value must be primitive"},
		},
		{
			name:    "nil target is a no-op",
			targets: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &UniqueRule{}
			ruleCtx := types.RuleFunctionContext{
				Rule: &types.Rule{
					Then: &types.RuleAction{
						Function:        "unique",
						FunctionOptions: tt.functionOptions,
					},
				},
			}

			var got []string
			for i := range tt.targets {
				target := tt.targets[i]
				ctx := ruleCtx
				ctx.Target = &target
				got = append(got, resultMessages(rule.RunRule(target.Node, ctx))...)
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected messages %+v, got %+v", tt.expected, got)
			}
		})
	}
}

func TestUniqueRuleScopeOption(t *testing.T) {
	document := map[string]interface{}{
		"info": map[string]interface{}{"title": "Shared"},
		"methods": []interface{}{
			map[string]interface{}{
				"name": "first",
				"params": []interface{}{
					map[string]interface{}{"name": "id"},
				},
			},
			map[string]interface{}{
				"name": "second",
				"params": []interface{}{
					map[string]interface{}{"name": "id"},
				},
			},
		},
	}

	ruleCtx := types.RuleFunctionContext{
		Rule: &types.Rule{
			Then: &types.RuleAction{
				Function: "unique",
				FunctionOptions: map[string]interface{}{
					"scope": "$.methods[*]",
				},
			},
		},
		Document: document,
	}

	paramName := func(method, param int) spec.NormalizedPath {
		return spec.NormalizedPath{
			spec.Name("methods"), spec.Index(method),
			spec.Name("params"), spec.Index(param),
			spec.Name("name"),
		}
	}
	infoField := func(field string) spec.NormalizedPath {
		return spec.NormalizedPath{spec.Name("info"), spec.Name(field)}
	}

	t.Run("does not collide across methods", func(t *testing.T) {
		rule := NewUniqueRule()
		targets := []selector.Target{
			valueTargetAt("id", paramName(0, 0)),
			valueTargetAt("id", paramName(1, 0)),
		}
		var got []string
		for i := range targets {
			target := targets[i]
			ctx := ruleCtx
			ctx.Target = &target
			got = append(got, resultMessages(rule.RunRule(target.Node, ctx))...)
		}
		if len(got) != 0 {
			t.Fatalf("expected no duplicates across methods, got %+v", got)
		}
	})

	t.Run("collides within the same method", func(t *testing.T) {
		rule := NewUniqueRule()
		targets := []selector.Target{
			valueTargetAt("id", paramName(1, 0)),
			valueTargetAt("id", paramName(1, 1)),
		}
		var got []string
		for i := range targets {
			target := targets[i]
			ctx := ruleCtx
			ctx.Target = &target
			got = append(got, resultMessages(rule.RunRule(target.Node, ctx))...)
		}
		if !reflect.DeepEqual(got, []string{`Duplicate value "id" (first seen at $['methods'][1]['params'][0]['name'])`}) {
			t.Fatalf("expected one in-method duplicate, got %+v", got)
		}
	})

	t.Run("skips targets outside every scope match", func(t *testing.T) {
		rule := NewUniqueRule()
		targets := []selector.Target{
			valueTargetAt("Shared", infoField("title")),
			valueTargetAt("Shared", infoField("version")),
		}
		var got []string
		for i := range targets {
			target := targets[i]
			ctx := ruleCtx
			ctx.Target = &target
			got = append(got, resultMessages(rule.RunRule(target.Node, ctx))...)
		}
		if len(got) != 0 {
			t.Fatalf("expected out-of-scope targets to be skipped, got %+v", got)
		}
	})

	t.Run("invalid scope JSONPath emits config diagnostic", func(t *testing.T) {
		rule := NewUniqueRule()
		ctx := ruleCtx
		ctx.Rule.Then.FunctionOptions = map[string]interface{}{
			"scope": "$.methods[",
		}
		target := valueTargetAt("id", paramName(0, 0))
		ctx.Target = &target
		got := resultMessages(rule.RunRule(target.Node, ctx))
		if len(got) != 1 || got[0] == "" {
			t.Fatalf("expected scope config diagnostic, got %+v", got)
		}
	})
}

func valueTargetAt(node interface{}, path spec.NormalizedPath) selector.Target {
	return selector.Target{
		Path:   path,
		Node:   node,
		Exists: true,
	}
}

func missingFieldTarget(field string, parentPath spec.NormalizedPath) selector.Target {
	childPath := append(append(spec.NormalizedPath{}, parentPath...), spec.Name(field))
	return selector.Target{
		Path:       childPath,
		Node:       nil,
		Exists:     false,
		Field:      field,
		Parent:     map[string]interface{}{},
		ParentPath: parentPath,
	}
}

func resultMessages(results []types.RuleFunctionResult) []string {
	if len(results) == 0 {
		return nil
	}

	messages := make([]string, 0, len(results))
	for _, result := range results {
		messages = append(messages, result.Message)
	}
	return messages
}
