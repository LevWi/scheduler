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
	"scheduler/appointment-service/internal/dbase/auth"
	slotsdb "scheduler/appointment-service/internal/dbase/backend/slots"

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
// TODO prepare error, prepare QueryId
func SlotsBusinessIdGetFunc(s *slotsdb.TimeSlotsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}

type AuthResult struct {
	Business common.ID
	Client   common.ID
}

type AddSlotsAuth interface {
	Authorization(r *http.Request) (AuthResult, error)
}

type AddSlotsAuthFromUrl struct {
}

func (AddSlotsAuthFromUrl) Authorization(r *http.Request) (AuthResult, error) {
	var result AuthResult
	businessID, ok := GetUserID(r.Context())
	if !ok {
		return result, fmt.Errorf("businessId: %w", common.ErrNotFound)
	}

	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		return result, fmt.Errorf("client_id: %w", common.ErrNotFound)
	}

	result.Business = businessID
	result.Client = clientID
	return result, nil
}

type AddSlotsAuthOneOffToken auth.OneOffTokenStorage

func (a *AddSlotsAuthOneOffToken) Authorization(r *http.Request) (AuthResult, error) {
	var result AuthResult
	token := r.URL.Query().Get("token")
	if token == "" {
		return result, fmt.Errorf("token: %w", common.ErrNotFound)
	}

	entry, err := (*auth.OneOffTokenStorage)(a).ExchangeToken(token)
	if err != nil {
		return result, err
	}
	result.Business = entry.BusinessID
	result.Client = entry.ClientID
	return result, err
}

// TODO Fix it, change swagger.Slot, prepare error, prepare QueryId
func SlotsBusinessIdPostFunc(au AddSlotsAuth, s *slotsdb.TimeSlotsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		authResult, err := au.Authorization(r)
		if err != nil {
			slog.WarnContext(r.Context(), "[SlotsBusinessIdPost]", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}

		var jsonSlots []swagger.Slot
		err = json.NewDecoder(r.Body).Decode(&jsonSlots)
		if err != nil {
			slog.WarnContext(r.Context(), err.Error())
			w.WriteHeader(http.StatusBadRequest) //TODO Use http.Error()?
			return
		}

		if len(jsonSlots) == 0 {
			slog.WarnContext(r.Context(), "Missed slots")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tpInterval := common.Interval{Start: jsonSlots[0].TpStart, End: jsonSlots[0].TpStart}
		slots := make(common.Intervals, 0, len(jsonSlots))
		for i := 0; i < len(jsonSlots); i++ {
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

		availableSlots, err := s.GetAvailableSlotsInRange(authResult.Business, tpInterval)
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

		s.AddSlots(slotsdb.AddSlotsData{
			Business: authResult.Business,
			Client:   authResult.Client,
			Slots:    slots,
		})

		w.WriteHeader(http.StatusOK)
	}
}
