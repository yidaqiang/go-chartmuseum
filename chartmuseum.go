package chartmuseum

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	userAgent        = "go-chartmuseum"
	defaultMediaType = "application/vnd.chartmuseum.v0+json"

	defaultBaseURL = "https://chart.ydq.io/"

	headerRateLimit = "RateLimit-Limit"
	headerRateReset = "RateLimit-Reset"
)

// AuthType represents an authentication type within ChartMuseum.
type AuthType int

// List of available authentication types.
const (
	DefaultNoAuth AuthType = iota
	BasicAuth
	// JobToken
	// OAuthToken
	//PrivateToken
)

type Client struct {
	// HTTP client used to communicate with the API.
	client *retryablehttp.Client

	// Base URL for API requests. Defaults to the public GitHub API, but can be
	// set to a domain endpoint to use with GitHub Enterprise. BaseURL should
	// always be specified with a trailing slash.
	baseURL *url.URL

	// disableRetries is used to disable the default retry logic.
	disableRetries bool

	// configureLimiterOnce is used to make sure the limiter is configured exactly
	// once and block all other calls until the initial (one) call is done.
	configureLimiterOnce sync.Once

	// Token type used to make authenticated API calls.
	authType AuthType

	// Username and password used for basic authentication.
	username, password string

	// Token used to make authenticated API calls.
	token string

	// Protects the token field from concurrent read/write accesses.
	tokenLock sync.RWMutex

	// User agent used when communicating with the GitHub API.
	UserAgent string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for talking to different parts of the ChartMuseum API.
	//Repositories *RepositoriesService
	Charts *ChartService
	Info   *InfoService
}

type service struct {
	client *Client
}

// Response wraps http.Response and decodes ChartMuseum response
type Response struct {
	*http.Response

	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Saved   bool   `json:"saved,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
	Healthy bool   `json:"healthy,omitempty"`
}

// NewClient returns a new ChartMuseum API client.
func NewClient(options ...ClientOptionFunc) (*Client, error) {
	client, err := newClient(options...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewBasicAuthClient returns a new GitLab API client. To use API methods which
// require authentication, provide a valid username and password.
func NewBasicAuthClient(username, password string, options ...ClientOptionFunc) (*Client, error) {
	client, err := newClient(options...)
	if err != nil {
		return nil, err
	}

	client.authType = BasicAuth
	client.username = username
	client.password = password

	return client, nil
}

func newClient(options ...ClientOptionFunc) (*Client, error) {
	c := &Client{UserAgent: userAgent}

	c.client = &retryablehttp.Client{
		Backoff:      c.retryHTTPBackoff,
		CheckRetry:   c.retryHTTPCheck,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   cleanhttp.DefaultPooledClient(),
		RetryMax:     3,
	}

	// Set the default base URL.
	c.setBaseURL(defaultBaseURL)

	// Apply any given client options.
	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(c); err != nil {
			return nil, err
		}
	}

	//c.Repositories = &RepositoriesService{client: c}
	c.Charts = &ChartService{client: c}
	c.Info = &InfoService{client: c}
	return c, nil
}

// retryHTTPCheck provides a callback for Client.CheckRetry which
// will retry both rate limit (429) and Server (>= 500) errors.
func (c *Client) retryHTTPCheck(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return false, err
	}
	if !c.disableRetries && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
		return true, nil
	}
	return false, nil
}

// retryHTTPBackoff provides a generic callback for Client.Backoff which
// will pass through all calls based on the status code of the response.
func (c *Client) retryHTTPBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// Use the rate limit backoff function when we are rate limited.
	if resp != nil && resp.StatusCode == 429 {
		return rateLimitBackoff(min, max, attemptNum, resp)
	}

	// Set custom duration's when we experience a service interruption.
	min = 700 * time.Millisecond
	max = 900 * time.Millisecond

	return retryablehttp.LinearJitterBackoff(min, max, attemptNum, resp)
}

// rateLimitBackoff provides a callback for Client.Backoff which will use the
// RateLimit-Reset header to determine the time to wait. We add some jitter
// to prevent a thundering herd.
//
// min and max are mainly used for bounding the jitter that will be added to
// the reset time retrieved from the headers. But if the final wait time is
// less then min, min will be used instead.
func rateLimitBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// rnd is used to generate pseudo-random numbers.
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// First create some jitter bounded by the min and max durations.
	jitter := time.Duration(rnd.Float64() * float64(max-min))

	if resp != nil {
		if v := resp.Header.Get(headerRateReset); v != "" {
			if reset, _ := strconv.ParseInt(v, 10, 64); reset > 0 {
				// Only update min if the given time to wait is longer.
				if wait := time.Until(time.Unix(reset, 0)); wait > min {
					min = wait
				}
			}
		}
	}

	return min + jitter
}

// BaseURL return a copy of the baseURL.
func (c *Client) BaseURL() *url.URL {
	u := *c.baseURL
	return &u
}

// setBaseURL sets the base URL for API requests to a custom endpoint.
func (c *Client) setBaseURL(urlStr string) error {
	// Make sure the given URL end with a slash
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// Update the base URL of the client.
	c.baseURL = baseURL

	return nil
}

// NewRequest creates a new API request. The method expects a relative URL
// path that will be resolved relative to the base URL of the Client.
// Relative URL paths should always be specified without a preceding slash.
// If specified, the value pointed to by body is JSON encoded and included
// as the request body.
func (c *Client) NewRequest(method, path string, opt interface{}, options []RequestOptionFunc) (*retryablehttp.Request, error) {
	u := *c.baseURL
	unescaped, err := url.PathUnescape(path)
	if err != nil {
		return nil, err
	}

	// Set the encoded path data
	u.RawPath = c.baseURL.Path + path
	u.Path = c.baseURL.Path + unescaped

	// Create a request specific headers map.
	reqHeaders := make(http.Header)

	if c.UserAgent != "" {
		reqHeaders.Set("User-Agent", c.UserAgent)
	}
	var body interface{}

	if _, ok := opt.(io.Reader); ok {

	}
	switch opt.(type) {
	case *os.File:
		// 上传 chart 包时，如果前面没有设置 RequestOptionFunc，添加 WithUpload
		// TOOD 应该校验是否有 WithUpload
		if len(options) < 1 {
			file := opt.(*os.File)
			stat, err := file.Stat()
			if err != nil {
				return nil, err
			}
			if stat.IsDir() {
				return nil, errors.New("can't be update a directory")
			}

			mediaType, _ := detectContentType(file)
			options = append(options, WithUpload(mediaType, stat.Size()))
		}
		body = opt
	case struct{}:
		if method == http.MethodPost || method == http.MethodPut {
			// 其他场景 json 数据
			reqHeaders.Set("Content-Type", "application/json")

			if opt != nil {
				body, err = json.Marshal(opt)
				if err != nil {
					return nil, err
				}
			}
		}
		if opt != nil {
			q, err := query.Values(opt)
			if err != nil {
				return nil, err
			}
			u.RawQuery = q.Encode()
		}
	default:
		// TODO noting
	}

	req, err := retryablehttp.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	// Set the request specific headers.
	for k, v := range reqHeaders {
		req.Header[k] = v
	}

	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(req); err != nil {
			return nil, err
		}
	}

	return req, nil
}

// newResponse creates a new Response for the provided http.Response.
// r must not be nil.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(req *retryablehttp.Request, v interface{}) (*Response, error) {
	// If not yet configured, try to configure the rate limiter. Fail
	// silently as the limiter will be disabled in case of an error.
	//c.configureLimiterOnce.Do(func() { c.configureLimiter(req.Context()) })

	// Wait will block until the limiter can obtain a new token.
	//err := c.limiter.Wait(req.Context())
	//if err != nil {
	//	return nil, err
	//}

	// Set the correct authentication header. If using basic auth, then check
	// if we already have a token and if not first authenticate and get one.
	//var basicAuthToken string
	switch c.authType {
	case DefaultNoAuth:
		logrus.Debug("The authentication mode does not exist")
	case BasicAuth:
		req.SetBasicAuth(c.username, c.password)
		/*c.tokenLock.RLock()
		basicAuthToken = c.token
		c.tokenLock.RUnlock()
		if basicAuthToken == "" {
			// If we don't have a token yet, we first need to request one.
			basicAuthToken, err = c.requestOAuthToken(req.Context(), basicAuthToken)
			if err != nil {
				return nil, err
			}
		}
		req.Header.Set("Authorization", "Bearer "+basicAuthToken)*/
	/*case JobToken:
		if values := req.Header.Values("JOB-TOKEN"); len(values) == 0 {
			req.Header.Set("JOB-TOKEN", c.token)
		}
	case OAuthToken:
		if values := req.Header.Values("Authorization"); len(values) == 0 {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
	case PrivateToken:
		if values := req.Header.Values("PRIVATE-TOKEN"); len(values) == 0 {
			req.Header.Set("PRIVATE-TOKEN", c.token)
		}*/
	default:
		return nil, errors.New("The authentication mode does not exist")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	/*if resp.StatusCode == http.StatusUnauthorized && c.authType == BasicAuth {
		resp.Body.Close()
		// The token most likely expired, so we need to request a new one and try again.
		if _, err := c.requestOAuthToken(req.Context(), basicAuthToken); err != nil {
			return nil, err
		}
		return c.Do(req, v)
	}
	defer resp.Body.Close()*/

	response := newResponse(resp)

	err = CheckResponse(resp)
	if err != nil {
		// Even though there was an error, we still return the response
		// in case the caller wants to inspect it further.
		return response, err
	}

	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}

	return response, err
}

// An ErrorResponse reports one or more errors caused by an API request.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/README.html#data-validation-and-error-reporting
type ErrorResponse struct {
	Body     []byte
	Response *http.Response
	Message  string
}

func (e *ErrorResponse) Error() string {
	path, _ := url.QueryUnescape(e.Response.Request.URL.Path)
	u := fmt.Sprintf("%s://%s%s", e.Response.Request.URL.Scheme, e.Response.Request.URL.Host, path)
	return fmt.Sprintf("%s %s: %d %s", e.Response.Request.Method, u, e.Response.StatusCode, e.Message)
}

// CheckResponse checks the API response for errors, and returns them if present.
func CheckResponse(r *http.Response) error {
	switch r.StatusCode {
	case 200, 201, 202, 204, 304:
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorResponse.Body = data

		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			errorResponse.Message = "failed to parse unknown error format"
		} else {
			errorResponse.Message = parseError(raw)
		}
	}

	return errorResponse
}

func parseError(raw interface{}) string {
	switch raw := raw.(type) {
	case string:
		return raw

	case []interface{}:
		var errs []string
		for _, v := range raw {
			errs = append(errs, parseError(v))
		}
		return fmt.Sprintf("[%s]", strings.Join(errs, ", "))

	case map[string]interface{}:
		var errs []string
		for k, v := range raw {
			errs = append(errs, fmt.Sprintf("{%s: %s}", k, parseError(v)))
		}
		sort.Strings(errs)
		return strings.Join(errs, ", ")

	default:
		return fmt.Sprintf("failed to parse unexpected error type: %T", raw)
	}
}

func parseRepoUrl(repo string) (string, error) {
	// TODO 检查 repo 是否符合 url 格式
	repoStr := strings.Trim(repo, "/")
	return repoStr, nil
}
