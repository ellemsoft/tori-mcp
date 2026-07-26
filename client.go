package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var apiLimiter = newOutboundLimiter(300)

var baseURL = "https://apps-gw-poc.svc.tori.fi"

var staticHeaders = map[string]string{
	"finn-device-info":         "iOS, mobile",
	"x-nmp-os-name":            "iOS",
	"x-nmp-app-brand":          "Tori",
	"x-nmp-os-version":         "26.3.1",
	"x-nmp-device":             "iPhone",
	"x-nmp-app-version-name":   "26.16.0",
	"x-nmp-app-build-number":   "26903",
	"buildnumber":              "26903",
	"finn-app-installation-id": "hiRMP4JIWqQ",
	"ab-test-device-id":        "632EB4DA-E226-4598-B6E9-44FA53B72BBD",
	"cmp-analytics":            "1",
	"cmp-personalisation":      "1",
	"cmp-marketing":            "1",
	"cmp-advertising":          "1",
	"accept":                   "application/json; charset=UTF-8",
	"accept-language":          "en-GB,en;q=0.9",
	"user-agent": "ToriApp_iOS/26.16.0-26903 (iPhone; CPU iPhone OS 26.3.1 like Mac OS X) " +
		"ToriNativeApp(UA spoofed for tracking) ToriApp_iOS",
}

type client struct {
	http *http.Client
}

func newClient() *client {
	return &client{http: &http.Client{}}
}

func (c *client) get(path, service string) (map[string]any, error) {
	start := time.Now()
	defer func() { serverLogf("[upstream] %s %v", service, time.Since(start).Round(time.Millisecond)) }()

	cleanPath := path
	query := ""
	if idx := strings.IndexByte(path, '?'); idx >= 0 {
		cleanPath = path[:idx]
		query = path[idx+1:]
	}

	apiLimiter.wait()

	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range staticHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("finn-gw-service", service)
	req.Header.Set("finn-gw-key", gwKey("GET", cleanPath, service, nil, query))

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		serverLogf("[api-error] %d from upstream service=%s", resp.StatusCode, service)
		return nil, fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return data, nil
}
