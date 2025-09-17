package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type botResult struct {
	BotId    string `json:"bot_id"`
	BotToken string `json:"token"`
}

func AddUserBotHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		businessId, ok := GetUserID(r.Context())
		if !ok {
			panic("businessId not found")
		}

		const secretLength = 32
		botToken := common.GenerateSecretKey(secretLength)
		botId := uuid.New().String()

		_, err := s.AddBot(botId, botToken, businessId)
		if err != nil {
			slog.WarnContext(r.Context(), "AddUserBot", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		result := botResult{
			BotId:    botId,
			BotToken: botToken,
		}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			slog.WarnContext(r.Context(), "AddUserBot. encode", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func DeleteUserBotHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		businessId, ok := GetUserID(r.Context())
		if !ok {
			panic("businessId not found")
		}

		vars := mux.Vars(r)

		err := s.DeleteBot(businessId, vars["bot_id"])
		if err != nil {
			slog.WarnContext(r.Context(), "DeleteBot", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
