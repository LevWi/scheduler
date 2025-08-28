package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"
	auth "scheduler/appointment-service/internal/auth"
)

type UserID = common.ID //TODO names conflict
type UserIdKey struct{}

var ErrSecurityRestriction = errors.New("security restriction")

type ExistingUserCheck interface {
	IsExist(uid UserID) error
}

type AuthUserCheck interface {
	Check(username string, password string) (UserID, error)
}

func GetUserID(c context.Context) (UserID, bool) {
	uid, ok := c.Value(UserIdKey{}).(string)
	return uid, ok
}

// Required application/x-www-form-urlencoded format
// TODO add Request Throttling , CSRF Protection
func LoginHandler(ses *auth.UserSessionStore, ac AuthUserCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			slog.WarnContext(r.Context(), "wrong method")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		err := r.ParseForm()
		if err != nil {
			slog.DebugContext(r.Context(), "parse request err")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		session, err := ses.Get(r)
		if err != nil {
			slog.WarnContext(r.Context(), "sessions", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		username := r.PostForm.Get("username")
		password := r.PostForm.Get("password")

		uid, err := ac.Check(username, password)
		if err != nil {
			slog.WarnContext(r.Context(), "try login", username, err.Error())
			switch {
			case errors.Is(err, common.ErrNotFound), errors.Is(err, common.ErrUnauthorized):
				session.DelAuthStatus()
				session.Save(r, w)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			case errors.Is(err, ErrSecurityRestriction):
				// TODO need notification about brute force attack
				http.Error(w, "please try later", http.StatusTooManyRequests)
			default:
				w.WriteHeader(http.StatusInternalServerError)
				slog.WarnContext(r.Context(), "[LoginHandler] user check", "err", err.Error())
			}
			return
		}

		session.SetUserID(uid)
		session.SetAuthStatus(auth.Status2faRequired)
		err = session.Save(r, w)
		if err != nil {
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized) //TODO?
	}
}

func LogoutHandler(store *auth.UserSessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := GetUserID(r.Context()); !ok {
			panic("uid not found")
		}
		err := store.Reset(w, r)
		if err != nil {
			slog.WarnContext(r.Context(), "LogoutHandler", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func CheckAuthHandler(store *auth.UserSessionStore, uc ExistingUserCheck, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := store.AuthenticationCheck(r)

		if err != nil {
			if errors.Is(err, common.ErrUnauthorized) {
				slog.WarnContext(r.Context(), "[CheckAuthHandler]", "err", err.Error())
				w.WriteHeader(http.StatusNetworkAuthenticationRequired)
				return
			}
			slog.WarnContext(r.Context(), "[CheckAuthHandler] unexpected", "err", err.Error())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		//TODO can we avoid it in the next steps by erasing uid field?
		err = uc.IsExist(uid)

		if err == common.ErrNotFound {
			w.WriteHeader(http.StatusNetworkAuthenticationRequired)
			return
		}

		if err != nil {
			slog.WarnContext(r.Context(), "sessions", "err", err.Error())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), UserIdKey{}, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// TODO if cache will be used make sure that user was cleared from there
func DeleteUserHandler(store *auth.UserSessionStore, deleteUser func(common.ID, string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := GetUserID(r.Context())
		if !ok {
			panic("uid not found")
		}

		err := r.ParseForm()
		if err != nil {
			slog.DebugContext(r.Context(), "parse request", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		password := r.Form.Get("password")
		if len(password) == 0 {
			slog.DebugContext(r.Context(), "password expected")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = deleteUser(uid, password)
		if err != nil {
			slog.WarnContext(r.Context(), "delete user error", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = store.Reset(w, r)
		if err != nil {
			slog.WarnContext(r.Context(), "[DeleteUserHandler]", "err", err.Error())
			http.Error(w, "Remove session", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
