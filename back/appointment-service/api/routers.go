package server

import (
	"context"
	"fmt"
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
	HandlerFunc http.HandlerFunc
}

type Routes []Route

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

func NewRouter(storage *storage.Storage, sesStore *auth.UserSessionStore) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	restrictionTable := common.NewLimitsTable[string](
		//TODO need more complex solution
		common.RequestLimitUpdateFunc(func(in *rate.Limiter) *rate.Limiter {
			return rate.NewLimiter(rate.Every(time.Second*15), 1)
		}))

	userCheck := userCheckWrap{storage, restrictionTable}

	oidcUserSignIn, err := NewUserSignIn(storage, sesStore)
	if err != nil {
		slog.Warn("[NewRouter]", "err", err.Error())
		panic(err)
	}

	//TODO add/remove business rules
	var routes = Routes{
		Route{
			"Index",
			"GET",
			"/",
			Index,
		},
		Route{
			"SlotsBusinessIdGet",
			"GET",
			"/slots/{business_id}",
			SlotsBusinessIdGetFunc(storage),
		},
		Route{
			"SlotsBusinessIdPost",
			"POST",
			"/slots/{business_id}",
			SlotsBusinessIdPostFunc(storage),
		},
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
			CheckAuthHandler(sesStore, userCheck, LogoutHandler(sesStore)),
		},
		Route{
			"DeleteUser",
			"DELETE",
			"/user",
			CheckAuthHandler(sesStore, userCheck, DeleteUserHandler(sesStore, storage.DeleteUserWithCheck)),
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

	return router
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}
