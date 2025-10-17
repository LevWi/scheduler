package api

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/auth/oidc"
	dbauth "scheduler/appointment-service/internal/dbase/auth"
	"scheduler/appointment-service/internal/dbase/backend/slots"
	"scheduler/appointment-service/internal/dbase/bots"
	"time"

	"scheduler/appointment-service/internal/auth"

	"github.com/jmoiron/sqlx"
	"golang.org/x/time/rate"
)

type api struct {
	storages struct {
		Auth         *dbauth.AuthStorage
		OneOffTokens *dbauth.OneOffTokenStorage
		TimeSlots    *slots.TimeSlotsStorage
		Bots         *bots.BotsStorage
	}

	cookieAuth        *CookieAuth
	userSignIn        *oidc.UserSignIn
	userSessionsStore *auth.UserSessionStore
}

func NewAPI(
	oauthCfgPath string,
	userSessionsStore *auth.UserSessionStore,
	db *sqlx.DB,
) (*api, error) {
	var a api

	a.userSessionsStore = userSessionsStore

	//For now only one Database used
	a.storages.Auth = &dbauth.AuthStorage{DB: db}
	a.storages.OneOffTokens = &dbauth.OneOffTokenStorage{DB: db}
	a.storages.TimeSlots = &slots.TimeSlotsStorage{DB: db}
	a.storages.Bots = &bots.BotsStorage{DB: db}

	oidcUserSignIn, err := newUserSignIn(a.storages.Auth, a.userSessionsStore, oauthCfgPath)
	if err != nil {
		return nil, err
	}
	a.userSignIn = oidcUserSignIn

	restrictionTable := common.NewLimitsTable[string](
		//TODO need more complex solution
		common.RequestLimitUpdateFunc(func(in *rate.Limiter) *rate.Limiter {
			return rate.NewLimiter(rate.Every(time.Second*15), 1)
		}))
	userCheck := userCheckWrap{a.storages.Auth, restrictionTable}
	a.cookieAuth = &CookieAuth{a.userSessionsStore, userCheck}

	return &a, nil
}
