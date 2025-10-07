package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/auth"
	"scheduler/appointment-service/internal/auth/oidc"
	"scheduler/appointment-service/internal/storage"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.Handler
}

// TODO need to cache IsExist(uid) result with periodic update. With mutex
type userCheckWrap struct {
	*storage.Storage
	table *common.LimitsTable[string]
}

func (uc userCheckWrap) Check(username string, password string) (UserID, error) {
	if !uc.table.Allow(username) {
		return "", ErrSecurityRestriction
	}
	return uc.CheckUserPassword(username, password)
}

func BotAuthMethod(storage *storage.Storage) AuthorizationMethodFunc {
	cache := auth.NewTokenCacheDefault((*auth.TokenStorage)(storage))
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

func newUserSignIn(storage *storage.Storage, sesStore *auth.UserSessionStore, configPath string) (*oidc.UserSignIn, error) {
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

type RouterBuilder struct {
	_          common.NoCopy
	storage    *storage.Storage
	sesStore   *auth.UserSessionStore
	cookieAuth *CookieAuth

	v *mux.Router
}

func NewRouterBuilder(storage *storage.Storage, sesStore *auth.UserSessionStore) *RouterBuilder {
	var rb RouterBuilder
	rb.v = mux.NewRouter().StrictSlash(true)

	restrictionTable := common.NewLimitsTable[string](
		//TODO need more complex solution
		common.RequestLimitUpdateFunc(func(in *rate.Limiter) *rate.Limiter {
			return rate.NewLimiter(rate.Every(time.Second*15), 1)
		}))
	userCheck := userCheckWrap{storage, restrictionTable}
	rb.cookieAuth = &CookieAuth{sesStore, userCheck}

	rb.storage = storage
	rb.sesStore = sesStore
	return &rb
}

// TODO need more logs
func (rb *RouterBuilder) AddUserAccountHandlers() *RouterBuilder {
	rb.addRoutes(
		Route{
			"Logout",
			"POST",
			"/logout",
			AuthHandler(rb.cookieAuth, LogoutHandler(rb.sesStore), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"DeleteUser",
			"DELETE",
			"/user",
			AuthHandler(rb.cookieAuth, DeleteUserHandler(rb.sesStore, rb.storage.DeleteUserWithCheck), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"UserBotAdd",
			"POST",
			"/user/bots",
			AuthHandler(rb.cookieAuth, AddUserBotHandler(rb.storage), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"UserBotDel",
			"DELETE",
			"/user/bots/{bot_id}",
			AuthHandler(rb.cookieAuth, DeleteUserBotHandler(rb.storage), http.HandlerFunc(LoginRequired)),
		})

	return rb
}

func (rb *RouterBuilder) addRoutes(routes ...Route) {
	for _, route := range routes {
		var handler http.Handler
		handler = Logger(route.HandlerFunc, route.Name)
		handler = PassRequestIdToCtx(handler)

		rb.v.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
}

func (rb *RouterBuilder) AddOIDCHandlers(oauthCfgPath string) *RouterBuilder {
	oidcUserSignIn, err := newUserSignIn(rb.storage, rb.sesStore, oauthCfgPath)
	if err != nil {
		slog.Warn("[NewRouter]", "err", err.Error())
		panic(err)
	}
	rb.addRoutes(
		Route{
			"OAuth2Redirect",
			"GET",
			"/oauth_login",
			oidc.OAuth2HTTPRedirectHandler(oidcUserSignIn),
		},
		Route{
			"OAuth2UserBack",
			"GET",
			"/callback", //"TODO fix it /oauth_callback"
			oidc.OAuth2HTTPUserBackHandler(oidcUserSignIn, nil),
		})
	return rb
}

func (rb *RouterBuilder) AddTimeSlotsHandlers() *RouterBuilder {
	rb.addRoutes(
		Route{
			"SlotsBusinessIdGet",
			"GET",
			"/slots/{business_id}",
			SlotsBusinessIdGetFunc(rb.storage),
		},
		Route{
			"SlotsBusinessIdPostFromBot",
			"POST",
			"/slots/bt",
			AuthHandler(BotAuthMethod(rb.storage), SlotsBusinessIdPostFunc(rb.storage, rb.storage), nil),
		},
		Route{
			"SlotsBusinessIdPost",
			"POST",
			"/slots",
			AuthHandler(rb.cookieAuth, SlotsBusinessIdPostFunc(rb.storage, rb.storage), http.HandlerFunc(LoginRequired)),
		})
	return rb
}

func (rb *RouterBuilder) AddBusinessRulesHandlers() *RouterBuilder {
	ruleStorage := rruleStorage{rb.storage}
	rb.addRoutes(
		Route{
			"AddBusinessRulePost",
			"POST",
			"/rrules",
			AuthHandler(rb.cookieAuth, AddBusinessRuleHandler(&ruleStorage), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"GetBusinessRule",
			"GET",
			"/rrules",
			AuthHandler(rb.cookieAuth, GetBusinessRulesHandler(&ruleStorage), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"DelBusinessRule",
			"DELETE",
			"/rrules/{id}",
			AuthHandler(rb.cookieAuth, DelBusinessRuleHandler(&ruleStorage), http.HandlerFunc(LoginRequired)),
		})
	return rb
}

func (rb *RouterBuilder) AppendFileServerLogic(dir string) *RouterBuilder {
	rb.v.Methods("GET").PathPrefix("/front/").Name("FileServer").
		Handler(http.StripPrefix("/front/", http.FileServer(http.Dir(dir))))
	return rb
}

func (rb *RouterBuilder) Done() *mux.Router {
	return rb.v
}

func LoginRequired(w http.ResponseWriter, r *http.Request) {
	slog.WarnContext(r.Context(), "[LoginRequired]", "RemoteAddr", r.RemoteAddr)
	http.Error(w, "Please login first", http.StatusNetworkAuthenticationRequired)
}
