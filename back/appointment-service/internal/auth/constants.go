package auth

type AuthorizedStatus string

var (
	StatusAuthorized   = AuthorizedStatus("authorized")
	StatusUnauthorized = AuthorizedStatus("unauthorized")
	Status2faRequired  = AuthorizedStatus("2fa_required")

	CookieUserSessionName = "sid"
	CookieKeyUserID       = "uid"
	CookieKeyAuthStatus   = "auth_stat"
	CookieKeyTimestamp    = "ts"
)
