package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	swagger "scheduler/appointment-service/api/types"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/storage"

	"github.com/gorilla/mux"
)

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func getTimeFromURL(key string, v url.Values) (time.Time, error) {
	dtStr := v.Get(key)
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

	dateStart, err := getTimeFromURL("date_start", r.URL.Query())
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	dateEnd, err := getTimeFromURL("date_end", r.URL.Query())
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slots, err := s.GetAvailableSlotsInRange(businessID, common.Interval{Start: dateStart, End: dateEnd})
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response swagger.AvailableSlots
	response.QueryId = r.Context().Value(RequestIdKey{}).(string)

	for _, slot := range slots {
		response.Slots = append(response.Slots, swagger.Slot{
			TpStart: slot.Start,
			Len:     int32(slot.End.Sub(slot.Start).Minutes()),
		})
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//TODO http: superfluous response.WriteHeader call
	w.WriteHeader(http.StatusOK)
}

// TODO Fix it, change swagger.Slot, prepare error, prepare QueryId
func SlotsBusinessIdPost(s *storage.Storage, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	businessID := vars["business_id"]
	if businessID == "" {
		slog.WarnContext(r.Context(), "business_id not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientID := vars["client_id"]
	if clientID == "" {
		slog.WarnContext(r.Context(), "client_id not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var jsonSlots []swagger.Slot
	err := json.NewDecoder(r.Body).Decode(&jsonSlots)
	if err != nil {
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusBadRequest) //TODO Use http.Error()?
		return
	}

	if len(jsonSlots) != 0 {
		slog.WarnContext(r.Context(), "Missed slots")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO where get client id ?

	tpInterval := common.Interval{Start: jsonSlots[0].TpStart, End: jsonSlots[0].TpStart}
	slots := make(common.Intervals, 0, len(jsonSlots))
	for i := 1; i < len(jsonSlots); i++ {
		if jsonSlots[i].TpStart.IsZero() {
			slog.WarnContext(r.Context(), "Nullable variable")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if jsonSlots[i].TpStart.Before(time.Now()) {
			slog.WarnContext(r.Context(), "slot in the past", slog.Any("slot", jsonSlots[i]))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		end := jsonSlots[i].TpStart.Add(time.Duration(jsonSlots[i].Len) * time.Minute)
		slots = append(slots, common.Interval{
			Start: jsonSlots[i].TpStart,
			End:   end})
		if tpInterval.Start.After(jsonSlots[i].TpStart) {
			tpInterval.Start = jsonSlots[i].TpStart
		}
		if tpInterval.End.Before(end) {
			tpInterval.End = end
		}
	}

	slots.SortByStart()

	if slots.HasOverlaps() {
		slog.ErrorContext(r.Context(), "Appointment slots have overlap")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slog.InfoContext(r.Context(), fmt.Sprint(slots))

	availableSlots, err := s.GetAvailableSlotsInRange(businessID, tpInterval)
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

	// TODO Not optimal
	for _, el := range slots {
		if !availableSlots.IsFit(el) {
			slog.WarnContext(r.Context(), "Conflict with available slot")
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	s.AddSlots(storage.AddSlotsData{
		Business: businessID,
		Client:   clientID,
		Slots:    slots,
	})

	w.WriteHeader(http.StatusOK)
}
