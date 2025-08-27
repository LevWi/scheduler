package auth

import (
	"errors"
	"fmt"
	"net/http"
	common "scheduler/appointment-service/internal"
	"time"

	"github.com/gorilla/sessions"
)

type UserSessionStore struct {
	S sessions.Store
	O []AuthCheckOptions
}

type AuthCheckOptions func(*sessions.Session) error

func (s *UserSessionStore) Get(r *http.Request) (*sessions.Session, error) {
	return s.S.Get(r, CookieUserSessionName)
}

func (s *UserSessionStore) Save(id common.ID, w http.ResponseWriter, r *http.Request) error {
	session, err := s.Get(r)
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

func (s *UserSessionStore) Reset(w http.ResponseWriter, r *http.Request) error {
	session, err := s.Get(r)
	if err != nil {
		return fmt.Errorf("[Reset] get session: %w", err)
	}

	delete(session.Values, CookieKeyUserID)
	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("[Reset] save session: %w", err)
	}
	return nil
}

func (s *UserSessionStore) Check(r *http.Request) (common.ID, error) {
	session, err := s.Get(r)
	if err != nil {
		return "", fmt.Errorf("[Check] get session: %w", err)
	}

	uid, ok := session.Values[CookieKeyUserID].(string)
	if !ok {
		return "", fmt.Errorf("%w: [Check] %s not found", common.ErrUnauthorized, CookieKeyUserID)
	}

	for _, opt := range s.O {
		err := opt(session)
		if err != nil {
			return "", fmt.Errorf("[Check] option check fail: %w", err)
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

func NewUserSessionStore(s sessions.Store, opts ...AuthCheckOptions) *UserSessionStore {
	return &UserSessionStore{
		S: s,
		O: opts,
	}
}
