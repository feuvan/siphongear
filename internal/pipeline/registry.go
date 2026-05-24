package pipeline

import (
	"fmt"
	"sort"
	"sync"
)

// StepMeta exposes step kind metadata for the UI.
type StepMeta struct {
	Kind        string         `json:"kind"`
	Stage       string         `json:"stage"` // input|fetch|transform|parse|extract|any
	Description string         `json:"description"`
	Schema      map[string]any `json:"schema"` // JSON schema-ish field descriptors
}

type registryEntry struct {
	factory Factory
	meta    StepMeta
}

var (
	regMu       sync.RWMutex
	registry    = map[string]registryEntry{}
)

func Register(meta StepMeta, f Factory) {
	regMu.Lock()
	defer regMu.Unlock()
	if _, dup := registry[meta.Kind]; dup {
		panic("duplicate step kind: " + meta.Kind)
	}
	registry[meta.Kind] = registryEntry{factory: f, meta: meta}
}

func HasFactory(kind string) bool {
	regMu.RLock()
	defer regMu.RUnlock()
	_, ok := registry[kind]
	return ok
}

func Build(kind string, cfg map[string]any) (Step, error) {
	regMu.RLock()
	entry, ok := registry[kind]
	regMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown step kind: %s", kind)
	}
	return entry.factory(cfg)
}

func ListMeta() []StepMeta {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]StepMeta, 0, len(registry))
	for _, e := range registry {
		out = append(out, e.meta)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Stage != out[j].Stage {
			return out[i].Stage < out[j].Stage
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}
