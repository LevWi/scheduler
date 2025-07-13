package server

import (
	"fmt"
	"net/http"

	"scheduler/appointment-service/internal/storage"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

type UserCheckWrap struct {
	*storage.Storage
}

func (uc UserCheckWrap) Check(username string, password string) (UserID, error) {
	return uc.CheckUserPassword(username, password)
}

func NewRouter(sto *storage.Storage, ses sessions.Store) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

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
			LoginHandler(ses, UserCheckWrap{sto}),
		},
		Route{
			"Logout",
			"POST",
			"/logout",
			CheckAuthHandler(ses, UserCheckWrap{sto}, LogoutHandler(ses)),
		},
		Route{
			"DeleteUser",
			"DELETE",
			"/user",
			CheckAuthHandler(ses, UserCheckWrap{sto}, DeleteUserHandler(ses, sto.DeleteUser)),
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
