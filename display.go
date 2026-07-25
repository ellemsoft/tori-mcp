package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func displaySearchJSON(r *searchResult) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(r)
}

func displayRawJSON(data map[string]any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

func displaySearchTable(r *searchResult, locale string) {
	w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	headers := []string{"ID", "Type", "Price", "Location", "Title", "URL"}
	if locale == "fi" {
		headers = []string{"ID", "Tyyppi", "Hinta", "Sijainti", "Otsikko", "URL"}
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
		headers[0], headers[1], headers[2], headers[3], headers[4], headers[5])

	if r.Total > 0 {
		fmt.Fprintf(w, "─ %d results ─\n", r.Total)
	}

	if r.Promoted != nil {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s [PROMOTED]\t%s\n",
			r.Promoted.ID,
			adTypeLabelLoc(r.Promoted, locale),
			formatPrice(r.Promoted.Price),
			truncate(r.Promoted.Location, 22),
			truncate(r.Promoted.Heading, 55),
			r.Promoted.CanonicalURL,
		)
	}

	for _, l := range r.Docs {
		prefix := ""
		if l.isRetailer() {
			prefix = "Y "
		}
		fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\t%s\t%s\n",
			prefix, l.ID,
			adTypeLabelLoc(&l, locale),
			formatPrice(l.Price),
			truncate(l.Location, 22),
			truncate(l.Heading, 55),
			l.CanonicalURL,
		)
	}
	w.Flush()
}

func displayListing(l *listing) {
	fmt.Printf("ID:       %s\n", l.ID)
	fmt.Printf("Title:    %s\n", l.Heading)
	fmt.Printf("Type:     %s\n", l.adTypeLabel())
	fmt.Printf("Price:    %s\n", formatPrice(l.Price))
	fmt.Printf("Location: %s\n", l.Location)

	if l.isRetailer() {
		fmt.Println("Seller:   Yritys")
	}

	for _, e := range l.Extras {
		if len(e.Values) > 0 {
			fmt.Printf("%s: %s\n", formatExtraID(e.ID), strings.Join(e.Values, ", "))
		}
	}

	if l.CanonicalURL != "" {
		fmt.Printf("URL:      %s\n", l.CanonicalURL)
	}
	if l.Image.URL != "" {
		fmt.Printf("Image:    %s\n", l.Image.URL)
	}
	if l.Timestamp > 0 {
		t := time.Unix(l.Timestamp/1000, 0)
		fmt.Printf("Posted:   %s\n", t.Format("2006-01-02"))
	}
}

func displayCategories(cats []categoryNode, locale string) {
	w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	hdr := "Category"
	if locale == "fi" {
		hdr = "Kategoria"
	}
	fmt.Fprintf(w, "Code\t%s\n", hdr)
	for _, c := range cats {
		name := c.Name
		if locale == "en" {
			name = translateCategory(c.Name)
		}
		fmt.Fprintf(w, "%s\t%s\n", c.Code, name)
	}
	w.Flush()
}

func displayLocations(locs []locationNode) {
	w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "Code\tLocation")
	for _, l := range locs {
		fmt.Fprintf(w, "%s\t%s\n", l.Code, l.Name)
	}
	w.Flush()
}

func formatPrice(p price) string {
	if p.Amount == 0 {
		return "-"
	}
	cur := p.Currency
	if cur == "" {
		cur = "EUR"
	}
	return fmt.Sprintf("%s %.0f", cur, p.Amount)
}
func formatExtraID(id string) string {
	labels := map[string]string{
		"brand":       "Brand",
		"memory_size": "Memory",
		"model":       "Model",
		"condition":   "Condition",
		"size":        "Size",
		"color":       "Color",
	}
	if l, ok := labels[id]; ok {
		return l
	}
	return id
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

func displayFilters(filters []filterItem) {
	for _, f := range filters {
		fmt.Printf("\n%s:\n", f.Name)
		w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
		for _, v := range f.Values {
			fmt.Fprintf(w, "  %s\t%s\t(%d)\n", v.Value, v.Label, v.Count)
		}
		w.Flush()
	}
}
