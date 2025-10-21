package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/auth"
	"scheduler/appointment-service/internal/auth/oidc"
	authdb "scheduler/appointment-service/internal/dbase/auth"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.Handler
}

// TODO need to cache IsExist(uid) result with periodic update. With mutex
type userCheckWrap struct {
	*authdb.AuthStorage
	table *common.LimitsTable[string]
}

func (uc userCheckWrap) Check(username string, password string) (UserID, error) {
	if !uc.table.Allow(username) {
		return "", ErrSecurityRestriction
	}
	return uc.CheckUserPassword(username, password)
}

// TODO Move logic to internal
func botAuthMethod(storage *auth.BotTokenStorage) AuthorizationMethodFunc {
	cache := auth.NewTokenCacheDefault(storage)
	type bearerAuthWrap struct {
		auth.BearerAuth
		shed *common.PeriodicCallback
	}
	wrp := bearerAuthWrap{
		BearerAuth: auth.BearerAuth{TC: cache},
		shed: common.NewPeriodicCallback(time.Minute*5, func() {
			cleared := cache.ForgetExpired()
			slog.Debug("[BotAuthMethod] PeriodicCallback", "cleared", cleared)
		})}
	wrp.shed.Start()
	return AuthorizationMethodFunc(func(_ http.ResponseWriter, r *http.Request) (common.ID, error) {
		return wrp.Authorization(r)
	})
}

func newUserSignIn(storage *authdb.AuthStorage, sesStore *auth.UserSessionStore, configPath string) (*oidc.UserSignIn, error) {
	//TODO move from here
	oidcCfgProvider, err := oidc.NewOAuth2CfgProviderFromFile(configPath)
	if err != nil {
		return nil, err
	}

	authCheck, err := oidc.NewOIDCAuthCheckDefault(context.Background(), storage)
	if err != nil {
		return nil, err
	}

	return &oidc.UserSignIn{
		OAuth2ValidateState: &oidc.OAuth2SessionsValidator{
			Store: sesStore.S,
		},
		OAuth2CfgProvider: oidcCfgProvider,
		OIDCAuthCheck:     authCheck,
		SaveUserCookie:    sesStore,
	}, nil
}

// TODO need more logs
func (a *api) addUserAccountHandlers(r *mux.Router) {
	addRoutes(
		r,
		Route{
			"Logout",
			"POST",
			"/logout",
			AuthHandler(a.cookieAuth, LogoutHandler(a.userSessionsStore), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"DeleteUser",
			"DELETE",
			"/user",
			AuthHandler(a.cookieAuth,
				DeleteUserHandler(a.userSessionsStore, a.storages.Auth.DeleteUserWithCheck),
				http.HandlerFunc(LoginRequired)),
		},
		Route{
			"UserBotAdd",
			"POST",
			"/user/bots",
			AuthHandler(a.cookieAuth, AddUserBotHandler(a.storages.Bots), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"UserBotDel",
			"DELETE",
			"/user/bots/{bot_id}",
			AuthHandler(a.cookieAuth, DeleteUserBotHandler(a.storages.Bots), http.HandlerFunc(LoginRequired)),
		})
}

func (a *api) Router() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	a.addTimeSlotsHandlers(r)
	a.addBusinessRulesHandlers(r)
	a.addUserAccountHandlers(r)
	a.addOIDCHandlers(r)

	r.Use(PassRequestIdToCtx)
	return r
}

func (a *api) addOIDCHandlers(r *mux.Router) {
	addRoutes(r,
		Route{
			"OAuth2Redirect",
			"GET",
			"/oauth_login",
			oidc.OAuth2HTTPRedirectHandler(a.userSignIn),
		},
		Route{
			"OAuth2UserBack",
			"GET",
			"/callback", //"TODO fix it /oauth_callback"
			oidc.OAuth2HTTPUserBackHandler(a.userSignIn, nil),
		})
}

func (a *api) addTimeSlotsHandlers(r *mux.Router) {
	oneOffAuth := (*AddSlotsAuthOneOffToken)(a.storages.Auth)
	bs := auth.BotTokenStorage{BotsStorage: a.storages.Bots}
	addRoutes(
		r,
		Route{
			"SlotsBusinessIdGet",
			"GET",
			"/slots/{business_id}",
			a.SlotsBusinessIdGetFunc(),
		},
		Route{
			"SlotsBusinessIdPostOneOff",
			"POST",
			"/slots/once",
			a.SlotsBusinessIdPostFunc(oneOffAuth),
		},
		Route{
			"SlotsBusinessIdPostFromBot",
			"POST",
			"/slots/bt",
			//SlotsBusinessIdPostFunc(&oneOffAuth, ts),
			AuthHandler(botAuthMethod(&bs), a.SlotsBusinessIdPostFunc(AddSlotsAuthFromUrl{}), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"SlotsBusinessIdPost",
			"POST",
			"/slots",
			AuthHandler(a.cookieAuth, a.SlotsBusinessIdPostFunc(AddSlotsAuthFromUrl{}), http.HandlerFunc(LoginRequired)),
		})
}

func (a *api) addBusinessRulesHandlers(r *mux.Router) {
	ruleStorage := rruleStorage{a.storages.TimeSlots}
	addRoutes(
		r,
		Route{
			"AddBusinessRulePost",
			"POST",
			"/rrules",
			AuthHandler(a.cookieAuth, AddBusinessRuleHandler(&ruleStorage), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"GetBusinessRule",
			"GET",
			"/rrules",
			AuthHandler(a.cookieAuth, GetBusinessRulesHandler(&ruleStorage), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"DelBusinessRule",
			"DELETE",
			"/rrules/{id}",
			AuthHandler(a.cookieAuth, DelBusinessRuleHandler(&ruleStorage), http.HandlerFunc(LoginRequired)),
		})
}

// TODO
// Deprecated: Move from service logic
func (a *api) AppendFileServerLogic(dir string, r *mux.Router) {
	r.Methods("GET").PathPrefix("/front/").Name("FileServer").
		Handler(http.StripPrefix("/front/", http.FileServer(http.Dir(dir))))
}

func LoginRequired(w http.ResponseWriter, r *http.Request) {
	slog.WarnContext(r.Context(), "[LoginRequired]", "RemoteAddr", r.RemoteAddr)
	http.Error(w, "Please login first", http.StatusNetworkAuthenticationRequired)
}

func addRoutes(r *mux.Router, routes ...Route) {
	for _, route := range routes {
		handler := Logger(route.HandlerFunc, route.Name)
		r.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
}
