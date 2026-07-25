#!/bin/bash
# tori-cli test cases — run from tori-cli directory
cd "$(dirname "$0")/.."
PASS=0; FAIL=0
t() { echo "TEST: $1"; if eval "$2" >/dev/null 2>&1; then echo "  PASS"; ((PASS++)); else echo "  FAIL"; ((FAIL++)); fi; echo; }

echo "==================== tori-cli test suite ===================="

# ── search ──────────────────────────────────────────────────
t "basic search"                './tori search iphone 2>&1 | grep -q "https://"'
t "price filter"                './tori search iphone --price-from 100 --price-to 500 2>&1 | grep -q "EUR"'
t "price from only"             './tori search iphone --price-from 500 2>&1 | grep -q "https://"'
t "price to only"               './tori search iphone --price-to 100 2>&1 | grep -q "https://"'
t "shipping only"               './tori search kenkä --shipping 2>&1 | grep -q "https://"'
t "category + location"         './tori search sohva --category 1.78.7756 --location 0.100018 2>&1 | grep -q "https://"'
t "category filter"             './tori search puhelin --category 1.93.3217 2>&1 | grep -q "https://"'
t "location filter"             './tori search auto --location 0.100018 2>&1 | grep -q "https://"'
t "--filter raw params"         './tori search sähköpyörä --filter bikes_type=8 2>&1 | grep -q "https://"'
t "multiple --filter"           './tori search pyörä --filter price_from=50 --filter price_to=200 2>&1 | grep -q "https://"'
t "filter dealer_segment"       './tori search sähköpyörä --filter dealer_segment=1 2>&1 | grep -q "https://"'
t "filter condition"            './tori search sähköpyörä --filter condition=2 2>&1 | grep -q "https://"'
t "filter condition + bikes"    './tori search sähköpyörä --filter condition=2 --filter bikes_type=8 2>&1 | grep -q "https://"'
t "empty results"               './tori search xyzzy_no_match_12345 2>&1 | grep -q "ID"'
t "page 2"                      './tori search iphone --page 2 2>&1 | grep -q "https://"'
t "page 3"                      './tori search iphone --page 3 2>&1 | grep -q "https://"'

# ── output formats ──────────────────────────────────────────
t "JSON output"                 './tori search iphone --json 2>&1 | python3 -c "import json,sys; json.load(sys.stdin)"'
t "raw output"                  './tori search iphone --raw 2>&1 | python3 -c "import json,sys; json.load(sys.stdin)"'
t "JSON parses correctly"       './tori search iphone --json 2>&1 | python3 -c "import json,sys; d=json.load(sys.stdin); assert len(d[\"Docs\"])>0"'
t "raw has docs array"          './tori search iphone --raw 2>&1 | python3 -c "import json,sys; d=json.load(sys.stdin); assert len(d[\"docs\"])>0"'

# ── categories ──────────────────────────────────────────────
t "category drill-down"         './tori categories 1.69 2>&1 | grep -q "Cycling"'
t "category drill-down 2"       './tori categories 1.93 2>&1 | grep -q "Phone"'
t "category name search"        './tori categories puhelin 2>&1 | grep -q "phone"'
t "category name search en"     './tori categories phone 2>&1 | grep -q "phone"'
t "category name search fi"     './tori categories puhelin --lang fi 2>&1 | grep -q "puhelin"'
t "category code drill"         './tori categories 1.69.3963 2>&1 | grep -q "Cycling"'
t "all categories"              './tori categories 2>&1 | grep -q "Antiikki"'

# ── locations ───────────────────────────────────────────────
t "location search"             './tori locations helsinki 2>&1 | grep -q "Helsinki"'
t "location search espoo"       './tori locations espoo 2>&1 | grep -q "Espoo"'
t "all locations"               './tori locations 2>&1 | grep -q "Uusimaa"'

# ── filters discovery ───────────────────────────────────────
t "filters discovery"           './tori filters iphone 2>&1 | grep -q "category"'
t "filters with category"       './tori filters iphone --category 1.93.3217 2>&1 | grep -q "brand"'
t "filters with location"       './tori filters iphone --location 0.100018 2>&1 | grep -q "category"'

# ── show ───────────────────────────────────────────────────
t "show listing"                './tori show 44583092 2>&1 | grep -qi "iphone"'
t "show --json"                 './tori show 44583092 --json 2>&1 | python3 -c "import json,sys; d=json.load(sys.stdin); assert d[\"id\"]"'
t "show --json has fields"      './tori show 44583092 --json 2>&1 | python3 -c "import json,sys; d=json.load(sys.stdin); assert d[\"canonical_url\"]; assert d[\"price\"]; assert d[\"heading\"]"'
t "show --fetch-body"           './tori show 34917144 --fetch-body --json 2>/dev/null | python3 -c "import json,sys; d=json.load(sys.stdin); assert d.get(\"description\")"'
t "show --fetch-body details"   './tori show 34917144 --fetch-body --json 2>/dev/null | python3 -c "import json,sys; d=json.load(sys.stdin); assert len(d.get(\"details\",{}))>=3"'
t "show --fetch-body multi-para" './tori show 43972911 --fetch-body --json 2>/dev/null | python3 -c "import json,sys; d=json.load(sys.stdin); assert len(d.get(\"description\",\"\"))>500"'

# ── locale ─────────────────────────────────────────────────
t "--lang fi"                   './tori search iphone --lang fi 2>&1 | grep -q "Tyyppi"'
t "--lang en"                   './tori search iphone 2>&1 | grep -q "Type"'
t "--lang fi categories"        './tori categories --lang fi 2>&1 | grep -q "Kategoria"'
t "--lang en categories"        './tori categories 2>&1 | grep -q "Category"'

# ── misc ───────────────────────────────────────────────────
t "help"                        './tori help 2>&1 | grep -q "LLM usage"'
t "go tests"                    'go test ./... -count=1'
t "go vet"                      'go vet ./...'

echo "==================== Results: $PASS pass, $FAIL fail ===================="
