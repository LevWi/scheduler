package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	swagger "scheduler/appointment-service/api/types"
	types "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/storage"

	"github.com/gorilla/mux"
)

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func getTimeFromHeader(key string, h http.Header) (time.Time, error) {
	dtStr := h.Get(key)
	if dtStr == "" {
		return time.Time{}, fmt.Errorf("%s not found", key)
	}

	dt, err := parseTime(dtStr)
	if err != nil {
		return time.Time{}, err
	}
	return dt, nil
}

func SlotsBusinessIdGetFunc(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		SlotsBusinessIdGet(s, w, r)
	}
}

func SlotsBusinessIdGet(s *storage.Storage, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	businessID, ok := vars["business_id"]
	if !ok {
		slog.Warn("business_id not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dateStart, err := getTimeFromHeader("date_start", r.Header)
	if err != nil {
		slog.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	dateEnd, err := getTimeFromHeader("date_end", r.Header)
	if err != nil {
		slog.Warn(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slots, err := s.GetSlotInRange(types.ID(businessID), dateStart, dateEnd)
	if err != nil {
		slog.Warn(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response swagger.AvailableSlots
	response.QueryId = "TBD" //TODO

	for _, slot := range slots {
		response.Slots = append(response.Slots, swagger.Slot{
			ClientId: string(slot.Client),
			TpStart:  slot.Start,
			Len:      int32(slot.Len),
		})
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Warn(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func SlotsBusinessIdPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
