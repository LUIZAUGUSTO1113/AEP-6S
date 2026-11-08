package station

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func ptr(v float64) *float64 { return &v }

func TestEvaluate(t *testing.T) {
	s := Station{Code: "IVA-001", Thresholds: []Threshold{
		{Parameter: "pH", Min: ptr(6), Max: ptr(9)},
		{Parameter: "Turbidity", Max: ptr(100)},
		{Parameter: "Dissolved Oxygen", Min: ptr(5.5)},
		{Parameter: "Zero Floor", Min: ptr(0)},
	}}

	for _, tc := range []struct {
		name, parameter string
		value           float64
		violated        bool
		message         string
	}{
		{"below min", "pH", 4.2, true, "pH abaixo do mínimo 6.0 da estação IVA-001"},
		{"above max", "pH", 9.5, true, "pH acima do máximo 9.0 da estação IVA-001"},
		{"inside range", "pH", 7.2, false, ""},
		{"on min boundary", "pH", 6, false, ""},
		{"on max boundary", "pH", 9, false, ""},
		{"only max set", "Turbidity", 150, true, "Turbidity acima do máximo 100.0 da estação IVA-001"},
		{"only max set, inside", "Turbidity", 20, false, ""},
		{"decimal bound", "Dissolved Oxygen", 5, true, "Dissolved Oxygen abaixo do mínimo 5.5 da estação IVA-001"},
		{"zero min is a real bound", "Zero Floor", -1, true, "Zero Floor abaixo do mínimo 0.0 da estação IVA-001"},
		{"case and spaces ignored", "  PH ", 3, true, "pH abaixo do mínimo 6.0 da estação IVA-001"},
		{"no threshold for parameter", "Temperature", 99, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			violated, message := s.Evaluate(tc.parameter, tc.value)
			require.Equal(t, tc.violated, violated)
			require.Equal(t, tc.message, message)
		})
	}

	violated, message := Station{}.Evaluate("pH", 1)
	require.False(t, violated, "station without thresholds never alerts")
	require.Empty(t, message)
}
