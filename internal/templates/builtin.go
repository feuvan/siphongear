package templates

import (
	"github.com/sunshow/siphongear/internal/pipeline"
)

func init() {
	Register(Template{
		Name:            "sub2api-balance",
		Description:     "sub2api / OneAPI / NewAPI 风格网关：登录 -> /api/v1/user/profile 取 balance",
		NeedsCredential: true,
		CredentialHint: &TemplateCredentialHint{
			Type: "password",
			Fields: []TemplateCredentialField{
				{Name: "email", Label: "Email", Type: "text", Required: true, Placeholder: "user@example.com"},
				{Name: "password", Label: "Password", Type: "password", Required: true},
			},
		},
		ScheduleType:    "interval",
		ScheduleSpec:    "30m",
		Timeout:         30,
		Variables: []TemplateVariable{
			{Name: "base_url", Label: "Base URL", Placeholder: "http://example.com:port", Required: true},
		},
		Pipeline: pipeline.Definition{
			Steps: []pipeline.StepConfig{
				{Kind: "input.credential", Name: "load credential",
					Config: map[string]any{
						"credential_id": 0,
						"var_name":      "cred",
					}},
				{Kind: "fetch.http", Name: "login",
					Config: map[string]any{
						"method":  "POST",
						"url":     "{{BASE_URL}}/api/v1/auth/login",
						"headers": map[string]any{"Content-Type": "application/json"},
						"body":    `{"email":"{{.vars.cred.email}}","password":"{{.vars.cred.password}}"}`,
						"timeout": 15,
					}},
				{Kind: "parse.json", Name: "parse login"},
				{Kind: "extract.jsonpath", Name: "extract token",
					Config: map[string]any{
						"mappings": []any{
							map[string]any{"name": "token", "path": "$.data.access_token", "type": "string"},
						},
					}},
				{Kind: "fetch.http", Name: "fetch profile",
					Config: map[string]any{
						"method":  "GET",
						"url":     "{{BASE_URL}}/api/v1/user/profile",
						"headers": map[string]any{"Authorization": "Bearer {{.vars.token}}"},
						"timeout": 15,
					}},
				{Kind: "parse.json", Name: "parse profile"},
				{Kind: "extract.jsonpath", Name: "extract balance",
					Config: map[string]any{
						"mappings": []any{
							map[string]any{"name": "balance", "path": "$.data.balance", "type": "number"},
						},
					}},
			},
			Indicators: []pipeline.IndicatorBind{{Key: "balance"}},
		},
		Indicators: []TemplateIndicator{
			{Key: "balance", Name: "余额", Type: "number", Display: "line"},
		},
	})
}
