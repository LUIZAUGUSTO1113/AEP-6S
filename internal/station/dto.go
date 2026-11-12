package station

// StationRequest is the body of POST /stations and PUT /stations/{id}.
// PUT replaces every editable field, so both use the same shape.
type StationRequest struct {
	Name       string             `json:"name" binding:"required" example:"Ponto Ponte do Rio Ivaí"`
	River      string             `json:"river" binding:"required" example:"Rio Ivaí"`
	Code       string             `json:"code" binding:"required" example:"IVA-001"`
	Location   LocationRequest    `json:"location" binding:"required"`
	Thresholds []ThresholdRequest `json:"thresholds"`
}

type LocationRequest struct {
	City  string  `json:"city" binding:"required" example:"Maringá"`
	State string  `json:"state" binding:"required" example:"PR"`
	Lat   float64 `json:"lat" example:"-23.42"`
	Lng   float64 `json:"lng" example:"-51.93"`
}

// ThresholdRequest bounds are optional; omit min or max to leave that side open.
type ThresholdRequest struct {
	Parameter string   `json:"parameter" binding:"required" example:"pH"`
	Min       *float64 `json:"min,omitempty" example:"6"`
	Max       *float64 `json:"max,omitempty" example:"9"`
}

func (r StationRequest) toStation() Station {
	station := Station{
		Name:  r.Name,
		River: r.River,
		Code:  r.Code,
		Location: Location{
			City:  r.Location.City,
			State: r.Location.State,
			Lat:   r.Location.Lat,
			Lng:   r.Location.Lng,
		},
	}
	if r.Thresholds != nil {
		station.Thresholds = make([]Threshold, len(r.Thresholds))
		for i, t := range r.Thresholds {
			station.Thresholds[i] = Threshold{Parameter: t.Parameter, Min: t.Min, Max: t.Max}
		}
	}
	return station
}
