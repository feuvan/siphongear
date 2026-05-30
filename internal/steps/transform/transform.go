package transform

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/dop251/goja"
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

type exprStep struct {
	expr      string
	saveAs    string
	timeoutMS int
}

func (s *exprStep) Kind() string { return "transform.expr" }

func (s *exprStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	rt := goja.New()
	out := in.Clone()
	_ = rt.Set("vars", out.Vars)

	timeout := time.Duration(s.timeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	timer := time.AfterFunc(timeout, func() {
		rt.Interrupt(fmt.Errorf("expr timeout after %s", timeout))
	})
	defer timer.Stop()

	val, err := rt.RunString("(function(){\nreturn (" + s.expr + ");\n})()")
	if err != nil {
		return nil, fmt.Errorf("expr error: %w", err)
	}
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil, fmt.Errorf("expr %q produced no value", s.expr)
	}
	exported := val.Export()

	if s.saveAs != "" {
		f, err := finiteNumber(exported)
		if err != nil {
			return nil, fmt.Errorf("expr %q: %w", s.expr, err)
		}
		out.Vars[s.saveAs] = f
		return out, nil
	}

	m, ok := exported.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expr %q must return an object when save_as is empty, got: %v", s.expr, exported)
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("expr %q produced no values", s.expr)
	}
	for k, v := range m {
		f, err := finiteNumber(v)
		if err != nil {
			return nil, fmt.Errorf("expr %q key %q: %w", s.expr, k, err)
		}
		out.Vars[k] = f
	}
	return out, nil
}

func finiteNumber(v any) (float64, error) {
	var f float64
	switch x := v.(type) {
	case float64:
		f = x
	case float32:
		f = float64(x)
	case int:
		f = float64(x)
	case int32:
		f = float64(x)
	case int64:
		f = float64(x)
	default:
		return 0, fmt.Errorf("result is not numeric: %v", v)
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, fmt.Errorf("non-finite result: %v", f)
	}
	return f, nil
}

func newExpr(cfg map[string]any) (pipeline.Step, error) {
	expr := pipeline.CfgString(cfg, "expr")
	if expr == "" {
		return nil, fmt.Errorf("expr required")
	}
	return &exprStep{
		expr:      expr,
		saveAs:    pipeline.CfgString(cfg, "save_as"),
		timeoutMS: pipeline.CfgInt(cfg, "timeout_ms", 5000),
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

	pipeline.Register(pipeline.StepMeta{
		Kind: "transform.expr", Stage: "transform",
		Description: "Evaluate a numeric expression over vars and save the result. Single mode: set Save as var. Batch mode: leave it empty and return an object like {balance: vars.balance / 100}",
		Schema: map[string]any{
			"expr":       map[string]any{"type": "code", "lang": "javascript", "label": "Expression", "required": true, "hint": "Reads vars.*. Single: vars.balance / 100. Batch (Save as var empty): {balance: vars.balance / 100, fee: vars.fee / 100}"},
			"save_as":    map[string]any{"type": "string", "label": "Save as var", "hint": "Single-value mode: output var name (use the source name to overwrite). Leave empty for batch object mode"},
			"timeout_ms": map[string]any{"type": "number", "label": "Timeout (ms)", "default": 5000},
		},
	}, newExpr)
}
