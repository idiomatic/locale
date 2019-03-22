Middleware for language/region locale handling.

Some websites need to show different translated or geographical
market-specific content.  This middleware makes use of a variety of
explicit and implicit factors to provide the most appropriate content.

## Query Parameters

A `?locale=` query parameter may be included in a page request that
indicate an explicit adjustment.  The `?locale=` query parameter can
be the resulting page URLs fetched in service to a locale menu.

Alternatively, the `?region=` or `?language=` query parameters may
provide vauge hints.

## Cookies

For the convenience of the end-user, explicit language selection
should be remembered across sessions.  The locale cookie values are
not PII.

## Initial URL Path Element

In order to "cache bust" and have an immediate effect upon changing a
locale, post-locale-selection page paths start with the new locale.

## Browser Language

Browsers and operating systems may provide language preference(s) via
the Accept-Language header.  The language (and subtag) can imply a
locale.

## GeoIP

The client IP address suggests a country, which further implies a locale.

## Some Combination

Combining vague query parameters (e.g., `?language=` or `?region=`),
browser languages (especially en or es), and GeoIP countries
(e.g. Canada) can have a combined effect in picking the most
appropriate language.
