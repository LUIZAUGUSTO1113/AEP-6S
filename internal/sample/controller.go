package sample

import (
	"encoding/json"
	"net/http"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// @Summary Create a new sample
// @Description Create a new sample with the provided data
// @Tags samples
// @Accept json
// @Produce json
// @Param sample body CreateSampleRequest true "Sample data"
// @Success 201 {object} Sample
// @Failure 400 {object} map[string]string "Bad Request (e.g. Invalid JSON, River is required)"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /samples [post]
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateSampleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON."})
		return
	}

	sample := Sample{
		River:     req.River,
		Parameter: req.Parameter,
		Value:     req.Value,
	}

	if req.CollectedAt != nil {
		sample.CollectedAt = *req.CollectedAt
	}

	if err := c.service.CreateSample(r.Context(), &sample); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sample)
}

// @Summary Get all samples
// @Description Retrieve all samples from the database
// @Tags samples
// @Produce json
// @Success 200 {array} Sample
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /samples [get]
func (c *Controller) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	samples, err := c.service.GetAllSamples(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(samples)
}
