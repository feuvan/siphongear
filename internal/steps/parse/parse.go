package parse

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/bytedance/sonic"

	"github.com/sunshow/siphongear/internal/pipeline"
)

type jsonStep struct {
	saveAs string
}

func (s *jsonStep) Kind() string { return "parse.json" }

func (s *jsonStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	if len(in.Body) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	var obj any
	if err := sonic.Unmarshal(in.Body, &obj); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	out := in.Clone()
	out.Object = obj
	if s.saveAs != "" {
		out.Vars[s.saveAs] = obj
	}
	return out, nil
}

func newJSON(cfg map[string]any) (pipeline.Step, error) {
	return &jsonStep{saveAs: pipeline.CfgString(cfg, "save_as")}, nil
}

type htmlStep struct{}

func (s *htmlStep) Kind() string { return "parse.html" }

func (s *htmlStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(in.Body))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}
	out := in.Clone()
	out.Object = doc
	return out, nil
}

func newHTML(_ map[string]any) (pipeline.Step, error) { return &htmlStep{}, nil }

func init() {
	pipeline.Register(pipeline.StepMeta{
		Kind: "parse.json", Stage: "parse",
		Description: "Parse body as JSON into the Object slot",
		Schema: map[string]any{
			"save_as": map[string]any{"type": "string", "label": "Also save into var"},
		},
	}, newJSON)

	pipeline.Register(pipeline.StepMeta{
		Kind: "parse.html", Stage: "parse",
		Description: "Parse body as HTML (goquery document) into the Object slot",
		Schema:      map[string]any{},
	}, newHTML)
}
