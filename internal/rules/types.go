package rules

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
)

const (
	ConditionCompare = "compare"

	ActionIndicatorColor = "indicator_color"

	TargetAll  = "all"
	TargetTags = "tags"

	SeverityWarning = "warning"
)

var compareOps = map[string]struct{}{
	"lt":  {},
	"lte": {},
	"gt":  {},
	"gte": {},
	"eq":  {},
	"ne":  {},
}

type Condition struct {
	Type  string  `json:"type"`
	Op    string  `json:"op,omitempty"`
	Value float64 `json:"value,omitempty"`
}

type Action struct {
	Type string `json:"type"`
}

func ParseConditions(s string) ([]Condition, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	var cs []Condition
	if err := sonic.UnmarshalString(s, &cs); err != nil {
		return nil, fmt.Errorf("parse conditions: %w", err)
	}
	return cs, nil
}

func ParseActions(s string) ([]Action, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	var as []Action
	if err := sonic.UnmarshalString(s, &as); err != nil {
		return nil, fmt.Errorf("parse actions: %w", err)
	}
	return as, nil
}

func ValidateConditions(cs []Condition) error {
	if len(cs) == 0 {
		return errors.New("at least one condition is required")
	}
	for i, c := range cs {
		switch c.Type {
		case ConditionCompare:
			if _, ok := compareOps[c.Op]; !ok {
				return fmt.Errorf("condition[%d]: unsupported op %q", i, c.Op)
			}
		default:
			return fmt.Errorf("condition[%d]: unsupported type %q", i, c.Type)
		}
	}
	return nil
}

func ValidateActions(as []Action) error {
	if len(as) == 0 {
		return errors.New("at least one action is required")
	}
	for i, a := range as {
		switch a.Type {
		case ActionIndicatorColor:
			// no extra fields today
		default:
			return fmt.Errorf("action[%d]: unsupported type %q", i, a.Type)
		}
	}
	return nil
}

func ParseTargetTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
