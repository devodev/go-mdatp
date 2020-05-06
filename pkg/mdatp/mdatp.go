package mdatp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	defaultBaseURLStr = "https://api.securitycenter.windows.com"
	defaultBaseURL, _ = url.Parse(defaultBaseURLStr)
	defaultVersion    = "v1.0"
	defaultUserAgent  = "go-mdatp"
	defaultTimeout    = 5 * time.Second

	microsoftTokenURL = "https://login.windows.net/%s/oauth2/token?api-version=1.0"
)

var (
	// ErrBadRequest is a 400 http error.
	ErrBadRequest = errors.New("bad request")
	// ErrNotFound is a 404 http error.
	ErrNotFound = errors.New("not found")
)

// service holds a pointer to the Client for service related
// methods to access Client methods, such as newRequest and do.
type service struct {
	client *Client
}

// A Client handles communication with the
// Microsoft Defender ATP API.
type Client struct {
	BaseURL   *url.URL
	UserAgent string
	version   string

	httpClient *http.Client
	tenantID   string

	// inspired by go-github:
	// https://github.com/google/go-github/blob/d913de9ce1e8ed5550283b448b37b721b61cc3b3/github/github.go#L159
	// Reuse a single struct instead of allocating one for each service on the heap.
	common service
}

// ClientOption provides a way to confgigure the client.
type ClientOption func(*Client) error

// WithHTTPClient sets the underlying http client to use.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("HTTP client is nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithOAuthClient sets the underlying http client to use.
func WithOAuthClient(creds *Credentials) ClientOption {
	return func(c *Client) error {
		if creds == nil {
			return fmt.Errorf("creds is nil")
		}
		c.httpClient = creds.OAuthClient()
		return nil
	}
}

// NewClient creates a Client using the provided httpClient.
// If nil is provided, a default httpClient with a default timeout value is created.
// Note that the default client has no way of authenticating itself against
// the Microsoft Defender ATP API.
// A convenience function is provided just for that: NewClientAuthenticated.
func NewClient(tenantID string, opts ...ClientOption) *Client {
	c := &Client{
		BaseURL:    defaultBaseURL,
		UserAgent:  defaultUserAgent,
		version:    defaultVersion,
		httpClient: &http.Client{Timeout: defaultTimeout},
		tenantID:   tenantID,
	}
	c.common.client = c
	return c
}

// Version returns the client version.
func (c *Client) Version() string {
	return c.version
}

// newRequest generates a http.Request based on the method
// and endpoint provided. Default headers are also set here.
func (c *Client) newRequest(method, path string, params url.Values, payload io.Reader) (*http.Request, error) {
	url := c.getURL(path, params)
	req, err := http.NewRequest(method, url.String(), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}

// getURL returns a URL based on the client version.
func (c *Client) getURL(path string, params url.Values) *url.URL {
	return &url.URL{
		Scheme:   c.BaseURL.Scheme,
		Host:     c.BaseURL.Host,
		Path:     fmt.Sprintf("/api/%s/%s", c.version, path),
		RawQuery: params.Encode(),
	}
}

// do performs a roundtrip using the underlying client
// and returns an error, if any.
// It will also try to decode the body into the provided out interface.
// It returns the response and any error from decoding.
func (c *Client) do(ctx context.Context, req *http.Request, out interface{}) (*Response, error) {
	if ctx == nil {
		return nil, errors.New("context must be non-nil")
	}
	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}
	defer resp.Body.Close()

	response := &Response{resp}

	err = CheckResponse(resp)

	if out != nil {
		decErr := json.NewDecoder(resp.Body).Decode(&out)
		if decErr == io.EOF {
			decErr = nil
		}
		err = decErr
	}
	return response, err
}

// CheckResponse validates the response returned from
// an API call and returns an error, if any.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, &errorResponse.Err)
	}
	return errorResponse
}

// Response encapsulates the http response received from
// a successful API call.
type Response struct {
	Response *http.Response
}

// ErrorResponse encapsulates the http response as well as the
// error returned in the body of an API call.
type ErrorResponse struct {
	Response *http.Response
	Err      *Error
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v. API Error: %+v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Response.Status, r.Err)
}

// Error represents the json object returned in the body
// of the response when an error is encountered.
type Error struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int { return &v }

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
