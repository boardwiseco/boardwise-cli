package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://app.boardwise.co"

// Rate-limit handling: a 429 is retried once, after Retry-After seconds.
const (
	defaultRetryAfter = 5 * time.Second
	maxRetryAfter     = 60 * time.Second
)

type Client struct {
	BaseURL    string
	Token      string
	httpClient *http.Client
	sleep      func(time.Duration)
}

func NewClient(baseURL, token string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		httpClient: &http.Client{},
		sleep:      time.Sleep,
	}
}

type APIError struct {
	StatusCode int
	Body       string
}

// errorBody is the shape of every Boardwise error response:
// {"error": code, "message": text}, plus "errors" on a 422.
type errorBody struct {
	Error   string          `json:"error"`
	Message string          `json:"message"`
	Errors  json.RawMessage `json:"errors"`
}

// Error returns the server's human-readable message with its code, and the
// failed fields of a validation error. A body that isn't a Boardwise error
// falls back to the status and raw body.
func (e *APIError) Error() string {
	var body errorBody
	if json.Unmarshal([]byte(e.Body), &body) != nil || body.Message == "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
	}

	msg := body.Message
	if body.Error != "" {
		msg += " (" + body.Error + ")"
	}

	var fields map[string][]string
	if len(body.Errors) > 0 && json.Unmarshal(body.Errors, &fields) == nil && len(fields) > 0 {
		names := make([]string, 0, len(fields))
		for name := range fields {
			names = append(names, name)
		}
		sort.Strings(names)
		parts := make([]string, 0, len(names))
		for _, name := range names {
			parts = append(parts, name+" "+strings.Join(fields[name], ", "))
		}
		msg += ": " + strings.Join(parts, "; ")
	}
	return msg
}

// Code returns the machine-readable "error" field of a Boardwise error body
// ({"error": "...", "message": "..."}), or "" when the body has none.
func (e *APIError) Code() string {
	var body errorBody
	if json.Unmarshal([]byte(e.Body), &body) != nil {
		return ""
	}
	return body.Error
}

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	return c.doURL(method, c.BaseURL+path, body)
}

// doURL sends one request to an absolute URL. A 429 is retried once after
// the server's Retry-After; the body is re-sent from the same bytes.
func (c *Client) doURL(method, rawURL string, body any) (*http.Response, error) {
	var data []byte
	if body != nil {
		var err error
		if data, err = json.Marshal(body); err != nil {
			return nil, err
		}
	}

	for attempt := 1; ; attempt++ {
		var bodyReader io.Reader
		if data != nil {
			bodyReader = bytes.NewReader(data)
		}

		req, err := http.NewRequest(method, rawURL, bodyReader)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/json")
		if data != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.Token != "" {
			req.Header.Set("Authorization", "Bearer "+c.Token)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt == 1 {
			wait := retryAfter(resp.Header.Get("Retry-After"))
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			c.sleep(wait)
			continue
		}

		if resp.StatusCode >= 400 {
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			return nil, &APIError{StatusCode: resp.StatusCode, Body: string(b)}
		}

		return resp, nil
	}
}

// retryAfter reads a Retry-After header given in seconds, capped at
// maxRetryAfter, defaulting when it is missing or unreadable.
func retryAfter(header string) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(header))
	if err != nil || seconds < 0 {
		return defaultRetryAfter
	}
	wait := time.Duration(seconds) * time.Second
	if wait > maxRetryAfter {
		return maxRetryAfter
	}
	return wait
}

func (c *Client) Get(path string, out any) error {
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) GetRaw(path string) ([]byte, error) {
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *Client) Post(path string, body, out any) error {
	resp, err := c.do("POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) PostRaw(path string, body any) ([]byte, error) {
	resp, err := c.do("POST", path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *Client) Delete(path string) error {
	resp, err := c.do("DELETE", path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// BuildPath constructs an org-scoped path.
func BuildPath(org, rest string) string {
	return "/" + url.PathEscape(org) + rest
}
