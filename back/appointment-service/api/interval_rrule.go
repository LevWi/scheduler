package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"
)

type RRuleWithTypeWrap struct {
	Value common.IntervalRRuleWithType
	JSON  string
}

type RRuleStorage interface {
	AddRRule(user common.ID, rule RRuleWithTypeWrap) error
}

// add if !json.Valid([]byte(ruleWrap.JSON))
func AddBusinessRuleHandler(rs RRuleStorage) http.HandlerFunc {
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
				slog.WarnContext(r.Context(), "Read HTTP body", "err", err.Error())
				return
			}
		}()

		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		var item RRuleWithTypeWrap
		if err := json.Unmarshal(bodyBytes, &item.Value); err != nil {
			slog.WarnContext(r.Context(), "decoding JSON", "err", err.Error())
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		item.JSON = string(bodyBytes)

		err = rs.AddRRule(uid, item)
		if err != nil {
			slog.WarnContext(r.Context(), "AddRule", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
