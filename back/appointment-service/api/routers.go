package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
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

type Routes []Route

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

func NewUserSignIn(storage *storage.Storage, sesStore *auth.UserSessionStore) (*oidc.UserSignIn, error) {
	//TODO move from here
	oidcCfgProvider, err := oidc.NewOAuth2CfgProviderFromFile("./oauth_cfg.json")
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
		OAuth2CfgProvider: &oidcCfgProvider,
		OIDCAuthCheck:     authCheck,
		SaveUserCookie:    sesStore,
	}, nil
}

// TODO need more logs
func NewRouter(storage *storage.Storage, sesStore *auth.UserSessionStore) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	restrictionTable := common.NewLimitsTable[string](
		//TODO need more complex solution
		common.RequestLimitUpdateFunc(func(in *rate.Limiter) *rate.Limiter {
			return rate.NewLimiter(rate.Every(time.Second*15), 1)
		}))

	userCheck := userCheckWrap{storage, restrictionTable}
	ruleStorage := rruleStorage{storage}

	oidcUserSignIn, err := NewUserSignIn(storage, sesStore)
	if err != nil {
		slog.Warn("[NewRouter]", "err", err.Error())
		panic(err)
	}

	cookieAuth := &CookieAuth{sesStore, userCheck}

	var routes = Routes{
		Route{
			"SlotsBusinessIdGet",
			"GET",
			"/slots/{business_id}",
			SlotsBusinessIdGetFunc(storage),
		},
		Route{
			"SlotsBusinessIdPostFromBot",
			"POST",
			"/slots/bt",
			AuthHandler(BotAuthMethod(storage), SlotsBusinessIdPostFunc(storage, storage), nil),
		},
		Route{
			"SlotsBusinessIdPost",
			"POST",
			"/slots",
			AuthHandler(cookieAuth, SlotsBusinessIdPostFunc(storage, storage), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"AddBusinessRulePost",
			"POST",
			"/rrules",
			AuthHandler(cookieAuth, AddBusinessRuleHandler(&ruleStorage), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"GetBusinessRule",
			"GET",
			"/rrules",
			AuthHandler(cookieAuth, GetBusinessRulesHandler(&ruleStorage), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"DelBusinessRule",
			"DELETE",
			"/rrules/{id}",
			AuthHandler(cookieAuth, DelBusinessRuleHandler(&ruleStorage), http.HandlerFunc(LoginRequired)),
		},
		//Disabled for security reasons
		// Route{
		// 	"Login",
		// 	"POST",
		// 	"/login",
		// 	//TODO add IP address check
		// 	PasswordLoginHandler(sesStore, userCheck),
		// },
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
		},
		Route{
			"Logout",
			"POST",
			"/logout",
			AuthHandler(cookieAuth, LogoutHandler(sesStore), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"DeleteUser",
			"DELETE",
			"/user",
			AuthHandler(cookieAuth, DeleteUserHandler(sesStore, storage.DeleteUserWithCheck), http.HandlerFunc(LoginRequired)),
		},

		Route{
			"UserBotAdd",
			"POST",
			"/user/bots",
			AuthHandler(cookieAuth, AddUserBotHandler(storage), http.HandlerFunc(LoginRequired)),
		},
		Route{
			"UserBotDel",
			"DELETE",
			"/user/bots/{bot_id}",
			AuthHandler(cookieAuth, DeleteUserBotHandler(storage), http.HandlerFunc(LoginRequired)),
		},
	}

	for _, route := range routes {
		var handler http.Handler
		handler = Logger(route.HandlerFunc, route.Name)
		handler = PassRequestIdToCtx(handler)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	//TODO Fix paths
	router.Methods("GET").PathPrefix("/front/").Name("FileServer").
		Handler(http.StripPrefix("/front/", http.FileServer(http.Dir(os.Getenv("FRONT_PATH")))))

	return router
}

func LoginRequired(w http.ResponseWriter, r *http.Request) {
	slog.WarnContext(r.Context(), "[LoginRequired]", "RemoteAddr", r.RemoteAddr)
	http.Error(w, "Please login first", http.StatusNetworkAuthenticationRequired)
}
