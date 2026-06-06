// Package client is the thin authenticated HTTP layer the commands call.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client issues authenticated requests against the API host.
type Client struct {
	Host  string
	Token string
	HTTP  *http.Client
}

// New builds a client with a sane default timeout.
func New(host, token string) *Client {
	return &Client{
		Host:  strings.TrimRight(host, "/"),
		Token: token,
		HTTP:  &http.Client{Timeout: 120 * time.Second},
	}
}

// Request describes a single call. Query is already URL-encoded.
type Request struct {
	Method      string
	Path        string
	Query       string
	Headers     map[string]string
	Body        io.Reader
	ContentType string
}

// Do sends the request, attaching the bearer token.
func (c *Client) Do(r Request) (*http.Response, error) {
	url := c.Host + r.Path
	if r.Query != "" {
		url += "?" + r.Query
	}
	req, err := http.NewRequest(strings.ToUpper(r.Method), url, r.Body)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Accept", "application/json, */*")
	if r.ContentType != "" {
		req.Header.Set("Content-Type", r.ContentType)
	}
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}
	return c.HTTP.Do(req)
}

// APIError is a non-2xx response, formatted from the JSON:API error payload.
type APIError struct {
	Status int
	Code   string
	Detail string
	Body   []byte
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("HTTP %d — %s", e.Status, e.Detail)
	}
	if body := strings.TrimSpace(string(e.Body)); body != "" {
		return fmt.Sprintf("HTTP %d — %s", e.Status, body)
	}
	return fmt.Sprintf("HTTP %d", e.Status)
}

// ParseError turns a non-2xx response body into an APIError.
func ParseError(resp *http.Response, body []byte) error {
	e := &APIError{Status: resp.StatusCode, Body: body}
	var parsed struct {
		Errors []struct {
			Code   string `json:"code"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		} `json:"errors"`
	}
	if json.Unmarshal(body, &parsed) == nil && len(parsed.Errors) > 0 {
		e.Code = parsed.Errors[0].Code
		var parts []string
		for _, er := range parsed.Errors {
			msg := er.Detail
			if msg == "" {
				msg = er.Title
			}
			seg := er.Code
			if seg != "" && msg != "" {
				seg += ": " + msg
			} else if seg == "" {
				seg = msg
			}
			parts = append(parts, seg)
		}
		e.Detail = strings.Join(parts, "; ")
	}
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		e.Detail = strings.TrimSpace(e.Detail + " (retry after " + ra + "s)")
	}
	return e
}
