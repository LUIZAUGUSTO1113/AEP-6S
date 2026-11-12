package station

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func validStation() *Station {
	return &Station{
		Name:     "Ponto Ponte do Rio Ivaí",
		River:    "Rio Ivaí",
		Code:     "IVA-001",
		Location: Location{City: "Maringá", State: "PR", Lat: -23.42, Lng: -51.93},
		Thresholds: []Threshold{
			{Parameter: "pH", Min: ptr(6), Max: ptr(9)},
			{Parameter: "Turbidity", Max: ptr(100)},
		},
	}
}

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Station)
		want   error
	}{
		{"valid", func(*Station) {}, nil},
		{"no thresholds", func(s *Station) { s.Thresholds = nil }, nil},
		{"blank name", func(s *Station) { s.Name = "  " }, ErrNameRequired},
		{"missing river", func(s *Station) { s.River = "" }, ErrRiverRequired},
		{"missing code", func(s *Station) { s.Code = "" }, ErrCodeRequired},
		{"missing city", func(s *Station) { s.Location.City = "" }, ErrCityRequired},
		{"missing state", func(s *Station) { s.Location.State = " " }, ErrStateRequired},
		{"threshold without parameter", func(s *Station) { s.Thresholds[1].Parameter = "" }, ErrThresholdParameterRequired},
		{"min greater than max", func(s *Station) { s.Thresholds[1].Min = ptr(150) }, ErrThresholdRange},
		{"min equal to max", func(s *Station) { s.Thresholds[1].Min = ptr(100) }, nil},
		{"pH min below 0", func(s *Station) { s.Thresholds[0].Min = ptr(-1) }, ErrInvalidPHThreshold},
		{"pH max above 14", func(s *Station) { s.Thresholds[0].Max = ptr(15) }, ErrInvalidPHThreshold},
		{"pH bounds on the edges", func(s *Station) { s.Thresholds[0].Min, s.Thresholds[0].Max = ptr(0), ptr(14) }, nil},
		{"pH matched case-insensitively", func(s *Station) { s.Thresholds[0].Parameter = " PH "; s.Thresholds[0].Max = ptr(20) }, ErrInvalidPHThreshold},
		{"non-pH bounds are free", func(s *Station) { s.Thresholds[1].Max = ptr(5000) }, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := validStation()
			tc.change(s)
			normalize(s)
			err := validate(s)
			if tc.want == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.want)
			}
		})
	}
}

func TestValidateReportsThresholdIndex(t *testing.T) {
	s := validStation()
	s.Thresholds[1].Parameter = ""
	require.EqualError(t, validate(s), "thresholds[1]: "+ErrThresholdParameterRequired.Error())
}

func TestNormalize(t *testing.T) {
	s := &Station{Name: " A ", River: " B ", Code: " C-1 ", Location: Location{City: " Maringá ", State: " pr "}}
	normalize(s)
	require.Equal(t, "A", s.Name)
	require.Equal(t, "B", s.River)
	require.Equal(t, "C-1", s.Code)
	require.Equal(t, "Maringá", s.Location.City)
	require.Equal(t, "PR", s.Location.State)
	require.NotNil(t, s.Thresholds)
	require.Empty(t, s.Thresholds)

	s.Thresholds = []Threshold{{Parameter: " pH "}}
	normalize(s)
	require.Equal(t, "pH", s.Thresholds[0].Parameter)
}
