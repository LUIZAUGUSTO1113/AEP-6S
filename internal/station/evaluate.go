package station

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Evaluate reports whether value breaks this station's threshold for
// parameter. Without a threshold for the parameter there is no violation.
// Parameter names are matched case-insensitively, ignoring surrounding spaces.
func (s Station) Evaluate(parameter string, value float64) (violated bool, message string) {
	name := strings.TrimSpace(parameter)
	for _, t := range s.Thresholds {
		if !strings.EqualFold(strings.TrimSpace(t.Parameter), name) {
			continue
		}
		if t.Min != nil && value < *t.Min {
			return true, fmt.Sprintf("%s abaixo do mínimo %s da estação %s", t.Parameter, formatBound(*t.Min), s.Code)
		}
		if t.Max != nil && value > *t.Max {
			return true, fmt.Sprintf("%s acima do máximo %s da estação %s", t.Parameter, formatBound(*t.Max), s.Code)
		}
	}
	return false, ""
}

// formatBound prints whole numbers as "6.0" and keeps every decimal otherwise.
func formatBound(v float64) string {
	if v == math.Trunc(v) {
		return strconv.FormatFloat(v, 'f', 1, 64)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}
