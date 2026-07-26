package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type filtersResult struct {
	Filters []filterItem `json:"filters"`
}

type categoriesResult struct {
	Categories []categoryNode `json:"categories"`
}

type locationsResult struct {
	Locations []locationNode `json:"locations"`
}

func newToriMCPServer() *server.MCPServer {
	s := server.NewMCPServer("tori-mcp", "1.0.2", server.WithToolCapabilities(true))

	s.AddTool(mcp.NewTool("search",
		mcp.WithDescription("Search Tori.fi. Returns id, heading, price, location, url. ALWAYS include the canonical_url when presenting a listing so the user can view the original listing and contact the seller. Use filters tool to discover filter codes."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithString("category", mcp.Description("Category code from categories tool")),
		mcp.WithString("location", mcp.Description("Location code from locations command")),
		mcp.WithNumber("price_from", mcp.Description("Min price EUR")),
		mcp.WithNumber("price_to", mcp.Description("Max price EUR")),
		mcp.WithBoolean("shipping", mcp.Description("ToriDiili items only")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithString("filter", mcp.Description("Raw filter key=value, comma-separated for multiple")),
		mcp.WithOutputSchema[searchResult](),
		readOnlyTool(),
	), searchHandler)

	s.AddTool(mcp.NewTool("show",
		mcp.WithDescription("Get listing details by ID. fetch_body=true for full description + detail tags."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Listing ID")),
		mcp.WithBoolean("fetch_body", mcp.Description("Include full description from listing page")),
		mcp.WithOutputSchema[listing](),
		readOnlyTool(),
	), showHandler)

	s.AddTool(mcp.NewTool("filters",
		mcp.WithDescription("Show available filter options for a query. Use to discover filter codes."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query to get filters for")),
		mcp.WithString("category", mcp.Description("Optional category code")),
		mcp.WithOutputSchema[filtersResult](),
		readOnlyTool(),
	), filtersHandler)

	s.AddTool(mcp.NewTool("categories",
		mcp.WithDescription("Browse category tree. Pass a code to drill down or a name to search."),
		mcp.WithString("query", mcp.Description("Category code to drill down, or name to search")),
		mcp.WithOutputSchema[categoriesResult](),
		readOnlyTool(),
	), categoriesHandler)

	s.AddTool(mcp.NewTool("locations",
		mcp.WithDescription("Browse Tori location codes. Pass a name or code to filter locations; use returned code with search.location."),
		mcp.WithString("query", mcp.Description("Location name or code to search")),
		mcp.WithOutputSchema[locationsResult](),
		readOnlyTool(),
	), locationsHandler)

	return s
}

func runMCP(port string) {
	enableServerLogs = true
	s := newToriMCPServer()

	if port == "" || port == "stdio" {
		server.ServeStdio(s)
		return
	}

	initStats("/var/lib/tori/stats.json")
	httpServer := server.NewStreamableHTTPServer(s, server.WithDisableLocalhostProtection(true))
	mux := http.NewServeMux()
	mux.Handle("/mcp", sessionMiddleware(httpServer))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><title>Tori MCP</title><style>body{font-family:system-ui;max-width:400px;margin:40px auto;padding:20px}</style></head><body><h1>Tori.fi MCP</h1><p>Connect at <code>/mcp</code> · <a href="/health">/health</a></p><hr><p style="color:#888;font-size:.85em">Unofficial. Docs: <a href="https://ellemsoft.com/mcps">ellemsoft.com/mcps</a></p>  <p style="color:#888;font-size:.75rem;margin-top:20px">Rate limited: 60 req/min per session. Outbound to source services: 300 req/min global.</p>
</body></html>`))
	})
	serveHTTP(port, mux)
}

func searchHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c := newClient()
	result, err := c.search(searchParams{
		Q: getStr(req, "query"), Category: getStr(req, "category"), Location: getStr(req, "location"),
		PriceFrom: getInt(req, "price_from"), PriceTo: getInt(req, "price_to"),
		Shipping: getBool(req, "shipping"), Page: getPage(req),
		ExtraFilters: splitCSV(getStr(req, "filter")),
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultStructuredOnly(result), nil
}

func showHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c := newClient()
	l, err := c.show(getStr(req, "id"))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if getBool(req, "fetch_body") && l.CanonicalURL != "" {
		body, details, _ := c.fetchBody(l.CanonicalURL)
		l.Description, l.Details = body, details
	}
	return mcp.NewToolResultStructuredOnly(l), nil
}

func filtersHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c := newClient()
	d, err := c.filters(searchParams{Q: getStr(req, "query"), Category: getStr(req, "category"), IncludeFilters: true})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultStructuredOnly(filtersResult{Filters: d}), nil
}

func categoriesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c := newClient()
	cats, err := c.categories()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	q := getStr(req, "query")
	if isCode(q) {
		cats = findChildren(cats, q)
	} else {
		cats = filterCategories(cats, q)
	}
	return mcp.NewToolResultStructuredOnly(categoriesResult{Categories: cats}), nil
}

func locationsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c := newClient()
	locs, err := c.locations()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultStructuredOnly(locationsResult{
		Locations: filterLocations(locs, getStr(req, "query")),
	}), nil
}

func getStr(req mcp.CallToolRequest, key string) string {
	if v, ok := getArgs(req)[key].(string); ok {
		return v
	}
	return ""
}
func getBool(req mcp.CallToolRequest, key string) bool {
	if v, ok := getArgs(req)[key].(bool); ok {
		return v
	}
	return false
}
func getInt(req mcp.CallToolRequest, key string) int {
	switch v := getArgs(req)[key].(type) {
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	}
	return 0
}
func getPage(req mcp.CallToolRequest) int {
	n := getInt(req, "page")
	if n < 1 {
		n = 1
	}
	return n
}

func readOnlyTool() mcp.ToolOption {
	return mcp.WithToolAnnotation(mcp.ToolAnnotation{
		ReadOnlyHint:    mcp.ToBoolPtr(true),
		DestructiveHint: mcp.ToBoolPtr(false),
		IdempotentHint:  mcp.ToBoolPtr(true),
		OpenWorldHint:   mcp.ToBoolPtr(true),
	})
}

func getArgs(req mcp.CallToolRequest) map[string]any {
	if m, ok := req.Params.Arguments.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}
func splitCSV(s string) []string {
	var r []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			r = append(r, t)
		}
	}
	return r
}
