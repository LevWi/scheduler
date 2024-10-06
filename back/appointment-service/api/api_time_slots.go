package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	types "scheduler/appointment-service/internal"
	storage "scheduler/appointment-service/internal/storage"

	"github.com/gorilla/mux"
)

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func getTimeFromHeader(key string, h http.Header) (time.Time, error) {
	dtStr := h.Get(key)
	if dtStr == "" {
		return time.Time{}, errors.New(fmt.Sprintf("%s not found", key))
	}

	dt, err := parseTime(dtStr)
	if err != nil {
		return time.Time{}, err
	}
	return dt, nil
}

func SlotsBusinessIdGet(w http.ResponseWriter, r *http.Request) {
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

	slots, err := storage.GetSlotInRange(storage.DB, types.ID(businessID), dateStart, dateEnd)

	w.WriteHeader(http.StatusOK)
}

func SlotsBusinessIdPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
