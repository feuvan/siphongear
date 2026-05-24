package script

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/go-resty/resty/v2"

	"github.com/sunshow/siphongear/internal/pipeline"
)

type jsStep struct {
	stage     string
	source    string
	allowHTTP bool
	timeoutMS int
}

func (s *jsStep) Kind() string { return "script.js." + s.stage }

func (s *jsStep) Run(ctx *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	rt := goja.New()
	rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	out := in.Clone()
	payload := map[string]any{
		"vars":   out.Vars,
		"meta":   out.Meta,
		"body":   string(out.Body),
		"object": out.Object,
	}
	_ = rt.Set("payload", payload)

	logger := ctx.Logger.With().Str("step", s.Kind()).Logger()
	_ = rt.Set("log", map[string]any{
		"info":  func(args ...any) { logger.Info().Msgf("%v", args) },
		"warn":  func(args ...any) { logger.Warn().Msgf("%v", args) },
		"error": func(args ...any) { logger.Error().Msgf("%v", args) },
	})

	if s.allowHTTP {
		client := resty.New().SetTimeout(20 * time.Second)
		_ = rt.Set("http", map[string]any{
			"get": func(url string) map[string]any {
				resp, err := client.R().Get(url)
				if err != nil {
					return map[string]any{"error": err.Error()}
				}
				return map[string]any{
					"status": resp.StatusCode(),
					"body":   string(resp.Body()),
				}
			},
			"post": func(url string, body string, headers map[string]string) map[string]any {
				req := client.R().SetBody(body)
				for k, v := range headers {
					req.SetHeader(k, v)
				}
				resp, err := req.Post(url)
				if err != nil {
					return map[string]any{"error": err.Error()}
				}
				return map[string]any{
					"status": resp.StatusCode(),
					"body":   string(resp.Body()),
				}
			},
		})
	}

	timeout := time.Duration(s.timeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	timer := time.AfterFunc(timeout, func() {
		rt.Interrupt(fmt.Errorf("script timeout after %s", timeout))
	})
	defer timer.Stop()

	val, err := rt.RunString("(function(){\n" + s.source + "\n})()")
	if err != nil {
		return nil, fmt.Errorf("script error: %w", err)
	}

	if val != nil && !goja.IsUndefined(val) && !goja.IsNull(val) {
		exported := val.Export()
		if m, ok := exported.(map[string]any); ok {
			if v, ok := m["vars"]; ok {
				if vm, ok := v.(map[string]any); ok {
					for k, vv := range vm {
						out.Vars[k] = vv
					}
				}
			}
			if v, ok := m["meta"]; ok {
				if mm, ok := v.(map[string]any); ok {
					for k, vv := range mm {
						out.Meta[k] = fmt.Sprint(vv)
					}
				}
			}
			if v, ok := m["body"]; ok {
				switch x := v.(type) {
				case string:
					out.Body = []byte(x)
				case []byte:
					out.Body = x
				}
			}
			if v, ok := m["object"]; ok {
				out.Object = v
			}
		}
	}
	return out, nil
}

func newJS(stage string) pipeline.Factory {
	return func(cfg map[string]any) (pipeline.Step, error) {
		src := pipeline.CfgString(cfg, "source")
		if src == "" {
			return nil, fmt.Errorf("source required")
		}
		return &jsStep{
			stage:     stage,
			source:    src,
			allowHTTP: pipeline.CfgBool(cfg, "allow_http", false),
			timeoutMS: pipeline.CfgInt(cfg, "timeout_ms", 5000),
		}, nil
	}
}

func init() {
	for _, stage := range []string{"input", "transform", "extract"} {
		pipeline.Register(pipeline.StepMeta{
			Kind:        "script.js." + stage,
			Stage:       stage,
			Description: "JavaScript (goja) snippet returning {vars, meta, body, object}",
			Schema: map[string]any{
				"source":     map[string]any{"type": "code", "lang": "javascript", "required": true},
				"allow_http": map[string]any{"type": "boolean", "label": "Allow http.get/post"},
				"timeout_ms": map[string]any{"type": "number", "label": "Timeout (ms)", "default": 5000},
			},
		}, newJS(stage))
	}
}
