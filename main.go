package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

const usage = `tori — search Tori.fi (Finnish marketplace) from the command line.

Commands:
  tori search <query> [flags]    Search listings
  tori filters <query>           Show available filter options
  tori categories [query|code]   Browse category tree; pass code to drill down
  tori locations [query]         Browse location tree
  tori show <id> [flags]         Show listing details

Search flags:
  --category CODE     Sub-category code (from "tori categories")
  --location CODE     Location code (from "tori locations")
  --price-from N      Min price EUR
  --price-to N        Max price EUR
  --shipping          ToriDiili (shipping) items only
  --filter K=V        Raw filter param (repeatable). Use "tori filters <q>" to discover codes.
                      e.g. --filter condition=2 --filter bikes_type=8
  --page N            Page number (default 1)
  --lang LANG         Output language: en (default) | fi
  --json              Structured JSON output
  --raw               Raw API JSON (passthrough)

Show flags:
  --json              Structured JSON output
  --fetch-body              Fetch description + detail tags (adds ~500ms; JSON keys: "description", "details")

Examples:
  tori search iphone --price-from 100 --price-to 500
  tori search sähköpyörä --filter bikes_type=8 --filter condition=2
  tori filters iphone --category 1.93.3217
  tori categories puhelin
  tori show 44583092 --fetch-body --json

LLM usage:
  Always include the URL (canonical_url) when sharing results.
  Discover filter codes with: tori filters <query>
  Known condition codes: 1=Uusi(New) 2=Kuin uusi(Like-new) 3=Hyvä(Good) 4=Tyydyttävä(Fair)
  Tables include a URL column — use it. JSON via --json or --raw.
  show --fetch-body --json returns fields: "description" (body text) and "details" (Kunto, Runkokoko, etc).
`

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "--mcp" {
		port := ""
		if len(os.Args) >= 3 {
			port = os.Args[2]
		}
		runMCP(port)
		return
	}
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "search":
		runSearch(os.Args[2:])
	case "filters":
		runFilters(os.Args[2:])
	case "categories", "cats":
		runCategories(os.Args[2:])
	case "locations", "locs":
		runLocations(os.Args[2:])
	case "show":
		runShow(os.Args[2:])
	case "help", "-h", "--help":
		flag.Usage()
	default:
		fmt.Fprintf(os.Stderr, "tori: unknown command %q. Run 'tori help'.\n", os.Args[1])
		os.Exit(1)
	}
}

type stringSlice []string

func (s *stringSlice) String() string     { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(v string) error { *s = append(*s, v); return nil }

func splitArgs(args []string) (flags, positional []string) {
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			if !strings.Contains(args[i], "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags[len(flags)-1] += "=" + args[i]
			}
		} else {
			positional = append(positional, args[i])
		}
	}
	return
}

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	category := fs.String("category", "", "Category code")
	location := fs.String("location", "", "Location code")
	priceFrom := fs.Int("price-from", 0, "Min price EUR")
	priceTo := fs.Int("price-to", 0, "Max price EUR")
	shipping := fs.Bool("shipping", false, "ToriDiili only")
	page := fs.Int("page", 1, "Page number")
	asJSON := fs.Bool("json", false, "JSON output")
	raw := fs.Bool("raw", false, "Raw API JSON")
	lang := fs.String("lang", "en", "Output language: fi or en")
	var filters stringSlice
	fs.Var(&filters, "filter", "Extra filter: key=value (repeatable)")

	flags, positional := splitArgs(args)
	if len(flags) > 0 {
		fs.Parse(flags)
	}
	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "tori search: query required")
		os.Exit(1)
	}

	c := newClient()
	params := searchParams{
		Q: strings.Join(positional, " "), Category: *category, Location: *location,
		PriceFrom: *priceFrom, PriceTo: *priceTo, Shipping: *shipping,
		Page: *page, ExtraFilters: []string(filters),
	}
	if *raw {
		data, err := c.searchRaw(params)
		if err != nil {
			exitErr(err)
		}
		displayRawJSON(data)
		return
	}
	result, err := c.search(params)
	if err != nil {
		exitErr(err)
	}
	if p, _ := c.polePosition(params); p != nil {
		result.Promoted = p
	}
	if *asJSON {
		displaySearchJSON(result)
		return
	}

	locale := "en"
	if strings.HasPrefix(strings.ToLower(*lang), "fi") {
		locale = "fi"
	}
	displaySearchTable(result, locale)
}

func runFilters(args []string) {
	fs := flag.NewFlagSet("filters", flag.ExitOnError)
	category := fs.String("category", "", "Category code")
	location := fs.String("location", "", "Location code")
	fs.Parse(args)
	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "tori filters: query required")
		os.Exit(1)
	}

	c := newClient()
	params := searchParams{Q: strings.Join(fs.Args(), " "), Category: *category, Location: *location, IncludeFilters: true}
	data, err := c.filters(params)
	if err != nil {
		exitErr(err)
	}
	displayFilters(data)
}

func runCategories(args []string) {
	fs := flag.NewFlagSet("categories", flag.ExitOnError)
	lang := fs.String("lang", "en", "Output language: fi or en")
	flags, positional := splitArgs(args)
	if len(flags) > 0 {
		fs.Parse(flags)
	}
	locale := "en"
	if strings.HasPrefix(strings.ToLower(*lang), "fi") {
		locale = "fi"
	}

	query := strings.Join(positional, " ")
	c := newClient()
	cats, err := c.categories()
	if err != nil {
		exitErr(err)
	}
	if isCode(query) {
		cats = findChildren(cats, query)
	} else {
		cats = filterCategories(cats, query)
	}
	if len(cats) == 0 {
		fmt.Println("No categories found.")
		return
	}
	displayCategories(cats, locale)
}

func isCode(s string) bool {
	for _, c := range s {
		if c != '.' && (c < '0' || c > '9') {
			return false
		}
	}
	return len(s) > 0
}

func runLocations(args []string) {
	c := newClient()
	locs, err := c.locations()
	if err != nil {
		exitErr(err)
	}
	if len(args) > 0 {
		locs = filterLocations(locs, strings.Join(args, " "))
	}
	if len(locs) == 0 {
		fmt.Println("No locations found.")
		return
	}
	displayLocations(locs)
}

func runShow(args []string) {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "JSON output")
	withFetchBody := fs.Bool("fetch-body", false, "Fetch full description from listing page")
	flags, positional := splitArgs(args)
	if len(flags) > 0 {
		fs.Parse(flags)
	}
	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "tori show: listing ID required")
		os.Exit(1)
	}

	c := newClient()
	l, err := c.show(positional[0])
	if err != nil {
		exitErr(err)
	}

	if *withFetchBody && l != nil && l.CanonicalURL != "" {
		body, details, err := c.fetchBody(l.CanonicalURL)
		if err != nil {
			exitErr(err)
		}
		if body != "" {
			l.Description = body
			l.Details = details
		}
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(l)
		return
	}
	displayListing(l)
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "tori: %v\n", err)
	os.Exit(1)
}
