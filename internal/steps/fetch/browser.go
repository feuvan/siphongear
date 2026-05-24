package fetch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/sunshow/siphongear/internal/pipeline"
)

type browserStep struct {
	url        string
	wsURL      string
	timeout    int
	userAgent  string
	waitMS     int
	evalScript string
	saveAs     string
}

func (s *browserStep) Kind() string { return "fetch.browser" }

func (s *browserStep) Run(ctx *pipeline.Context, in *pipeline.Payload) (*pipeline.Payload, error) {
	urlStr, err := pipeline.RenderTemplate(s.url, in, nil)
	if err != nil {
		return nil, err
	}
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return nil, fmt.Errorf("url required")
	}

	timeout := time.Duration(s.timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	var allocCtx context.Context
	var allocCancel context.CancelFunc
	if s.wsURL != "" {
		allocCtx, allocCancel = chromedp.NewRemoteAllocator(ctx, s.wsURL)
	} else {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)
		if s.userAgent != "" {
			opts = append(opts, chromedp.UserAgent(s.userAgent))
		}
		allocCtx, allocCancel = chromedp.NewExecAllocator(ctx, opts...)
	}
	defer allocCancel()

	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	tCtx, tCancel := context.WithTimeout(browserCtx, timeout)
	defer tCancel()

	actions := []chromedp.Action{chromedp.Navigate(urlStr)}
	if s.waitMS > 0 {
		actions = append(actions, chromedp.Sleep(time.Duration(s.waitMS)*time.Millisecond))
	}

	out := in.Clone()

	if s.evalScript != "" {
		var result any
		actions = append(actions, chromedp.Evaluate(s.evalScript, &result))
		if err := chromedp.Run(tCtx, actions...); err != nil {
			return nil, err
		}
		if s.saveAs != "" {
			out.Vars[s.saveAs] = result
		}
		if str, ok := result.(string); ok {
			out.Body = []byte(str)
		}
		return out, nil
	}

	var html string
	actions = append(actions, chromedp.OuterHTML("html", &html))
	if err := chromedp.Run(tCtx, actions...); err != nil {
		return nil, err
	}
	out.Body = []byte(html)
	if s.saveAs != "" {
		out.Vars[s.saveAs] = html
	}
	return out, nil
}

func newBrowser(cfg map[string]any) (pipeline.Step, error) {
	return &browserStep{
		url:        pipeline.CfgString(cfg, "url"),
		wsURL:      pipeline.CfgString(cfg, "ws_url"),
		timeout:    pipeline.CfgInt(cfg, "timeout", 30),
		userAgent:  pipeline.CfgString(cfg, "user_agent"),
		waitMS:     pipeline.CfgInt(cfg, "wait_ms", 0),
		evalScript: pipeline.CfgString(cfg, "evaluate"),
		saveAs:     pipeline.CfgString(cfg, "save_as"),
	}, nil
}

func init() {
	pipeline.Register(pipeline.StepMeta{
		Kind:        "fetch.browser",
		Stage:       "fetch",
		Description: "Load a URL with chromedp; optionally evaluate JS",
		Schema: map[string]any{
			"url":        map[string]any{"type": "string", "label": "URL", "required": true},
			"ws_url":     map[string]any{"type": "string", "label": "Remote browser WebSocket URL"},
			"timeout":    map[string]any{"type": "number", "label": "Timeout (s)", "default": 30},
			"user_agent": map[string]any{"type": "string", "label": "User agent"},
			"wait_ms":    map[string]any{"type": "number", "label": "Wait after navigate (ms)"},
			"evaluate":   map[string]any{"type": "text", "label": "JS expression to evaluate"},
			"save_as":    map[string]any{"type": "string", "label": "Save result/html as var"},
		},
	}, newBrowser)
}
