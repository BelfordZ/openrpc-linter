package functions

import (
	"reflect"
	"testing"

	"github.com/open-rpc/openrpc-linter/selector"
	"github.com/open-rpc/openrpc-linter/types"
	"github.com/theory/jsonpath/spec"
)

func TestTruthyRuleMissingFieldMessage(t *testing.T) {
	target := selector.Target{
		Path:   spec.NormalizedPath{spec.Name("info"), spec.Name("description")},
		Field:  "description",
		Exists: false,
	}

	results := (&TruthyRule{}).RunRule(nil, types.RuleFunctionContext{Target: &target})
	if len(results) != 1 {
		t.Fatalf("expected one result, got %+v", results)
	}
	if results[0].Message != "missing field 'description'" {
		t.Fatalf("expected missing field message, got %q", results[0].Message)
	}
	if !reflect.DeepEqual(results[0].Path, []string{"$['info']['description']"}) {
		t.Fatalf("expected missing field path, got %+v", results[0].Path)
	}
}
