package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// Payload is the data passed between steps.
type Payload struct {
	Vars   map[string]any
	Body   []byte
	Object any
	Meta   map[string]string
}

func NewPayload() *Payload {
	return &Payload{
		Vars: map[string]any{},
		Meta: map[string]string{},
	}
}

func (p *Payload) Clone() *Payload {
	out := NewPayload()
	for k, v := range p.Vars {
		out.Vars[k] = v
	}
	for k, v := range p.Meta {
		out.Meta[k] = v
	}
	if p.Body != nil {
		body := make([]byte, len(p.Body))
		copy(body, p.Body)
		out.Body = body
	}
	out.Object = p.Object
	return out
}

// CredentialResolver resolves a credential's plaintext payload by ID.
type CredentialResolver interface {
	Resolve(ctx context.Context, credentialID uint) (map[string]any, error)
}

// Context is shared across a pipeline run.
type Context struct {
	context.Context
	RunID         uint64
	CollectorID   uint
	Trigger       string
	Logger        zerolog.Logger
	Credentials   CredentialResolver
	Now           time.Time
	StepListener  StepListener
}

// StepListener observes step lifecycle events.
type StepListener interface {
	OnStepStart(index int, kind string)
	OnStepEnd(index int, kind string, duration time.Duration, snippet string, err error)
}

// Step is the unit of work in a pipeline.
type Step interface {
	Kind() string
	Run(ctx *Context, in *Payload) (*Payload, error)
}

// Factory builds a Step from configuration.
type Factory func(cfg map[string]any) (Step, error)

// StepConfig is the persisted shape of one step inside a pipeline JSON.
type StepConfig struct {
	Kind    string         `json:"kind"`
	Name    string         `json:"name,omitempty"`
	Config  map[string]any `json:"config,omitempty"`
	Enabled *bool          `json:"enabled,omitempty"`
}

func (s StepConfig) IsEnabled() bool {
	if s.Enabled == nil {
		return true
	}
	return *s.Enabled
}

// Definition is the persisted shape of a full pipeline.
type Definition struct {
	Steps      []StepConfig    `json:"steps"`
	Indicators []IndicatorBind `json:"indicators,omitempty"`
}

// IndicatorBind binds a payload var to an indicator (resolved later from DB).
type IndicatorBind struct {
	Key  string `json:"key"`
	Path string `json:"path,omitempty"` // optional dotted path into Vars/Object
}

// Validate ensures all step kinds exist in the registry.
func (d Definition) Validate() error {
	for i, s := range d.Steps {
		if s.Kind == "" {
			return fmt.Errorf("step %d: empty kind", i)
		}
		if !HasFactory(s.Kind) {
			return fmt.Errorf("step %d: unknown kind %q", i, s.Kind)
		}
	}
	return nil
}
