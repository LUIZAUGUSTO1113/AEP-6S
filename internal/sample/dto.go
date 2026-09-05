package sample

import "time"

type CreateSampleRequest struct {
	River       string     `json:"river" binding:"required" example:"Rio Paraná"`
	Parameter   string     `json:"parameter" binding:"required" example:"pH"`
	Value       float64    `json:"value" binding:"required" example:"7.2"`
	CollectedAt *time.Time `json:"collected_at,omitempty" example:"2026-09-04T21:00:00Z"`
}
