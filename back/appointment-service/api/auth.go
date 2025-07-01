package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"

	"github.com/gorilla/sessions"
)

var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

type UserID = common.ID
type UserIdKey struct{}

type UserChecker interface {
	IsExist(uid UserID) (bool, error)
	Check(username string, password string) (UserID, error)
}

func GetUserID(c context.Context) UserID {
	uid, _ := c.Value(UserIdKey{}).(string)
	return uid
}

func LoginHandler(store sessions.Store, uc UserChecker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			if errors.Is(err, ErrNotFound) || errors.Is(err, ErrUnauthorized) {
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
	})
}

func LogoutHandler(store sessions.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "sid")
		delete(session.Values, "uid")
		session.Save(r, w)
		w.WriteHeader(http.StatusOK)
	})
}

func CheckAuthHandler(store sessions.Store, uc UserChecker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "sid")
		if err != nil {
			slog.WarnContext(r.Context(), "sessions", "err", err.Error())
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
			//TODO print request_id ?
			slog.WarnContext(r.Context(), "sessions", "err", "uid cast error")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		//TODO can we avoid it in the next steps by erasing uid field?
		exist, err := uc.IsExist(uid)
		if err != nil {
			slog.WarnContext(r.Context(), "sessions", "err", err.Error())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if !exist {
			w.WriteHeader(http.StatusNetworkAuthenticationRequired)
			return
		}

		ctx := context.WithValue(r.Context(), UserIdKey{}, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
