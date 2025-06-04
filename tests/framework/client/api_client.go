package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type APIClient struct {
	client  *resty.Client
	BaseURL string
	Token   string
}

type APIResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Raw        *resty.Response
	JSON       map[string]interface{}
}

// NewAPIClient creates a new API client instance
func NewAPIClient(baseURL string) *APIClient {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)
	
	// Add request/response logging for debugging
	client.SetLogger(&APILogger{})
	
	return &APIClient{
		client:  client,
		BaseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

// SetAuth sets authentication token
func (c *APIClient) SetAuth(token string) *APIClient {
	c.Token = token
	c.client.SetAuthToken(token)
	return c
}

// SetHeaders sets default headers
func (c *APIClient) SetHeaders(headers map[string]string) *APIClient {
	c.client.SetHeaders(headers)
	return c
}

// GET performs HTTP GET request
func (c *APIClient) GET(endpoint string) *APIResponse {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Get(c.BaseURL + endpoint)
	
	return c.buildResponse(resp, err)
}

// POST performs HTTP POST request
func (c *APIClient) POST(endpoint string, payload interface{}) *APIResponse {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(c.BaseURL + endpoint)
	
	return c.buildResponse(resp, err)
}

// PUT performs HTTP PUT request
func (c *APIClient) PUT(endpoint string, payload interface{}) *APIResponse {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Put(c.BaseURL + endpoint)
	
	return c.buildResponse(resp, err)
}

// PATCH performs HTTP PATCH request
func (c *APIClient) PATCH(endpoint string, payload interface{}) *APIResponse {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Patch(c.BaseURL + endpoint)
	
	return c.buildResponse(resp, err)
}

// DELETE performs HTTP DELETE request
func (c *APIClient) DELETE(endpoint string) *APIResponse {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Delete(c.BaseURL + endpoint)
	
	return c.buildResponse(resp, err)
}

// buildResponse creates APIResponse from resty response
func (c *APIClient) buildResponse(resp *resty.Response, err error) *APIResponse {
	apiResp := &APIResponse{
		Raw: resp,
	}
	
	if err != nil {
		apiResp.StatusCode = 0
		return apiResp
	}
	
	apiResp.StatusCode = resp.StatusCode()
	apiResp.Body = resp.Body()
	apiResp.Headers = resp.Header()
	
	// Parse JSON if possible
	if len(apiResp.Body) > 0 {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(apiResp.Body, &jsonData); err == nil {
			apiResp.JSON = jsonData
		}
	}
	
	return apiResp
}

// APILogger implements resty.Logger interface
type APILogger struct{}

func (l *APILogger) Errorf(format string, v ...interface{}) {
	fmt.Printf("[API ERROR] "+format+"\n", v...)
}

func (l *APILogger) Warnf(format string, v ...interface{}) {
	fmt.Printf("[API WARN] "+format+"\n", v...)
}

func (l *APILogger) Debugf(format string, v ...interface{}) {
	fmt.Printf("[API DEBUG] "+format+"\n", v...)
}