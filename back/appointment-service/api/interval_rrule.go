package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"

	"github.com/gorilla/mux"
)

type RRuleWithType struct {
	Value common.IntervalRRuleWithType
	JSON  string
}

func (r *RRuleWithType) MarshalJSON() ([]byte, error) {
	if len(r.JSON) != 0 {
		return []byte(r.JSON), nil
	}

	return json.Marshal(r.Value)
}

func (r *RRuleWithType) UnmarshalJSON(in []byte) error {
	err := json.Unmarshal(in, &r.Value)
	if err != nil {
		return err
	}

	r.JSON = string(in)
	return nil
}

type RRuleResult struct {
	Id    common.ID
	Rrule RRuleWithType
}

type RRuleStorageI interface {
	AddRRule(user common.ID, rule RRuleWithType) error
	RemoveRRule(user common.ID, ruleId common.ID) error
	GetRRules(user common.ID) ([]RRuleResult, error)
}

// add if !json.Valid([]byte(ruleWrap.JSON))
func AddBusinessRuleHandler(rs RRuleStorageI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := GetUserID(r.Context())
		if !ok {
			panic("uid not found")
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			slog.WarnContext(r.Context(), "Read HTTP body", "err", err.Error())
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
		defer func() {
			err = r.Body.Close()
			if err != nil {
				slog.WarnContext(r.Context(), "Close HTTP body", "err", err.Error())
				return
			}
		}()

		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		var rule RRuleWithType
		if err := json.Unmarshal(bodyBytes, &rule); err != nil {
			slog.WarnContext(r.Context(), "decoding JSON", "err", err.Error())
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		err = rs.AddRRule(uid, rule)
		if err != nil {
			slog.WarnContext(r.Context(), "AddRule", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetBusinessRulesHandler(rs RRuleStorageI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := GetUserID(r.Context())
		if !ok {
			panic("uid not found")
		}

		rules, err := rs.GetRRules(uid)
		if err != nil {
			slog.WarnContext(r.Context(), "GetRules", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		err = json.NewEncoder(w).Encode(rules)
		if err != nil {
			slog.WarnContext(r.Context(), "encoding JSON", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func DelBusinessRuleHandler(rs RRuleStorageI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := GetUserID(r.Context())
		if !ok {
			panic("uid not found")
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err := rs.RemoveRRule(uid, id)
		if err != nil {
			slog.WarnContext(r.Context(), "DeleteBusinessRule", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
