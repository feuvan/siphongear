package extract

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/ohler55/ojg/jp"

	"github.com/sunshow/siphongear/internal/pipeline"
)

type mapping struct {
	Name string
	Path string
	Type string // number|string|bool|json
}

func loadMappings(cfg map[string]any) []mapping {
	out := []mapping{}
	for _, v := range pipeline.CfgSlice(cfg, "mappings") {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, mapping{
			Name: pipeline.CfgString(m, "name"),
			Path: pipeline.CfgString(m, "path"),
			Type: pipeline.CfgString(m, "type"),
		})
	}
	return out
}

func castValue(v any, kind string) any {
	switch kind {
	case "", "auto":
		return v
	case "number":
		switch x := v.(type) {
		case float64:
			return x
		case int:
			return float64(x)
		case string:
			f, err := strconv.ParseFloat(x, 64)
			if err == nil {
				return f
			}
		}
		return v
	case "string":
		return fmt.Sprint(v)
	case "bool":
		switch x := v.(type) {
		case bool:
			return x
		case string:
			b, err := strconv.ParseBool(x)
			if err == nil {
				return b
			}
		}
		return v
	}
	return v
}

type jsonpathStep struct {
	mappings []mapping
}

func (s *jsonpathStep) Kind() string { return "extract.jsonpath" }

func (s *jsonpathStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	if in.Object == nil {
		return nil, fmt.Errorf("object is nil; run parse step first")
	}
	out := in.Clone()
	for _, m := range s.mappings {
		expr, err := jp.ParseString(m.Path)
		if err != nil {
			return nil, fmt.Errorf("jsonpath %s: %w", m.Path, err)
		}
		results := expr.Get(in.Object)
		var v any
		if len(results) == 1 {
			v = results[0]
		} else if len(results) > 1 {
			v = results
		}
		out.Vars[m.Name] = castValue(v, m.Type)
	}
	return out, nil
}

func newJSONPath(cfg map[string]any) (pipeline.Step, error) {
	return &jsonpathStep{mappings: loadMappings(cfg)}, nil
}

type cssStep struct {
	mappings []mapping
	attrs    map[string]string // name -> attribute name (empty = text)
}

func (s *cssStep) Kind() string { return "extract.css" }

func (s *cssStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	doc, ok := in.Object.(*goquery.Document)
	if !ok {
		return nil, fmt.Errorf("object is not goquery.Document; run parse.html first")
	}
	out := in.Clone()
	for _, m := range s.mappings {
		sel := doc.Find(m.Path).First()
		var v any
		if attr, ok := s.attrs[m.Name]; ok && attr != "" {
			v, _ = sel.Attr(attr)
		} else {
			v = sel.Text()
		}
		out.Vars[m.Name] = castValue(v, m.Type)
	}
	return out, nil
}

func newCSS(cfg map[string]any) (pipeline.Step, error) {
	mappings := loadMappings(cfg)
	attrs := map[string]string{}
	for _, v := range pipeline.CfgSlice(cfg, "mappings") {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		name := pipeline.CfgString(m, "name")
		if a := pipeline.CfgString(m, "attr"); a != "" {
			attrs[name] = a
		}
	}
	return &cssStep{mappings: mappings, attrs: attrs}, nil
}

type regexStep struct {
	pattern *regexp.Regexp
	saveAs  string
	groups  []string
}

func (s *regexStep) Kind() string { return "extract.regex" }

func (s *regexStep) Run(_ *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	out := in.Clone()
	matches := s.pattern.FindStringSubmatch(string(in.Body))
	if matches == nil {
		return out, nil
	}
	if s.saveAs != "" {
		out.Vars[s.saveAs] = matches[0]
	}
	for i, name := range s.pattern.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		if i < len(matches) {
			out.Vars[name] = matches[i]
		}
	}
	for i, g := range s.groups {
		if i+1 < len(matches) {
			out.Vars[g] = matches[i+1]
		}
	}
	return out, nil
}

func newRegex(cfg map[string]any) (pipeline.Step, error) {
	pat := pipeline.CfgString(cfg, "pattern")
	if pat == "" {
		return nil, fmt.Errorf("pattern required")
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}
	groups := []string{}
	for _, v := range pipeline.CfgSlice(cfg, "groups") {
		groups = append(groups, fmt.Sprint(v))
	}
	return &regexStep{pattern: re, saveAs: pipeline.CfgString(cfg, "save_as"), groups: groups}, nil
}

func init() {
	pipeline.Register(pipeline.StepMeta{
		Kind: "extract.jsonpath", Stage: "extract",
		Description: "Extract values from a parsed JSON object via jsonpath",
		Schema: map[string]any{
			"mappings": map[string]any{
				"type": "array", "label": "Mappings",
				"itemSchema": map[string]any{
					"name": map[string]any{"type": "string", "required": true},
					"path": map[string]any{"type": "string", "required": true},
					"type": map[string]any{"type": "string", "options": []string{"auto", "number", "string", "bool"}},
				},
			},
		},
	}, newJSONPath)

	pipeline.Register(pipeline.StepMeta{
		Kind: "extract.css", Stage: "extract",
		Description: "Extract values from an HTML document via CSS selectors",
		Schema: map[string]any{
			"mappings": map[string]any{
				"type": "array", "label": "Mappings",
				"itemSchema": map[string]any{
					"name": map[string]any{"type": "string", "required": true},
					"path": map[string]any{"type": "string", "label": "CSS selector", "required": true},
					"attr": map[string]any{"type": "string", "label": "Attribute (empty = text)"},
					"type": map[string]any{"type": "string", "options": []string{"auto", "number", "string", "bool"}},
				},
			},
		},
	}, newCSS)

	pipeline.Register(pipeline.StepMeta{
		Kind: "extract.regex", Stage: "extract",
		Description: "Extract values from the body via regex (named groups become vars)",
		Schema: map[string]any{
			"pattern": map[string]any{"type": "string", "required": true},
			"groups":  map[string]any{"type": "array", "label": "Group names (positional)"},
			"save_as": map[string]any{"type": "string", "label": "Save full match as var"},
		},
	}, newRegex)
}
