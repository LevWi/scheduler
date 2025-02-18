package server

import (
	"errors"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"

	"github.com/gorilla/sessions"
)

var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

type UserID = common.ID

type UserChecker interface {
	Check(username string, password string) (UserID, error)
}

func loginHandler(store sessions.Store, uc UserChecker, h HttpIO) {
	if h.Req.Method != "POST" {
		slog.WarnContext(h.Req.Context(), "wrong method")
		h.Wrt.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	err := h.Req.ParseForm()
	if err != nil {
		slog.DebugContext(h.Req.Context(), "parse request err")
		h.Wrt.WriteHeader(http.StatusBadRequest)
		return
	}

	username := h.Req.PostForm.Get("username")
	password := h.Req.PostForm.Get("password")

	uid, err := uc.Check(username, password)
	if err != nil {
		slog.WarnContext(h.Req.Context(), "try login", username, err.Error())
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrUnauthorized) {
			http.Error(h.Wrt, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		h.Wrt.WriteHeader(http.StatusInternalServerError)
		return
	}

	session, err := store.Get(h.Req, "sid")
	if err != nil {
		slog.ErrorContext(h.Req.Context(), "sessions", "err", err.Error())
	}

	session.Values["uid"] = uid
	session.Save(h.Req, h.Wrt)

	h.Wrt.WriteHeader(http.StatusOK)
}

func logoutHandler(store sessions.Store, h HttpIO) {
	session, _ := store.Get(h.Req, "sid")
	delete(session.Values, "uid")
	session.Save(h.Req, h.Wrt)
	h.Wrt.WriteHeader(http.StatusOK)
}
