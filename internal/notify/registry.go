package notify

import (
	"fmt"
	"sort"
	"sync"
)

type registryEntry struct {
	factory Factory
	meta    Meta
}

var (
	regMu    sync.RWMutex
	registry = map[string]registryEntry{}
)

// Register installs a notifier type. Call from package init().
func Register(meta Meta, f Factory) {
	regMu.Lock()
	defer regMu.Unlock()
	if _, dup := registry[meta.Type]; dup {
		panic("duplicate notify type: " + meta.Type)
	}
	registry[meta.Type] = registryEntry{factory: f, meta: meta}
}

// HasFactory reports whether the type is registered.
func HasFactory(t string) bool {
	regMu.RLock()
	defer regMu.RUnlock()
	_, ok := registry[t]
	return ok
}

// Build constructs a Notifier instance from the given decrypted payload.
func Build(t string, payload map[string]any) (Notifier, error) {
	regMu.RLock()
	entry, ok := registry[t]
	regMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown notify type: %s", t)
	}
	return entry.factory(payload)
}

// ListMeta returns all registered type metadata sorted by type.
func ListMeta() []Meta {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]Meta, 0, len(registry))
	for _, e := range registry {
		out = append(out, e.meta)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Type < out[j].Type })
	return out
}
