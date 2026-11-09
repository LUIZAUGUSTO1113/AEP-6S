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
// @Description Record a measurement at an existing station and generate an alert if it violates a threshold
// @Tags samples
// @Accept json
// @Produce json
// @Param sample body SampleRequest true "Sample data"
// @Success 201 {object} Sample
// @Failure 400 {object} map[string]string "Invalid sample data"
// @Failure 404 {object} map[string]string "Station not found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /samples [post]
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	sample, err := readSample(r)
	if err == nil {
		err = c.service.CreateSample(r.Context(), sample)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sample)
}

// @Summary List samples
// @Description Retrieve all samples or only the samples linked to a station
// @Tags samples
// @Produce json
// @Param station_id query string false "Station ID (hexadecimal ObjectID)"
// @Success 200 {array} Sample
// @Failure 400 {object} map[string]string "Invalid station ID"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /samples [get]
func (c *Controller) GetAll(w http.ResponseWriter, r *http.Request) {
	samples, err := c.service.GetAllSamples(r.Context(), r.URL.Query().Get("station_id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, samples)
}

// @Summary Delete a sample
// @Description Delete a sample and its alert, if any
// @Tags samples
// @Produce json
// @Param id path string true "Sample ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Invalid sample ID"
// @Failure 404 {object} map[string]string "Sample not found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /samples/{id} [delete]
func (c *Controller) DeleteById(w http.ResponseWriter, r *http.Request) {
	if err := c.service.DeleteSample(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Update a sample
// @Description Replace an existing sample and reevaluate its alert
// @Tags samples
// @Accept json
// @Produce json
// @Param id path string true "Sample ID"
// @Param sample body SampleRequest true "Full sample replacement data"
// @Success 200 {object} Sample
// @Failure 400 {object} map[string]string "Invalid sample data or ID"
// @Failure 404 {object} map[string]string "Sample or station not found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /samples/{id} [put]
func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	sample, err := readSample(r)
	if err == nil {
		sample, err = c.service.UpdateSample(r.Context(), r.PathValue("id"), sample)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sample)
}

var errInvalidJSON = errors.New("Invalid JSON.")

func readSample(r *http.Request) (*Sample, error) {
	var req SampleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errInvalidJSON
	}
	return req.toSample()
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, errInvalidJSON), errors.Is(err, ErrInvalidStationID), errors.Is(err, ErrInvalidID),
		errors.Is(err, ErrValueRequired), errors.Is(err, ErrParameterRequired), errors.Is(err, ErrInvalidPH):
		status = http.StatusBadRequest
	case errors.Is(err, ErrSampleNotFound), errors.Is(err, ErrStationNotFound):
		status = http.StatusNotFound
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
