package server

import (
	"fmt"
	"net/http"
	"time"

	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/storage"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
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

func NewRouter(sto *storage.Storage, ses sessions.Store) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	restrictionTable := common.NewLimitsTable[string](
		//TODO need more complex solution
		common.RequestLimitUpdateFunc(func(in *rate.Limiter) *rate.Limiter {
			return rate.NewLimiter(rate.Every(time.Second*15), 1)
		}))

	userCheck := userCheckWrap{sto, restrictionTable}

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
			SlotsBusinessIdGetFunc(sto),
		},
		Route{
			"SlotsBusinessIdPost",
			"POST",
			"/slots/{business_id}",
			SlotsBusinessIdPostFunc(sto),
		},
		Route{
			"Login",
			"POST",
			"/login",
			//TODO add IP address check
			LoginHandler(ses, userCheck),
		},
		Route{
			"Logout",
			"POST",
			"/logout",
			CheckAuthHandler(ses, userCheck, LogoutHandler(ses)),
		},
		Route{
			"DeleteUser",
			"DELETE",
			"/user",
			CheckAuthHandler(ses, userCheck, DeleteUserHandler(ses, sto.DeleteUser)),
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
