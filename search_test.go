package main

import "testing"

func TestGwKey(t *testing.T) {
	// Known test vector from torium Python library.
	result := gwKey("GET", "/public/users/697554341/unreadmessagecount", "MESSAGING-API", nil, "")
	expected := "bbAqA7PQNmE6YbhPHwTmhasqW/n2rXnHl+f2UTJjxQWcIDynRvYR2sDCBxDpWgJkTfVfPOkbjzVR78rnn/1ojg=="
	if result != expected {
		t.Errorf("gwKey: got %q, want %q", result, expected)
	}
}

func TestGwKeyRootPath(t *testing.T) {
	// Root path "/" should be treated as empty.
	r1 := gwKey("GET", "/", "SVC", nil, "")
	r2 := gwKey("GET", "", "SVC", nil, "")
	if r1 != r2 {
		t.Errorf("gwKey with / and empty path should match: %q vs %q", r1, r2)
	}
}

func TestSearchParamsQuery(t *testing.T) {
	p := searchParams{
		Q:         "iphone",
		Category:  "1.93.3217",
		Location:  "0.100018",
		PriceFrom: 100,
		PriceTo:   500,
		Shipping:  true,
		Page:      2,
	}
	qs := p.query()
	if qs == "" {
		t.Fatal("query returned empty")
	}
	// Spot-check key parameters.
	for _, want := range []string{
		"q=iphone",
		"sub_category=1.93.3217",
		"location=0.100018",
		"price_from=100",
		"price_to=500",
		"shipping_exists=true",
		"page=2",
		"client=NMP-IOS",
		"include_results=true",
	} {
		if !contains(qs, want) {
			t.Errorf("query missing %q in: %s", want, qs)
		}
	}
}

func TestSearchParamsDefaults(t *testing.T) {
	p := searchParams{Q: "test", Page: 1}
	qs := p.query()
	for _, unwanted := range []string{"sub_category", "location", "price_from", "price_to", "shipping_exists"} {
		if contains(qs, unwanted) {
			t.Errorf("query should not contain %q when unset: %s", unwanted, qs)
		}
	}
}

func TestAdTypeLabel(t *testing.T) {
	tests := []struct{ heading, want string }{
		{"Myydään iPhone", "Myydään"},
		{"OSTETAAN auto", "Ostetaan"},
		{"ostetaan kaikki", "Ostetaan"},
		{"annetaan sohva", "Annetaan"},
		{"ANNETAAN ilmaiseksi", "Annetaan"},
		{"iPhone 16", "Myydään"},
		{"", "Myydään"},
	}
	for _, tt := range tests {
		l := listing{Heading: tt.heading}
		if got := l.adTypeLabel(); got != tt.want {
			t.Errorf("adTypeLabel(%q) = %q, want %q", tt.heading, got, tt.want)
		}
	}
}

func TestAdTypeLabelLoc(t *testing.T) {
	l := &listing{Heading: "Ostetaan iPhone"}
	if got := adTypeLabelLoc(l, "en"); got != "Wanted" {
		t.Errorf("adTypeLabelLoc en Ostetaan = %q, want Wanted", got)
	}
	if got := adTypeLabelLoc(l, "fi"); got != "Ostetaan" {
		t.Errorf("adTypeLabelLoc fi Ostetaan = %q, want Ostetaan", got)
	}

	l2 := &listing{Heading: "iPhone myydään"}
	if got := adTypeLabelLoc(l2, "en"); got != "For sale" {
		t.Errorf("adTypeLabelLoc en Myydään = %q, want For sale", got)
	}
}

func TestIsRetailer(t *testing.T) {
	l := listing{Flags: []string{"retailer"}}
	if !l.isRetailer() {
		t.Error("isRetailer should be true for retailer flag")
	}
	l2 := listing{Flags: []string{}}
	if l2.isRetailer() {
		t.Error("isRetailer should be false without retailer flag")
	}
}

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		p    price
		want string
	}{
		{price{Amount: 0, Currency: "EUR"}, "-"},
		{price{Amount: 229, Currency: "EUR"}, "EUR 229"},
		{price{Amount: 229, Currency: ""}, "EUR 229"},
		{price{Amount: 59.5, Currency: "EUR"}, "EUR 60"},
	}
	for _, tt := range tests {
		if got := formatPrice(tt.p); got != tt.want {
			t.Errorf("formatPrice(%+v) = %q, want %q", tt.p, got, tt.want)
		}
	}
}

func TestParseFilters(t *testing.T) {
	data := map[string]any{
		"filters": []any{
			map[string]any{
				"name": "q",
				"filter_items": []any{
					map[string]any{"display_name": "iphone", "value": "iphone", "hits": float64(-1)},
				},
			},
			map[string]any{
				"name": "category",
				"filter_items": []any{
					map[string]any{"display_name": "Electronics", "value": "0.93", "hits": float64(4951)},
				},
			},
		},
	}
	result := parseFilters(data)
	if len(result) != 1 {
		t.Fatalf("expected 1 filter (q skipped), got %d", len(result))
	}
	if result[0].Name != "category" {
		t.Errorf("expected category filter, got %q", result[0].Name)
	}
	if result[0].Values[0].Count != 4951 {
		t.Errorf("expected 4951 hits, got %d", result[0].Values[0].Count)
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
