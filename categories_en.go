package main

// categoryTranslations maps Finnish category names to English.
// Full list fetched from the Tori.fi API category tree.
// When new categories appear, they're stored as Finnish fallback.
var categoryTranslations = map[string]string{
	// ── Antiikki ja taide (0.76) ────────────────────────────────────────
	"Antiikkihuonekalut":           "Antique furniture",
	"Aterimet ja pöytähopeat":      "Cutlery & silverware",
	"Keramiikka, posliini ja lasi": "Ceramics, porcelain & glass",
	"Taide":                        "Art",
	"Muu antiikki":                 "Other antiques",

	// ── Auto-, vene- ja moottoripyörätarvikkeet (0.90) ─────────────────
	"Asuntoauto- ja matkailuautotarvikkeet": "RV & camper accessories",
	"Auton osat":                            "Car parts",
	"Mototarvikkeet ja varaosat":            "Motorcycle parts & accessories",
	"Mönkijän varaosat":                     "ATV parts",
	"Trailerit":                             "Trailers",
	"Veneen varaosat":                       "Boat parts",
	"Muut autotarvikkeet":                   "Other vehicle accessories",

	// ── Elektroniikka ja kodinkoneet (0.93) ─────────────────────────────
	"Kodin pienkoneet":                 "Small kitchen appliances",
	"Kodinkoneet":                      "Home appliances",
	"Puhelimet ja tarvikkeet":          "Phones & accessories",
	"Terveys ja hyvinvointi":           "Health & wellness",
	"Tietotekniikka":                   "Computers & IT",
	"Valokuvaus ja video":              "Photography & video",
	"Videopelit ja konsolit":           "Video games & consoles",
	"Ääni ja kuva":                     "Audio & video",
	"Muu elektroniikka ja kodinkoneet": "Other electronics & appliances",

	// ── Eläimet ja eläintarvikkeet (0.77) ────────────────────────────────
	"Akvaariot": "Aquariums",
	"Eläinten ruokinta, hoito, jalostus ja tallipaikat": "Animal feed, care & stables",
	"Hevoset":                                "Horses",
	"Hevostarvikkeet ja ratsastustarvikkeet": "Horse & riding equipment",
	"Hyönteiset ja hämähäkit":                "Insects & spiders",
	"Häkit":                                  "Cages",
	"Jyväsijät ja kanit":                     "Rodents & rabbits",
	"Kalat":                                  "Fish",
	"Karja":                                  "Livestock",
	"Kissat":                                 "Cats",
	"Kissatarvikkeet":                        "Cat supplies",
	"Koirat":                                 "Dogs",
	"Koiratarvikkeet":                        "Dog supplies",
	"Linnut":                                 "Birds",
	"Matelijat":                              "Reptiles",
	"Muut eläimet":                           "Other animals",
	"Muut eläintarvikkeet":                   "Other pet supplies",

	// ── Koti ja sisustus (0.78) ─────────────────────────────────────────
	"Hyllyt ja lipastot":          "Shelves & dressers",
	"Kaapit":                      "Cabinets",
	"Keittiötarvikkeet ja astiat": "Kitchenware & dishes",
	"Koriste- ja sisustusesineet": "Decor & interior items",
	"Makuuhuone":                  "Bedroom",
	"Matot ja tekstiilit":         "Rugs & textiles",
	"Pöydät ja tuolit":            "Tables & chairs",
	"Sesonkitarvikkeet":           "Seasonal items",
	"Sohvat ja lepotuolit":        "Sofas & lounge chairs",
	"Valaisimet":                  "Lighting",
	"Muut huonekalut ja sisustus": "Other furniture & decor",

	// ── Lapset ja vanhemmat (0.68) ──────────────────────────────────────
	"Lasten kalusteet": "Children's furniture",
	"Lasten kengät":    "Children's shoes",
	"Lasten kirjat":    "Children's books",
	"Lasten tekstiilit ja sisustustarvikkeet": "Children's textiles & decor",
	"Lasten vaatteet":                         "Children's clothing",
	"Lastentarvikkeet ja turvallisuus":        "Baby items & safety",
	"Lastenvaunut ja rattaat":                 "Strollers & prams",
	"Lelut":                                   "Toys",
	"Turvaistuimet":                           "Car seats",
	"Äitiysvaatteet":                          "Maternity wear",
	"Muut":                                    "Other",

	// ── Liiketoiminta ja palvelut (0.91) ─────────────────────────────────
	"Esitystekniikka":                         "Presentation equipment",
	"Kauppa ja jälleenmyynti":                 "Retail & resale",
	"Konetekniikka ja varaosat":               "Machinery & parts",
	"Kontit ja työmaakopit":                   "Containers & site huts",
	"Maatalous":                               "Agriculture",
	"Rahti ja tavarankuljetus":                "Freight & transport",
	"Rakentaminen ja remontointi":             "Construction & renovation",
	"Suurtalouskeittiö ja ravintola-ala":      "Commercial kitchen & restaurant",
	"Terveys ja ensiapu":                      "Health & first aid",
	"Toimistotarvikkeet ja toimistokalusteet": "Office supplies & furniture",
	"Web-domainit ja puhelinnumerot":          "Web domains & phone numbers",
	"Muu liiketoiminta ja palvelut":           "Other business & services",

	// ── Piha ja remontointi (0.67) ──────────────────────────────────────
	"Autotallin ovet ja kalusteet":       "Garage doors & fixtures",
	"Hälyttimet ja turvallisuus":         "Alarms & security",
	"Keittiöt":                           "Kitchens",
	"Kylpyhuone ja sauna":                "Bathroom & sauna",
	"Lämmitys ja ilmanvaihto":            "Heating & ventilation",
	"Mökkitarvikkeet":                    "Cottage supplies",
	"Piha ja puutarha":                   "Garden & yard",
	"Rakennustarvikkeet ja remontointi":  "Building materials & renovation",
	"Työkalut":                           "Tools",
	"Muu koti, puutarha ja rakentaminen": "Other home, garden & construction",

	// ── Urheilu ja ulkoilu (0.69) ───────────────────────────────────────
	"Extreme-urheilu":                        "Extreme sports",
	"Fanituotteet":                           "Fan merchandise",
	"Golf":                                   "Golf",
	"Hiihto ja laskettelu":                   "Skiing & snowboarding",
	"Jääkiekko ja luistelu":                  "Ice hockey & skating",
	"Kuntosalilaitteet":                      "Gym equipment",
	"Pallopelit":                             "Ball sports",
	"Pyöräily":                               "Cycling",
	"Ravintolisät":                           "Supplements",
	"Retkeily, kalastus ja metsästys":        "Outdoors, fishing & hunting",
	"Urheilukellot ja aktiivisuusrannekkeet": "Sports watches & trackers",
	"Urheiluvaatteet ja -kengät":             "Sportswear & shoes",
	"Vesiurheilu":                            "Water sports",
	"Muu urheilu":                            "Other sports",

	// ── Vaatteet, kosmetiikka ja asusteet (0.71) ────────────────────────
	"Asusteet":                               "Accessories",
	"Ihonhoito ja hiustenhoito":              "Skincare & haircare",
	"Kellot ja rannekellot":                  "Watches",
	"Kengät":                                 "Shoes",
	"Korut ja korurasiat":                    "Jewelry & boxes",
	"Kosmetiikka":                            "Cosmetics",
	"Laukut ja lompakot":                     "Bags & wallets",
	"Miesten vaatteet":                       "Men's clothing",
	"Naamiaisasut":                           "Costumes",
	"Naisten vaatteet":                       "Women's clothing",
	"Silmälasit ja linssit":                  "Glasses & lenses",
	"Muut vaatteet, kosmetiikka ja asusteet": "Other clothing & accessories",

	// ── Viihde ja harrastukset (0.86) ───────────────────────────────────
	"Elintarvikkeet":                  "Food & groceries",
	"Keräily":                         "Collectibles",
	"Kirjat ja lehdet":                "Books & magazines",
	"Käsityöt":                        "Crafts",
	"Matkat ja matkaliput":            "Travel & tickets",
	"Musiikki ja elokuvat":            "Music & movies",
	"Pienoismallit ja rakennussarjat": "Models & building kits",
	"Radio-ohjattavat":                "Radio-controlled",
	"Seurapelit":                      "Board games",
	"Soittimet":                       "Musical instruments",
	"Muu viihde ja harrastukset":      "Other entertainment & hobbies",
}

// translateCategory returns the English name for a Finnish category,
// or the original if no translation is available.
func translateCategory(fi string) string {
	if en, ok := categoryTranslations[fi]; ok {
		return en
	}
	return fi
}
