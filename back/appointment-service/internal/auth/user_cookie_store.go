package auth

import (
	"errors"
	"fmt"
	"net/http"
	common "scheduler/appointment-service/internal"
	"time"

	"github.com/gorilla/sessions"
)

type UserCookieStore struct {
	S sessions.Store
	O []AuthCheckOptions
}

type AuthCheckOptions func(*sessions.Session) error

func (s *UserCookieStore) Save(id common.ID, w http.ResponseWriter, r *http.Request) error {
	session, err := s.S.Get(r, CookieUserSessionName)
	if err != nil {
		return fmt.Errorf("[Save] get session: %w", err)
	}

	session.Values[CookieKeyUserID] = string(id)
	session.Values[CookieKeyTimestamp] = time.Now().Unix()
	err = s.S.Save(r, w, session)
	if err != nil {
		return fmt.Errorf("[Save] save session: %w", err)
	}

	return nil
}

func (s *UserCookieStore) AuthCheck(r *http.Request) (common.ID, error) {
	session, err := s.S.Get(r, CookieUserSessionName)
	if err != nil {
		return "", fmt.Errorf("[AuthCheck] get session: %w", err)
	}

	uid, ok := session.Values[CookieKeyUserID].(string)
	if !ok {
		return "", fmt.Errorf("[AuthCheck] %s %w", CookieKeyUserID, common.ErrNotFound)
	}

	for _, opt := range s.O {
		err := opt(session)
		if err != nil {
			return "", fmt.Errorf("[AuthCheck] option check fail: %w", err)
		}
	}

	return common.ID(uid), nil
}

var ErrSessionExpired = errors.New("session expired")

func WithSessionLifeTime(d time.Duration) AuthCheckOptions {
	return func(s *sessions.Session) error {
		tp, ok := s.Values[CookieKeyTimestamp].(int64)
		if !ok {
			return fmt.Errorf("[WithSessionLifeTime] %s %w", CookieKeyTimestamp, common.ErrNotFound)
		}

		if time.Since(time.Unix(tp, 0)) > d {
			return ErrSessionExpired
		}

		return nil
	}
}
