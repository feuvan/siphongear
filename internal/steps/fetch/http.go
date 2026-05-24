package fetch

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/sunshow/siphongear/internal/pipeline"
)

type httpStep struct {
	method      string
	urlTpl      string
	headers     map[string]string
	bodyTpl     string
	queryTpl    map[string]string
	timeout     int
	proxyURL    string
	saveBody    string // var name to save response body string under (default "")
	saveStatus  string // var name to save status code under
	saveHeader  string // var name to save response headers
	insecureTLS bool
	expectJSON  bool
}

func (s *httpStep) Kind() string { return "fetch.http" }

func (s *httpStep) Run(ctx *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	urlStr, err := pipeline.RenderTemplate(s.urlTpl, in, nil)
	if err != nil {
		return nil, fmt.Errorf("render url: %w", err)
	}
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return nil, fmt.Errorf("url is required")
	}

	client := resty.New()
	if s.timeout > 0 {
		client.SetTimeout(time.Duration(s.timeout) * time.Second)
	}
	if s.proxyURL != "" {
		client.SetProxy(s.proxyURL)
	}
	if s.insecureTLS {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	req := client.R().SetContext(ctx)
	for k, v := range s.headers {
		rendered, err := pipeline.RenderTemplate(v, in, nil)
		if err != nil {
			return nil, fmt.Errorf("render header %s: %w", k, err)
		}
		req.SetHeader(k, rendered)
	}
	for k, v := range s.queryTpl {
		rendered, err := pipeline.RenderTemplate(v, in, nil)
		if err != nil {
			return nil, fmt.Errorf("render query %s: %w", k, err)
		}
		req.SetQueryParam(k, rendered)
	}
	if s.bodyTpl != "" {
		rendered, err := pipeline.RenderTemplate(s.bodyTpl, in, nil)
		if err != nil {
			return nil, fmt.Errorf("render body: %w", err)
		}
		req.SetBody(rendered)
	}

	method := strings.ToUpper(s.method)
	if method == "" {
		method = "GET"
	}
	resp, err := req.Execute(method, urlStr)
	if err != nil {
		return nil, fmt.Errorf("http %s %s: %w", method, urlStr, err)
	}

	out := in.Clone()
	out.Body = append([]byte(nil), resp.Body()...)
	out.Meta["status"] = fmt.Sprintf("%d", resp.StatusCode())
	out.Meta["content_type"] = resp.Header().Get("Content-Type")
	if s.saveBody != "" {
		out.Vars[s.saveBody] = string(resp.Body())
	}
	if s.saveStatus != "" {
		out.Vars[s.saveStatus] = resp.StatusCode()
	}
	if s.saveHeader != "" {
		hdr := map[string]any{}
		for k, v := range resp.Header() {
			if len(v) == 1 {
				hdr[k] = v[0]
			} else {
				vals := make([]any, len(v))
				for i, vv := range v {
					vals[i] = vv
				}
				hdr[k] = vals
			}
		}
		out.Vars[s.saveHeader] = hdr
	}
	if resp.StatusCode() >= 400 {
		return out, fmt.Errorf("http %d: %s", resp.StatusCode(), truncate(string(resp.Body()), 200))
	}
	return out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func newHTTP(cfg map[string]any) (pipeline.Step, error) {
	return &httpStep{
		method:      pipeline.CfgString(cfg, "method"),
		urlTpl:      pipeline.CfgString(cfg, "url"),
		headers:     pipeline.CfgStringMap(cfg, "headers"),
		bodyTpl:     pipeline.CfgString(cfg, "body"),
		queryTpl:    pipeline.CfgStringMap(cfg, "query"),
		timeout:     pipeline.CfgInt(cfg, "timeout", 30),
		proxyURL:    pipeline.CfgString(cfg, "proxy"),
		saveBody:    pipeline.CfgString(cfg, "save_body_as"),
		saveStatus:  pipeline.CfgString(cfg, "save_status_as"),
		saveHeader:  pipeline.CfgString(cfg, "save_headers_as"),
		insecureTLS: pipeline.CfgBool(cfg, "insecure_tls", false),
		expectJSON:  pipeline.CfgBool(cfg, "expect_json", false),
	}, nil
}

func init() {
	pipeline.Register(pipeline.StepMeta{
		Kind:        "fetch.http",
		Stage:       "fetch",
		Description: "Perform an HTTP request via resty (templates supported in url/headers/body)",
		Schema: map[string]any{
			"method":          map[string]any{"type": "string", "label": "Method", "default": "GET"},
			"url":             map[string]any{"type": "string", "label": "URL", "required": true},
			"headers":         map[string]any{"type": "object", "label": "Headers"},
			"query":           map[string]any{"type": "object", "label": "Query params"},
			"body":            map[string]any{"type": "text", "label": "Body"},
			"timeout":         map[string]any{"type": "number", "label": "Timeout (s)", "default": 30},
			"proxy":           map[string]any{"type": "string", "label": "Proxy URL"},
			"save_body_as":    map[string]any{"type": "string", "label": "Save body as var"},
			"save_status_as":  map[string]any{"type": "string", "label": "Save status as var"},
			"save_headers_as": map[string]any{"type": "string", "label": "Save headers as var"},
			"insecure_tls":    map[string]any{"type": "boolean", "label": "Skip TLS verify"},
		},
	}, newHTTP)
}
