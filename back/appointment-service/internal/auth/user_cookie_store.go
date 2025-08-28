package auth

import (
	"errors"
	"fmt"
	"net/http"
	common "scheduler/appointment-service/internal"
	"time"

	"github.com/gorilla/sessions"
)

var (
	StatusAuthenticated = "authenticated"
	Status2faRequired   = "2fa_required"

	CookieUserSessionName = "sid"
	CookieKeyUserID       = "uid"
	CookieKeyAuthStatus   = "auth_stat"
	CookieKeyTimestamp    = "ts"
)

type UserSessionStore struct {
	S sessions.Store
	O []AuthCheckOptions
}

type AuthCheckOptions func(*UserSession) error

type UserSession struct {
	*sessions.Session
}

func NewUserSessionStore(s sessions.Store, opts ...AuthCheckOptions) *UserSessionStore {
	return &UserSessionStore{
		S: s,
		O: opts,
	}
}

func (us *UserSessionStore) Authenticate(id common.ID, w http.ResponseWriter, r *http.Request) error {
	session, err := us.Get(r)
	if err != nil {
		return fmt.Errorf("[Authenticate] get session: %w", err)
	}

	return session.Authenticate(id, w, r)
}

func (us *UserSessionStore) Reset(w http.ResponseWriter, r *http.Request) error {
	session, err := us.Get(r)
	if err != nil {
		return fmt.Errorf("[Reset] get session: %w", err)
	}

	session.Options.MaxAge = -1 // Delete login cookie
	session.DelUserID()
	session.DelAuthStatus()
	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("[Reset] save session: %w", err)
	}
	return nil
}

func (us *UserSessionStore) AuthenticationCheck(r *http.Request) (common.ID, error) {
	session, err := us.Get(r)
	if err != nil {
		return "", fmt.Errorf("[AuthenticationCheck] get session: %w", err)
	}

	uid, err := session.GetUserID()
	if err != nil {
		return "", fmt.Errorf("[AuthenticationCheck]: %w", err)
	}

	for _, opt := range us.O {
		err := opt(session)
		if err != nil {
			return "", fmt.Errorf("[AuthenticationCheck] option check fail: %w", err)
		}
	}

	return common.ID(uid), nil
}

var ErrSessionExpired = errors.Join(errors.New("session expired"), common.ErrUnauthorized)

func WithSessionLifeTime(d time.Duration) AuthCheckOptions {
	return func(s *UserSession) error {
		tp, err := s.GetTimeStamp()
		if err != nil {
			return err
		}

		if time.Since(time.Unix(tp, 0)) > d {
			return ErrSessionExpired
		}

		return nil
	}
}

func WithAuthStatusCheck() AuthCheckOptions {
	return func(s *UserSession) error {
		status, err := s.GetAuthStatus()
		if err != nil {
			return err
		}

		if status != StatusAuthenticated {
			return common.ErrUnauthorized
		}

		return nil
	}
}

func (us *UserSessionStore) Get(r *http.Request) (*UserSession, error) {
	sp, e := us.S.Get(r, CookieUserSessionName)
	return &UserSession{sp}, e
}

func (s *UserSession) SetUserID(id common.ID) {
	s.Values[CookieKeyUserID] = string(id)
	s.Values[CookieKeyTimestamp] = time.Now().Unix()
}

func (s *UserSession) DelUserID() {
	delete(s.Values, CookieKeyUserID)
	delete(s.Values, CookieKeyTimestamp)
}

func GetKey[Type any](s *sessions.Session, key string) (Type, error) {
	v, ok := s.Values[key].(Type)
	if !ok {
		return v, fmt.Errorf("[Session] key %s %w", key, common.ErrNotFound)
	}
	return v, nil
}

func (s *UserSession) GetUserID() (common.ID, error) {
	return GetKey[string](s.Session, CookieKeyUserID)
}

func (s *UserSession) GetTimeStamp() (int64, error) {
	return GetKey[int64](s.Session, CookieKeyTimestamp)
}

func (s *UserSession) GetAuthStatus() (string, error) {
	return GetKey[string](s.Session, CookieKeyAuthStatus)
}

func (s *UserSession) SetAuthStatus(status string) {
	s.Values[CookieKeyAuthStatus] = status
}

func (s *UserSession) DelAuthStatus() {
	delete(s.Values, CookieKeyAuthStatus)
}

func (s *UserSession) Authenticate(id common.ID, w http.ResponseWriter, r *http.Request) error {
	s.SetUserID(id)
	s.SetAuthStatus(StatusAuthenticated)
	err := s.Save(r, w)
	if err != nil {
		return fmt.Errorf("[Authenticate] save session: %w", err)
	}

	return nil
}
