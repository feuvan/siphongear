package transform

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"golang.org/x/net/html/charset"

	"github.com/sunshow/siphongear/internal/pipeline"
)

type gunzipStep struct{}

func (s *gunzipStep) Kind() string { return "transform.gunzip" }

func (s *gunzipStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	if len(in.Body) == 0 {
		return in, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(in.Body))
	if err != nil {
		return nil, fmt.Errorf("gunzip: %w", err)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	out := in.Clone()
	out.Body = data
	return out, nil
}

func newGunzip(_ map[string]any) (pipeline.Step, error) { return &gunzipStep{}, nil }

type charsetStep struct {
	srcLabel string
}

func (s *charsetStep) Kind() string { return "transform.charset" }

func (s *charsetStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	if len(in.Body) == 0 {
		return in, nil
	}
	contentType := in.Meta["content_type"]
	r, err := charset.NewReader(bytes.NewReader(in.Body), contentType)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	out := in.Clone()
	out.Body = data
	return out, nil
}

func newCharset(cfg map[string]any) (pipeline.Step, error) {
	return &charsetStep{srcLabel: pipeline.CfgString(cfg, "label")}, nil
}

type templateStep struct {
	tpl string
	out string
}

func (s *templateStep) Kind() string { return "transform.template" }

func (s *templateStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	r, err := pipeline.RenderTemplate(s.tpl, in, nil)
	if err != nil {
		return nil, err
	}
	out := in.Clone()
	if s.out == "" {
		out.Body = []byte(r)
	} else {
		out.Vars[s.out] = r
	}
	return out, nil
}

func newTemplate(cfg map[string]any) (pipeline.Step, error) {
	return &templateStep{
		tpl: pipeline.CfgString(cfg, "template"),
		out: pipeline.CfgString(cfg, "save_as"),
	}, nil
}

func init() {
	pipeline.Register(pipeline.StepMeta{
		Kind: "transform.gunzip", Stage: "transform",
		Description: "Decompress gzip-encoded body",
		Schema:      map[string]any{},
	}, newGunzip)

	pipeline.Register(pipeline.StepMeta{
		Kind: "transform.charset", Stage: "transform",
		Description: "Convert body charset to UTF-8 based on Content-Type",
		Schema:      map[string]any{"label": map[string]any{"type": "string", "label": "Override charset label"}},
	}, newCharset)

	pipeline.Register(pipeline.StepMeta{
		Kind: "transform.template", Stage: "transform",
		Description: "Render a Go text/template using payload data",
		Schema: map[string]any{
			"template": map[string]any{"type": "text", "label": "Template", "required": true},
			"save_as":  map[string]any{"type": "string", "label": "Save as var (default: replace body)"},
		},
	}, newTemplate)
}
