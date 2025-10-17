package api

import (
	"log/slog"
	"net/http"
	auth "scheduler/appointment-service/internal/auth"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const kLoginTimeOut = 2 * time.Minute

func GenerateOTPKeyBasic(login, company string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      company,
		AccountName: login,
	})
}

type OTPSecret = string

type OTPSecretGetter interface {
	Get(UserID) (OTPSecret, error)
}

type OTPSecretSetter interface {
	Set(UserID, OTPSecret) error
}

func GenerateOTPKey(secretStore OTPSecretGetter) http.HandlerFunc {
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
func ValidateOTPassword(sesStore *auth.UserSessionStore, secretStore OTPSecretGetter) http.HandlerFunc {
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

		s, err := sesStore.Get(r)
		if err != nil {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword] session", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if status, err := s.GetAuthStatus(); err != nil {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword]", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if status == auth.StatusAuthenticated {
			slog.WarnContext(r.Context(), "[ValidateOTPassword] already authorized")
			http.Error(w, "already authorized?", http.StatusBadRequest)
		} else if status != auth.Status2faRequired {
			slog.WarnContext(r.Context(), "[ValidateOTPassword]", "unexpected status", status)
			http.Error(w, "login step missed?", http.StatusUnauthorized)
			return
		}

		if tp, err := s.GetTimeStamp(); err != nil {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword]", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if time.Since(time.Unix(tp, 0)) > kLoginTimeOut {
			slog.DebugContext(r.Context(), "[ValidateOTPassword] Timeout")

			err = sesStore.Reset(w, r)
			if err != nil {
				slog.ErrorContext(r.Context(), "[ValidateOTPassword] session", "err", err.Error())
			}
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}

		uid, err := s.GetUserID()
		if err != nil {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword]", "err", err.Error())
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

		err = sesStore.Authenticate(uid, w, r)
		if err != nil {
			slog.ErrorContext(r.Context(), "[ValidateOTPassword] session", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
