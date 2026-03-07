package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/dbase/backend/slots"

	"github.com/gorilla/mux"
)

type RRuleWithType = common.IntervalRRuleWithType
type RRuleResult = slots.DbBusinessRule

type RRuleStorageI interface {
	AddBusinessRule(user common.ID, rule RRuleWithType) (slots.RuleID, error)
	DeleteBusinessRule(user common.ID, ruleId common.ID) error
	GetBusinessRules(user common.ID) ([]RRuleResult, error)
}

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

		_, err = rs.AddBusinessRule(uid, rule)
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

		rules, err := rs.GetBusinessRules(uid)
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

		err := rs.DeleteBusinessRule(uid, id)
		if err != nil {
			slog.WarnContext(r.Context(), "DeleteBusinessRule", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
