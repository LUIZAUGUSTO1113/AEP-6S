package sample

import (
	"encoding/json"
	"errors"
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

// @Summary Delete a sample
// @Description Delete a sample from the database
// @Tags samples
// @Produce json
// @Param id path string true "Sample ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Bad Request (e.g. Invalid ObjectID format)"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /samples/{id} [delete]
func (c *Controller) DeleteById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")
	err := c.service.DeleteSample(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidID):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, ErrSampleNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Update a sample
// @Description Fully replace an existing sample by ID (idempotent PUT)
// @Tags samples
// @Accept json
// @Produce json
// @Param id path string true "Sample ID"
// @Param sample body UpdateSampleRequest true "Full sample replacement data"
// @Success 200 {object} Sample
// @Failure 400 {object} map[string]string "Bad Request (e.g. Invalid JSON, River is required, Parameter is required)"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /samples/{id} [put]
func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")

	var req UpdateSampleRequest
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

	updated, err := c.service.UpdateSample(r.Context(), id, &sample)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidID):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, ErrSampleNotFound):
			w.WriteHeader(http.StatusNotFound)
		case err.Error() == "River is required." || err.Error() == "Parameter is required.":
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(updated)
}
