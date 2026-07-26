package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const searchKey = "SEARCH_ID_BAP_COMMON"

type listing struct {
	ID           string            `json:"id"`
	Heading      string            `json:"heading"`
	Location     string            `json:"location"`
	Price        price             `json:"price"`
	AdType       int               `json:"ad_type"`
	TradeType    string            `json:"trade_type"`
	Timestamp    int64             `json:"timestamp"`
	Image        image             `json:"image"`
	CanonicalURL string            `json:"canonical_url"`
	Flags        []string          `json:"flags"`
	Extras       []extra           `json:"extras"`
	Labels       []label           `json:"labels"`
	Brand        string            `json:"brand"`
	MemorySize   string            `json:"memory_size,omitempty"`
	OrgName      string            `json:"organisation_name,omitempty"`
	Description  string            `json:"description,omitempty"`
	Details      map[string]string `json:"details,omitempty"`
}

type price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency_code"`
}

type image struct {
	URL string `json:"url"`
}

type extra struct {
	ID     string   `json:"id"`
	Values []string `json:"values"`
}

type label struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

func (l listing) adTypeLabel() string {
	t := strings.TrimSpace(strings.ToLower(l.Heading))
	if strings.HasPrefix(t, "ostetaan") {
		return "Ostetaan"
	}
	if strings.HasPrefix(t, "annetaan") {
		return "Annetaan"
	}
	return "Myydään"
}

func adTypeLabelLoc(l *listing, locale string) string {
	label := l.adTypeLabel()
	if locale == "fi" {
		return label
	}
	switch label {
	case "Myydään":
		return "For sale"
	case "Ostetaan":
		return "Wanted"
	case "Annetaan":
		return "Free"
	}
	return label
}

func (l listing) isRetailer() bool {
	for _, f := range l.Flags {
		if f == "retailer" {
			return true
		}
	}
	return false
}

type searchParams struct {
	Q              string
	Category       string
	Location       string
	PriceFrom      int
	PriceTo        int
	Shipping       bool
	Page           int
	IncludeFilters bool
	ExtraFilters   []string
}

func (p searchParams) query() string {
	v := url.Values{
		"q": {p.Q}, "page": {strconv.Itoa(p.Page)},
		"client": {"NMP-IOS"}, "include_results": {"true"},
	}
	if p.Category != "" {
		v.Set("sub_category", p.Category)
	}
	if p.Location != "" {
		v.Set("location", p.Location)
	}
	if p.PriceFrom > 0 {
		v.Set("price_from", strconv.Itoa(p.PriceFrom))
	}
	if p.PriceTo > 0 {
		v.Set("price_to", strconv.Itoa(p.PriceTo))
	}
	if p.Shipping {
		v.Set("shipping_exists", "true")
	}
	if p.IncludeFilters {
		v.Set("include_filters", "true")
	}
	for _, f := range p.ExtraFilters {
		if idx := strings.IndexByte(f, '='); idx > 0 {
			v.Set(f[:idx], f[idx+1:])
		}
	}
	return v.Encode()
}

func (c *client) search(p searchParams) (*searchResult, error) {
	qs := p.query()
	data, err := c.get(fmt.Sprintf("/search/%s?%s", searchKey, qs), "SEARCH-QUEST-RC")
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	return parseSearchResult(data, p.Page), nil
}

func (c *client) searchRaw(p searchParams) (map[string]any, error) {
	qs := p.query()
	return c.get(fmt.Sprintf("/search/%s?%s", searchKey, qs), "SEARCH-QUEST-RC")
}

type searchResult struct {
	Docs     []listing `json:"docs"`
	Promoted *listing  `json:"promoted,omitempty"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
}

func parseSearchResult(data map[string]any, page int) *searchResult {
	r := &searchResult{Page: page}
	if docs, ok := data["docs"].([]any); ok {
		r.Docs = make([]listing, 0, len(docs))
		for _, d := range docs {
			r.Docs = append(r.Docs, parseListing(d))
		}
	}
	if total, ok := data["numFound"].(float64); ok {
		r.Total = int(total)
	}
	return r
}

// ── Filters ──────────────────────────────────────────────────────────────────

type filterItem struct {
	Name   string        `json:"name"`
	Values []filterValue `json:"values"`
}
type filterValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Count int    `json:"count"`
}

func (c *client) filters(p searchParams) ([]filterItem, error) {
	p.IncludeFilters = true
	p.Page = 1
	qs := p.query()
	data, err := c.get(fmt.Sprintf("/search/%s?%s", searchKey, qs), "SEARCH-QUEST-RC")
	if err != nil {
		return nil, fmt.Errorf("filters: %w", err)
	}
	return parseFilters(data), nil
}

func parseFilters(data map[string]any) []filterItem {
	rawFilters, _ := data["filters"].([]any)
	result := make([]filterItem, 0, len(rawFilters))
	for _, rf := range rawFilters {
		fm, _ := rf.(map[string]any)
		fi := filterItem{}
		fi.Name, _ = fm["name"].(string)
		if fi.Name == "" || fi.Name == "q" {
			continue
		}
		items, _ := fm["filter_items"].([]any)
		fi.Values = make([]filterValue, 0, len(items))
		for _, item := range items {
			im, _ := item.(map[string]any)
			fv := filterValue{}
			fv.Label, _ = im["display_name"].(string)
			fv.Value, _ = im["value"].(string)
			if c, ok := im["hits"].(float64); ok {
				fv.Count = int(c)
			}
			if fv.Value != "" {
				fi.Values = append(fi.Values, fv)
			}
		}
		if len(fi.Values) > 0 {
			result = append(result, fi)
		}
	}
	return result
}

// ── Pole position ────────────────────────────────────────────────────────────

func (c *client) polePosition(p searchParams) (*listing, error) {
	v := url.Values{
		"q": {p.Q}, "page": {strconv.Itoa(p.Page)},
		"client": {"NMP-IOS"}, "include_results": {"true"},
	}
	if p.Category != "" {
		v.Set("sub_category", p.Category)
	}
	if p.Location != "" {
		v.Set("location", p.Location)
	}
	qs := v.Encode()
	data, err := c.get(fmt.Sprintf("/pole-position/api/search/%s?%s", searchKey, qs), "POLE-POSITION-API")
	if err != nil {
		return nil, err
	}
	result, ok := data["result"].(map[string]any)
	if !ok || result == nil {
		return nil, nil
	}
	entry, _ := result["searchEntry"].(map[string]any)
	if entry == nil {
		return nil, nil
	}
	l := parseListing(entry)
	return &l, nil
}

// ── Parse helpers ────────────────────────────────────────────────────────────

func parseListing(raw any) listing {
	m, _ := raw.(map[string]any)
	l := listing{}
	if v, ok := m["id"].(string); ok {
		l.ID = v
	} else if v, ok := m["id"].(float64); ok {
		l.ID = strconv.FormatFloat(v, 'f', 0, 64)
	}
	l.Heading, _ = m["heading"].(string)
	l.Location, _ = m["location"].(string)
	l.CanonicalURL, _ = m["canonical_url"].(string)
	l.Brand, _ = m["brand"].(string)
	l.MemorySize, _ = m["memory_size"].(string)
	l.OrgName, _ = m["organisation_name"].(string)
	l.TradeType, _ = m["trade_type"].(string)
	if p, ok := m["price"].(map[string]any); ok {
		l.Price.Amount, _ = p["amount"].(float64)
		l.Price.Currency, _ = p["currency_code"].(string)
	}
	if ad, ok := m["ad_type"].(float64); ok {
		l.AdType = int(ad)
	}
	if ts, ok := m["timestamp"].(float64); ok {
		l.Timestamp = int64(ts)
	}
	if img, ok := m["image"].(map[string]any); ok {
		l.Image.URL, _ = img["url"].(string)
	}
	l.Description, _ = m["description"].(string)
	l.TradeType, _ = m["trade_type"].(string)
	parseStringSlice(m, "flags", &l.Flags)
	parseExtras(m, &l.Extras)
	parseLabels(m, &l.Labels)
	if dm, ok := m["details"].(map[string]any); ok {
		l.Details = make(map[string]string, len(dm))
		for k, v := range dm {
			if s, ok := v.(string); ok {
				l.Details[k] = s
			}
		}
	}
	return l
}

func parseStringSlice(m map[string]any, key string, dst *[]string) {
	items, _ := m[key].([]any)
	*dst = make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			*dst = append(*dst, s)
		}
	}
}

func parseExtras(m map[string]any, dst *[]extra) {
	items, _ := m["extras"].([]any)
	*dst = make([]extra, 0, len(items))
	for _, item := range items {
		em, _ := item.(map[string]any)
		e := extra{}
		e.ID, _ = em["id"].(string)
		if vals, ok := em["values"].([]any); ok {
			e.Values = make([]string, 0, len(vals))
			for _, val := range vals {
				if s, ok := val.(string); ok {
					e.Values = append(e.Values, s)
				}
			}
		}
		*dst = append(*dst, e)
	}
}

func parseLabels(m map[string]any, dst *[]label) {
	items, _ := m["labels"].([]any)
	*dst = make([]label, 0, len(items))
	for _, item := range items {
		lm, _ := item.(map[string]any)
		lab := label{}
		lab.Text, _ = lm["text"].(string)
		lab.Type, _ = lm["type"].(string)
		*dst = append(*dst, lab)
	}
}

// ── Categories ───────────────────────────────────────────────────────────────

type categoryNode struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (c *client) categories() ([]categoryNode, error) {
	data, err := c.get("/public/v3/category-explorer?profile=mobile", "MARKETPLACE-NAV-BAR")
	if err != nil {
		return nil, fmt.Errorf("categories: %w", err)
	}
	tree, _ := data["catex_tree"].([]any)
	for _, t := range tree {
		tn, _ := t.(map[string]any)
		if id, _ := tn["id"].(string); strings.Contains(id, "all") {
			return walkCategoryTree(tn["subtree"]), nil
		}
	}
	return nil, nil
}

func walkCategoryTree(raw any) []categoryNode {
	nodes, _ := raw.([]any)
	var result []categoryNode
	for _, n := range nodes {
		m, _ := n.(map[string]any)
		label, _ := m["label"].(string)
		dest, _ := m["destinations"].(map[string]any)
		search, _ := dest["search"].(map[string]any)
		params, _ := search["search_parameters"].([]any)
		for _, p := range params {
			pm, _ := p.(map[string]any)
			if key, _ := pm["key"].(string); key == "sub_category" {
				vals, _ := pm["values"].([]any)
				if len(vals) > 0 {
					code, _ := vals[0].(string)
					result = append(result, categoryNode{Code: code, Name: label})
				}
			}
		}
		result = append(result, walkCategoryTree(m["subtree"])...)
	}
	return result
}

func filterCategories(all []categoryNode, query string) []categoryNode {
	if query == "" {
		return all
	}
	q := strings.ToLower(query)
	var result []categoryNode
	for _, cat := range all {
		if strings.Contains(strings.ToLower(cat.Name), q) ||
			strings.Contains(strings.ToLower(cat.Code), q) ||
			strings.Contains(strings.ToLower(translateCategory(cat.Name)), q) {
			result = append(result, cat)
		}
	}
	return result
}

func findChildren(cats []categoryNode, code string) []categoryNode {
	var result []categoryNode
	prefix := code + "."
	for _, c := range cats {
		if strings.HasPrefix(c.Code, prefix) {
			result = append(result, c)
		}
	}
	return result
}

// ── Locations ────────────────────────────────────────────────────────────────

type locationNode struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (c *client) locations() ([]locationNode, error) {
	p := searchParams{Q: "", Page: 1, IncludeFilters: true}
	qs := p.query()
	data, err := c.get(fmt.Sprintf("/search/%s?%s", searchKey, qs), "SEARCH-QUEST-RC")
	if err != nil {
		return nil, fmt.Errorf("locations: %w", err)
	}
	filters, _ := data["filters"].([]any)
	for _, f := range filters {
		fm, _ := f.(map[string]any)
		if name, _ := fm["name"].(string); name == "location" {
			return walkLocationTree(fm["filter_items"]), nil
		}
	}
	return nil, nil
}

func walkLocationTree(raw any) []locationNode {
	items, _ := raw.([]any)
	var result []locationNode
	for _, item := range items {
		m, _ := item.(map[string]any)
		name, _ := m["display_name"].(string)
		code, _ := m["value"].(string)
		if name != "" && code != "" {
			result = append(result, locationNode{Code: code, Name: name})
		}
		result = append(result, walkLocationTree(m["filter_items"])...)
	}
	return result
}

func filterLocations(all []locationNode, query string) []locationNode {
	if query == "" {
		return all
	}
	q := strings.ToLower(query)
	var result []locationNode
	for _, loc := range all {
		if strings.Contains(strings.ToLower(loc.Name), q) ||
			strings.Contains(strings.ToLower(loc.Code), q) {
			result = append(result, loc)
		}
	}
	return result
}

// ── Show ─────────────────────────────────────────────────────────────────────

func (c *client) show(id string) (*listing, error) {
	p := searchParams{Q: id, Page: 1}
	qs := p.query()
	data, err := c.get(fmt.Sprintf("/search/%s?%s", searchKey, qs), "SEARCH-QUEST-RC")
	if err != nil {
		return nil, fmt.Errorf("show: %w", err)
	}
	docs, _ := data["docs"].([]any)
	for _, d := range docs {
		l := parseListing(d)
		if l.ID == id {
			return &l, nil
		}
	}
	return nil, fmt.Errorf("listing %s not found", id)
}
func (c *client) fetchBody(pageURL string) (body string, details map[string]string, err error) {
	apiLimiter.wait()
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read page: %w", err)
	}
	html := string(b)
	// Extract description body — all <p> tags within whitespace-pre-wrap div.
	if start := strings.Index(html, `class="whitespace-pre-wrap"`); start >= 0 {
		// Find the closing </div> of whitespace-pre-wrap (next </div> after the class).
		divEnd := strings.Index(html[start:], "</div>")
		if divEnd >= 0 {
			section := html[start : start+divEnd]
			// Extract all text between <p> and </p> tags, joining with newlines.
			var parts []string
			for pos := 0; pos < len(section); {
				pStart := strings.Index(section[pos:], "<p>")
				if pStart < 0 {
					break
				}
				pStart += pos + len("<p>")
				pEnd := strings.Index(section[pStart:], "</p>")
				if pEnd < 0 {
					break
				}
				text := strings.TrimSpace(section[pStart : pStart+pEnd])
				// Skip empty/whitespace-only paragraphs and non-content (like links-only)
				if text != "" && !strings.HasPrefix(text, "http") {
					parts = append(parts, text)
				}
				pos = pStart + pEnd + len("</p>")
			}
			body = strings.Join(parts, "\n\n")
			body = unescapeHTML(body)
		}
	}

	// Extract detail chips: <p class="mb-0">KEY: <b>VALUE</b></p>
	details = make(map[string]string)
	needle := `<p class="mb-0">`
	for pos := 0; pos < len(html); {
		idx := strings.Index(html[pos:], needle)
		if idx < 0 {
			break
		}
		start := pos + idx + len(needle)
		end := strings.Index(html[start:], "</p>")
		text := html[start : start+end]
		// Split "KEY: <b>VALUE</b>" or "KEY: VALUE", strip HTML comments.
		text = strings.ReplaceAll(text, "<!-- -->", "")
		if colonIdx := strings.Index(text, ":"); colonIdx > 0 {
			key := strings.TrimSpace(text[:colonIdx])
			val := strings.TrimSpace(text[colonIdx+1:])
			val = strings.TrimPrefix(strings.TrimSpace(val), "<b>")
			val = strings.TrimSuffix(val, "</b>")
			val = strings.TrimSpace(val)
			if key != "" && val != "" {
				details[key] = val
			}
		}
		pos = start + end + len("</p>")
	}
	return
}

func unescapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	return s
}
