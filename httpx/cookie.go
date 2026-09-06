package httpx

import (
	"net/http"
	"time"
)

type CookieOptions struct {
	Path     string
	Domain   string
	MaxAge   int
	Expires  time.Time
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

func SetCookie(
	w Res,
	name string,
	value string,
	options CookieOptions,
) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     name,
			Value:    value,
			Path:     options.Path,
			Domain:   options.Domain,
			MaxAge:   options.MaxAge,
			Expires:  options.Expires,
			Secure:   options.Secure,
			HttpOnly: options.HTTPOnly,
			SameSite: options.SameSite,
		},
	)
}

func GetCookie(
	r *Req,
	name string,
) (string, bool) {
	cookie, err := r.Cookie(name)

	if err != nil {
		return "", false
	}

	return cookie.Value, true
}

func ClearCookie(
	w Res,
	name string,
	options CookieOptions,
) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     name,
			Value:    "",
			Path:     options.Path,
			Domain:   options.Domain,
			MaxAge:   -1,
			Expires:  time.Unix(1, 0),
			Secure:   options.Secure,
			HttpOnly: options.HTTPOnly,
			SameSite: options.SameSite,
		},
	)
}

type AuthCookieConfig struct {
	Name     string
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

func (a *App) ConfigureAuthCookie(
	config AuthCookieConfig,
) {
	a.authCookie = config
}

func (a *App) SetAuthCookie(
	w Res,
	token string,
) {
	SetCookie(
		w,
		a.authCookie.Name,
		token,
		CookieOptions{
			Path:     a.authCookie.Path,
			Domain:   a.authCookie.Domain,
			MaxAge:   a.authCookie.MaxAge,
			Secure:   a.authCookie.Secure,
			HTTPOnly: a.authCookie.HTTPOnly,
			SameSite: a.authCookie.SameSite,
		},
	)
}

func (a *App) ClearAuthCookie(
	w Res,
) {
	ClearCookie(
		w,
		a.authCookie.Name,
		CookieOptions{
			Path:     a.authCookie.Path,
			Domain:   a.authCookie.Domain,
			Secure:   a.authCookie.Secure,
			HTTPOnly: a.authCookie.HTTPOnly,
			SameSite: a.authCookie.SameSite,
			MaxAge:   a.authCookie.MaxAge,
		},
	)
}
