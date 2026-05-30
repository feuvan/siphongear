package pipeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/sunshow/siphongear/internal/pipeline"
	_ "github.com/sunshow/siphongear/internal/steps"
)

func TestEngine_StaticAndJSON(t *testing.T) {
	def := pipeline.Definition{
		Steps: []pipeline.StepConfig{
			{Kind: "input.static", Config: map[string]any{
				"vars": map[string]any{"hello": "world"},
			}},
		},
		Indicators: []pipeline.IndicatorBind{
			{Key: "hello"},
		},
	}
	if err := def.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	pCtx := &pipeline.Context{
		Context: context.Background(),
		Logger:  zerolog.Nop(),
		Now:     time.Now(),
	}
	res, err := pipeline.NewEngine().Run(pCtx, def, pipeline.NewPayload())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if v, ok := res.Indicators["hello"]; !ok || v != "world" {
		t.Fatalf("indicator hello = %v", v)
	}
}

func TestEngine_JSONPath(t *testing.T) {
	def := pipeline.Definition{
		Steps: []pipeline.StepConfig{
			{Kind: "input.static", Config: map[string]any{
				"vars": map[string]any{"raw": `{"balance": 12.5}`},
			}},
			{Kind: "transform.template", Config: map[string]any{
				"template": "{{.vars.raw}}",
			}},
			{Kind: "parse.json"},
			{Kind: "extract.jsonpath", Config: map[string]any{
				"mappings": []any{
					map[string]any{"name": "balance", "path": "$.balance", "type": "number"},
				},
			}},
		},
		Indicators: []pipeline.IndicatorBind{{Key: "balance"}},
	}
	pCtx := &pipeline.Context{Context: context.Background(), Logger: zerolog.Nop()}
	res, err := pipeline.NewEngine().Run(pCtx, def, pipeline.NewPayload())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	v, ok := res.Indicators["balance"].(float64)
	if !ok || v != 12.5 {
		t.Fatalf("expected 12.5, got %v", res.Indicators["balance"])
	}
}

func TestEngine_TransformExpr(t *testing.T) {
	def := pipeline.Definition{
		Steps: []pipeline.StepConfig{
			{Kind: "input.static", Config: map[string]any{
				"vars": map[string]any{"raw": 1250.0},
			}},
			{Kind: "transform.expr", Config: map[string]any{
				"expr":    "vars.raw / 100",
				"save_as": "balance",
			}},
		},
		Indicators: []pipeline.IndicatorBind{{Key: "balance"}},
	}
	pCtx := &pipeline.Context{Context: context.Background(), Logger: zerolog.Nop()}
	res, err := pipeline.NewEngine().Run(pCtx, def, pipeline.NewPayload())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	v, ok := res.Indicators["balance"].(float64)
	if !ok || v != 12.5 {
		t.Fatalf("expected 12.5, got %v", res.Indicators["balance"])
	}
}

func TestEngine_TransformExprNonFiniteFails(t *testing.T) {
	def := pipeline.Definition{
		Steps: []pipeline.StepConfig{
			{Kind: "input.static", Config: map[string]any{
				"vars": map[string]any{"raw": 5.0},
			}},
			{Kind: "transform.expr", Config: map[string]any{
				"expr":    "vars.missing / 100",
				"save_as": "balance",
			}},
		},
		Indicators: []pipeline.IndicatorBind{{Key: "balance"}},
	}
	pCtx := &pipeline.Context{Context: context.Background(), Logger: zerolog.Nop()}
	if _, err := pipeline.NewEngine().Run(pCtx, def, pipeline.NewPayload()); err == nil {
		t.Fatalf("expected error for non-finite result, got nil")
	}
}
