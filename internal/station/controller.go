package station

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

// @Summary Create a station
// @Description Create a monitoring station with nested location and thresholds. The code must be unique.
// @Tags stations
// @Accept json
// @Produce json
// @Param station body StationRequest true "Station data"
// @Success 201 {object} Station
// @Failure 400 {object} map[string]string "Bad Request (e.g. Invalid JSON, Name is required, pH threshold out of range)"
// @Failure 409 {object} map[string]string "Conflict (station code already exists)"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /stations [post]
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req StationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidJSON)
		return
	}

	station := req.toStation()
	if err := c.service.CreateStation(r.Context(), &station); err != nil {
		writeError(w, statusFor(err), err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(station)
}

// @Summary List stations
// @Description Retrieve every monitoring station
// @Tags stations
// @Produce json
// @Success 200 {array} Station
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /stations [get]
func (c *Controller) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stations, err := c.service.ListStations(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stations)
}

// @Summary Get a station
// @Description Retrieve one station by ID
// @Tags stations
// @Produce json
// @Param id path string true "Station ID"
// @Success 200 {object} Station
// @Failure 400 {object} map[string]string "Bad Request (invalid ObjectID format)"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /stations/{id} [get]
func (c *Controller) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	station, err := c.service.GetStation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, statusFor(err), err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(station)
}

// @Summary Update a station
// @Description Replace every editable field of a station (idempotent PUT)
// @Tags stations
// @Accept json
// @Produce json
// @Param id path string true "Station ID"
// @Param station body StationRequest true "Full station replacement data"
// @Success 200 {object} Station
// @Failure 400 {object} map[string]string "Bad Request (e.g. Invalid JSON, invalid ObjectID, validation error)"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 409 {object} map[string]string "Conflict (station code already exists)"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /stations/{id} [put]
func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req StationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidJSON)
		return
	}

	station := req.toStation()
	updated, err := c.service.UpdateStation(r.Context(), r.PathValue("id"), &station)
	if err != nil {
		writeError(w, statusFor(err), err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(updated)
}

// @Summary Delete a station
// @Description Delete a station. Refused with 409 while any sample still references it.
// @Tags stations
// @Produce json
// @Param id path string true "Station ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Bad Request (invalid ObjectID format)"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 409 {object} map[string]string "Conflict (station still has samples)"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /stations/{id} [delete]
func (c *Controller) DeleteByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := c.service.DeleteStation(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, statusFor(err), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var errInvalidJSON = errors.New("Invalid JSON.")

var validationErrors = []error{
	ErrInvalidID,
	ErrNameRequired,
	ErrRiverRequired,
	ErrCodeRequired,
	ErrCityRequired,
	ErrStateRequired,
	ErrThresholdParameterRequired,
	ErrThresholdRange,
	ErrInvalidPHThreshold,
}

func statusFor(err error) int {
	for _, target := range validationErrors {
		if errors.Is(err, target) {
			return http.StatusBadRequest
		}
	}
	switch {
	case errors.Is(err, ErrStationNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrStationHasSamples), errors.Is(err, ErrDuplicateCode):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
