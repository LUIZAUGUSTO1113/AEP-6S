package station

import (
	"fmt"
	"strings"
)

// normalize trims text fields, upper-cases the state (UF) and turns nil
// thresholds into an empty list so the API never returns null for it.
func normalize(s *Station) {
	s.Name = strings.TrimSpace(s.Name)
	s.River = strings.TrimSpace(s.River)
	s.Code = strings.TrimSpace(s.Code)
	s.Location.City = strings.TrimSpace(s.Location.City)
	s.Location.State = strings.ToUpper(strings.TrimSpace(s.Location.State))
	if s.Thresholds == nil {
		s.Thresholds = []Threshold{}
	}
	for i := range s.Thresholds {
		s.Thresholds[i].Parameter = strings.TrimSpace(s.Thresholds[i].Parameter)
	}
}

func validate(s *Station) error {
	switch {
	case s.Name == "":
		return ErrNameRequired
	case s.River == "":
		return ErrRiverRequired
	case s.Code == "":
		return ErrCodeRequired
	case s.Location.City == "":
		return ErrCityRequired
	case s.Location.State == "":
		return ErrStateRequired
	}
	for i, t := range s.Thresholds {
		if err := validateThreshold(t); err != nil {
			return fmt.Errorf("thresholds[%d]: %w", i, err)
		}
	}
	return nil
}

func validateThreshold(t Threshold) error {
	if t.Parameter == "" {
		return ErrThresholdParameterRequired
	}
	if t.Min != nil && t.Max != nil && *t.Min > *t.Max {
		return ErrThresholdRange
	}
	if strings.EqualFold(t.Parameter, "pH") {
		for _, bound := range []*float64{t.Min, t.Max} {
			if bound != nil && !(*bound >= 0 && *bound <= 14) {
				return ErrInvalidPHThreshold
			}
		}
	}
	return nil
}
