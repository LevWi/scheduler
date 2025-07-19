package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const kLoginTimeOut = int64(45)

func GenerateOTPKeyBasic(login, company string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      company,
		AccountName: login,
	})
}

type OTPSecret = string

type OTPSecretStore interface {
	Get(UserID) (OTPSecret, error)
	Set(UserID, OTPSecret) error
}

func GenerateOTPKey(secretStore OTPSecretStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// uid, ok := GetUserID(r.Context())
		// if !ok {
		// 	panic("uid not found")
		// }
		//TBD
	}
}

// Required application/x-www-form-urlencoded format
// Expected POST method
func ValidateOTPassword(sesStore sessions.Store, secretStore OTPSecretStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.DebugContext(r.Context(), "[ValidateOTPassword] parse request err")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		passcode := r.PostForm.Get("passcode")
		if passcode == "" {
			slog.DebugContext(r.Context(), "[ValidateOTPassword] passcode not found")
			http.Error(w, "passcode not found", http.StatusBadRequest)
			return
		}

		s, err := sesStore.Get(r, UserSessionName)
		if err != nil {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword] session", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if status, ok := s.Values[KeyAuthStatus].(string); !ok {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword] KeyAuthStatus not found")
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if status == StatusAuthorized {
			slog.WarnContext(r.Context(), "[ValidateOTPassword] already authorized")
			http.Error(w, "already authorized?", http.StatusBadRequest)
		} else if status != Status2faRequired {
			slog.WarnContext(r.Context(), "[ValidateOTPassword]", "unexpected status", status)
			http.Error(w, "login step missed?", http.StatusUnauthorized)
			return
		}

		if ts, ok := s.Values[KeyTimestamp].(int64); !ok {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword] KeyTimestamp not found")
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if (time.Now().UTC().Unix() - ts) > kLoginTimeOut {
			slog.DebugContext(r.Context(), "[ValidateOTPassword] Timeout")

			s.Options.MaxAge = -1
			err = s.Save(r, w)
			if err != nil {
				slog.ErrorContext(r.Context(), "[ValidateOTPassword] session", "err", err.Error())
			}
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}

		uid, ok := s.Values[KeyUserID].(string)
		if !ok {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword] uid not found")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		secret, err := secretStore.Get(uid)
		if err != nil {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword] secret", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !totp.Validate(passcode, secret) {
			slog.DebugContext(r.Context(), "[ValidateOTPassword] passcode wrong")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		s.Options.MaxAge = -1 // Delete login cookie

		s, err = sesStore.Get(r, UserSessionName)
		if err != nil {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword] user session", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.Values[KeyUserID] = uid
		s.Values[KeyAuthStatus] = StatusAuthorized
		delete(s.Values, KeyTimestamp)

		err = s.Save(r, w)
		if err != nil {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword] session", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
