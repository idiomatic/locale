package locale

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/language"
)

var (
	QueryLanguage = "lang"
	QueryRegion   = "region"
	QueryLocale   = "locale"

	CookieName = "locale"

	GeoipCountryHeaders = []string{
		"CloudFront-Viewer-Country",
		"CF-IPCountry",
		"GEOIP_COUNTRY_CODE",
	}
)

// Get hosting provider inferred country code according to client IP address.
func GeoipCountryCode(r *http.Request) string {
	for _, header := range GeoipCountryHeaders {
		code := r.Header.Get(header)
		if code != "" {
			return code
		}
	}
	return "unk"
}

// Retrieve locale from request.
// See SetRequestLocale().
func GetRequestLocale(r *http.Request) string {
	return r.Header.Get("X-Locale")
}

// Save locale with request.  Overwrites X-Locale header.
// See GetRequestLocale().
func SetRequestLocale(r *http.Request, locale string) {
	r.Header.Set("X-Locale", locale)
}

// Retrieve URL path without locale from request.
// See SetRequestDelocalizedPath().
func GetRequestDelocalizedPath(r *http.Request) string {
	return r.Header.Get("X-Delocalized-Path")
}

// Save URL path without locale with request.  Overwrites X-Delocalized-Path header.
// See GetRequestDelocalizedPath().
func SetRequestDelocalizedPath(r *http.Request, path string) {
	r.Header.Set("X-Delocalized-Path", path)
}

// Remember user-specified locale override for a month.
func SetCookieLocale(w http.ResponseWriter, locale string) {
	c := http.Cookie{
		Name:    "locale",
		Value:   locale,
		Expires: time.Now().AddDate(0, 1, 0),
		Path:    "/",
	}
	http.SetCookie(w, &c)
}

// Infer the language based on query params, cookies, request headers, or geoIP.
// Intentionally ignores X-Locale.
func DeduceLanguage(matcher language.Matcher, r *http.Request) language.Tag {
	geoipCountry := RemapCountryCode(GeoipCountryCode(r))

	var candidates []string

	// imprecise query parameter hints (unlike the precise locale)
	{
		var (
			queryLanguage = ""
			queryRegion   = ""
			query         = r.URL.Query()
		)

		// explicit locale query parameter
		if l, ok := query[QueryLocale]; ok {
			candidates = append(candidates, l[0])
		}

		if l, ok := query[QueryLanguage]; ok {
			queryLanguage = l[0]
		}

		if r, ok := query[QueryRegion]; ok {
			queryRegion = RemapCountryCode(r[0])

			// expclit language and country query parameters (equivalent to explicit locale)
			if queryLanguage != "" {
				candidates = append(candidates, fmt.Sprintf("%s-%s", queryLanguage, queryRegion))
			}
		}

		if queryLanguage != "" {
			// explicit language query parameter and implicit IP address
			candidates = append(candidates, fmt.Sprintf("%s-%s", queryLanguage, geoipCountry))
		}

		// explicit country query parameter; country often implies a language
		if queryRegion != "" {
			candidates = append(candidates, fmt.Sprintf("und-%s", queryRegion))
		}
	}

	// user-configured browser/OS setting of language-regionalDialect priorities
	acceptLanguageHeader := r.Header.Get("Accept-Language")
	if acceptLanguageHeader != "" {
		candidates = append(candidates, acceptLanguageHeader)
	}

	// IP address implies a country, which may further imply a language
	candidates = append(candidates, fmt.Sprintf("und-%s", geoipCountry))

	tag, _ := language.MatchStrings(matcher, candidates...)

	return tag
}

// Find a prefix among many.
func matchPrefix(s string, prefixes []string) (string, bool) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return prefix, true
		}
	}
	return "", false
}

// Remove any of these prefixes from a string.
func trimPrefixes(s string, prefixes []string) string {
	if p, found := matchPrefix(s, prefixes); found {
		return strings.TrimPrefix(s, p)
	}
	return s
}

func contains(locales []string, locale string) bool {
	for _, l := range locales {
		if l == locale {
			return true
		}
	}
	return false
}

// Remove the locale part from the URL Path.  See SetRequestDelocalizedPath().
func StripLocalePrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := GetRequestDelocalizedPath(r); p != "" {
			r.URL.Path = p
		}
		next.ServeHTTP(w, r)
	})
}

// Chaining HTTP handler that alters request per query parameters, cookies, locale-prefixed URL paths, or headers.
// Request path is localized and revertable through StripLocalePrefix().
// Request locale is accessed through GetRequestLocale().
func Handler(locales []string, matcher language.Matcher, next http.Handler) http.Handler {
	var pathPrefixes []string

	for _, locale := range locales {
		pathPrefixes = append(pathPrefixes, "/"+locale)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			locale          string
			pathLocale      string
			delocalizedPath = r.URL.Path
			newPath         = func() string { return fmt.Sprintf("/%s%s", locale, delocalizedPath) }
		)

		if prefix, found := matchPrefix(r.URL.Path, pathPrefixes); found {
			delocalizedPath = strings.TrimPrefix(r.URL.Path, prefix)
			pathLocale = strings.TrimPrefix(prefix, "/")
		}

		if q, ok := r.URL.Query()["locale"]; ok && len(q) > 0 {
			localeQuery := q[0]

			// HACK transform Java to IETF
			localeQuery = strings.Replace(localeQuery, "_", "-", -1)

			if contains(locales, localeQuery) {
				locale = localeQuery
				SetCookieLocale(w, locale)
				if locale != pathLocale {
					http.Redirect(w, r, newPath(), http.StatusFound)
					return
				}
			}
		}

		if locale == "" {
			locale = pathLocale
			// no need to redirect since the path is the locale source.
		}

		if locale == "" {
			if cookie, err := r.Cookie(CookieName); err == nil {
				locale = cookie.Value
				// locale was explicitly changed at some point in the past month.
				// assume implict "short" paths are cached with "wrong" locale content, so go explicit.
				if locale != pathLocale {
					http.Redirect(w, r, newPath(), http.StatusFound)
					return
				}
			}
		}

		if locale == "" {
			locale = BaseRegion(DeduceLanguage(matcher, r))
			// preserve path to keep them tidy when correctly guessing locale.
		}

		SetRequestLocale(r, locale)
		SetRequestDelocalizedPath(r, delocalizedPath)
		r.URL.Path = newPath()

		next.ServeHTTP(w, r)
	})
}
