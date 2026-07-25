package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientSearchMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("finn-gw-service") != "SEARCH-QUEST-RC" {
			t.Errorf("wrong finn-gw-service: %s", r.Header.Get("finn-gw-service"))
		}
		if r.Header.Get("finn-gw-key") == "" {
			t.Error("missing finn-gw-key")
		}
		// Return mock response
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte(`{"docs":[{"id":"12345","heading":"iPhone 16","location":"Helsinki","price":{"amount":500,"currency_code":"EUR"},"ad_type":67,"timestamp":1700000000000,"canonical_url":"https://www.tori.fi/12345","image":{"url":"https://img.tori.net/test.jpg"}}],"numFound":42}`))
	}))
	defer ts.Close()

	// Override base URL via a test-only variable
	origBase := baseURL
	baseURL = ts.URL
	defer func() { baseURL = origBase }()
	origStatic := staticHeaders
	staticHeaders = map[string]string{} // empty headers for mock
	defer func() { staticHeaders = origStatic }()

	c := newClient()
	result, err := c.search(searchParams{Q: "iphone", Page: 1})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if result.Total != 42 {
		t.Errorf("total = %d, want 42", result.Total)
	}
	if len(result.Docs) != 1 {
		t.Fatalf("docs = %d, want 1", len(result.Docs))
	}
	if result.Docs[0].Heading != "iPhone 16" {
		t.Errorf("heading = %q", result.Docs[0].Heading)
	}
	if result.Docs[0].Price.Amount != 500 {
		t.Errorf("price = %f", result.Docs[0].Price.Amount)
	}
	if result.Docs[0].CanonicalURL != "https://www.tori.fi/12345" {
		t.Errorf("url = %q", result.Docs[0].CanonicalURL)
	}
}

func TestClientFilterParams(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte(`{"docs":[],"numFound":0}`))
	}))
	defer ts.Close()

	origBase := baseURL
	baseURL = ts.URL
	defer func() { baseURL = origBase }()
	origStatic := staticHeaders
	staticHeaders = map[string]string{}
	defer func() { staticHeaders = origStatic }()

	c := newClient()
	c.search(searchParams{
		Q: "pyörä", Category: "1.69.3963",
		PriceFrom: 100, PriceTo: 1000,
		Shipping: true, Page: 2,
		ExtraFilters: []string{"bikes_type=8", "condition=2"},
	})

	for _, want := range []string{
		"q=py%C3%B6r%C3%A4", "sub_category=1.69.3963",
		"price_from=100", "price_to=1000",
		"shipping_exists=true", "page=2",
		"bikes_type=8", "condition=2",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query missing %q in: %s", want, gotQuery)
		}
	}
}

func TestClientShowMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte(`{"docs":[{"id":"99999","heading":"Test bike","location":"Espoo","price":{"amount":1200,"currency_code":"EUR"},"canonical_url":"https://www.tori.fi/99999","image":{"url":"https://img.tori.net/bike.jpg"},"extras":[{"id":"brand","values":["Tunturi"]}]}]}`))
	}))
	defer ts.Close()

	origBase := baseURL
	baseURL = ts.URL
	defer func() { baseURL = origBase }()
	origStatic := staticHeaders
	staticHeaders = map[string]string{}
	defer func() { staticHeaders = origStatic }()

	c := newClient()
	l, err := c.show("99999")
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if l.ID != "99999" {
		t.Errorf("id = %q", l.ID)
	}
	if l.Heading != "Test bike" {
		t.Errorf("heading = %q", l.Heading)
	}
	if len(l.Extras) != 1 || l.Extras[0].Values[0] != "Tunturi" {
		t.Errorf("extras = %v", l.Extras)
	}
}

func TestClientShowNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte(`{"docs":[{"id":"12345","heading":"Other listing","price":{"amount":100,"currency_code":"EUR"},"canonical_url":"https://www.tori.fi/12345","image":{"url":""}}]}`))
	}))
	defer ts.Close()

	origBase := baseURL
	baseURL = ts.URL
	defer func() { baseURL = origBase }()
	origStatic := staticHeaders
	staticHeaders = map[string]string{}
	defer func() { staticHeaders = origStatic }()

	c := newClient()
	_, err := c.show("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent listing")
	}
}

func TestClientFetchBodyMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><div class="whitespace-pre-wrap"><p>This is a test description.</p><p>Second paragraph with more detail.</p></div><p class="mb-0">Kunto<!-- -->: <b>Hyvä</b></p><p class="mb-0">Merkki: <b>Tunturi</b></p></html>`))
	}))
	defer ts.Close()

	c := newClient()
	body, details, err := c.fetchBody(ts.URL)
	if err != nil {
		t.Fatalf("fetchBody failed: %v", err)
	}
	if !strings.Contains(body, "test description") {
		t.Errorf("body = %q", body)
	}
	if !strings.Contains(body, "Second paragraph") {
		t.Errorf("missing second paragraph in body")
	}
	if details["Kunto"] != "Hyvä" {
		t.Errorf("details[Kunto] = %q", details["Kunto"])
	}
	if details["Merkki"] != "Tunturi" {
		t.Errorf("details[Merkki] = %q", details["Merkki"])
	}
}

var origBase string
var origStatic map[string]string

func setMockBase(ts *httptest.Server) {
	origBase = baseURL
	origStatic = staticHeaders
	baseURL = ts.URL
	staticHeaders = map[string]string{}
}

func restoreBase() {
	baseURL = origBase
	staticHeaders = origStatic
}
