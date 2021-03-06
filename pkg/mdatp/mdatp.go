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

	"github.com/sirupsen/logrus"
)

var (
	defaultBaseURLStr = "https://api.securitycenter.windows.com"
	defaultBaseURL, _ = url.Parse(defaultBaseURLStr)
	defaultVersion    = "v1.0"
	defaultUserAgent  = "go-mdatp"
	defaultTimeout    = 30 * time.Second
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
	BaseURL *url.URL

	userAgent string
	version   string

	logger     *logrus.Logger
	httpClient *http.Client

	// inspired by go-github:
	// https://github.com/google/go-github/blob/d913de9ce1e8ed5550283b448b37b721b61cc3b3/github/github.go#L159
	// Reuse a single struct instead of allocating one for each service on the heap.
	common service

	Alert *AlertService
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

// WithOAuthClient creates a oauth credentials config from
// provided oauth attributes and uses it to create an authenticated HTTP client
// that will be applied as the underlying http client.
func WithOAuthClient(clientID, clientSecret, tenantID string) ClientOption {
	return func(c *Client) error {
		conf := OAuthConfig(clientID, clientSecret, tenantID)
		httpClient := conf.Client(context.Background())
		httpClient.Timeout = defaultTimeout
		c.httpClient = httpClient
		return nil
	}
}

// WithHTTPTimeout sets the Timeout value on the underlying http client.
func WithHTTPTimeout(t time.Duration) ClientOption {
	return func(c *Client) error {
		c.httpClient.Timeout = t
		return nil
	}
}

// WithLogger sets the logger used by the client.
func WithLogger(l *logrus.Logger) ClientOption {
	return func(c *Client) error {
		c.logger = l
		return nil
	}
}

// NewClient creates a Client that hosts services
// to interact with the Microsoft Defender ATP SIEM API.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		BaseURL:    defaultBaseURL,
		userAgent:  defaultUserAgent,
		version:    defaultVersion,
		logger:     logrus.New(),
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	c.common.client = c
	c.Alert = (*AlertService)(&c.common)
	return c, nil
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
	req.Header.Set("User-Agent", c.userAgent)
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

	response, err := validateResponse(resp)
	if err == nil && out != nil {
		if decErr := json.NewDecoder(resp.Body).Decode(&out); decErr != io.EOF {
			err = decErr
		}
	}
	return response, err
}

// validateResponse validates the response returned from
// an API call.
func validateResponse(r *http.Response) (*Response, error) {
	resp := &Response{HTTPResponse: r}
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return resp, nil
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return resp, err
	}
	if err := json.Unmarshal(data, &resp.APIError); err != nil {
		return resp, err
	}
	return resp, nil
}

// Response encapsulates the http response received from
// a successful API call.
type Response struct {
	HTTPResponse *http.Response
	APIError     *APIError
}

// APIError represents the JSON returned by the API
// when an error is encountered.
type APIError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Target  string `json:"target"`
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
