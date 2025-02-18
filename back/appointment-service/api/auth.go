package server

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	common "scheduler/appointment-service/internal"

	"github.com/gorilla/sessions"
)

// Note: Don't store your key in your source code. Pass it via an
// environmental variable, or flag (or both), and don't accidentally commit it
// alongside your code. Ensure your key is sufficiently random - i.e. use Go's
// crypto/rand or securecookie.GenerateRandomKey(32) and persist the result.
var store sessions.Store

func init() {
	const UserCookieAge = 85400 * 5 // 5 days
	s := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	s.MaxAge(UserCookieAge)
	store = s
}

var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

type UserID = common.ID

type UserChecker interface {
	Check(username string, password string) (UserID, error)
}

func loginHandler(uc UserChecker, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		slog.WarnContext(r.Context(), "Wrong method")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseForm()
	if err != nil {
		slog.DebugContext(r.Context(), "Parse request err")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	uid, err := uc.Check(username, password)
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrUnauthorized) {
			slog.WarnContext(r.Context(), "Wrong user/password")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		slog.WarnContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session, err := store.Get(r, "sid")
	if err != nil {
		slog.ErrorContext(r.Context(), "sessions", "err", err.Error())
	}

	session.Values["uid"] = uid
	session.Save(r, w)

	w.WriteHeader(http.StatusOK)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "sid")
	delete(session.Values, "uid")
	session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}
