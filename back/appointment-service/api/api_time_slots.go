package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	swagger "scheduler/appointment-service/api/types"
	common "scheduler/appointment-service/internal"
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

func SlotsBusinessIdPostFunc(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		SlotsBusinessIdPost(s, w, r)
	}
}

// TODO prepare error, prepare QueryId
func SlotsBusinessIdGet(s *storage.Storage, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	businessID := vars["business_id"]
	if businessID == "" {
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

	slots, err := s.GetBusySlotsInRange(businessID, dateStart, dateEnd)
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

func SlotsBusinessIdPost(s *storage.Storage, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	businessID := vars["business_id"]
	if businessID == "" {
		slog.Warn("business_id not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var jsonSlots []swagger.Slot
	err := json.NewDecoder(r.Body).Decode(&jsonSlots)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	appointment := common.Appointment{Business: businessID, Slots: make([]common.Slot, 0, len(jsonSlots))}
	for _, jsSlot := range jsonSlots {
		appointment.Slots = append(appointment.Slots, common.Slot{
			Client: jsSlot.ClientId,
			Start:  jsSlot.TpStart,
			Len:    int(jsSlot.Len),
		})
	}

	slog.Info(fmt.Sprintf("%v", appointment))

	//TODO work with database

	w.WriteHeader(http.StatusOK)
}
