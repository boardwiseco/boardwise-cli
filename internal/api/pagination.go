package api

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

// MaxPerPage is the largest page the API serves.
const MaxPerPage = 100

// EachPage GETs path and every following page, calling fn with each page's
// body. The first request asks for MaxPerPage records unless path already
// sets per_page; later pages follow the RFC 8288 Link rel="next" URL until
// there is none.
//
// A next URL is followed only when its scheme and host are the API's own:
// the bearer token is never sent to another host.
func (c *Client) EachPage(path string, fn func(page []byte) error) error {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid API URL %q: %w", c.BaseURL, err)
	}

	current, err := url.Parse(c.BaseURL + withPerPage(path))
	if err != nil {
		return err
	}
	seen := map[string]bool{}

	for {
		seen[current.String()] = true

		resp, err := c.doURL("GET", current.String(), nil)
		if err != nil {
			return err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if err := fn(body); err != nil {
			return err
		}

		link := nextLink(resp.Header.Values("Link"))
		if link == "" {
			return nil
		}
		ref, err := url.Parse(link)
		if err != nil {
			return fmt.Errorf("invalid next-page link %q: %w", link, err)
		}
		next := current.ResolveReference(ref)
		if !sameOrigin(base, next) {
			return fmt.Errorf("refusing to follow a next-page link to %s://%s: it is not the API host %s://%s",
				next.Scheme, next.Host, base.Scheme, base.Host)
		}
		if seen[next.String()] {
			return fmt.Errorf("next-page link %s repeats a page already read", next)
		}
		current = next
	}
}

// withPerPage adds per_page=MaxPerPage to path's query unless it is set.
func withPerPage(path string) string {
	p, rawQuery, _ := strings.Cut(path, "?")
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return path
	}
	if query.Has("per_page") {
		return path
	}
	query.Set("per_page", fmt.Sprint(MaxPerPage))
	return p + "?" + query.Encode()
}

// nextLink returns the target of the rel="next" link in RFC 8288 Link
// header values, or "".
func nextLink(values []string) string {
	for _, header := range values {
		rest := header
		for {
			start := strings.Index(rest, "<")
			if start < 0 {
				break
			}
			end := strings.Index(rest[start:], ">")
			if end < 0 {
				break
			}
			target := rest[start+1 : start+end]
			rest = rest[start+end+1:]

			params := rest
			if i := strings.Index(rest, "<"); i >= 0 {
				params = rest[:i]
			}
			for _, param := range strings.Split(params, ";") {
				name, value, ok := strings.Cut(strings.TrimSpace(param), "=")
				if !ok || !strings.EqualFold(strings.TrimSpace(name), "rel") {
					continue
				}
				value = strings.Trim(strings.TrimSpace(strings.TrimRight(strings.TrimSpace(value), ",")), `"`)
				for _, rel := range strings.Fields(value) {
					if strings.EqualFold(rel, "next") {
						return target
					}
				}
			}
		}
	}
	return ""
}

// sameOrigin reports whether u has the same scheme, host and port as base.
func sameOrigin(base, u *url.URL) bool {
	return strings.EqualFold(base.Scheme, u.Scheme) &&
		strings.EqualFold(base.Hostname(), u.Hostname()) &&
		effectivePort(base) == effectivePort(u)
}

func effectivePort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return "443"
	case "http":
		return "80"
	}
	return ""
}
