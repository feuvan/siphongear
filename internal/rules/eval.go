package rules

func Evaluate(conds []Condition, value *float64) bool {
	if len(conds) == 0 || value == nil {
		return false
	}
	for _, c := range conds {
		if c.Type != ConditionCompare {
			return false
		}
		if !compare(c.Op, *value, c.Value) {
			return false
		}
	}
	return true
}

func compare(op string, a, b float64) bool {
	switch op {
	case "lt":
		return a < b
	case "lte":
		return a <= b
	case "gt":
		return a > b
	case "gte":
		return a >= b
	case "eq":
		return a == b
	case "ne":
		return a != b
	}
	return false
}

func TagsIntersect(rule []string, card []string) bool {
	if len(rule) == 0 || len(card) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(rule))
	for _, t := range rule {
		set[t] = struct{}{}
	}
	for _, t := range card {
		if _, ok := set[t]; ok {
			return true
		}
	}
	return false
}
