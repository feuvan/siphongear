package input

import (
	"fmt"

	"github.com/sunshow/siphongear/internal/pipeline"
)

// staticStep merges configured vars into the payload.
type staticStep struct {
	vars map[string]any
}

func (s *staticStep) Kind() string { return "input.static" }

func (s *staticStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	out := in.Clone()
	for k, v := range s.vars {
		out.Vars[k] = v
	}
	return out, nil
}

func newStatic(cfg map[string]any) (pipeline.Step, error) {
	return &staticStep{vars: pipeline.CfgMap(cfg, "vars")}, nil
}

// credentialStep loads a credential and exposes its plaintext payload as cred.* vars.
type credentialStep struct {
	id    uint
	field string
}

func (s *credentialStep) Kind() string { return "input.credential" }

func (s *credentialStep) Run(ctx *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	if ctx.Credentials == nil {
		return nil, fmt.Errorf("no credential resolver in context")
	}
	if s.id == 0 {
		return nil, fmt.Errorf("credential_id is required")
	}
	data, err := ctx.Credentials.Resolve(ctx, s.id)
	if err != nil {
		return nil, err
	}
	out := in.Clone()
	if s.field == "" {
		out.Vars["cred"] = data
	} else {
		out.Vars[s.field] = data
	}
	return out, nil
}

func newCredential(cfg map[string]any) (pipeline.Step, error) {
	return &credentialStep{
		id:    uint(pipeline.CfgInt(cfg, "credential_id", 0)),
		field: pipeline.CfgString(cfg, "var_name"),
	}, nil
}

func init() {
	pipeline.Register(pipeline.StepMeta{
		Kind:        "input.static",
		Stage:       "input",
		Description: "Set static variables on the payload",
		Schema: map[string]any{
			"vars": map[string]any{"type": "object", "label": "Variables", "description": "Key/value pairs to inject"},
		},
	}, newStatic)

	pipeline.Register(pipeline.StepMeta{
		Kind:        "input.credential",
		Stage:       "input",
		Description: "Load a stored credential into payload vars",
		Schema: map[string]any{
			"credential_id": map[string]any{"type": "number", "label": "Credential ID", "required": true},
			"var_name":      map[string]any{"type": "string", "label": "Variable name", "default": "cred"},
		},
	}, newCredential)
}
