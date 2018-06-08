package cookie

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// CookieExpireDelete may be set on Cookie.Expire for expiring the given cookie.
	CookieExpireDelete = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	// CookieExpireUnlimited indicates that the cookie doesn't expire.
	CookieExpireUnlimited = time.Now().AddDate(50, 0, 0)
)

// Get returns cookie's value by it's name
// returns empty string if nothing was found
func Get(name string, req *http.Request) string {
	if c, err := req.Cookie(name); err == nil {
		return c.Value
	}
	return ""
}

// Remove deletes a cookie by it's name/key
func Remove(name string, res http.ResponseWriter, req *http.Request) {
	c, err := req.Cookie(name)
	if err != nil {
		return
	}

	c.Expires = CookieExpireDelete
	c.MaxAge = -1
	c.Value = ""
	c.Path = "/"
	Add(c, res)
}

// Add adds a cookie
func Add(cookie *http.Cookie, res http.ResponseWriter) {
	if v := cookie.String(); v != "" {
		http.SetCookie(res, cookie)
	}
	http.SetCookie(res, cookie)
}

var cookiePool sync.Pool

// AcquireCookie returns an empty Cookie object from the pool.
//
// The returned object may be returned back to the pool with ReleaseCookie.
// This allows reducing GC load.
func AcquireCookie() *http.Cookie {
	v := cookiePool.Get()
	if v == nil {
		return &http.Cookie{}
	}

	cookie := v.(*http.Cookie)
	cookie.HttpOnly = true
	cookie.Path = ""
	cookie.Name = ""
	cookie.Raw = ""
	cookie.Value = ""
	cookie.Domain = ""
	cookie.MaxAge = -1
	cookie.Expires = CookieExpireUnlimited
	return cookie
}

// ReleaseCookie returns the Cookie object acquired with AcquireCookie back
// to the pool.
//
// Do not access released Cookie object, otherwise data races may occur.
func ReleaseCookie(cookie *http.Cookie) {
	cookiePool.Put(cookie)
}

// IsValidCookieDomain returns true if the receiver is a valid domain to set
// valid means that is recognised as 'domain' by the browser, so it(the cookie) can be shared with subdomains also
func IsValidCookieDomain(domain string) bool {
	if domain == "0.0.0.0" || domain == "127.0.0.1" {
		// for these type of hosts, we can't allow subdomains persistance,
		// the web browser doesn't understand the mysubdomain.0.0.0.0 and mysubdomain.127.0.0.1 mysubdomain.32.196.56.181. as scorrectly ubdomains because of the many dots
		// so don't set a cookie domain here, let browser handle this
		return false
	}

	dotLen := strings.Count(domain, ".")
	if dotLen == 0 {
		// we don't have a domain, maybe something like 'localhost', browser doesn't see the .localhost as wildcard subdomain+domain
		return false
	}
	if dotLen >= 3 {
		if lastDotIdx := strings.LastIndexByte(domain, '.'); lastDotIdx != -1 {
			// chekc the last part, if it's number then propably it's ip
			if len(domain) > lastDotIdx+1 {
				_, err := strconv.Atoi(domain[lastDotIdx+1:])
				if err == nil {
					return false
				}
			}
		}
	}

	return true
}
