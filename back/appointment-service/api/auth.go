package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"

	"github.com/gorilla/sessions"
)

type UserID = common.ID
type UserIdKey struct{}

type UserChecker interface {
	IsExist(uid UserID) error
	Check(username string, password string) (UserID, error)
}

func GetUserID(c context.Context) UserID {
	uid, _ := c.Value(UserIdKey{}).(string)
	return uid
}

// Required application/x-www-form-urlencoded format
// TODO add protection against brute force
func LoginHandler(store sessions.Store, uc UserChecker) http.HandlerFunc {
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

		username := r.PostForm.Get("username")
		password := r.PostForm.Get("password")

		uid, err := uc.Check(username, password)
		if err != nil {
			slog.WarnContext(r.Context(), "try login", username, err.Error())
			if errors.Is(err, common.ErrNotFound) || errors.Is(err, common.ErrUnauthorized) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		session, err := store.Get(r, "sid")
		if err != nil {
			slog.WarnContext(r.Context(), "sessions", "err", err.Error())
			// TODO where error handling?
		}

		session.Values["uid"] = uid
		session.Save(r, w)

		w.WriteHeader(http.StatusOK)
	}
}

func LogoutHandler(store sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(GetUserID(r.Context())) == 0 {
			panic("uid not found")
		}
		session, _ := store.Get(r, "sid")

		delete(session.Values, "uid")
		session.Save(r, w)
		w.WriteHeader(http.StatusOK)
	}
}

func CheckAuthHandler(store sessions.Store, uc UserChecker, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "sid")
		if err != nil {
			slog.WarnContext(r.Context(), "[CheckAuthHandler] sessions", "err", err.Error())
			http.Error(w, "sid internal error", http.StatusInternalServerError)
			return
		}

		uidi, ok := session.Values["uid"]
		if !ok {
			w.WriteHeader(http.StatusNetworkAuthenticationRequired)
			return
		}

		uid, ok := uidi.(string)
		if !ok {
			delete(session.Values, "uid")
			session.Save(r, w)

			//TODO print request_id ?
			slog.WarnContext(r.Context(), "sessions", "err", "uid cast error")
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
func DeleteUserHandler(store sessions.Store, deleteUser func(common.ID, string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := GetUserID(r.Context())
		if len(uid) == 0 {
			panic("uid not found")
		}

		err := r.ParseForm()
		if err != nil {
			slog.DebugContext(r.Context(), "parse request err")
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

		session, err := store.Get(r, "sid")
		if err != nil {
			slog.WarnContext(r.Context(), "[DeleteUserHandler] sessions", "err", err.Error())
			http.Error(w, "sid internal error", http.StatusInternalServerError)
			return
		}

		delete(session.Values, "uid")
		session.Save(r, w)
		w.WriteHeader(http.StatusOK)
	}
}
