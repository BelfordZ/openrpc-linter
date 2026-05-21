package selector

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	v1_4 "github.com/open-rpc/spec-types/generated/packages/go/v1_4"
)

// MetaSchema is the minimal contract the selector needs from any OpenRPC
// version.
//
// Root returns the parsed root schema object (already JSON-decoded). It is
// always a map[string]any so callers can navigate without type assertions
// scattered across the codebase.
//
// Resolve dereferences "#/definitions/<name>" local references inside the
// meta-schema. External refs (notably "https://meta.json-schema.tools/")
// return nil — that's the natural OpenRPC-vs-JSON-Schema boundary where
// we stop indexing.
type MetaSchema interface {
	Root() map[string]any
	Resolve(ref string) map[string]any
}

// V14MetaSchema loads the embedded v1.4 meta-schema from the
// open-rpc/spec-types Go package, exactly once.
type V14MetaSchema struct {
	root map[string]any
}

// v14Singleton caches the parsed root schema so repeated lint runs (or
// per-test instances) don't re-parse the ~50KB JSON string.
//
//nolint:gochecknoglobals // intentional: meta-schema is immutable.
var (
	v14Singleton *V14MetaSchema
	v14Once      sync.Once
	v14Err       error
)

// NewV14 returns the singleton v1.4 meta-schema, parsing on first call.
func NewV14() *V14MetaSchema {
	v14Once.Do(func() {
		var root map[string]any
		if err := json.Unmarshal([]byte(v1_4.RawOpenrpcDocument), &root); err != nil {
			v14Err = fmt.Errorf("parsing embedded v1.4 meta-schema: %w", err)
			return
		}
		v14Singleton = &V14MetaSchema{root: root}
	})
	if v14Err != nil {
		panic(v14Err)
	}
	return v14Singleton
}

// Root returns the root schema object.
func (m *V14MetaSchema) Root() map[string]any { return m.root }

// Resolve follows a local "#/definitions/<name>" ref. Returns nil for any
// non-local ref
func (m *V14MetaSchema) Resolve(ref string) map[string]any {
	const prefix = "#/definitions/"
	if !strings.HasPrefix(ref, prefix) {
		return nil
	}
	defs, _ := m.root["definitions"].(map[string]any)
	if defs == nil {
		return nil
	}
	target, _ := defs[strings.TrimPrefix(ref, prefix)].(map[string]any)
	return target
}
