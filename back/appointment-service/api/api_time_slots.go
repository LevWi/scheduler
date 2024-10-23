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

// TODO add lock?
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
		slog.WarnContext(r.Context(), "business_id not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dateStart, err := getTimeFromHeader("date_start", r.Header)
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	dateEnd, err := getTimeFromHeader("date_end", r.Header)
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slots, err := s.GetAvailableSlotsInRange(businessID, dateStart, dateEnd)
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response swagger.AvailableSlots
	response.QueryId = r.Context().Value(RequestIdKey{}).(string)

	for _, slot := range slots {
		response.Slots = append(response.Slots, swagger.Slot{
			ClientId: string(slot.Client),
			TpStart:  slot.Start,
			Len:      int32(slot.End.Sub(slot.Start).Minutes()),
		})
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// TODO prepare error, prepare QueryId
func SlotsBusinessIdPost(s *storage.Storage, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	businessID := vars["business_id"]
	if businessID == "" {
		slog.WarnContext(r.Context(), "business_id not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var jsonSlots []swagger.Slot
	err := json.NewDecoder(r.Body).Decode(&jsonSlots)
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(jsonSlots) != 0 {
		slog.WarnContext(r.Context(), "Missed slots")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	{
		tmp := make(map[int64]struct{}, len(jsonSlots))
		for _, slot := range jsonSlots {
			if slot.TpStart.IsZero() {
				slog.WarnContext(r.Context(), "Nullable variable")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if slot.TpStart.Compare(time.Now()) == -1 {
				slog.WarnContext(r.Context(), "slot in the past", slog.Any("slot", slot))
			}
			tmp[slot.TpStart.Unix()] = struct{}{}
		}
		if len(tmp) != len(jsonSlots) {
			slog.WarnContext(r.Context(), "Duplicated slots", slog.Any("payload", jsonSlots))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	tpStart := jsonSlots[0].TpStart
	tpEnd := jsonSlots[0].TpStart
	appointment := common.Appointment{Business: businessID, Slots: make([]common.Slot, 0, len(jsonSlots))}
	for i := 1; i < len(jsonSlots); i++ {
		end := jsonSlots[i].TpStart.Add(time.Duration(jsonSlots[i].Len) * time.Minute)
		appointment.Slots = append(appointment.Slots, common.Slot{
			Client: jsonSlots[i].ClientId,
			Interval: common.Interval{
				Start: jsonSlots[i].TpStart,
				End:   end},
		})
		if tpStart.Compare(jsonSlots[i].TpStart) == 1 {
			tpStart = jsonSlots[i].TpStart
		}
		if tpEnd.Compare(end) == -1 {
			tpEnd = end
		}
	}

	slog.InfoContext(r.Context(), fmt.Sprint(appointment))

	availableSlots, err := s.GetAvailableSlotsInRange(businessID, tpStart, tpEnd)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(availableSlots) == 0 {
		slog.WarnContext(r.Context(), "No available slots")
		w.WriteHeader(http.StatusConflict)
		return
	}

	availableSlotsMap := make(map[int64]common.Slot, len(availableSlots))
	for _, slot := range availableSlots {
		availableSlotsMap[slot.Start.Unix()] = slot
	}

	if len(availableSlots) != len(availableSlotsMap) {
		slog.WarnContext(r.Context(), "Duplicate detected", slog.Any("slot", availableSlots))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, requested := range appointment.Slots {
		unxTime := requested.Start.Unix()
		if available, exist := availableSlotsMap[unxTime]; exist {
			if available.Len < requested.Len {
				slog.WarnContext(r.Context(), "Slot is too big", slog.Any("slot", requested),
					slog.Any("available", available))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			slog.InfoContext(r.Context(), "Approved", slog.Any("slot", requested))
			delete(availableSlotsMap, unxTime)
		} else {
			slog.WarnContext(r.Context(), "Slot is not available", slog.Any("slot", requested))
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
